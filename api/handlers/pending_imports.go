package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

// GetPendingImports handles GET /me/imports/pending
func GetPendingImports(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		records, err := app.FindRecordsByFilter(
			"pending_imports",
			"user = {:user} && status = 'unmatched'",
			"-created",
			200, 0,
			map[string]any{"user": user.Id},
		)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to fetch pending imports"})
		}

		items := make([]map[string]any, 0, len(records))
		for _, r := range records {
			var customShelves []string
			raw := r.Get("custom_shelves")
			if raw != nil {
				if b, err := json.Marshal(raw); err == nil {
					_ = json.Unmarshal(b, &customShelves)
				}
			}
			if customShelves == nil {
				customShelves = []string{}
			}

			items = append(items, map[string]any{
				"id":              r.Id,
				"title":           r.GetString("title"),
				"author":          r.GetString("author"),
				"isbn13":          r.GetString("isbn13"),
				"exclusive_shelf": r.GetString("exclusive_shelf"),
				"custom_shelves":  customShelves,
				"rating":          r.Get("rating"),
				"review_text":     r.GetString("review_text"),
				"date_read":       r.GetString("date_read"),
				"date_added":      r.GetString("date_added"),
				"created":         r.GetString("created"),
			})
		}

		return e.JSON(http.StatusOK, items)
	}
}

// ResolvePendingImport handles PATCH /me/imports/pending/:id
// Resolves a pending import by matching it to an OL work ID, or dismisses it.
func ResolvePendingImport(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		id := e.Request.PathValue("id")
		record, err := app.FindRecordById("pending_imports", id)
		if err != nil || record.GetString("user") != user.Id {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Pending import not found"})
		}
		if record.GetString("status") != "unmatched" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Already resolved"})
		}

		data := struct {
			Action string `json:"action"` // "resolve" or "dismiss"
			OLID   string `json:"ol_id"`  // required if action=resolve
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}

		switch data.Action {
		case "dismiss":
			record.Set("status", "resolved")
			if err := app.Save(record); err != nil {
				return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to update"})
			}
			return e.JSON(http.StatusOK, map[string]any{"ok": true})

		case "resolve":
			if data.OLID == "" {
				return e.JSON(http.StatusBadRequest, map[string]any{"error": "ol_id is required for resolve action"})
			}

			// Look up the book from OL and upsert
			book, err := upsertBook(app, data.OLID, record.GetString("title"), "", record.GetString("isbn13"), record.GetString("author"), 0, "")
			if err != nil {
				return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to save book"})
			}

			// Create user_book
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
					return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to create user_book"})
				}
				ub = core.NewRecord(coll)
				ub.Set("user", user.Id)
				ub.Set("book", book.Id)
			}

			rating := record.Get("rating")
			if rating != nil && rating != 0.0 {
				ub.Set("rating", rating)
			}
			if review := record.GetString("review_text"); review != "" {
				ub.Set("review_text", review)
			}
			if dr := record.GetString("date_read"); dr != "" {
				ub.Set("date_read", dr)
			}
			if da := record.GetString("date_added"); da != "" {
				ub.Set("date_added", da)
			}

			if err := app.Save(ub); err != nil {
				return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to save user_book"})
			}

			// Map exclusive shelf to status tag
			statusSlug := mapGoodreadsShelf(record.GetString("exclusive_shelf"))
			if statusSlug != "" {
				setStatusTag(app, user.Id, book.Id, statusSlug)
			}

			// Mark resolved
			record.Set("status", "resolved")
			if err := app.Save(record); err != nil {
				return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to update"})
			}

			refreshBookStats(app, book.Id)

			return e.JSON(http.StatusOK, map[string]any{"ok": true, "book_id": book.Id})

		default:
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "action must be 'resolve' or 'dismiss'"})
		}
	}
}

// DeletePendingImport handles DELETE /me/imports/pending/:id
func DeletePendingImport(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		id := e.Request.PathValue("id")
		record, err := app.FindRecordById("pending_imports", id)
		if err != nil || record.GetString("user") != user.Id {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Pending import not found"})
		}

		if err := app.Delete(record); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to delete"})
		}

		return e.JSON(http.StatusOK, map[string]any{"ok": true})
	}
}

