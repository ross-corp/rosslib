package handlers

import (
	"net/http"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

// SearchSeries handles GET /series/search?q=<name>
func SearchSeries(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		q := strings.TrimSpace(e.Request.URL.Query().Get("q"))
		if q == "" {
			return e.JSON(http.StatusOK, map[string]any{"results": []any{}})
		}

		type seriesRow struct {
			ID          string  `db:"id" json:"id"`
			Name        string  `db:"name" json:"name"`
			Description *string `db:"description" json:"description"`
			BookCount   int     `db:"book_count" json:"book_count"`
		}

		var rows []seriesRow
		err := app.DB().NewQuery(`
			SELECT s.id, s.name, s.description,
				   (SELECT COUNT(*) FROM book_series bs WHERE bs.series = s.id) as book_count
			FROM series s
			WHERE s.name LIKE {:pattern}
			ORDER BY book_count DESC, s.name
			LIMIT 20
		`).Bind(map[string]any{"pattern": "%" + q + "%"}).All(&rows)
		if err != nil {
			return e.JSON(http.StatusOK, map[string]any{"results": []any{}})
		}

		if len(rows) == 0 {
			return e.JSON(http.StatusOK, map[string]any{"results": []any{}})
		}

		return e.JSON(http.StatusOK, map[string]any{"results": rows})
	}
}

// GetAuthorSeries handles GET /authors/{authorKey}/series?name=...
// Finds series containing books whose authors field matches the given name.
func GetAuthorSeries(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		authorName := strings.TrimSpace(e.Request.URL.Query().Get("name"))
		if authorName == "" {
			return e.JSON(http.StatusOK, []any{})
		}

		type seriesRow struct {
			ID          string  `db:"id" json:"id"`
			Name        string  `db:"name" json:"name"`
			Description *string `db:"description" json:"description"`
			BookCount   int     `db:"book_count" json:"book_count"`
		}

		var rows []seriesRow
		err := app.DB().NewQuery(`
			SELECT DISTINCT s.id, s.name, s.description,
				   (SELECT COUNT(*) FROM book_series bs2 WHERE bs2.series = s.id) as book_count
			FROM series s
			JOIN book_series bs ON bs.series = s.id
			JOIN books b ON bs.book = b.id
			WHERE b.authors LIKE {:pattern}
			ORDER BY s.name
		`).Bind(map[string]any{"pattern": "%" + authorName + "%"}).All(&rows)
		if err != nil || len(rows) == 0 {
			return e.JSON(http.StatusOK, []any{})
		}

		return e.JSON(http.StatusOK, rows)
	}
}

// GetBookSeries handles GET /books/{workId}/series
func GetBookSeries(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		workID := e.Request.PathValue("workId")

		books, _ := app.FindRecordsByFilter("books",
			"open_library_id = {:id}", "", 1, 0,
			map[string]any{"id": workID},
		)
		if len(books) == 0 {
			return e.JSON(http.StatusOK, []any{})
		}

		type seriesRow struct {
			SeriesID    string  `db:"series_id" json:"series_id"`
			Name        string  `db:"name" json:"name"`
			Description *string `db:"description" json:"description"`
			Position    *int    `db:"position" json:"position"`
		}

		var rows []seriesRow
		err := app.DB().NewQuery(`
			SELECT s.id as series_id, s.name, s.description, bs.position
			FROM book_series bs
			JOIN series s ON bs.series = s.id
			WHERE bs.book = {:book}
			ORDER BY s.name
		`).Bind(map[string]any{"book": books[0].Id}).All(&rows)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}

		if len(rows) == 0 {
			return e.JSON(http.StatusOK, []any{})
		}

		return e.JSON(http.StatusOK, rows)
	}
}

// GetSeriesDetail handles GET /series/{seriesId}
func GetSeriesDetail(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		seriesID := e.Request.PathValue("seriesId")

		series, err := app.FindRecordById("series", seriesID)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Series not found"})
		}

		type bookRow struct {
			BookID         string  `db:"book_id" json:"book_id"`
			OpenLibraryID  string  `db:"open_library_id" json:"open_library_id"`
			Title          string  `db:"title" json:"title"`
			CoverURL       *string `db:"cover_url" json:"cover_url"`
			Authors        *string `db:"authors" json:"authors"`
			Position       *int    `db:"position" json:"position"`
		}

		var books []bookRow
		err = app.DB().NewQuery(`
			SELECT b.id as book_id, b.open_library_id, b.title, b.cover_url, b.authors, bs.position
			FROM book_series bs
			JOIN books b ON bs.book = b.id
			WHERE bs.series = {:series}
			ORDER BY CASE WHEN bs.position IS NULL THEN 1 ELSE 0 END, bs.position, b.title
		`).Bind(map[string]any{"series": seriesID}).All(&books)
		if err != nil {
			books = []bookRow{}
		}

		// Check viewer's reading progress for these books
		viewerID := ""
		if e.Auth != nil {
			viewerID = e.Auth.Id
		}

		progressMap := map[string]string{}
		if viewerID != "" && len(books) > 0 {
			var bookIDs []any
			for _, b := range books {
				bookIDs = append(bookIDs, b.BookID)
			}
			type statusRow struct {
				BookID string `db:"book_id"`
				Slug   string `db:"slug"`
			}
			var statuses []statusRow
			_ = app.DB().NewQuery(`
				SELECT btv.book AS book_id, tv.slug
				FROM book_tag_values btv
				JOIN tag_values tv ON btv.tag_value = tv.id
				JOIN tag_keys tk ON tv.tag_key = tk.id
				WHERE btv.user = {:user} AND btv.book IN {:bookIds}
				  AND tk.slug = 'status'
			`).Bind(map[string]any{
				"user":    viewerID,
				"bookIds": bookIDs,
			}).All(&statuses)
			for _, s := range statuses {
				progressMap[s.BookID] = s.Slug
			}
		}

		var result []map[string]any
		for _, b := range books {
			entry := map[string]any{
				"book_id":         b.BookID,
				"open_library_id": b.OpenLibraryID,
				"title":           b.Title,
				"cover_url":       b.CoverURL,
				"authors":         b.Authors,
				"position":        b.Position,
			}
			if status, ok := progressMap[b.BookID]; ok {
				entry["viewer_status"] = status
			}
			result = append(result, entry)
		}
		if result == nil {
			result = []map[string]any{}
		}

		return e.JSON(http.StatusOK, map[string]any{
			"id":          series.Id,
			"name":        series.GetString("name"),
			"description": series.GetString("description"),
			"books":       result,
		})
	}
}

