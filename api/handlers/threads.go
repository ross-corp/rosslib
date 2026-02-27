package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

var mentionRegex = regexp.MustCompile(`@([a-zA-Z0-9_]+)`)

// GetBookThreads handles GET /books/{workId}/threads
func GetBookThreads(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		workID := e.Request.PathValue("workId")

		books, _ := app.FindRecordsByFilter("books",
			"open_library_id = {:id}", "", 1, 0,
			map[string]any{"id": workID},
		)
		if len(books) == 0 {
			return e.JSON(http.StatusOK, []any{})
		}

		type threadRow struct {
			ID          string  `db:"id" json:"id"`
			BookID      string  `db:"book_id" json:"book_id"`
			UserID      string  `db:"user_id" json:"user_id"`
			Username    string  `db:"username" json:"username"`
			DisplayName *string `db:"display_name" json:"display_name"`
			Avatar      *string `db:"avatar" json:"avatar"`
			Title       string  `db:"title" json:"title"`
			Body        string  `db:"body" json:"body"`
			Spoiler     bool    `db:"spoiler" json:"spoiler"`
			CreatedAt   string  `db:"created_at" json:"created_at"`
		}

		var threads []threadRow
		err := app.DB().NewQuery(`
			SELECT t.id, t.book as book_id, t.user as user_id, u.username,
				   u.display_name, u.avatar,
				   t.title, t.body, t.spoiler, t.created as created_at
			FROM threads t
			JOIN users u ON t.user = u.id
			WHERE t.book = {:book} AND (t.deleted_at IS NULL OR t.deleted_at = '')
			ORDER BY t.created DESC
		`).Bind(map[string]any{"book": books[0].Id}).All(&threads)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}

		var result []map[string]any
		for _, t := range threads {
			var avatarURL *string
			if t.Avatar != nil && *t.Avatar != "" {
				url := fmt.Sprintf("/api/files/users/%s/%s", t.UserID, *t.Avatar)
				avatarURL = &url
			}

			type countResult struct {
				Count int `db:"count"`
			}
			var cnt countResult
			_ = app.DB().NewQuery(`
				SELECT COUNT(*) as count FROM thread_comments
				WHERE thread = {:thread} AND (deleted_at IS NULL OR deleted_at = '')
			`).Bind(map[string]any{"thread": t.ID}).One(&cnt)

			result = append(result, map[string]any{
				"id":            t.ID,
				"book_id":       t.BookID,
				"user_id":       t.UserID,
				"username":      t.Username,
				"display_name":  t.DisplayName,
				"avatar_url":    avatarURL,
				"title":         t.Title,
				"body":          t.Body,
				"spoiler":       t.Spoiler,
				"created_at":    t.CreatedAt,
				"comment_count": cnt.Count,
			})
		}
		if result == nil {
			result = []map[string]any{}
		}

		return e.JSON(http.StatusOK, result)
	}
}

// CreateThread handles POST /books/{workId}/threads
func CreateThread(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		workID := e.Request.PathValue("workId")

		books, _ := app.FindRecordsByFilter("books",
			"open_library_id = {:id}", "", 1, 0,
			map[string]any{"id": workID},
		)
		if len(books) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Book not found"})
		}

		data := struct {
			Title   string `json:"title"`
			Body    string `json:"body"`
			Spoiler bool   `json:"spoiler"`
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}
		if data.Title == "" || data.Body == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "title and body required"})
		}
		if len(data.Title) > 500 {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "title must be 500 characters or fewer"})
		}
		if len(data.Body) > 10000 {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "body must be 10,000 characters or fewer"})
		}

		coll, err := app.FindCollectionByNameOrId("threads")
		if err != nil {
			return err
		}
		rec := core.NewRecord(coll)
		rec.Set("book", books[0].Id)
		rec.Set("user", user.Id)
		rec.Set("title", data.Title)
		rec.Set("body", data.Body)
		rec.Set("spoiler", data.Spoiler)
		if err := app.Save(rec); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}

		recordActivity(app, user.Id, "created_thread", map[string]any{
			"book":   books[0].Id,
			"thread": rec.Id,
		})

		return e.JSON(http.StatusOK, map[string]any{
			"id":    rec.Id,
			"title": data.Title,
		})
	}
}

