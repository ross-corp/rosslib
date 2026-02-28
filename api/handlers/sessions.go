package handlers

import (
	"net/http"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// GetSessions handles GET /me/books/{olId}/sessions
func GetSessions(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		olID := e.Request.PathValue("olId")

		books, _ := app.FindRecordsByFilter("books",
			"open_library_id = {:id}", "", 1, 0,
			map[string]any{"id": olID},
		)
		if len(books) == 0 {
			return e.JSON(http.StatusOK, []any{})
		}
		book := books[0]

		type sessionRow struct {
			ID           string   `db:"id" json:"id"`
			DateStarted  *string  `db:"date_started" json:"date_started"`
			DateFinished *string  `db:"date_finished" json:"date_finished"`
			Rating       *float64 `db:"rating" json:"rating"`
			Notes        *string  `db:"notes" json:"notes"`
			Created      string   `db:"created" json:"created"`
		}
		var sessions []sessionRow
		err := app.DB().NewQuery(`
			SELECT id, date_started, date_finished, rating, notes, created
			FROM reading_sessions
			WHERE "user" = {:user} AND book = {:book}
			ORDER BY created DESC
		`).Bind(map[string]any{"user": user.Id, "book": book.Id}).All(&sessions)
		if err != nil || sessions == nil {
			sessions = []sessionRow{}
		}

		return e.JSON(http.StatusOK, sessions)
	}
}

// CreateSession handles POST /me/books/{olId}/sessions
func CreateSession(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		olID := e.Request.PathValue("olId")

		books, _ := app.FindRecordsByFilter("books",
			"open_library_id = {:id}", "", 1, 0,
			map[string]any{"id": olID},
		)
		if len(books) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Book not found"})
		}
		book := books[0]

		// Verify the user has this book in their library
		ubs, _ := app.FindRecordsByFilter("user_books",
			"user = {:user} && book = {:book}",
			"", 1, 0,
			map[string]any{"user": user.Id, "book": book.Id},
		)
		if len(ubs) == 0 {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Book not in your library"})
		}

		data := struct {
			DateStarted  *string  `json:"date_started"`
			DateFinished *string  `json:"date_finished"`
			Rating       *float64 `json:"rating"`
			Notes        *string  `json:"notes"`
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}

		if data.Rating != nil && (*data.Rating < 1 || *data.Rating > 5) {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Rating must be between 1 and 5"})
		}

		// Validate dates if provided
		if data.DateStarted != nil && *data.DateStarted != "" {
			if _, err := time.Parse(time.RFC3339, *data.DateStarted); err != nil {
				if _, err2 := time.Parse("2006-01-02", *data.DateStarted); err2 != nil {
					return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid date_started format"})
				}
			}
		}
		if data.DateFinished != nil && *data.DateFinished != "" {
			if _, err := time.Parse(time.RFC3339, *data.DateFinished); err != nil {
				if _, err2 := time.Parse("2006-01-02", *data.DateFinished); err2 != nil {
					return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid date_finished format"})
				}
			}
		}

		coll, err := app.FindCollectionByNameOrId("reading_sessions")
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to find collection"})
		}

		rec := core.NewRecord(coll)
		rec.Set("user", user.Id)
		rec.Set("book", book.Id)
		if data.DateStarted != nil {
			rec.Set("date_started", *data.DateStarted)
		}
		if data.DateFinished != nil {
			rec.Set("date_finished", *data.DateFinished)
		}
		if data.Rating != nil {
			rec.Set("rating", *data.Rating)
		}
		if data.Notes != nil {
			rec.Set("notes", *data.Notes)
		}

		if err := app.Save(rec); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to save session"})
		}

		return e.JSON(http.StatusOK, map[string]any{
			"id":            rec.Id,
			"date_started":  rec.GetString("date_started"),
			"date_finished": rec.GetString("date_finished"),
			"rating":        rec.Get("rating"),
			"notes":         rec.GetString("notes"),
		})
	}
}

// UpdateSession handles PATCH /me/sessions/{sessionId}
func UpdateSession(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		sessionId := e.Request.PathValue("sessionId")

		rec, err := app.FindRecordById("reading_sessions", sessionId)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Session not found"})
		}

		if rec.GetString("user") != user.Id {
			return e.JSON(http.StatusForbidden, map[string]any{"error": "Not your session"})
		}

		data := struct {
			DateStarted  *string  `json:"date_started"`
			DateFinished *string  `json:"date_finished"`
			Rating       *float64 `json:"rating"`
			Notes        *string  `json:"notes"`
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}

		if data.Rating != nil && *data.Rating != 0 && (*data.Rating < 1 || *data.Rating > 5) {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Rating must be between 1 and 5"})
		}

		if data.DateStarted != nil {
			rec.Set("date_started", *data.DateStarted)
		}
		if data.DateFinished != nil {
			rec.Set("date_finished", *data.DateFinished)
		}
		if data.Rating != nil {
			rec.Set("rating", *data.Rating)
		}
		if data.Notes != nil {
			rec.Set("notes", *data.Notes)
		}

		if err := app.Save(rec); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to update session"})
		}

		return e.JSON(http.StatusOK, map[string]any{"message": "Session updated"})
	}
}

// DeleteSession handles DELETE /me/sessions/{sessionId}
func DeleteSession(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		sessionId := e.Request.PathValue("sessionId")

		rec, err := app.FindRecordById("reading_sessions", sessionId)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Session not found"})
		}

		if rec.GetString("user") != user.Id {
			return e.JSON(http.StatusForbidden, map[string]any{"error": "Not your session"})
		}

		if err := app.Delete(rec); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to delete session"})
		}

		return e.JSON(http.StatusOK, map[string]any{"message": "Session deleted"})
	}
}