// UpdateSeries handles PATCH /series/{seriesId}
func UpdateSeries(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
		}

		seriesID := e.Request.PathValue("seriesId")

		series, err := app.FindRecordById("series", seriesID)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Series not found"})
		}

		data := struct {
			Name        *string `json:"name"`
			Description *string `json:"description"`
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}

		if data.Name != nil {
			trimmed := strings.TrimSpace(*data.Name)
			if trimmed == "" {
				return e.JSON(http.StatusBadRequest, map[string]any{"error": "name cannot be empty"})
			}
			if len(trimmed) > 255 {
				return e.JSON(http.StatusBadRequest, map[string]any{"error": "name must be 255 characters or fewer"})
			}
			series.Set("name", trimmed)
		}
		if data.Description != nil {
			desc := strings.TrimSpace(*data.Description)
			if len(desc) > 5000 {
				return e.JSON(http.StatusBadRequest, map[string]any{"error": "description must be 5000 characters or fewer"})
			}
			series.Set("description", desc)
		}

		if err := app.Save(series); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to update series"})
		}

		return e.JSON(http.StatusOK, map[string]any{
			"id":          series.Id,
			"name":        series.GetString("name"),
			"description": series.GetString("description"),
		})
	}
}

// DeleteSeries handles DELETE /series/{seriesId}
func DeleteSeries(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
		}

		seriesID := e.Request.PathValue("seriesId")

		series, err := app.FindRecordById("series", seriesID)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Series not found"})
		}

		// Check that the series has zero book_series links
		var count struct {
			N int `db:"n"`
		}
		err = app.DB().NewQuery(`
			SELECT COUNT(*) as n FROM book_series WHERE series = {:series}
		`).Bind(map[string]any{"series": seriesID}).One(&count)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to check series books"})
		}
		if count.N > 0 {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Cannot delete series that still has books"})
		}

		if err := app.Delete(series); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to delete series"})
		}

		return e.JSON(http.StatusOK, map[string]any{"ok": true})
	}
}

// AddBookToSeries handles POST /books/{workId}/series
func AddBookToSeries(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
		}

		workID := e.Request.PathValue("workId")

		data := struct {
			SeriesName string `json:"series_name"`
			Position   *int   `json:"position"`
		}{}
		if err := e.BindBody(&data); err != nil || data.SeriesName == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "series_name required"})
		}
		if len(data.SeriesName) > 255 {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "series_name must be 255 characters or fewer"})
		}

		// Find the book
		books, _ := app.FindRecordsByFilter("books",
			"open_library_id = {:id}", "", 1, 0,
			map[string]any{"id": workID},
		)
		if len(books) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Book not found"})
		}

		// Find or create series by name
		seriesName := strings.TrimSpace(data.SeriesName)
		existing, _ := app.FindRecordsByFilter("series",
			"name = {:name}", "", 1, 0,
			map[string]any{"name": seriesName},
		)

		var seriesRec *core.Record
		if len(existing) > 0 {
			seriesRec = existing[0]
		} else {
			coll, err := app.FindCollectionByNameOrId("series")
			if err != nil {
				return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to find series collection"})
			}
			seriesRec = core.NewRecord(coll)
			seriesRec.Set("name", seriesName)
			if err := app.Save(seriesRec); err != nil {
				return e.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
			}
		}

		// Check if book-series link already exists
		existingLink, _ := app.FindRecordsByFilter("book_series",
			"book = {:book} && series = {:series}", "", 1, 0,
			map[string]any{"book": books[0].Id, "series": seriesRec.Id},
		)
		if len(existingLink) > 0 {
			// Update position if provided
			if data.Position != nil {
				existingLink[0].Set("position", *data.Position)
				if err := app.Save(existingLink[0]); err != nil {
					return e.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
				}
			}
			return e.JSON(http.StatusOK, map[string]any{
				"series_id": seriesRec.Id,
				"name":      seriesRec.GetString("name"),
				"position":  data.Position,
			})
		}

		// Create book_series link
		bsColl, err := app.FindCollectionByNameOrId("book_series")
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to find book_series collection"})
		}
		bs := core.NewRecord(bsColl)
		bs.Set("book", books[0].Id)
		bs.Set("series", seriesRec.Id)
		if data.Position != nil {
			bs.Set("position", *data.Position)
		}
		if err := app.Save(bs); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}

		return e.JSON(http.StatusCreated, map[string]any{
			"series_id": seriesRec.Id,
			"name":      seriesRec.GetString("name"),
			"position":  data.Position,
		})
	}
}
