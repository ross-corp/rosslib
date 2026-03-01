package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

// SearchBooks handles GET /books/search?q=...&page=1
func SearchBooks(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		q := e.Request.URL.Query().Get("q")
		if q == "" {
			return e.JSON(http.StatusOK, map[string]any{"total": 0, "page": 1, "results": []any{}})
		}

		const perPage = 20
		page := 1
		if p := e.Request.URL.Query().Get("page"); p != "" {
			if n, err := strconv.Atoi(p); err == nil && n > 0 {
				page = n
			}
		}
		offset := (page - 1) * perPage

		// Search local books first
		localBooks, _ := app.FindRecordsByFilter("books",
			"title LIKE {:q} || authors LIKE {:q}",
			"-created", perPage, offset,
			map[string]any{"q": "%" + q + "%"},
		)

		var results []map[string]any
		seenOLIDs := map[string]bool{}

		// Batch-fetch book_stats for all local books
		statsMap := map[string]*core.Record{} // bookId -> stats record
		if len(localBooks) > 0 {
			var bookIDs []any
			for _, b := range localBooks {
				bookIDs = append(bookIDs, b.Id)
			}
			allStats, _ := app.FindRecordsByFilter("book_stats",
				"book IN {:ids}", "", len(localBooks), 0,
				map[string]any{"ids": bookIDs},
			)
			for _, s := range allStats {
				statsMap[s.GetString("book")] = s
			}
		}

		for _, b := range localBooks {
			olid := b.GetString("open_library_id")
			seenOLIDs[olid] = true

			var avgRating *float64
			var ratingCount, alreadyReadCount int
			if s, ok := statsMap[b.Id]; ok {
				if rc := s.GetInt("rating_count"); rc > 0 {
					avg := s.GetFloat("rating_sum") / float64(rc)
					avgRating = &avg
				}
				ratingCount = s.GetInt("rating_count")
				alreadyReadCount = s.GetInt("reads_count")
			}

			var subjects []string
			if s := b.GetString("subjects"); s != "" {
				for _, part := range strings.Split(s, ",") {
					part = strings.TrimSpace(part)
					if part != "" {
						subjects = append(subjects, part)
						if len(subjects) >= 3 {
							break
						}
					}
				}
			}

			results = append(results, map[string]any{
				"key":                olid,
				"title":              b.GetString("title"),
				"authors":           splitAuthors(b.GetString("authors")),
				"publish_year":       b.GetInt("publication_year"),
				"cover_url":         b.GetString("cover_url"),
				"edition_count":     0,
				"average_rating":    avgRating,
				"rating_count":      ratingCount,
				"already_read_count": alreadyReadCount,
				"subjects":          subjects,
			})
		}

		// Supplement with Open Library
		ol := newOLClient()
		olOffset := offset
		if len(localBooks) > 0 {
			// If local results filled this page, push OL offset further to avoid overlap
			olOffset = offset
		}
		olData, err := ol.get(fmt.Sprintf("/search.json?q=%s&limit=%d&offset=%d&fields=key,title,author_name,first_publish_year,isbn,cover_i,edition_count,subject,numFound", url.QueryEscape(q), perPage, olOffset))
		if err == nil {
			if docs, ok := olData["docs"].([]any); ok {
				for _, d := range docs {
					doc, ok := d.(map[string]any)
					if !ok {
						continue
					}
					key := ""
					if k, ok := doc["key"].(string); ok {
						// OL returns "/works/OL123W", strip prefix
						key = strings.TrimPrefix(k, "/works/")
					}
					if key == "" || seenOLIDs[key] {
						continue
					}

					var coverURL *string
					if coverI, ok := doc["cover_i"].(float64); ok && coverI > 0 {
						url := fmt.Sprintf("https://covers.openlibrary.org/b/id/%.0f-M.jpg", coverI)
						coverURL = &url
					}

					var authors []string
					if authorNames, ok := doc["author_name"].([]any); ok {
						for _, a := range authorNames {
							if s, ok := a.(string); ok {
								authors = append(authors, s)
							}
						}
					}

					var pubYear *float64
					if y, ok := doc["first_publish_year"].(float64); ok {
						pubYear = &y
					}

					// Extract subjects from OL search results (take first 3)
					var olSubjects []string
					if subjectList, ok := doc["subject"].([]any); ok {
						for _, s := range subjectList {
							if str, ok := s.(string); ok && str != "" {
								olSubjects = append(olSubjects, str)
								if len(olSubjects) >= 3 {
									break
								}
							}
						}
					}

					results = append(results, map[string]any{
						"key":                key,
						"title":              doc["title"],
						"authors":           authors,
						"publish_year":       pubYear,
						"cover_url":         coverURL,
						"edition_count":     doc["edition_count"],
						"average_rating":    nil,
						"rating_count":      0,
						"already_read_count": 0,
						"subjects":          olSubjects,
					})
				}
			}
		}

		if results == nil {
			results = []map[string]any{}
		}
		// Cap to perPage results
		if len(results) > perPage {
			results = results[:perPage]
		}

		// Estimate total from OL numFound (which is the most complete source)
		total := len(results)
		if olData != nil {
			if numFound, ok := olData["numFound"].(float64); ok && int(numFound) > total {
				total = int(numFound)
			}
		}

		return e.JSON(http.StatusOK, map[string]any{
			"total":   total,
			"page":    page,
			"results": results,
		})
	}
}

