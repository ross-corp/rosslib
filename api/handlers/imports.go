package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// PreviewGoodreadsImport handles POST /me/import/goodreads/preview
// Streams NDJSON: progress lines followed by a final result line.
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

		requiredCols := []string{"Title"}
		for _, c := range requiredCols {
			if _, ok := colIdx[c]; !ok {
				return e.JSON(http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("Missing column: %s", c)})
			}
		}

		// Types matching the webapp's PreviewRow / BookCandidate
		type bookCandidate struct {
			OLID     string   `json:"ol_id"`
			Title    string   `json:"title"`
			Authors  []string `json:"authors"`
			CoverURL *string  `json:"cover_url"`
			Year     *int     `json:"year"`
		}
		type previewRow struct {
			RowID              int            `json:"row_id"`
			Title              string         `json:"title"`
			Author             string         `json:"author"`
			ISBN13             string         `json:"isbn13"`
			Rating             *float64       `json:"rating"`
			ReviewText         *string        `json:"review_text"`
			Spoiler            bool           `json:"spoiler"`
			DateRead           *string        `json:"date_read"`
			DateAdded          *string        `json:"date_added"`
			ExclusiveShelfSlug string         `json:"exclusive_shelf_slug"`
			CustomShelves      []string       `json:"custom_shelves"`
			Status             string         `json:"status"`
			Match              *bookCandidate `json:"match,omitempty"`
		}

		var csvRows [][]string
		for {
			row, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				continue
			}
			csvRows = append(csvRows, row)
		}

		results := make([]previewRow, len(csvRows))
		ol := newOLClient()

		var wg sync.WaitGroup
		sem := make(chan struct{}, 5)

		getCol := func(row []string, name string) string {
			if idx, ok := colIdx[name]; ok && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}

		// Set up NDJSON streaming
		w := e.Response
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		flusher, ok := w.(http.Flusher)
		if !ok {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Streaming not supported"})
		}
		enc := json.NewEncoder(w)

		type progressMsg struct {
			Index int
			Row   previewRow
		}
		ch := make(chan progressMsg, len(csvRows))

		for i, row := range csvRows {
			wg.Add(1)
			go func(i int, row []string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				pr := previewRow{
					RowID:  i,
					Title:  getCol(row, "Title"),
					Author: getCol(row, "Author"),
					ISBN13: strings.Trim(getCol(row, "ISBN13"), "=\""),
				}

				// Map exclusive shelf — pass through Goodreads name
				pr.ExclusiveShelfSlug = strings.ToLower(getCol(row, "Exclusive Shelf"))

				if r, err := strconv.ParseFloat(getCol(row, "My Rating"), 64); err == nil && r > 0 {
					pr.Rating = &r
				}
				if review := getCol(row, "My Review"); review != "" {
					pr.ReviewText = &review
				}
				if dr := getCol(row, "Date Read"); dr != "" {
					pr.DateRead = &dr
				}
				if da := getCol(row, "Date Added"); da != "" {
					pr.DateAdded = &da
				}

				// Parse custom shelves from Bookshelves column
				if bookshelves := getCol(row, "Bookshelves"); bookshelves != "" {
					for _, s := range strings.Split(bookshelves, ",") {
						s = strings.TrimSpace(s)
						if s != "" && s != "to-read" && s != "currently-reading" && s != "read" {
							pr.CustomShelves = append(pr.CustomShelves, s)
						}
					}
				}
				if pr.CustomShelves == nil {
					pr.CustomShelves = []string{}
				}

				// Try to find via ISBN
				isbn := pr.ISBN13
				if isbn == "" {
					isbn = strings.Trim(getCol(row, "ISBN"), "=\"")
				}

				found := false
				var olID, matchTitle, coverURL string
				var authors []string

				if isbn != "" {
					data, err := ol.get(fmt.Sprintf("/isbn/%s.json", isbn))
					if err == nil {
						if works, ok := data["works"].([]any); ok && len(works) > 0 {
							if w, ok := works[0].(map[string]any); ok {
								if key, ok := w["key"].(string); ok {
									olID = strings.TrimPrefix(key, "/works/")
									found = true
								}
							}
						}
						if t, ok := data["title"].(string); ok {
							matchTitle = t
						}
						if covers, ok := data["covers"].([]any); ok && len(covers) > 0 {
							if coverID, ok := covers[0].(float64); ok {
								coverURL = fmt.Sprintf("https://covers.openlibrary.org/b/id/%.0f-M.jpg", coverID)
							}
						}
					}
				}

				// Fallback: search by title
				if !found && pr.Title != "" {
					data, err := ol.get(fmt.Sprintf("/search.json?title=%s&limit=1", pr.Title))
					if err == nil {
						if docs, ok := data["docs"].([]any); ok && len(docs) > 0 {
							if doc, ok := docs[0].(map[string]any); ok {
								if key, ok := doc["key"].(string); ok {
									olID = strings.TrimPrefix(key, "/works/")
									found = true
								}
								if t, ok := doc["title"].(string); ok {
									matchTitle = t
								}
								if coverI, ok := doc["cover_i"].(float64); ok && coverI > 0 {
									coverURL = fmt.Sprintf("https://covers.openlibrary.org/b/id/%.0f-M.jpg", coverI)
								}
								if authorNames, ok := doc["author_name"].([]any); ok {
									for _, a := range authorNames {
										if name, ok := a.(string); ok {
											authors = append(authors, name)
										}
									}
								}
							}
						}
					}
				}

				if found {
					pr.Status = "matched"
					match := bookCandidate{
						OLID:    olID,
						Title:   matchTitle,
						Authors: authors,
					}
					if match.Title == "" {
						match.Title = pr.Title
					}
					if len(match.Authors) == 0 && pr.Author != "" {
						match.Authors = []string{pr.Author}
					}
					if coverURL != "" {
						match.CoverURL = &coverURL
					}
					pr.Match = &match
				} else {
					pr.Status = "unmatched"
				}

				ch <- progressMsg{Index: i, Row: pr}
			}(i, row)
		}

		go func() {
			wg.Wait()
			close(ch)
		}()

		// Stream progress as results come in
		processed := 0
		for msg := range ch {
			results[msg.Index] = msg.Row
			processed++
			_ = enc.Encode(map[string]any{
				"type":    "progress",
				"current": processed,
				"total":   len(csvRows),
				"title":   msg.Row.Title,
			})
			flusher.Flush()
		}

		matched := 0
		for _, r := range results {
			if r.Status == "matched" {
				matched++
			}
		}

		// Build deduplicated shelf summary with counts
		shelfCounts := map[string]int{}
		for _, r := range results {
			for _, s := range r.CustomShelves {
				shelfCounts[s]++
			}
		}
		type shelfSummary struct {
			Name  string `json:"name"`
			Count int    `json:"count"`
		}
		shelves := make([]shelfSummary, 0, len(shelfCounts))
		for name, count := range shelfCounts {
			shelves = append(shelves, shelfSummary{Name: name, Count: count})
		}
		sort.Slice(shelves, func(i, j int) bool {
			return shelves[i].Count > shelves[j].Count
		})

		// Final result line
		_ = enc.Encode(map[string]any{
			"type":      "result",
			"total":     len(results),
			"matched":   matched,
			"ambiguous":  0,
			"unmatched": len(results) - matched,
			"rows":      results,
			"shelves":   shelves,
		})
		flusher.Flush()

		return nil
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
			Rows []struct {
				RowID              int      `json:"row_id"`
				OLID               string   `json:"ol_id"`
				Title              string   `json:"title"`
				CoverURL           string   `json:"cover_url"`
				Authors            string   `json:"authors"`
				PublicationYear    int      `json:"publication_year"`
				ISBN13             string   `json:"isbn13"`
				Rating             float64  `json:"rating"`
				ReviewText         string   `json:"review_text"`
				Spoiler            bool     `json:"spoiler"`
				DateRead           string   `json:"date_read"`
				DateAdded          string   `json:"date_added"`
				ExclusiveShelfSlug string   `json:"exclusive_shelf_slug"`
				CustomShelves      []string `json:"custom_shelves"`
			} `json:"rows"`
			ShelfMappings []struct {
				Shelf      string `json:"shelf"`
				Action     string `json:"action"` // "tag", "skip", "create_label", "existing_label"
				LabelName  string `json:"label_name"`
				LabelKeyID string `json:"label_key_id"`
			} `json:"shelf_mappings"`
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}

		// Build tag map from shelf mappings before processing books
		type tagMapping struct {
			KeyID   string
			ValueID string
		}
		tagMap := map[string]tagMapping{}
		for _, sm := range data.ShelfMappings {
			switch sm.Action {
			case "tag":
				key, val, err := ensureTagKey(app, user.Id, sm.Shelf, "select_multiple")
				if err != nil {
					continue
				}
				tagMap[sm.Shelf] = tagMapping{KeyID: key.Id, ValueID: val.Id}

			case "create_label":
				if sm.LabelName == "" {
					continue
				}
				key, _, err := ensureTagKey(app, user.Id, sm.LabelName, "select_multiple")
				if err != nil {
					continue
				}
				val, err := ensureTagValue(app, key.Id, sm.Shelf)
				if err != nil {
					continue
				}
				tagMap[sm.Shelf] = tagMapping{KeyID: key.Id, ValueID: val.Id}

			case "existing_label":
				if sm.LabelKeyID == "" {
					continue
				}
				key, err := app.FindRecordById("tag_keys", sm.LabelKeyID)
				if err != nil || key.GetString("user") != user.Id {
					continue
				}
				val, err := ensureTagValue(app, key.Id, sm.Shelf)
				if err != nil {
					continue
				}
				tagMap[sm.Shelf] = tagMapping{KeyID: key.Id, ValueID: val.Id}
			}
		}

		imported := 0
		failed := 0
		var errors []string

		for _, b := range data.Rows {
			if b.OLID == "" {
				failed++
				errors = append(errors, fmt.Sprintf("No match for row %d: %s", b.RowID, b.Title))
				continue
			}

			// Upsert book
			book, err := upsertBook(app, b.OLID, b.Title, b.CoverURL, b.ISBN13, b.Authors, b.PublicationYear)
			if err != nil {
				failed++
				errors = append(errors, fmt.Sprintf("Failed to save book: %s — %v", b.Title, err))
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
					failed++
					errors = append(errors, fmt.Sprintf("Failed to create user_book: %s — %v", b.Title, err))
					continue
				}
				ub = core.NewRecord(coll)
				ub.Set("user", user.Id)
				ub.Set("book", book.Id)
			}

			if b.Rating > 0 {
				ub.Set("rating", b.Rating)
			}
			if b.ReviewText != "" {
				ub.Set("review_text", b.ReviewText)
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
				failed++
				errors = append(errors, fmt.Sprintf("Failed to save user_book: %s — %v", b.Title, err))
				continue
			}

			// Map exclusive shelf to status tag
			statusSlug := mapGoodreadsShelf(b.ExclusiveShelfSlug)
			if statusSlug != "" {
				setStatusTag(app, user.Id, book.Id, statusSlug)
			}

			// Assign custom shelf tags
			for _, shelf := range b.CustomShelves {
				tm, ok := tagMap[shelf]
				if !ok {
					continue
				}
				dup, _ := app.FindRecordsByFilter("book_tag_values",
					"user = {:user} && book = {:book} && tag_key = {:key} && tag_value = {:val}",
					"", 1, 0,
					map[string]any{"user": user.Id, "book": book.Id, "key": tm.KeyID, "val": tm.ValueID},
				)
				if len(dup) > 0 {
					continue
				}
				btvColl, err := app.FindCollectionByNameOrId("book_tag_values")
				if err != nil {
					continue
				}
				btv := core.NewRecord(btvColl)
				btv.Set("user", user.Id)
				btv.Set("book", book.Id)
				btv.Set("tag_key", tm.KeyID)
				btv.Set("tag_value", tm.ValueID)
				_ = app.Save(btv)
			}

			imported++
			refreshBookStats(app, book.Id)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"imported": imported,
			"failed":   failed,
			"errors":   errors,
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
