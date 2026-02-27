package handlers

import (
	"net/http"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

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
			for _, b := range books {
				type statusResult struct {
					Slug string `db:"slug"`
				}
				var result statusResult
				err = app.DB().NewQuery(`
					SELECT tv.slug
					FROM book_tag_values btv
					JOIN tag_values tv ON btv.tag_value = tv.id
					JOIN tag_keys tk ON tv.tag_key = tk.id
					WHERE btv.user = {:user} AND btv.book = {:book}
					  AND tk.slug = 'status'
					LIMIT 1
				`).Bind(map[string]any{
					"user": viewerID,
					"book": b.BookID,
				}).One(&result)
				if err == nil && result.Slug != "" {
					progressMap[b.BookID] = result.Slug
				}
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