// LookupBook handles GET /books/lookup?isbn=...&ol_id=...
func LookupBook(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		isbn := e.Request.URL.Query().Get("isbn")
		olID := e.Request.URL.Query().Get("ol_id")

		if isbn == "" && olID == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "isbn or ol_id required"})
		}

		ol := newOLClient()

		if isbn != "" {
			data, err := ol.get(fmt.Sprintf("/isbn/%s.json", isbn))
			if err != nil {
				return e.JSON(http.StatusNotFound, map[string]any{"error": "Book not found"})
			}
			// Extract work ID from the ISBN response
			if works, ok := data["works"].([]any); ok && len(works) > 0 {
				if w, ok := works[0].(map[string]any); ok {
					if key, ok := w["key"].(string); ok {
						olID = strings.TrimPrefix(key, "/works/")
					}
				}
			}
		}

		if olID == "" {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Book not found"})
		}

		// Fetch work details from OL
		workData, err := ol.get(fmt.Sprintf("/works/%s.json", olID))
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Book not found"})
		}

		title, _ := workData["title"].(string)
		var coverURL string
		if covers, ok := workData["covers"].([]any); ok && len(covers) > 0 {
			if coverID, ok := covers[0].(float64); ok {
				coverURL = fmt.Sprintf("https://covers.openlibrary.org/b/id/%.0f-M.jpg", coverID)
			}
		}

		// Fetch authors
		var authorNames []string
		if authorList, ok := workData["authors"].([]any); ok {
			for _, a := range authorList {
				if aMap, ok := a.(map[string]any); ok {
					if authorRef, ok := aMap["author"].(map[string]any); ok {
						if key, ok := authorRef["key"].(string); ok {
							aData, err := ol.get(key + ".json")
							if err == nil {
								if name, ok := aData["name"].(string); ok {
									authorNames = append(authorNames, name)
								}
							}
						}
					}
				}
			}
		}

		authors := strings.Join(authorNames, ", ")

		// Upsert local book
		book, err := upsertBook(app, olID, title, coverURL, isbn, authors, 0)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to save book"})
		}

		return e.JSON(http.StatusOK, map[string]any{
			"key":       olID,
			"book_id":   book.Id,
			"title":     title,
			"authors":   authorNames,
			"cover_url": coverURL,
			"isbn":      isbn,
		})
	}
}

// ScanBook handles POST /books/scan (barcode scan)
func ScanBook(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		data := struct {
			ISBN string `json:"isbn"`
		}{}
		if err := e.BindBody(&data); err != nil || data.ISBN == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "isbn required"})
		}

		// Delegate to lookup
		e.Request.URL.RawQuery = "isbn=" + data.ISBN
		return LookupBook(app)(e)
	}
}

// seriesPositionRegex matches patterns like "Book 3", "#5", "Vol. 2", etc.
var seriesPositionRegex = regexp.MustCompile(`(?i)(?:book|#|no\.?|vol\.?|volume|part)\s*(\d+)`)

