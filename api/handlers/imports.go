package handlers

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// PreviewGoodreadsImport handles POST /me/import/goodreads/preview
func PreviewGoodreadsImport(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		file, _, err := e.Request.FormFile("file")
		if err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "CSV file required"})
		}
		defer file.Close()

		reader := csv.NewReader(file)
		headers, err := reader.Read()
		if err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid CSV"})
		}

		// Map header indices
		colIdx := map[string]int{}
		for i, h := range headers {
			colIdx[strings.TrimSpace(h)] = i
		}

		// Required columns
		requiredCols := []string{"Title"}
		for _, c := range requiredCols {
			if _, ok := colIdx[c]; !ok {
				return e.JSON(http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("Missing column: %s", c)})
			}
		}

		type previewBook struct {
			Title         string  `json:"title"`
			Authors       string  `json:"authors"`
			ISBN          string  `json:"isbn"`
			ISBN13        string  `json:"isbn13"`
			Rating        float64 `json:"rating"`
			Shelf         string  `json:"shelf"`
			DateRead      string  `json:"date_read"`
			DateAdded     string  `json:"date_added"`
			Review        string  `json:"review"`
			OpenLibraryID string  `json:"open_library_id"`
			CoverURL      string  `json:"cover_url"`
			Found         bool    `json:"found"`
		}

		var rows [][]string
		for {
			row, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				continue
			}
			rows = append(rows, row)
		}

		results := make([]previewBook, len(rows))
		ol := newOLClient()

		// Worker pool for OL lookups
		var wg sync.WaitGroup
		sem := make(chan struct{}, 5) // 5 concurrent workers

		getCol := func(row []string, name string) string {
			if idx, ok := colIdx[name]; ok && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}

		for i, row := range rows {
			wg.Add(1)
			go func(i int, row []string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				pb := previewBook{
					Title:     getCol(row, "Title"),
					Authors:   getCol(row, "Author"),
					ISBN:      strings.Trim(getCol(row, "ISBN"), "=\""),
					ISBN13:    strings.Trim(getCol(row, "ISBN13"), "=\""),
					Shelf:     getCol(row, "Exclusive Shelf"),
					DateRead:  getCol(row, "Date Read"),
					DateAdded: getCol(row, "Date Added"),
					Review:    getCol(row, "My Review"),
				}

				if r, err := strconv.ParseFloat(getCol(row, "My Rating"), 64); err == nil {
					pb.Rating = r
				}

				// Try to find via ISBN
				isbn := pb.ISBN13
				if isbn == "" {
					isbn = pb.ISBN
				}
				if isbn != "" {
					data, err := ol.get(fmt.Sprintf("/isbn/%s.json", isbn))
					if err == nil {
						if works, ok := data["works"].([]any); ok && len(works) > 0 {
							if w, ok := works[0].(map[string]any); ok {
								if key, ok := w["key"].(string); ok {
									pb.OpenLibraryID = strings.TrimPrefix(key, "/works/")
									pb.Found = true
								}
							}
						}
						if covers, ok := data["covers"].([]any); ok && len(covers) > 0 {
							if coverID, ok := covers[0].(float64); ok {
								pb.CoverURL = fmt.Sprintf("https://covers.openlibrary.org/b/id/%.0f-M.jpg", coverID)
							}
						}
					}
				}

				// Fallback: search by title
				if !pb.Found && pb.Title != "" {
					data, err := ol.get(fmt.Sprintf("/search.json?title=%s&limit=1", pb.Title))
					if err == nil {
						if docs, ok := data["docs"].([]any); ok && len(docs) > 0 {
							if doc, ok := docs[0].(map[string]any); ok {
								if key, ok := doc["key"].(string); ok {
									pb.OpenLibraryID = strings.TrimPrefix(key, "/works/")
									pb.Found = true
								}
								if coverI, ok := doc["cover_i"].(float64); ok && coverI > 0 {
									pb.CoverURL = fmt.Sprintf("https://covers.openlibrary.org/b/id/%.0f-M.jpg", coverI)
								}
							}
						}
					}
				}

				results[i] = pb
			}(i, row)
		}

		wg.Wait()

		found := 0
		for _, r := range results {
			if r.Found {
				found++
			}
		}

		return e.JSON(http.StatusOK, map[string]any{
			"total":   len(results),
			"found":   found,
			"missing": len(results) - found,
			"books":   results,
		})
	}
}

// CommitGoodreadsImport handles POST /me/import/goodreads/commit
func CommitGoodreadsImport(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		data := struct {
			Books []struct {
				OpenLibraryID string   `json:"open_library_id"`
				Title         string   `json:"title"`
				Authors       string   `json:"authors"`
				ISBN13        string   `json:"isbn13"`
				CoverURL      string   `json:"cover_url"`
				Rating        float64  `json:"rating"`
				Shelf         string   `json:"shelf"`
				DateRead      string   `json:"date_read"`
				DateAdded     string   `json:"date_added"`
				Review        string   `json:"review"`
			} `json:"books"`
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}

		imported := 0
		skipped := 0

		for _, b := range data.Books {
			if b.OpenLibraryID == "" {
				skipped++
				continue
			}

			// Upsert book
			book, err := upsertBook(app, b.OpenLibraryID, b.Title, b.CoverURL, b.ISBN13, b.Authors, 0)
			if err != nil {
				skipped++
				continue
			}

			// Upsert user_book
			existing, _ := app.FindRecordsByFilter("user_books",
				"user = {:user} && book = {:book}",
				"", 1, 0,
				map[string]any{"user": user.Id, "book": book.Id},
			)
			var ub *core.Record
			if len(existing) > 0 {
				ub = existing[0]
			} else {
				coll, err := app.FindCollectionByNameOrId("user_books")
				if err != nil {
					skipped++
					continue
				}
				ub = core.NewRecord(coll)
				ub.Set("user", user.Id)
				ub.Set("book", book.Id)
			}

			if b.Rating > 0 {
				ub.Set("rating", b.Rating)
			}
			if b.Review != "" {
				ub.Set("review_text", b.Review)
			}
			if b.DateRead != "" {
				ub.Set("date_read", b.DateRead)
			}
			dateAdded := b.DateAdded
			if dateAdded == "" {
				dateAdded = time.Now().UTC().Format(time.RFC3339)
			}
			ub.Set("date_added", dateAdded)

			if err := app.Save(ub); err != nil {
				skipped++
				continue
			}

			// Map Goodreads shelf to status
			statusSlug := mapGoodreadsShelf(b.Shelf)
			if statusSlug != "" {
				setStatusTag(app, user.Id, book.Id, statusSlug)
			}

			imported++
			refreshBookStats(app, book.Id)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"imported": imported,
			"skipped":  skipped,
		})
	}
}

func mapGoodreadsShelf(shelf string) string {
	switch strings.ToLower(shelf) {
	case "to-read":
		return "want-to-read"
	case "currently-reading":
		return "currently-reading"
	case "read":
		return "finished"
	default:
		return ""
	}
}

