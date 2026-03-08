package handlers

import (
	"fmt"
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

// GetBookQuotes handles GET /books/{workId}/quotes
// Returns public quotes for a book, paginated.
func GetBookQuotes(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		workID := e.Request.PathValue("workId")

		books, _ := app.FindRecordsByFilter("books",
			"open_library_id = {:id}", "", 1, 0,
			map[string]any{"id": workID},
		)
		if len(books) == 0 {
			return e.JSON(http.StatusOK, []any{})
		}

		page := 1
		if p := e.Request.URL.Query().Get("page"); p != "" {
			_, _ = fmt.Sscanf(p, "%d", &page)
		}
		if page < 1 {
			page = 1
		}
		limit := 20
		offset := (page - 1) * limit

		type quoteRow struct {
			ID          string  `db:"id" json:"id"`
			UserID      string  `db:"user_id" json:"user_id"`
			Username    string  `db:"username" json:"username"`
			DisplayName *string `db:"display_name" json:"display_name"`
			Avatar      *string `db:"avatar" json:"avatar"`
			Text        string  `db:"text" json:"text"`
			PageNumber  *int    `db:"page_number" json:"page_number"`
			Note        *string `db:"note" json:"note"`
			CreatedAt   string  `db:"created_at" json:"created_at"`
		}

		var quotes []quoteRow
		err := app.DB().NewQuery(`
			SELECT q.id, q.user as user_id, u.username, u.display_name, u.avatar,
				   q.text, q.page_number, q.note, q.created as created_at
			FROM book_quotes q
			JOIN users u ON q.user = u.id
			WHERE q.book = {:book} AND q.is_public = true
			ORDER BY q.created DESC
			LIMIT {:limit} OFFSET {:offset}
		`).Bind(map[string]any{
			"book":   books[0].Id,
			"limit":  limit,
			"offset": offset,
		}).All(&quotes)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}

		var result []map[string]any
		for _, q := range quotes {
			var avatarURL *string
			if q.Avatar != nil && *q.Avatar != "" {
				url := fmt.Sprintf("/api/files/users/%s/%s", q.UserID, *q.Avatar)
				avatarURL = &url
			}
			row := map[string]any{
				"id":           q.ID,
				"user_id":      q.UserID,
				"username":     q.Username,
				"display_name": q.DisplayName,
				"avatar_url":   avatarURL,
				"text":         q.Text,
				"page_number":  q.PageNumber,
				"note":         q.Note,
				"created_at":   q.CreatedAt,
			}
			result = append(result, row)
		}
		if result == nil {
			result = []map[string]any{}
		}

		return e.JSON(http.StatusOK, result)
	}
}

// GetMyBookQuotes handles GET /me/books/{olId}/quotes
// Returns the authenticated user's quotes for a book.
func GetMyBookQuotes(app core.App) func(e *core.RequestEvent) error {
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

		type quoteRow struct {
			ID         string  `db:"id" json:"id"`
			Text       string  `db:"text" json:"text"`
			PageNumber *int    `db:"page_number" json:"page_number"`
			Note       *string `db:"note" json:"note"`
			IsPublic   bool    `db:"is_public" json:"is_public"`
			CreatedAt  string  `db:"created_at" json:"created_at"`
		}

		var quotes []quoteRow
		err := app.DB().NewQuery(`
			SELECT id, text, page_number, note, is_public, created as created_at
			FROM book_quotes
			WHERE user = {:user} AND book = {:book}
			ORDER BY created DESC
		`).Bind(map[string]any{
			"user": user.Id,
			"book": books[0].Id,
		}).All(&quotes)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}

		var result []map[string]any
		for _, q := range quotes {
			result = append(result, map[string]any{
				"id":          q.ID,
				"text":        q.Text,
				"page_number": q.PageNumber,
				"note":        q.Note,
				"is_public":   q.IsPublic,
				"created_at":  q.CreatedAt,
			})
		}
		if result == nil {
			result = []map[string]any{}
		}

		return e.JSON(http.StatusOK, result)
	}
}

// CreateBookQuote handles POST /me/books/{olId}/quotes
// Creates a new quote for a book.
func CreateBookQuote(app core.App) func(e *core.RequestEvent) error {
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

		data := struct {
			Text       string `json:"text"`
			PageNumber *int   `json:"page_number"`
			Note       string `json:"note"`
			IsPublic   *bool  `json:"is_public"`
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}
		if data.Text == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "text is required"})
		}
		if len(data.Text) > 2000 {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "text must be 2000 characters or fewer"})
		}
		if len(data.Note) > 500 {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "note must be 500 characters or fewer"})
		}

		isPublic := true
		if data.IsPublic != nil {
			isPublic = *data.IsPublic
		}

		coll, err := app.FindCollectionByNameOrId("book_quotes")
		if err != nil {
			return err
		}
		rec := core.NewRecord(coll)
		rec.Set("user", user.Id)
		rec.Set("book", books[0].Id)
		rec.Set("text", data.Text)
		if data.PageNumber != nil {
			rec.Set("page_number", *data.PageNumber)
		}
		rec.Set("note", data.Note)
		rec.Set("is_public", isPublic)

		if err := app.Save(rec); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}

		return e.JSON(http.StatusOK, map[string]any{
			"id":         rec.Id,
			"text":       data.Text,
			"created_at": rec.GetString("created"),
		})
	}
}

// DeleteBookQuote handles DELETE /me/quotes/{quoteId}
// Deletes a quote owned by the authenticated user.
func DeleteBookQuote(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		quoteID := e.Request.PathValue("quoteId")

		quote, err := app.FindRecordById("book_quotes", quoteID)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Quote not found"})
		}
		if quote.GetString("user") != user.Id {
			return e.JSON(http.StatusForbidden, map[string]any{"error": "Not your quote"})
		}

		if err := app.Delete(quote); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to delete"})
		}

		e.Response.WriteHeader(http.StatusNoContent)
		return nil
	}
}