// populateSeriesFromOL checks Open Library edition data for series information
// and creates series + book_series records if none exist yet. This is best-effort;
// errors are logged but never surface to the caller.
func populateSeriesFromOL(app core.App, ol *cachedOLClient, bookRec *core.Record, workID string, subjects []string) {
	// Skip if book already has series links
	type countResult struct {
		Count int `db:"count"`
	}
	var existing countResult
	_ = app.DB().NewQuery(`SELECT COUNT(*) as count FROM book_series WHERE book = {:book}`).
		Bind(map[string]any{"book": bookRec.Id}).One(&existing)
	if existing.Count > 0 {
		return
	}

	type seriesCandidate struct {
		Name     string
		Position *int
	}
	var candidates []seriesCandidate
	seen := map[string]bool{}

	// 1. Check OL editions for series field
	editionsData, err := ol.get(fmt.Sprintf("/works/%s/editions.json?limit=50", workID))
	if err == nil {
		if entries, ok := editionsData["entries"].([]any); ok {
			for _, entry := range entries {
				ed, ok := entry.(map[string]any)
				if !ok {
					continue
				}
				if seriesList, ok := ed["series"].([]any); ok {
					for _, s := range seriesList {
						name, ok := s.(string)
						if !ok || name == "" {
							continue
						}
						name = strings.TrimSpace(name)
						// Try to extract position number from series string
						// e.g. "Harry Potter #3" or "Discworld Book 15"
						var pos *int
						if matches := seriesPositionRegex.FindStringSubmatch(name); len(matches) > 1 {
							if n, err := strconv.Atoi(matches[1]); err == nil {
								pos = &n
							}
							// Strip the position part to get the clean series name
							cleanName := strings.TrimSpace(seriesPositionRegex.ReplaceAllString(name, ""))
							// Remove trailing punctuation like commas, semicolons
							cleanName = strings.TrimRight(cleanName, " ,;:-")
							if cleanName != "" {
								name = cleanName
							}
						}
						nameKey := strings.ToLower(name)
						if seen[nameKey] {
							continue
						}
						seen[nameKey] = true
						candidates = append(candidates, seriesCandidate{Name: name, Position: pos})
					}
				}
			}
		}
	}

	// 2. Check subjects for series-like patterns (e.g. "Harry Potter" as a subject
	// that matches a known series naming pattern). Only use subjects if we found
	// nothing from editions, since subjects are less reliable.
	// Skip generic/common subjects that aren't series names.
	if len(candidates) == 0 && len(subjects) > 0 {
		seriesKeywords := []string{"series", "trilogy", "saga", "chronicles", "cycle", "quartet", "duology"}
		for _, subj := range subjects {
			lower := strings.ToLower(subj)
			for _, kw := range seriesKeywords {
				if strings.Contains(lower, kw) {
					// Clean up: remove the keyword suffix to get the series name
					name := strings.TrimSpace(subj)
					nameKey := strings.ToLower(name)
					if !seen[nameKey] {
						seen[nameKey] = true
						candidates = append(candidates, seriesCandidate{Name: name, Position: nil})
					}
					break
				}
			}
		}
	}

	if len(candidates) == 0 {
		log.Printf("[Series] No series data found for work %s", workID)
		return
	}

	log.Printf("[Series] Found %d series candidate(s) for work %s: %v", len(candidates), workID, func() []string {
		names := make([]string, len(candidates))
		for i, c := range candidates {
			if c.Position != nil {
				names[i] = fmt.Sprintf("%s (#%d)", c.Name, *c.Position)
			} else {
				names[i] = c.Name
			}
		}
		return names
	}())

	// Find-or-create series and book_series links
	for _, c := range candidates {
		// Find or create the series record
		existingSeries, _ := app.FindRecordsByFilter("series",
			"name = {:name}", "", 1, 0,
			map[string]any{"name": c.Name},
		)

		var seriesRec *core.Record
		if len(existingSeries) > 0 {
			seriesRec = existingSeries[0]
		} else {
			coll, err := app.FindCollectionByNameOrId("series")
			if err != nil {
				log.Printf("[Series] Failed to find series collection: %v", err)
				continue
			}
			seriesRec = core.NewRecord(coll)
			seriesRec.Set("name", c.Name)
			if err := app.Save(seriesRec); err != nil {
				log.Printf("[Series] Failed to create series %q: %v", c.Name, err)
				continue
			}
		}

		// Create book_series link
		bsColl, err := app.FindCollectionByNameOrId("book_series")
		if err != nil {
			log.Printf("[Series] Failed to find book_series collection: %v", err)
			continue
		}
		bs := core.NewRecord(bsColl)
		bs.Set("book", bookRec.Id)
		bs.Set("series", seriesRec.Id)
		if c.Position != nil {
			bs.Set("position", *c.Position)
		}
		if err := app.Save(bs); err != nil {
			log.Printf("[Series] Failed to link book to series %q: %v", c.Name, err)
			continue
		}
		log.Printf("[Series] Linked book %s to series %q (position: %v)", workID, c.Name, c.Position)
	}
}

