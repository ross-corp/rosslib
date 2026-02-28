package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// GetReviewComments handles GET /books/{workId}/reviews/{userId}/comments
func GetReviewComments(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		workID := e.Request.PathValue("workId")
		reviewUserID := e.Request.PathValue("userId")

		books, _ := app.FindRecordsByFilter("books",
			"open_library_id = {:id}", "", 1, 0,
			map[string]any{"id": workID},
		)
		if len(books) == 0 {
			return e.JSON(http.StatusOK, []any{})
		}

		type commentRow struct {
			ID          string  `db:"id" json:"id"`
			UserID      string  `db:"user_id" json:"user_id"`
			Username    string  `db:"username" json:"username"`
			DisplayName *string `db:"display_name" json:"display_name"`
			Avatar      *string `db:"avatar" json:"avatar"`
			Body        string  `db:"body" json:"body"`
			CreatedAt   string  `db:"created_at" json:"created_at"`
		}

		var comments []commentRow
		err := app.DB().NewQuery(`
			SELECT rc.id, rc.user as user_id, u.username, u.display_name, u.avatar,
				   rc.body, rc.created as created_at
			FROM review_comments rc
			JOIN users u ON rc.user = u.id
			WHERE rc.book = {:book} AND rc.review_user = {:review_user}
				AND (rc.deleted_at IS NULL OR rc.deleted_at = '')
			ORDER BY rc.created ASC
		`).Bind(map[string]any{"book": books[0].Id, "review_user": reviewUserID}).All(&comments)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}

		var result []map[string]any
		for _, c := range comments {
			var avatarURL *string
			if c.Avatar != nil && *c.Avatar != "" {
				url := fmt.Sprintf("/api/files/users/%s/%s", c.UserID, *c.Avatar)
				avatarURL = &url
			}
			result = append(result, map[string]any{
				"id":           c.ID,
				"user_id":      c.UserID,
				"username":     c.Username,
				"display_name": c.DisplayName,
				"avatar_url":   avatarURL,
				"body":         c.Body,
				"created_at":   c.CreatedAt,
			})
		}
		if result == nil {
			result = []map[string]any{}
		}

		return e.JSON(http.StatusOK, result)
	}
}

// AddReviewComment handles POST /books/{workId}/reviews/{userId}/comments
func AddReviewComment(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		workID := e.Request.PathValue("workId")
		reviewUserID := e.Request.PathValue("userId")

		books, _ := app.FindRecordsByFilter("books",
			"open_library_id = {:id}", "", 1, 0,
			map[string]any{"id": workID},
		)
		if len(books) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Book not found"})
		}

		// Verify the review exists
		userBooks, err := app.FindRecordsByFilter("user_books",
			"book = {:book} && user = {:user} && review_text != '' && review_text IS NOT NULL",
			"", 1, 0,
			map[string]any{"book": books[0].Id, "user": reviewUserID},
		)
		if err != nil || len(userBooks) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Review not found"})
		}

		data := struct {
			Body string `json:"body"`
		}{}
		if err := e.BindBody(&data); err != nil || data.Body == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "body required"})
		}
		if len(data.Body) > 2000 {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "comment must be 2,000 characters or fewer"})
		}

		coll, err := app.FindCollectionByNameOrId("review_comments")
		if err != nil {
			return err
		}
		rec := core.NewRecord(coll)
		rec.Set("user", user.Id)
		rec.Set("book", books[0].Id)
		rec.Set("review_user", reviewUserID)
		rec.Set("body", data.Body)
		if err := app.Save(rec); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}

		// Notify the review author (unless commenting on own review)
		if user.Id != reviewUserID {
			go notifyReviewComment(app, user, reviewUserID, workID, data.Body)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"id":   rec.Id,
			"body": data.Body,
		})
	}
}

// DeleteReviewComment handles DELETE /review-comments/{commentId}
func DeleteReviewComment(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		commentID := e.Request.PathValue("commentId")

		comment, err := app.FindRecordById("review_comments", commentID)
		if err != nil || comment.GetString("deleted_at") != "" {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Comment not found"})
		}

		// Allow comment author or moderator
		if comment.GetString("user") != user.Id && !user.GetBool("is_moderator") {
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

// notifyReviewComment sends a review_comment notification to the review author.
func notifyReviewComment(app core.App, commenter *core.Record, reviewUserID, bookOLID, body string) {
	if !ShouldNotify(app, reviewUserID, "review_comment") {
		return
	}

	commenterName := commenter.GetString("display_name")
	if commenterName == "" {
		commenterName = commenter.GetString("username")
	}

	preview := body
	if len(preview) > 120 {
		preview = preview[:120] + "..."
	}

	notifColl, err := app.FindCollectionByNameOrId("notifications")
	if err != nil {
		return
	}
	rec := core.NewRecord(notifColl)
	rec.Set("user", reviewUserID)
	rec.Set("notif_type", "review_comment")
	rec.Set("title", fmt.Sprintf("%s commented on your review", commenterName))
	rec.Set("body", preview)
	rec.Set("metadata", map[string]any{
		"book_ol_id":   bookOLID,
		"commenter_id": commenter.Id,
	})
	rec.Set("read", false)
	_ = app.Save(rec)
}
