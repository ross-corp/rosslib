package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

// SendRecommendation handles POST /me/recommendations
func SendRecommendation(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		var body struct {
			Username  string `json:"username"`
			BookOlID  string `json:"book_ol_id"`
			Note      string `json:"note"`
		}
		if err := json.NewDecoder(e.Request.Body).Decode(&body); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}
		if body.Username == "" || body.BookOlID == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "username and book_ol_id are required"})
		}

		// Cannot recommend to self
		if body.Username == user.GetString("username") {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Cannot recommend to yourself"})
		}

		// Find recipient user
		recipients, err := app.FindRecordsByFilter("users",
			"username = {:username}", "", 1, 0,
			map[string]any{"username": body.Username},
		)
		if err != nil || len(recipients) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "User not found"})
		}
		recipient := recipients[0]

		// Find or create the book record
		books, err := app.FindRecordsByFilter("books",
			"open_library_id = {:olid}", "", 1, 0,
			map[string]any{"olid": body.BookOlID},
		)
		if err != nil || len(books) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Book not found"})
		}
		book := books[0]

		// Check for duplicate recommendation
		existing, err := app.FindRecordsByFilter("recommendations",
			"sender = {:sender} && recipient = {:recipient} && book = {:book}",
			"", 1, 0,
			map[string]any{"sender": user.Id, "recipient": recipient.Id, "book": book.Id},
		)
		if err == nil && len(existing) > 0 {
			return e.JSON(http.StatusConflict, map[string]any{"error": "You already recommended this book to this user"})
		}

		// Create recommendation
		coll, err := app.FindCollectionByNameOrId("recommendations")
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to create recommendation"})
		}
		rec := core.NewRecord(coll)
		rec.Set("sender", user.Id)
		rec.Set("recipient", recipient.Id)
		rec.Set("book", book.Id)
		rec.Set("note", body.Note)
		rec.Set("status", "pending")
		if err := app.Save(rec); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to create recommendation"})
		}

		// Create notification for recipient
		senderName := user.GetString("display_name")
		if senderName == "" {
			senderName = user.GetString("username")
		}
		bookTitle := book.GetString("title")

		go func() {
			if !ShouldNotify(app, recipient.Id, "book_recommendation") {
				return
			}

			notifsColl, err := app.FindCollectionByNameOrId("notifications")
			if err != nil {
				return
			}
			notif := core.NewRecord(notifsColl)
			notif.Set("user", recipient.Id)
			notif.Set("notif_type", "book_recommendation")
			notif.Set("title", fmt.Sprintf("%s recommended a book", senderName))
			notif.Set("body", fmt.Sprintf("%s recommended \"%s\"", senderName, bookTitle))
			notif.Set("metadata", map[string]any{
				"sender_username": user.GetString("username"),
				"book_ol_id":     body.BookOlID,
				"book_title":     bookTitle,
				"note":           body.Note,
			})
			notif.Set("read", false)
			_ = app.Save(notif)
		}()

		// Record activity
		recordActivity(app, user.Id, "sent_recommendation", map[string]any{
			"book":        book.Id,
			"target_user": recipient.Id,
			"metadata": map[string]any{
				"note": body.Note,
			},
		})

		return e.JSON(http.StatusCreated, map[string]any{
			"id":      rec.Id,
			"message": "Recommendation sent",
		})
	}
}

// GetRecommendations handles GET /me/recommendations?status=pending
func GetRecommendations(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		status := e.Request.URL.Query().Get("status")
		if status == "" {
			status = "pending"
		}

		type recRow struct {
			ID              string  `db:"id"`
			Note            *string `db:"note"`
			Status          string  `db:"status"`
			Created         string  `db:"created"`
			SenderID        string  `db:"sender_id"`
			SenderUsername   string  `db:"sender_username"`
			SenderDisplay   *string `db:"sender_display"`
			SenderAvatar    *string `db:"sender_avatar"`
			BookID          string  `db:"book_id"`
			BookOLID        string  `db:"book_olid"`
			BookTitle       string  `db:"book_title"`
			BookCoverURL    *string `db:"book_cover_url"`
			BookAuthors     *string `db:"book_authors"`
		}

		query := `
			SELECT r.id, r.note, r.status, r.created,
				   u.id as sender_id, u.username as sender_username, u.display_name as sender_display, u.avatar as sender_avatar,
				   b.id as book_id, b.open_library_id as book_olid, b.title as book_title, b.cover_url as book_cover_url, b.authors as book_authors
			FROM recommendations r
			JOIN users u ON r.sender = u.id
			JOIN books b ON r.book = b.id
			WHERE r.recipient = {:user}
		`
		params := map[string]any{"user": user.Id}

		if status != "all" {
			query += " AND r.status = {:status}"
			params["status"] = status
		}
		query += " ORDER BY r.created DESC LIMIT 50"

		var rows []recRow
		err := app.DB().NewQuery(query).Bind(params).All(&rows)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}

		result := make([]map[string]any, 0, len(rows))
		for _, row := range rows {
			var senderAvatarURL *string
			if row.SenderAvatar != nil && *row.SenderAvatar != "" {
				url := fmt.Sprintf("/api/files/users/%s/%s", row.SenderID, *row.SenderAvatar)
				senderAvatarURL = &url
			}

			item := map[string]any{
				"id":         row.ID,
				"note":       row.Note,
				"status":     row.Status,
				"created_at": row.Created,
				"sender": map[string]any{
					"user_id":      row.SenderID,
					"username":     row.SenderUsername,
					"display_name": row.SenderDisplay,
					"avatar_url":   senderAvatarURL,
				},
				"book": map[string]any{
					"open_library_id": row.BookOLID,
					"title":           row.BookTitle,
					"cover_url":       row.BookCoverURL,
					"authors":         row.BookAuthors,
				},
			}
			result = append(result, item)
		}

		return e.JSON(http.StatusOK, result)
	}
}

// UpdateRecommendation handles PATCH /me/recommendations/{recId}
func UpdateRecommendation(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		recID := e.Request.PathValue("recId")

		rec, err := app.FindRecordById("recommendations", recID)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Recommendation not found"})
		}

		// Only recipient can update
		if rec.GetString("recipient") != user.Id {
			return e.JSON(http.StatusForbidden, map[string]any{"error": "Not your recommendation"})
		}

		var body struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(e.Request.Body).Decode(&body); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}

		if body.Status != "seen" && body.Status != "dismissed" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Status must be 'seen' or 'dismissed'"})
		}

		rec.Set("status", body.Status)
		if err := app.Save(rec); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to update recommendation"})
		}

		return e.JSON(http.StatusOK, map[string]any{"message": "Recommendation updated"})
	}
}