// GetThread handles GET /threads/{threadId}
func GetThread(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		threadID := e.Request.PathValue("threadId")

		thread, err := app.FindRecordById("threads", threadID)
		if err != nil || thread.GetString("deleted_at") != "" {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Thread not found"})
		}

		threadUser, _ := app.FindRecordById("users", thread.GetString("user"))
		var avatarURL *string
		username := ""
		var displayName *string
		if threadUser != nil {
			username = threadUser.GetString("username")
			dn := threadUser.GetString("display_name")
			if dn != "" {
				displayName = &dn
			}
			if av := threadUser.GetString("avatar"); av != "" {
				url := "/api/files/" + threadUser.Collection().Id + "/" + threadUser.Id + "/" + av
				avatarURL = &url
			}
		}

		// Get comments
		type commentRow struct {
			ID          string  `db:"id" json:"id"`
			UserID      string  `db:"user_id" json:"user_id"`
			Username    string  `db:"username" json:"username"`
			DisplayName *string `db:"display_name" json:"display_name"`
			Avatar      *string `db:"avatar" json:"avatar"`
			Parent      *string `db:"parent" json:"parent"`
			Body        string  `db:"body" json:"body"`
			CreatedAt   string  `db:"created_at" json:"created_at"`
		}
		var comments []commentRow
		_ = app.DB().NewQuery(`
			SELECT tc.id, tc.user as user_id, u.username, u.display_name, u.avatar,
				   tc.parent, tc.body, tc.created as created_at
			FROM thread_comments tc
			JOIN users u ON tc.user = u.id
			WHERE tc.thread = {:thread} AND (tc.deleted_at IS NULL OR tc.deleted_at = '')
			ORDER BY tc.created ASC
		`).Bind(map[string]any{"thread": threadID}).All(&comments)

		var commentResults []map[string]any
		for _, c := range comments {
			var cAvatarURL *string
			if c.Avatar != nil && *c.Avatar != "" {
				url := fmt.Sprintf("/api/files/users/%s/%s", c.UserID, *c.Avatar)
				cAvatarURL = &url
			}
			commentResults = append(commentResults, map[string]any{
				"id":           c.ID,
				"user_id":      c.UserID,
				"username":     c.Username,
				"display_name": c.DisplayName,
				"avatar_url":   cAvatarURL,
				"parent":       c.Parent,
				"body":         c.Body,
				"created_at":   c.CreatedAt,
			})
		}
		if commentResults == nil {
			commentResults = []map[string]any{}
		}

		return e.JSON(http.StatusOK, map[string]any{
			"id":           thread.Id,
			"book":         thread.GetString("book"),
			"user_id":      thread.GetString("user"),
			"username":     username,
			"display_name": displayName,
			"avatar_url":   avatarURL,
			"title":        thread.GetString("title"),
			"body":         thread.GetString("body"),
			"spoiler":      thread.GetBool("spoiler"),
			"created_at":   thread.GetString("created"),
			"comments":     commentResults,
		})
	}
}

// DeleteThread handles DELETE /threads/{threadId}
func DeleteThread(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		threadID := e.Request.PathValue("threadId")

		thread, err := app.FindRecordById("threads", threadID)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Thread not found"})
		}
		if thread.GetString("user") != user.Id {
			return e.JSON(http.StatusForbidden, map[string]any{"error": "Not your thread"})
		}

		// Soft delete
		thread.Set("deleted_at", time.Now().UTC().Format(time.RFC3339))
		if err := app.Save(thread); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to delete"})
		}

		e.Response.WriteHeader(http.StatusNoContent)
		return nil
	}
}