// GetBookDetail handles GET /books/{workId}
func GetBookDetail(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		workID := e.Request.PathValue("workId")

		// Try local first
		localBooks, _ := app.FindRecordsByFilter("books",
			"open_library_id = {:id}", "", 1, 0,
			map[string]any{"id": workID},
		)

		// Fetch from OL for enriched data
		ol := newOLClient()
		workData, olErr := ol.get(fmt.Sprintf("/works/%s.json", workID))

		title := ""
		var description *string
		var coverURL *string
		var authors []map[string]any
		var firstPubYear *float64
		var pageCount *int
		var publisher *string
		var subjects []string

		if olErr == nil {
			if t, ok := workData["title"].(string); ok {
				title = t
			}
			if desc, ok := workData["description"].(string); ok {
				description = &desc
			} else if descMap, ok := workData["description"].(map[string]any); ok {
				if v, ok := descMap["value"].(string); ok {
					description = &v
				}
			}
			if covers, ok := workData["covers"].([]any); ok && len(covers) > 0 {
				if coverID, ok := covers[0].(float64); ok {
					url := fmt.Sprintf("https://covers.openlibrary.org/b/id/%.0f-L.jpg", coverID)
					coverURL = &url
				}
			}
			if authorList, ok := workData["authors"].([]any); ok {
				for _, a := range authorList {
					if aMap, ok := a.(map[string]any); ok {
						if authorRef, ok := aMap["author"].(map[string]any); ok {
							if key, ok := authorRef["key"].(string); ok {
								aData, err := ol.get(key + ".json")
								if err == nil {
									if name, ok := aData["name"].(string); ok {
										authorKey := strings.TrimPrefix(key, "/authors/")
										authors = append(authors, map[string]any{
											"name": name,
											"key":  authorKey,
										})
									}
								}
							}
						}
					}
				}
			}
			// Extract subjects (up to 10)
			if subjectList, ok := workData["subjects"].([]any); ok {
				for _, s := range subjectList {
					if str, ok := s.(string); ok && str != "" {
						subjects = append(subjects, str)
						if len(subjects) >= 10 {
							break
						}
					}
				}
			}
		}

		// Fallback to local data
		if title == "" && len(localBooks) > 0 {
			title = localBooks[0].GetString("title")
			if cv := localBooks[0].GetString("cover_url"); cv != "" {
				coverURL = &cv
			}
			if a := localBooks[0].GetString("authors"); a != "" {
				for _, name := range splitAuthors(a) {
					authors = append(authors, map[string]any{
						"name": name,
						"key":  nil,
					})
				}
			}
		}

		// Fetch edition count from OL
		var editionCount int
		editionsData, edErr := ol.get(fmt.Sprintf("/works/%s/editions.json?limit=0", workID))
		if edErr == nil {
			if size, ok := editionsData["size"].(float64); ok {
				editionCount = int(size)
			}
		}
		// Fallback subjects from local book record
		if len(subjects) == 0 && len(localBooks) > 0 {
			if s := localBooks[0].GetString("subjects"); s != "" {
				for _, part := range strings.Split(s, ",") {
					part = strings.TrimSpace(part)
					if part != "" {
						subjects = append(subjects, part)
						if len(subjects) >= 10 {
							break
						}
					}
				}
			}
		}

		// Populate page_count and publisher from local data (OL work data doesn't include these)
		if len(localBooks) > 0 {
			if pc := localBooks[0].GetInt("page_count"); pc > 0 {
				pageCount = &pc
			}
			if pub := localBooks[0].GetString("publisher"); pub != "" {
				publisher = &pub
			}
		}

		// Get local stats
		var avgRating *float64
		var ratingCount, readsCount, wtrCount int
		if len(localBooks) > 0 {
			stats, err := app.FindRecordsByFilter("book_stats",
				"book = {:book}", "", 1, 0,
				map[string]any{"book": localBooks[0].Id},
			)
			if err == nil && len(stats) > 0 {
				if rc := stats[0].GetInt("rating_count"); rc > 0 {
					avg := stats[0].GetFloat("rating_sum") / float64(rc)
					avgRating = &avg
				}
				ratingCount = stats[0].GetInt("rating_count")
				readsCount = stats[0].GetInt("reads_count")
				wtrCount = stats[0].GetInt("want_to_read_count")
			}
		}

		if subjects == nil {
			subjects = []string{}
		}

		// Auto-populate series data from Open Library if not already present
		if len(localBooks) > 0 {
			populateSeriesFromOL(app, ol, localBooks[0], workID, subjects)
		}

		// Get series memberships
		type seriesMembership struct {
			SeriesID string `db:"series_id" json:"series_id"`
			Name     string `db:"name" json:"name"`
			Position *int   `db:"position" json:"position"`
		}
		var seriesList []seriesMembership
		if len(localBooks) > 0 {
			_ = app.DB().NewQuery(`
				SELECT s.id as series_id, s.name, bs.position
				FROM book_series bs
				JOIN series s ON bs.series = s.id
				WHERE bs.book = {:book}
				ORDER BY s.name
			`).Bind(map[string]any{"book": localBooks[0].Id}).All(&seriesList)
		}

		var seriesOut any
		if len(seriesList) > 0 {
			seriesOut = seriesList
		}

		return e.JSON(http.StatusOK, map[string]any{
			"key":                     workID,
			"title":                   title,
			"authors":                 authors,
			"description":             description,
			"cover_url":               coverURL,
			"average_rating":          avgRating,
			"rating_count":            ratingCount,
			"local_reads_count":       readsCount,
			"local_want_to_read_count": wtrCount,
			"publisher":               publisher,
			"page_count":              pageCount,
			"first_publish_year":      firstPubYear,
			"edition_count":           editionCount,
			"subjects":                subjects,
			"series":                  seriesOut,
		})
	}
}

// GetBookEditions handles GET /books/{workId}/editions
func GetBookEditions(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		workID := e.Request.PathValue("workId")
		ol := newOLClient()
		data, err := ol.get(fmt.Sprintf("/works/%s/editions.json?limit=20", workID))
		if err != nil {
			return e.JSON(http.StatusOK, map[string]any{"entries": []any{}})
		}
		return e.JSON(http.StatusOK, data)
	}
}