// RetryPendingImport handles POST /me/imports/pending/:id/retry
// Re-runs the lookup chain for a single pending import and returns the result.
// If a match is found, it auto-resolves (creates user_book) and removes the pending import.
func RetryPendingImport(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		id := e.Request.PathValue("id")
		record, err := app.FindRecordById("pending_imports", id)
		if err != nil || record.GetString("user") != user.Id {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Pending import not found"})
		}
		if record.GetString("status") != "unmatched" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Already resolved"})
		}

		title := record.GetString("title")
		author := record.GetString("author")
		isbn := record.GetString("isbn13")

		ol := newOLClient()
		gb := newGBClient()

		found := false
		var olID, matchTitle, coverURL string
		var authors []string

		// 1. Check local DB by ISBN
		if isbn != "" {
			existing, localErr := app.FindRecordsByFilter("books",
				"isbn13 = {:isbn}",
				"", 1, 0,
				map[string]any{"isbn": isbn},
			)
			if localErr == nil && len(existing) > 0 {
				rec := existing[0]
				olID = rec.GetString("open_library_id")
				matchTitle = rec.GetString("title")
				coverURL = rec.GetString("cover_url")
				if a := rec.GetString("authors"); a != "" {
					authors = strings.Split(a, ", ")
				}
				found = olID != ""
			}
		}

		// 2. OL direct ISBN endpoint
		if isbn != "" && !found {
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

		// 3. OL search by ISBN
		if isbn != "" && !found {
			data, err := ol.get(fmt.Sprintf("/search.json?isbn=%s&fields=key,title,author_name,first_publish_year,cover_i&limit=1", url.QueryEscape(isbn)))
			if err == nil {
				olID, matchTitle, coverURL, authors, found = extractOLSearchResult(data)
			}
		}

		// 4. OL search by cleaned title+author
		if !found && title != "" {
			cleaned := cleanGoodreadsTitle(title)
			cleanedAuthor := cleanGoodreadsAuthor(author)
			q := "title=" + url.QueryEscape(cleaned)
			if cleanedAuthor != "" {
				q += "&author=" + url.QueryEscape(cleanedAuthor)
			}
			data, err := ol.get(fmt.Sprintf("/search.json?%s&fields=key,title,author_name,first_publish_year,cover_i&limit=1", q))
			if err == nil {
				olID, matchTitle, coverURL, authors, found = extractOLSearchResult(data)
			}
		}

		// 5. OL search by title only
		if !found && title != "" {
			cleaned := cleanGoodreadsTitle(title)
			cleaned = stripAuthorPrefix(cleaned, author)
			data, err := ol.get(fmt.Sprintf("/search.json?title=%s&fields=key,title,author_name,first_publish_year,cover_i&limit=1", url.QueryEscape(cleaned)))
			if err == nil {
				id, mt, cu, au, ok := extractOLSearchResult(data)
				if ok && titleMatchesResult(cleaned, mt) {
					olID, matchTitle, coverURL, authors, found = id, mt, cu, au, ok
				}
			}
		}

		// 6. Strip comma-subtitles and retry
		if !found && title != "" {
			cleaned := cleanGoodreadsTitle(title)
			if idx := strings.Index(cleaned, ","); idx > 0 {
				shortened := strings.TrimSpace(cleaned[:idx])
				if len(shortened) > 3 {
					cleanedAuthor := cleanGoodreadsAuthor(author)
					q := "title=" + url.QueryEscape(shortened)
					if cleanedAuthor != "" {
						q += "&author=" + url.QueryEscape(cleanedAuthor)
					}
					data, err := ol.get(fmt.Sprintf("/search.json?%s&fields=key,title,author_name,first_publish_year,cover_i&limit=1", q))
					if err == nil {
						id, mt, cu, au, ok := extractOLSearchResult(data)
						if ok && titleMatchesResult(shortened, mt) {
							olID, matchTitle, coverURL, authors, found = id, mt, cu, au, ok
						}
					}
				}
			}
		}

		// 7. Google Books fallback
		if !found {
			olID, matchTitle, coverURL, authors, found = googleBooksLookup(gb, ol, isbn, cleanGoodreadsTitle(title), cleanGoodreadsAuthor(author))
		}

		// 8. LLM-powered fuzzy matching
		type candidate struct {
			OLID     string   `json:"ol_id"`
			Title    string   `json:"title"`
			Authors  []string `json:"authors"`
			CoverURL *string  `json:"cover_url"`
		}

		if !found && title != "" {
			if result := llmFuzzyMatch(ol, title, author); result != nil {
				var candidates []candidate
				for _, c := range result.Candidates {
					bc := candidate{
						OLID:    c.OLID,
						Title:   c.Title,
						Authors: c.Authors,
					}
					if c.CoverURL != "" {
						bc.CoverURL = &c.CoverURL
					}
					candidates = append(candidates, bc)
				}
				return e.JSON(http.StatusOK, map[string]any{
					"status":     "ambiguous",
					"candidates": candidates,
				})
			}
		}

		if !found {
			return e.JSON(http.StatusOK, map[string]any{
				"status": "unmatched",
			})
		}

		// Match found — auto-resolve: upsert book + user_book
		book, err := upsertBook(app, olID, matchTitle, coverURL, isbn, strings.Join(authors, ", "), 0, "")
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to save book"})
		}

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
				return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to create user_book"})
			}
			ub = core.NewRecord(coll)
			ub.Set("user", user.Id)
			ub.Set("book", book.Id)
		}

		rating := record.Get("rating")
		if rating != nil && rating != 0.0 {
			ub.Set("rating", rating)
		}
		if review := record.GetString("review_text"); review != "" {
			ub.Set("review_text", review)
		}
		if dr := record.GetString("date_read"); dr != "" {
			ub.Set("date_read", dr)
		}
		if da := record.GetString("date_added"); da != "" {
			ub.Set("date_added", da)
		}

		if err := app.Save(ub); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to save user_book"})
		}

		// Map exclusive shelf to status tag
		statusSlug := mapGoodreadsShelf(record.GetString("exclusive_shelf"))
		if statusSlug != "" {
			setStatusTag(app, user.Id, book.Id, statusSlug)
		}

		// Mark resolved
		record.Set("status", "resolved")
		_ = app.Save(record)

		refreshBookStats(app, book.Id)

		return e.JSON(http.StatusOK, map[string]any{
			"status":  "matched",
			"book_id": book.Id,
			"match": map[string]any{
				"ol_id":     olID,
				"title":     matchTitle,
				"authors":   authors,
				"cover_url": coverURL,
			},
		})
	}
}
