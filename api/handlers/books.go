package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

// SearchBooks handles GET /books/search?q=...
func SearchBooks(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		q := e.Request.URL.Query().Get("q")
		if q == "" {
			return e.JSON(http.StatusOK, map[string]any{"total": 0, "results": []any{}})
		}

		// Search local books first
		localBooks, _ := app.FindRecordsByFilter("books",
			"title LIKE {:q} || authors LIKE {:q}",
			"-created", 20, 0,
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
			})
		}

		// Supplement with Open Library
		ol := newOLClient()
		olData, err := ol.get(fmt.Sprintf("/search.json?q=%s&limit=20&fields=key,title,author_name,first_publish_year,isbn,cover_i,edition_count", q))
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
					})
				}
			}
		}

		if results == nil {
			results = []map[string]any{}
		}
		total := len(results)
		if total > 20 {
			results = results[:20]
			total = 20
		}

		return e.JSON(http.StatusOK, map[string]any{
			"total":   total,
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
		var authorNames []string
		var firstPubYear *float64
		var pageCount *int

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
										authorNames = append(authorNames, name)
									}
								}
							}
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
				authorNames = splitAuthors(a)
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

		return e.JSON(http.StatusOK, map[string]any{
			"key":                     workID,
			"title":                   title,
			"authors":                 authorNames,
			"description":             description,
			"cover_url":               coverURL,
			"average_rating":          avgRating,
			"rating_count":            ratingCount,
			"local_reads_count":       readsCount,
			"local_want_to_read_count": wtrCount,
			"publisher":               nil,
			"page_count":              pageCount,
			"first_publish_year":      firstPubYear,
			"edition_count":           editionCount,
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
			UserID      string   `db:"user_id" json:"user_id"`
			Username    string   `db:"username" json:"username"`
			DisplayName *string  `db:"display_name" json:"display_name"`
			Avatar      *string  `db:"avatar" json:"avatar"`
			Rating      *float64 `db:"rating" json:"rating"`
			ReviewText  string   `db:"review_text" json:"review_text"`
			Spoiler     bool     `db:"spoiler" json:"spoiler"`
			DateRead    *string  `db:"date_read" json:"date_read"`
			DateAdded   string   `db:"date_added" json:"date_added"`
		}

		var reviews []reviewRow
		err := app.DB().NewQuery(`
			SELECT ub.user as user_id, u.username, u.display_name, u.avatar,
				   ub.rating, ub.review_text, ub.spoiler, ub.date_read,
				   ub.date_added as date_added
			FROM user_books ub
			JOIN users u ON ub.user = u.id
			WHERE ub.book = {:book} AND ub.review_text != '' AND ub.review_text IS NOT NULL
			ORDER BY CASE WHEN ub.user = {:viewer} THEN 0 ELSE 1 END, ub.date_added DESC
		`).Bind(map[string]any{"book": books[0].Id, "viewer": viewerID}).All(&reviews)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}

		// Build response with avatar URLs
		var result []map[string]any
		for _, r := range reviews {
			var avatarURL *string
			if r.Avatar != nil && *r.Avatar != "" {
				// We need the user's collection ID for avatar URL
				url := fmt.Sprintf("/api/files/users/%s/%s", r.UserID, *r.Avatar)
				avatarURL = &url
			}

			// Check if viewer follows this reviewer
			isFollowed := false
			if viewerID != "" && viewerID != r.UserID {
				follows, err := app.FindRecordsByFilter("follows",
					"follower = {:f} && followee = {:t} && status = 'active'",
					"", 1, 0,
					map[string]any{"f": viewerID, "t": r.UserID},
				)
				isFollowed = err == nil && len(follows) > 0
			}

			result = append(result, map[string]any{
				"username":     r.Username,
				"display_name": r.DisplayName,
				"avatar_url":   avatarURL,
				"rating":       r.Rating,
				"review_text":  r.ReviewText,
				"spoiler":      r.Spoiler,
				"date_read":    r.DateRead,
				"date_added":   r.DateAdded,
				"is_followed":  isFollowed,
			})
		}
		if result == nil {
			result = []map[string]any{}
		}

		return e.JSON(http.StatusOK, result)
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

// GetAuthorDetail handles GET /authors/{authorKey}
func GetAuthorDetail(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		authorKey := e.Request.PathValue("authorKey")
		ol := newOLClient()
		data, err := ol.get(fmt.Sprintf("/authors/%s.json", authorKey))
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Author not found"})
		}
		return e.JSON(http.StatusOK, data)
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
			CoverURL *string `db:"cover_url" json:"cover_url"`
		}
		var books []bookRow
		err := app.DB().NewQuery(`
			SELECT b.open_library_id, b.title, b.cover_url
			FROM book_follows bf
			JOIN books b ON bf.book = b.id
			WHERE bf.user = {:user}
			ORDER BY bf.created DESC
		`).Bind(map[string]any{"user": user.Id}).All(&books)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}
		return e.JSON(http.StatusOK, books)
	}
}