// GetBookStats handles GET /books/{workId}/stats
func GetBookStats(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		workID := e.Request.PathValue("workId")

		books, _ := app.FindRecordsByFilter("books",
			"open_library_id = {:id}", "", 1, 0,
			map[string]any{"id": workID},
		)
		if len(books) == 0 {
			return e.JSON(http.StatusOK, map[string]any{
				"reads_count": 0, "want_to_read_count": 0,
				"rating_sum": 0, "rating_count": 0, "review_count": 0,
			})
		}

		stats, err := app.FindRecordsByFilter("book_stats",
			"book = {:book}", "", 1, 0,
			map[string]any{"book": books[0].Id},
		)
		if err != nil || len(stats) == 0 {
			return e.JSON(http.StatusOK, map[string]any{
				"reads_count": 0, "want_to_read_count": 0,
				"rating_sum": 0, "rating_count": 0, "review_count": 0,
			})
		}

		s := stats[0]
		return e.JSON(http.StatusOK, map[string]any{
			"reads_count":        s.GetInt("reads_count"),
			"want_to_read_count": s.GetInt("want_to_read_count"),
			"rating_sum":         s.GetFloat("rating_sum"),
			"rating_count":       s.GetInt("rating_count"),
			"review_count":       s.GetInt("review_count"),
		})
	}
}

// GetBookReviews handles GET /books/{workId}/reviews
func GetBookReviews(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		workID := e.Request.PathValue("workId")

		books, _ := app.FindRecordsByFilter("books",
			"open_library_id = {:id}", "", 1, 0,
			map[string]any{"id": workID},
		)
		if len(books) == 0 {
			return e.JSON(http.StatusOK, []any{})
		}

		viewerID := ""
		if e.Auth != nil {
			viewerID = e.Auth.Id
		}

		type reviewRow struct {
			UserBookID  string   `db:"user_book_id" json:"user_book_id"`
			UserID      string   `db:"user_id" json:"user_id"`
			Username    string   `db:"username" json:"username"`
			DisplayName *string  `db:"display_name" json:"display_name"`
			Avatar      *string  `db:"avatar" json:"avatar"`
			Rating      *float64 `db:"rating" json:"rating"`
			ReviewText  string   `db:"review_text" json:"review_text"`
			Spoiler     bool     `db:"spoiler" json:"spoiler"`
			DateRead    *string  `db:"date_read" json:"date_read"`
			DateAdded   string   `db:"date_added" json:"date_added"`
			LikeCount   int      `db:"like_count" json:"like_count"`
			LikedByMe   int      `db:"liked_by_me" json:"liked_by_me"`
		}

		var reviews []reviewRow
		query := `
			SELECT ub.id as user_book_id, ub.user as user_id, u.username, u.display_name, u.avatar,
				   ub.rating, ub.review_text, ub.spoiler, ub.date_read,
				   ub.date_added as date_added,
				   COALESCE((SELECT COUNT(*) FROM review_likes rl WHERE rl.book = ub.book AND rl.review_user = ub.user), 0) as like_count,
				   COALESCE((SELECT COUNT(*) FROM review_likes rl WHERE rl.book = ub.book AND rl.review_user = ub.user AND rl.user = {:viewer}), 0) as liked_by_me
			FROM user_books ub
			JOIN users u ON ub.user = u.id
			WHERE ub.book = {:book} AND ub.review_text != '' AND ub.review_text IS NOT NULL`
		params := map[string]any{"book": books[0].Id, "viewer": viewerID}

		if viewerID != "" {
			query += `
			AND ub.user NOT IN (SELECT blocked FROM blocks WHERE blocker = {:viewer})
			AND ub.user NOT IN (SELECT blocker FROM blocks WHERE blocked = {:viewer})`
		}
		query += `
			ORDER BY CASE WHEN ub.user = {:viewer} THEN 0 ELSE 1 END, ub.date_added DESC`

		err := app.DB().NewQuery(query).Bind(params).All(&reviews)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}

		// Batch-check which reviewers the viewer follows
		followedSet := map[string]bool{}
		if viewerID != "" {
			var reviewerIDs []string
			for _, r := range reviews {
				if r.UserID != viewerID {
					reviewerIDs = append(reviewerIDs, r.UserID)
				}
			}
			if len(reviewerIDs) > 0 {
				type followRow struct {
					Followee string `db:"followee"`
				}
				var followRows []followRow
				placeholders := ""
				bindParams := map[string]any{"viewer": viewerID}
				for i, id := range reviewerIDs {
					key := fmt.Sprintf("uid%d", i)
					if i > 0 {
						placeholders += ", "
					}
					placeholders += "{:" + key + "}"
					bindParams[key] = id
				}
				followQuery := fmt.Sprintf(`SELECT followee FROM follows WHERE follower = {:viewer} AND followee IN (%s) AND status = 'active'`, placeholders)
				_ = app.DB().NewQuery(followQuery).Bind(bindParams).All(&followRows)
				for _, f := range followRows {
					followedSet[f.Followee] = true
				}
			}
		}

		// Build response with avatar URLs
		var result []map[string]any
		for _, r := range reviews {
			var avatarURL *string
			if r.Avatar != nil && *r.Avatar != "" {
				url := fmt.Sprintf("/api/files/users/%s/%s", r.UserID, *r.Avatar)
				avatarURL = &url
			}

			result = append(result, map[string]any{
				"user_book_id": r.UserBookID,
				"user_id":      r.UserID,
				"username":     r.Username,
				"display_name": r.DisplayName,
				"avatar_url":   avatarURL,
				"rating":       r.Rating,
				"review_text":  r.ReviewText,
				"spoiler":      r.Spoiler,
				"date_read":    r.DateRead,
				"date_added":   r.DateAdded,
				"is_followed":  followedSet[r.UserID],
				"like_count":   r.LikeCount,
				"liked_by_me":  r.LikedByMe > 0,
			})
		}
		if result == nil {
			result = []map[string]any{}
		}

		return e.JSON(http.StatusOK, result)
	}
}

