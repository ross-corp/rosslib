package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
		gb := newGBClient()

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

				// 1. Check local DB by ISBN (skip external API if we already have it)
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

				// 3. OL search by ISBN (searches across all editions via search index)
				if isbn != "" && !found {
					data, err := ol.get(fmt.Sprintf("/search.json?isbn=%s&fields=key,title,author_name,first_publish_year,cover_i&limit=1", url.QueryEscape(isbn)))
					if err == nil {
						olID, matchTitle, coverURL, authors, found = extractOLSearchResult(data)
					}
				}

				// 4. OL search by cleaned title+author
				// Goodreads titles often include junk: series info "(Series, #N)",
				// subtitles "Title: Subtitle", format tags "[Hardcover]",
				// embedded author "by Author Name (2005)", pipe separators, etc.
				if !found && pr.Title != "" {
					cleaned := cleanGoodreadsTitle(pr.Title)
					cleanedAuthor := cleanGoodreadsAuthor(pr.Author)
					q := "title=" + url.QueryEscape(cleaned)
					if cleanedAuthor != "" {
						q += "&author=" + url.QueryEscape(cleanedAuthor)
					}
					data, err := ol.get(fmt.Sprintf("/search.json?%s&fields=key,title,author_name,first_publish_year,cover_i&limit=1", q))
					if err == nil {
						olID, matchTitle, coverURL, authors, found = extractOLSearchResult(data)
					}
				}

				// 5. OL search by cleaned title only (author may be wrong/missing)
				// Also strips author name from front of title
				// ("Arthur C. Clark Expedition to Earth" → "Expedition to Earth")
				if !found && pr.Title != "" {
					cleaned := cleanGoodreadsTitle(pr.Title)
					cleaned = stripAuthorPrefix(cleaned, pr.Author)
					data, err := ol.get(fmt.Sprintf("/search.json?title=%s&fields=key,title,author_name,first_publish_year,cover_i&limit=1", url.QueryEscape(cleaned)))
					if err == nil {
						id, mt, cu, au, ok := extractOLSearchResult(data)
						if ok && titleMatchesResult(cleaned, mt) {
							olID, matchTitle, coverURL, authors, found = id, mt, cu, au, ok
						}
					}
				}

				// 6. Last resort: strip comma-subtitles and retry
				if !found && pr.Title != "" {
					cleaned := cleanGoodreadsTitle(pr.Title)
					if idx := strings.Index(cleaned, ","); idx > 0 {
						shortened := strings.TrimSpace(cleaned[:idx])
						if len(shortened) > 3 {
							cleanedAuthor := cleanGoodreadsAuthor(pr.Author)
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

				// 7. Google Books fallback — search GB, then map result back to OL
				if !found {
					olID, matchTitle, coverURL, authors, found = googleBooksLookup(gb, ol, isbn, cleanGoodreadsTitle(pr.Title), cleanGoodreadsAuthor(pr.Author))
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
			UnmatchedRows []struct {
				Title              string   `json:"title"`
				Author             string   `json:"author"`
				ISBN13             string   `json:"isbn13"`
				Rating             float64  `json:"rating"`
				ReviewText         string   `json:"review_text"`
				DateRead           string   `json:"date_read"`
				DateAdded          string   `json:"date_added"`
				ExclusiveShelfSlug string   `json:"exclusive_shelf_slug"`
				CustomShelves      []string `json:"custom_shelves"`
			} `json:"unmatched_rows"`
			ShelfMappings []struct {
				Shelf      string `json:"shelf"`
				Action     string `json:"action"` // "tag", "skip", "create_label", "existing_label", "map_dnf"
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

		// Collect shelves mapped to DNF status
		dnfShelves := map[string]bool{}
		for _, sm := range data.ShelfMappings {
			if sm.Action == "map_dnf" {
				dnfShelves[sm.Shelf] = true
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

			// Override status to DNF if any custom shelf is mapped to DNF
			for _, shelf := range b.CustomShelves {
				if dnfShelves[shelf] {
					setStatusTag(app, user.Id, book.Id, "dnf")
					break
				}
			}

			imported++
			refreshBookStats(app, book.Id)
		}

		// Persist unmatched rows to pending_imports
		pendingSaved := 0
		if len(data.UnmatchedRows) > 0 {
			piColl, piErr := app.FindCollectionByNameOrId("pending_imports")
			if piErr == nil {
				for _, u := range data.UnmatchedRows {
					if u.Title == "" {
						continue
					}
					pi := core.NewRecord(piColl)
					pi.Set("user", user.Id)
					pi.Set("source", "goodreads")
					pi.Set("title", u.Title)
					pi.Set("author", u.Author)
					pi.Set("isbn13", u.ISBN13)
					pi.Set("exclusive_shelf", u.ExclusiveShelfSlug)
					if u.CustomShelves != nil {
						pi.Set("custom_shelves", u.CustomShelves)
					} else {
						pi.Set("custom_shelves", []string{})
					}
					if u.Rating > 0 {
						pi.Set("rating", u.Rating)
					}
					pi.Set("review_text", u.ReviewText)
					pi.Set("date_read", u.DateRead)
					pi.Set("date_added", u.DateAdded)
					pi.Set("status", "unmatched")
					if err := app.Save(pi); err == nil {
						pendingSaved++
					}
				}
			}
		}

		return e.JSON(http.StatusOK, map[string]any{
			"imported":      imported,
			"failed":        failed,
			"errors":        errors,
			"pending_saved": pendingSaved,
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

// cleanGoodreadsTitle strips junk that Goodreads CSV titles commonly include.
//
// Examples:
//
//	"Children of Ruin (Children of Time, #2)" → "Children of Ruin"
//	"Skunk Works: A Personal Memoir of My Years at Lockheed" → "Skunk Works"
//	"Red Harvest by Dashiell Hammett (July 17 1989)" → "Red Harvest"
//	"All Men Are Brothers by Gandhi,Mohandas K.. [2005] Paperback" → "All Men Are Brothers"
//	"Man Plus (S.F. MASTERWORKS) by Pohl, Frederik New Edition (2000)" → "Man Plus"
//	"Survival in Auschwitz[SURVIVAL IN AUSCHWITZ][Paperback]" → "Survival in Auschwitz"
//	"The Wealth of Nations/Books I-III" → "The Wealth of Nations"
//	"Destinies, Feb 1980 | The Science Fiction Magazine" → "Destinies"
//	"[(The Soul of a Butterfly)] [by: Muhammad Ali]" → "The Soul of a Butterfly"
func cleanGoodreadsTitle(title string) string {
	// Strip [ and ] characters but keep inner content. This handles both
	// metadata tags like [Hardcover] and decorative brackets like [(Title)].
	// The inner content (format words, repeated titles, etc.) is cleaned
	// by the later steps.
	title = strings.NewReplacer("[", "", "]", "").Replace(title)
	title = strings.TrimSpace(title)

	// Strip everything from " by " or " by:" onward (embedded author/edition info).
	// Only strip if " by" appears after at least two words so we don't break
	// titles like "Stand by Me".
	lower := strings.ToLower(title)
	for _, sep := range []string{" by ", " by:"} {
		if idx := strings.Index(lower, sep); idx > 0 {
			before := strings.TrimSpace(title[:idx])
			if strings.Contains(before, " ") {
				title = before
				break
			}
		}
	}

	// Strip leading "By Author Name " prefix when title starts with "By "
	// and the real title follows. Detect by checking if stripping "By <words>"
	// still leaves a multi-word remainder.
	if strings.HasPrefix(title, "By ") {
		rest := title[3:]
		for _, sep := range []string{" The ", " A ", " An "} {
			if idx := strings.Index(rest, sep); idx > 0 {
				candidate := strings.TrimSpace(rest[idx:])
				if len(candidate) > 3 {
					title = candidate
					break
				}
			}
		}
	}

	// Strip trailing parenthetical content: "(Series Name, #N)", "(2000)", etc.
	// Repeat to handle nested cases like "Title (A) by Author (2000)"
	for {
		idx := strings.LastIndex(title, "(")
		if idx <= 0 {
			break
		}
		title = strings.TrimSpace(title[:idx])
	}

	// Strip wrapping parens left over from decorative brackets: "(Title)" → "Title"
	if strings.HasPrefix(title, "(") && strings.HasSuffix(title, ")") {
		title = title[1 : len(title)-1]
	}

	// Strip " | " pipe-separated subtitles
	if idx := strings.Index(title, " | "); idx > 0 {
		title = strings.TrimSpace(title[:idx])
	}
	if idx := strings.Index(title, "|"); idx > 0 {
		title = strings.TrimSpace(title[:idx])
	}

	// Strip "/subtitle" (but not mid-word slashes)
	if idx := strings.Index(title, "/"); idx > 0 {
		title = strings.TrimSpace(title[:idx])
	}

	// Strip subtitle after colon
	if idx := strings.Index(title, ":"); idx > 0 {
		title = strings.TrimSpace(title[:idx])
	}

	// Strip trailing junk words: "Paperback", "Hardcover", "Mass Market", edition years
	for {
		trimmed := strings.TrimRight(title, " .,")
		changed := false
		for _, suffix := range []string{"Paperback", "Hardcover", "Mass Market"} {
			if strings.HasSuffix(trimmed, suffix) {
				trimmed = strings.TrimSpace(trimmed[:len(trimmed)-len(suffix)])
				changed = true
			}
		}
		title = trimmed
		if !changed {
			break
		}
	}

	return strings.TrimSpace(title)
}

// cleanGoodreadsAuthor returns a cleaned author name suitable for OL search,
// or empty string if the author should be skipped (unknown, mangled, etc.).
func cleanGoodreadsAuthor(author string) string {
	author = strings.TrimSpace(author)
	if author == "" {
		return ""
	}

	lower := strings.ToLower(author)

	// Skip placeholder/unknown authors
	if lower == "unknown author" || lower == "unknown" || lower == "various" || lower == "anonymous" {
		return ""
	}

	// Skip mangled multi-author with semicolons ("Leo ; Bradbury Margulies")
	if strings.Contains(author, ";") {
		return ""
	}

	// Skip concatenated names with no space ("PrimoLevi") — real names have spaces
	if !strings.Contains(author, " ") && len(author) > 1 {
		return ""
	}

	// Strip single-letter middle initials: "Balaji S. Srinivasan" → "Balaji Srinivasan"
	// OL often doesn't index middle initials.
	parts := strings.Fields(author)
	var cleaned []string
	for _, p := range parts {
		trimmed := strings.TrimRight(p, ".")
		if len(trimmed) == 1 && trimmed == strings.ToUpper(trimmed) {
			continue // skip single-letter initials
		}
		cleaned = append(cleaned, p)
	}
	if len(cleaned) > 0 {
		author = strings.Join(cleaned, " ")
	}

	return author
}

// stripAuthorPrefix removes author name from the beginning of a title.
// Goodreads sometimes produces titles like "Arthur C. Clark Expedition to Earth"
// where the author name is prepended to the actual title.
func stripAuthorPrefix(title, author string) string {
	if author == "" || len(title) <= len(author) {
		return title
	}

	// Normalize for comparison: lowercase, strip periods/commas
	norm := func(s string) string {
		s = strings.ToLower(s)
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", "")
		return s
	}

	normTitle := norm(title)
	normAuthor := norm(author)

	if strings.HasPrefix(normTitle, normAuthor) {
		rest := strings.TrimSpace(title[len(author):])
		if len(rest) > 3 {
			return rest
		}
	}

	// Try matching just the first and last name words against the title prefix.
	// Handles misspellings like "Arthur C. Clark" (title) vs "Arthur C. Clarke" (author)
	// by comparing word-by-word and stopping when the author name diverges.
	titleWords := strings.Fields(norm(title))
	authorWords := strings.Fields(normAuthor)
	if len(authorWords) >= 2 && len(titleWords) > len(authorWords) {
		firstName := authorWords[0]
		lastName := authorWords[len(authorWords)-1]

		// Check if title starts with first name
		if titleWords[0] == firstName {
			// Scan forward to find where last name approximately appears
			for i := 1; i < len(titleWords) && i <= len(authorWords)+1; i++ {
				w := titleWords[i]
				// Match last name with some tolerance (prefix match for misspellings)
				if len(w) >= 3 && len(lastName) >= 3 &&
					(strings.HasPrefix(w, lastName[:3]) || strings.HasPrefix(lastName, w[:3])) {
					// Everything after this word is the real title
					// Reconstruct from original title by counting words
					origWords := strings.Fields(title)
					if i+1 < len(origWords) {
						rest := strings.Join(origWords[i+1:], " ")
						if len(rest) > 3 {
							return rest
						}
					}
					break
				}
			}
		}
	}

	return title
}

// titleMatchesResult checks if an OL search result title is a reasonable match
// for the search query. Used to reject false positives from aggressive search steps.
func titleMatchesResult(searchTitle, resultTitle string) bool {
	s := strings.ToLower(searchTitle)
	r := strings.ToLower(resultTitle)

	// Direct containment either way
	if strings.Contains(r, s) || strings.Contains(s, r) {
		return true
	}

	// Check if most significant words overlap
	skip := map[string]bool{"the": true, "a": true, "an": true, "of": true, "and": true, "in": true, "to": true, "for": true}
	searchWords := map[string]bool{}
	for _, w := range strings.Fields(s) {
		if !skip[w] && len(w) > 2 {
			searchWords[w] = true
		}
	}
	if len(searchWords) == 0 {
		return false
	}
	matches := 0
	for _, w := range strings.Fields(r) {
		if searchWords[strings.ToLower(w)] {
			matches++
		}
	}
	return matches > 0 && float64(matches)/float64(len(searchWords)) >= 0.5
}

// extractOLSearchResult pulls work ID, title, cover, and authors from an
// Open Library /search.json response.
func extractOLSearchResult(data map[string]any) (olID, matchTitle, coverURL string, authors []string, found bool) {
	docs, ok := data["docs"].([]any)
	if !ok || len(docs) == 0 {
		return
	}
	doc, ok := docs[0].(map[string]any)
	if !ok {
		return
	}
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
	return
}

// PreviewStoryGraphImport handles POST /me/import/storygraph/preview
// Streams NDJSON: progress lines followed by a final result line.
func PreviewStoryGraphImport(app core.App) func(e *core.RequestEvent) error {
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
		gb := newGBClient()

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
					Author: getCol(row, "Authors"),
				}

				// StoryGraph uses "ISBN/UID" column
				isbn := strings.TrimSpace(getCol(row, "ISBN/UID"))

				// Map StoryGraph read status
				pr.ExclusiveShelfSlug = strings.ToLower(getCol(row, "Read Status"))

				if r, err := strconv.ParseFloat(getCol(row, "Star Rating"), 64); err == nil && r > 0 {
					pr.Rating = &r
				}
				if review := getCol(row, "Review"); review != "" {
					pr.ReviewText = &review
				}

				// Parse Read Dates — may be a range "2024/01/15-2024/02/20" or single date
				if dates := getCol(row, "Read Dates"); dates != "" {
					parts := strings.SplitN(dates, "-", 2)
					// Take the last date as date_read (finish date)
					lastDate := strings.TrimSpace(parts[len(parts)-1])
					if lastDate != "" {
						// Normalize slashes to dashes for consistency
						lastDate = strings.ReplaceAll(lastDate, "/", "-")
						pr.DateRead = &lastDate
					}
				}

				// Parse Tags column — comma-separated tags become custom shelves
				if tags := getCol(row, "Tags"); tags != "" {
					for _, t := range strings.Split(tags, ",") {
						t = strings.TrimSpace(t)
						if t != "" {
							pr.CustomShelves = append(pr.CustomShelves, t)
						}
					}
				}
				if pr.CustomShelves == nil {
					pr.CustomShelves = []string{}
				}

				pr.ISBN13 = isbn

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

				// 4. OL search by title+author
				if !found && pr.Title != "" {
					cleaned := cleanStoryGraphTitle(pr.Title)
					cleanedAuthor := cleanStoryGraphAuthor(pr.Author)
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
				if !found && pr.Title != "" {
					cleaned := cleanStoryGraphTitle(pr.Title)
					data, err := ol.get(fmt.Sprintf("/search.json?title=%s&fields=key,title,author_name,first_publish_year,cover_i&limit=1", url.QueryEscape(cleaned)))
					if err == nil {
						id, mt, cu, au, ok := extractOLSearchResult(data)
						if ok && titleMatchesResult(cleaned, mt) {
							olID, matchTitle, coverURL, authors, found = id, mt, cu, au, ok
						}
					}
				}

				// 6. Google Books fallback
				if !found {
					olID, matchTitle, coverURL, authors, found = googleBooksLookup(gb, ol, isbn, cleanStoryGraphTitle(pr.Title), cleanStoryGraphAuthor(pr.Author))
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

		// Build deduplicated shelf summary (tags) with counts
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

// CommitStoryGraphImport handles POST /me/import/storygraph/commit
func CommitStoryGraphImport(app core.App) func(e *core.RequestEvent) error {
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
			UnmatchedRows []struct {
				Title              string   `json:"title"`
				Author             string   `json:"author"`
				ISBN13             string   `json:"isbn13"`
				Rating             float64  `json:"rating"`
				ReviewText         string   `json:"review_text"`
				DateRead           string   `json:"date_read"`
				DateAdded          string   `json:"date_added"`
				ExclusiveShelfSlug string   `json:"exclusive_shelf_slug"`
				CustomShelves      []string `json:"custom_shelves"`
			} `json:"unmatched_rows"`
			ShelfMappings []struct {
				Shelf      string `json:"shelf"`
				Action     string `json:"action"`
				LabelName  string `json:"label_name"`
				LabelKeyID string `json:"label_key_id"`
			} `json:"shelf_mappings"`
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}

		// Build tag map from shelf mappings
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

		// Collect shelves mapped to DNF status
		dnfShelves := map[string]bool{}
		for _, sm := range data.ShelfMappings {
			if sm.Action == "map_dnf" {
				dnfShelves[sm.Shelf] = true
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

			book, err := upsertBook(app, b.OLID, b.Title, b.CoverURL, b.ISBN13, b.Authors, b.PublicationYear)
			if err != nil {
				failed++
				errors = append(errors, fmt.Sprintf("Failed to save book: %s — %v", b.Title, err))
				continue
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

			// Map StoryGraph status to status tag
			statusSlug := mapStoryGraphStatus(b.ExclusiveShelfSlug)
			if statusSlug != "" {
				setStatusTag(app, user.Id, book.Id, statusSlug)
			}

			// Assign custom shelf tags (StoryGraph tags)
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

			// Override status to DNF if any custom shelf is mapped to DNF
			for _, shelf := range b.CustomShelves {
				if dnfShelves[shelf] {
					setStatusTag(app, user.Id, book.Id, "dnf")
					break
				}
			}

			imported++
			refreshBookStats(app, book.Id)
		}

		// Persist unmatched rows to pending_imports
		pendingSaved := 0
		if len(data.UnmatchedRows) > 0 {
			piColl, piErr := app.FindCollectionByNameOrId("pending_imports")
			if piErr == nil {
				for _, u := range data.UnmatchedRows {
					if u.Title == "" {
						continue
					}
					pi := core.NewRecord(piColl)
					pi.Set("user", user.Id)
					pi.Set("source", "storygraph")
					pi.Set("title", u.Title)
					pi.Set("author", u.Author)
					pi.Set("isbn13", u.ISBN13)
					pi.Set("exclusive_shelf", u.ExclusiveShelfSlug)
					if u.CustomShelves != nil {
						pi.Set("custom_shelves", u.CustomShelves)
					} else {
						pi.Set("custom_shelves", []string{})
					}
					if u.Rating > 0 {
						pi.Set("rating", u.Rating)
					}
					pi.Set("review_text", u.ReviewText)
					pi.Set("date_read", u.DateRead)
					pi.Set("date_added", u.DateAdded)
					pi.Set("status", "unmatched")
					if err := app.Save(pi); err == nil {
						pendingSaved++
					}
				}
			}
		}

		return e.JSON(http.StatusOK, map[string]any{
			"imported":      imported,
			"failed":        failed,
			"errors":        errors,
			"pending_saved": pendingSaved,
		})
	}
}

func mapStoryGraphStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "to-read":
		return "want-to-read"
	case "currently-reading":
		return "currently-reading"
	case "read":
		return "finished"
	case "did-not-finish":
		return "dnf"
	default:
		return ""
	}
}

// cleanStoryGraphTitle strips subtitle/series info from StoryGraph titles.
// StoryGraph titles are generally cleaner than Goodreads but may include subtitles.
func cleanStoryGraphTitle(title string) string {
	title = strings.TrimSpace(title)

	// Strip trailing parenthetical (series info)
	for {
		idx := strings.LastIndex(title, "(")
		if idx <= 0 {
			break
		}
		title = strings.TrimSpace(title[:idx])
	}

	// Strip subtitle after colon
	if idx := strings.Index(title, ":"); idx > 0 {
		title = strings.TrimSpace(title[:idx])
	}

	return strings.TrimSpace(title)
}

// cleanStoryGraphAuthor returns a cleaned author name for OL search.
// StoryGraph may list multiple authors comma-separated; use the first.
func cleanStoryGraphAuthor(author string) string {
	author = strings.TrimSpace(author)
	if author == "" {
		return ""
	}

	// Take first author if comma-separated
	if idx := strings.Index(author, ","); idx > 0 {
		author = strings.TrimSpace(author[:idx])
	}

	return author
}