// AddComment handles POST /threads/{threadId}/comments
func AddComment(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		threadID := e.Request.PathValue("threadId")

		thread, err := app.FindRecordById("threads", threadID)
		if err != nil || thread.GetString("deleted_at") != "" {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Thread not found"})
		}

		data := struct {
			Body   string  `json:"body"`
			Parent *string `json:"parent"`
		}{}
		if err := e.BindBody(&data); err != nil || data.Body == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "body required"})
		}
		if len(data.Body) > 5000 {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "comment must be 5,000 characters or fewer"})
		}

		// Enforce max 1-level nesting
		if data.Parent != nil && *data.Parent != "" {
			parentComment, err := app.FindRecordById("thread_comments", *data.Parent)
			if err != nil {
				return e.JSON(http.StatusBadRequest, map[string]any{"error": "Parent comment not found"})
			}
			if parentComment.GetString("parent") != "" {
				return e.JSON(http.StatusBadRequest, map[string]any{"error": "Cannot nest comments more than 1 level"})
			}
		}

		coll, err := app.FindCollectionByNameOrId("thread_comments")
		if err != nil {
			return err
		}
		rec := core.NewRecord(coll)
		rec.Set("thread", threadID)
		rec.Set("user", user.Id)
		rec.Set("body", data.Body)
		if data.Parent != nil {
			rec.Set("parent", *data.Parent)
		}
		if err := app.Save(rec); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}

		// Fan out @mention notifications
		go fanOutMentionNotifications(app, user, thread, rec.Id, data.Body)

		return e.JSON(http.StatusOK, map[string]any{
			"id":   rec.Id,
			"body": data.Body,
		})
	}
}

// DeleteComment handles DELETE /threads/{threadId}/comments/{commentId}
func DeleteComment(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		commentID := e.Request.PathValue("commentId")

		comment, err := app.FindRecordById("thread_comments", commentID)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Comment not found"})
		}
		if comment.GetString("user") != user.Id {
			return e.JSON(http.StatusForbidden, map[string]any{"error": "Not your comment"})
		}

		// Soft delete
		comment.Set("deleted_at", time.Now().UTC().Format(time.RFC3339))
		if err := app.Save(comment); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to delete"})
		}

		e.Response.WriteHeader(http.StatusNoContent)
		return nil
	}
}

// fanOutMentionNotifications scans a comment body for @username mentions and
// creates a thread_mention notification for each valid, distinct, non-self user.
func fanOutMentionNotifications(app core.App, commenter *core.Record, thread *core.Record, commentID, body string) {
	matches := mentionRegex.FindAllStringSubmatch(body, -1)
	if len(matches) == 0 {
		return
	}

	// Resolve the book's open_library_id for notification metadata.
	bookOLID := ""
	book, err := app.FindRecordById("books", thread.GetString("book"))
	if err == nil {
		bookOLID = book.GetString("open_library_id")
	}

	commenterName := commenter.GetString("display_name")
	if commenterName == "" {
		commenterName = commenter.GetString("username")
	}

	// Truncate body for notification preview.
	preview := body
	if len(preview) > 120 {
		preview = preview[:120] + "..."
	}

	seen := map[string]bool{}
	for _, m := range matches {
		username := strings.ToLower(m[1])
		if seen[username] {
			continue
		}
		seen[username] = true

		// Skip self-mentions.
		if strings.EqualFold(username, commenter.GetString("username")) {
			continue
		}

		// Look up the mentioned user (case-insensitive).
		users, err := app.FindRecordsByFilter("users",
			"LOWER(username) = {:username}", "", 1, 0,
			map[string]any{"username": username},
		)
		if err != nil || len(users) == 0 {
			continue
		}
		mentioned := users[0]

		notifColl, err := app.FindCollectionByNameOrId("notifications")
		if err != nil {
			continue
		}
		rec := core.NewRecord(notifColl)
		rec.Set("user", mentioned.Id)
		rec.Set("notif_type", "thread_mention")
		rec.Set("title", fmt.Sprintf("%s mentioned you in a thread", commenterName))
		rec.Set("body", preview)
		rec.Set("metadata", map[string]any{
			"thread_id":  thread.Id,
			"comment_id": commentID,
			"book_ol_id": bookOLID,
		})
		rec.Set("read", false)
		_ = app.Save(rec)
	}
}