// GetPopularBooks handles GET /books/popular
func GetPopularBooks(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		type row struct {
			OLID      string  `db:"open_library_id" json:"key"`
			Title     string  `db:"title" json:"title"`
			Authors   string  `db:"authors" json:"authors_raw"`
			CoverURL  *string `db:"cover_url" json:"cover_url"`
			PubYear   *int    `db:"publication_year" json:"publish_year"`
			AvgRating *float64 `db:"avg_rating" json:"average_rating"`
			RatCount  int     `db:"rating_count" json:"rating_count"`
			Reads     int     `db:"reads_count" json:"already_read_count"`
		}
		var rows []row
		err := app.DB().NewQuery(`
			SELECT b.open_library_id, b.title, b.authors, b.cover_url, b.publication_year,
				CASE WHEN bs.rating_count > 0
					THEN ROUND(bs.rating_sum * 1.0 / bs.rating_count, 2)
					ELSE NULL END AS avg_rating,
				bs.rating_count, bs.reads_count
			FROM book_stats bs
			JOIN books b ON bs.book = b.id
			WHERE bs.reads_count > 0 OR bs.rating_count > 0
			ORDER BY bs.reads_count DESC, bs.rating_count DESC
			LIMIT 12
		`).All(&rows)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}

		var results []map[string]any
		for _, r := range rows {
			results = append(results, map[string]any{
				"key":                r.OLID,
				"title":              r.Title,
				"authors":            splitAuthors(r.Authors),
				"cover_url":          r.CoverURL,
				"publish_year":       r.PubYear,
				"average_rating":     r.AvgRating,
				"rating_count":       r.RatCount,
				"already_read_count": r.Reads,
			})
		}
		if results == nil {
			results = []map[string]any{}
		}
		return e.JSON(http.StatusOK, results)
	}
}

// SearchAuthors handles GET /authors/search?q=...
func SearchAuthors(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		q := e.Request.URL.Query().Get("q")
		if q == "" {
			return e.JSON(http.StatusOK, map[string]any{"total": 0, "results": []any{}})
		}

		ol := newOLClient()
		raw, err := ol.getRaw(fmt.Sprintf("/search/authors.json?q=%s&limit=20", q))
		if err != nil {
			return e.JSON(http.StatusOK, map[string]any{"total": 0, "results": []any{}})
		}

		var data map[string]any
		if err := json.Unmarshal(raw, &data); err != nil {
			return e.JSON(http.StatusOK, map[string]any{"total": 0, "results": []any{}})
		}

		total := 0
		if t, ok := data["numFound"].(float64); ok {
			total = int(t)
		}

		var results []map[string]any
		if docs, ok := data["docs"].([]any); ok {
			for _, d := range docs {
				doc, ok := d.(map[string]any)
				if !ok {
					continue
				}
				key, _ := doc["key"].(string)
				name, _ := doc["name"].(string)

				var photoURL *string
				// OL doesn't return photo in search, construct from key
				if key != "" {
					url := fmt.Sprintf("https://covers.openlibrary.org/a/olid/%s-M.jpg", key)
					photoURL = &url
				}

				results = append(results, map[string]any{
					"key":          key,
					"name":         name,
					"birth_date":   doc["birth_date"],
					"death_date":   doc["death_date"],
					"top_work":     doc["top_work"],
					"work_count":   doc["work_count"],
					"top_subjects": doc["top_subjects"],
					"photo_url":    photoURL,
				})
			}
		}
		if results == nil {
			results = []map[string]any{}
		}

		return e.JSON(http.StatusOK, map[string]any{
			"total":   total,
			"results": results,
		})
	}
}

// GetAuthorDetail handles GET /authors/{authorKey}?limit=24&offset=0
// Fetches author info and a paginated slice of their works from Open Library.
func GetAuthorDetail(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		authorKey := e.Request.PathValue("authorKey")
		ol := newOLClient()

		// Parse pagination params (defaults: limit=24, offset=0).
		limit := 24
		offset := 0
		if v := e.Request.URL.Query().Get("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
				limit = n
			}
		}
		if v := e.Request.URL.Query().Get("offset"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n >= 0 {
				offset = n
			}
		}

		// Fetch author metadata.
		authorData, err := ol.get(fmt.Sprintf("/authors/%s.json", authorKey))
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Author not found"})
		}

		// Extract author fields.
		name, _ := authorData["name"].(string)

		var bio *string
		switch b := authorData["bio"].(type) {
		case string:
			bio = &b
		case map[string]any:
			if v, ok := b["value"].(string); ok {
				bio = &v
			}
		}

		var birthDate, deathDate *string
		if v, ok := authorData["birth_date"].(string); ok {
			birthDate = &v
		}
		if v, ok := authorData["death_date"].(string); ok {
			deathDate = &v
		}

		var photoURL *string
		if photos, ok := authorData["photos"].([]any); ok && len(photos) > 0 {
			if id, ok := photos[0].(float64); ok && int(id) > 0 {
				u := fmt.Sprintf("https://covers.openlibrary.org/a/id/%d-L.jpg", int(id))
				photoURL = &u
			}
		}

		var links []map[string]any
		if rawLinks, ok := authorData["links"].([]any); ok {
			for _, rl := range rawLinks {
				if lm, ok := rl.(map[string]any); ok {
					title, _ := lm["title"].(string)
					url, _ := lm["url"].(string)
					if title != "" && url != "" {
						links = append(links, map[string]any{"title": title, "url": url})
					}
				}
			}
		}

		// Fetch works with pagination.
		worksData, _ := ol.get(fmt.Sprintf("/authors/%s/works.json?limit=%d&offset=%d", authorKey, limit, offset))

		workCount := 0
		var works []map[string]any
		if worksData != nil {
			if sz, ok := worksData["size"].(float64); ok {
				workCount = int(sz)
			}
			if entries, ok := worksData["entries"].([]any); ok {
				for _, entry := range entries {
					e, ok := entry.(map[string]any)
					if !ok {
						continue
					}
					key, _ := e["key"].(string)
					title, _ := e["title"].(string)

					// OL work keys are like "/works/OL12345W"
					key = strings.TrimPrefix(key, "/works/")

					var coverURL *string
					if covers, ok := e["covers"].([]any); ok {
						for _, c := range covers {
							if id, ok := c.(float64); ok && int(id) > 0 {
								u := fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-M.jpg", int(id))
								coverURL = &u
								break
							}
						}
					}

					works = append(works, map[string]any{
						"key":       key,
						"title":     title,
						"cover_url": coverURL,
					})
				}
			}
		}
		if works == nil {
			works = []map[string]any{}
		}

		return e.JSON(http.StatusOK, map[string]any{
			"key":        authorKey,
			"name":       name,
			"bio":        bio,
			"birth_date": birthDate,
			"death_date": deathDate,
			"photo_url":  photoURL,
			"links":      links,
			"work_count": workCount,
			"works":      works,
		})
	}
}

// splitAuthors splits a comma-separated authors string into a slice.
func splitAuthors(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ", ")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// CheckAuthorFollow handles GET /authors/{authorKey}/follow
func CheckAuthorFollow(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		authorKey := e.Request.PathValue("authorKey")

		existing, _ := app.FindRecordsByFilter("author_follows",
			"user = {:user} && author_key = {:key}",
			"", 1, 0,
			map[string]any{"user": user.Id, "key": authorKey},
		)
		return e.JSON(http.StatusOK, map[string]any{"following": len(existing) > 0})
	}
}

// FollowAuthor handles POST /authors/{authorKey}/follow
func FollowAuthor(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		authorKey := e.Request.PathValue("authorKey")

		data := struct {
			AuthorName string `json:"author_name"`
		}{}
		_ = e.BindBody(&data)

		existing, _ := app.FindRecordsByFilter("author_follows",
			"user = {:user} && author_key = {:key}",
			"", 1, 0,
			map[string]any{"user": user.Id, "key": authorKey},
		)
		if len(existing) > 0 {
			return e.JSON(http.StatusOK, map[string]any{"message": "Already following"})
		}

		coll, err := app.FindCollectionByNameOrId("author_follows")
		if err != nil {
			return err
		}
		rec := core.NewRecord(coll)
		rec.Set("user", user.Id)
		rec.Set("author_key", authorKey)
		rec.Set("author_name", data.AuthorName)
		if err := app.Save(rec); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}

		return e.JSON(http.StatusOK, map[string]any{"message": "Following author"})
	}
}

// UnfollowAuthor handles DELETE /authors/{authorKey}/follow
func UnfollowAuthor(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		authorKey := e.Request.PathValue("authorKey")

		existing, err := app.FindRecordsByFilter("author_follows",
			"user = {:user} && author_key = {:key}",
			"", 1, 0,
			map[string]any{"user": user.Id, "key": authorKey},
		)
		if err != nil || len(existing) == 0 {
			return e.JSON(http.StatusOK, map[string]any{"message": "Not following"})
		}
		_ = app.Delete(existing[0])
		return e.JSON(http.StatusOK, map[string]any{"message": "Unfollowed author"})
	}
}

// GetFollowedAuthors handles GET /me/followed-authors
func GetFollowedAuthors(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		records, err := app.FindRecordsByFilter("author_follows",
			"user = {:user}", "-created", 100, 0,
			map[string]any{"user": user.Id},
		)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}

		var results []map[string]any
		for _, r := range records {
			results = append(results, map[string]any{
				"author_key":  r.GetString("author_key"),
				"author_name": r.GetString("author_name"),
			})
		}
		if results == nil {
			results = []map[string]any{}
		}
		return e.JSON(http.StatusOK, results)
	}
}

// CheckBookFollow handles GET /books/{workId}/follow
func CheckBookFollow(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		workID := e.Request.PathValue("workId")

		books, _ := app.FindRecordsByFilter("books",
			"open_library_id = {:id}", "", 1, 0,
			map[string]any{"id": workID},
		)
		if len(books) == 0 {
			return e.JSON(http.StatusOK, map[string]any{"following": false})
		}

		existing, _ := app.FindRecordsByFilter("book_follows",
			"user = {:user} && book = {:book}",
			"", 1, 0,
			map[string]any{"user": user.Id, "book": books[0].Id},
		)
		return e.JSON(http.StatusOK, map[string]any{"following": len(existing) > 0})
	}
}

// FollowBook handles POST /books/{workId}/follow
func FollowBook(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		workID := e.Request.PathValue("workId")

		books, _ := app.FindRecordsByFilter("books",
			"open_library_id = {:id}", "", 1, 0,
			map[string]any{"id": workID},
		)
		if len(books) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Book not found"})
		}

		existing, _ := app.FindRecordsByFilter("book_follows",
			"user = {:user} && book = {:book}",
			"", 1, 0,
			map[string]any{"user": user.Id, "book": books[0].Id},
		)
		if len(existing) > 0 {
			return e.JSON(http.StatusOK, map[string]any{"message": "Already following"})
		}

		coll, err := app.FindCollectionByNameOrId("book_follows")
		if err != nil {
			return err
		}
		rec := core.NewRecord(coll)
		rec.Set("user", user.Id)
		rec.Set("book", books[0].Id)
		if err := app.Save(rec); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}

		return e.JSON(http.StatusOK, map[string]any{"message": "Following book"})
	}
}

// UnfollowBook handles DELETE /books/{workId}/follow
func UnfollowBook(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		workID := e.Request.PathValue("workId")

		books, _ := app.FindRecordsByFilter("books",
			"open_library_id = {:id}", "", 1, 0,
			map[string]any{"id": workID},
		)
		if len(books) == 0 {
			return e.JSON(http.StatusOK, map[string]any{"message": "Not following"})
		}

		existing, err := app.FindRecordsByFilter("book_follows",
			"user = {:user} && book = {:book}",
			"", 1, 0,
			map[string]any{"user": user.Id, "book": books[0].Id},
		)
		if err != nil || len(existing) == 0 {
			return e.JSON(http.StatusOK, map[string]any{"message": "Not following"})
		}
		_ = app.Delete(existing[0])
		return e.JSON(http.StatusOK, map[string]any{"message": "Unfollowed book"})
	}
}

// GetFollowedBooks handles GET /me/followed-books
func GetFollowedBooks(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		type bookRow struct {
			OLID     string  `db:"open_library_id" json:"open_library_id"`
			Title    string  `db:"title" json:"title"`
			Authors  string  `db:"authors" json:"-"`
			CoverURL *string `db:"cover_url" json:"cover_url"`
		}
		var books []bookRow
		err := app.DB().NewQuery(`
			SELECT b.open_library_id, b.title, b.authors, b.cover_url
			FROM book_follows bf
			JOIN books b ON bf.book = b.id
			WHERE bf.user = {:user}
			ORDER BY bf.created DESC
		`).Bind(map[string]any{"user": user.Id}).All(&books)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}
		result := make([]map[string]any, len(books))
		for i, b := range books {
			result[i] = map[string]any{
				"open_library_id": b.OLID,
				"title":           b.Title,
				"authors":         splitAuthors(b.Authors),
				"cover_url":       b.CoverURL,
			}
		}
		return e.JSON(http.StatusOK, result)
	}
}
