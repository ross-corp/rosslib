package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// AddBook handles POST /me/books
func AddBook(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		data := struct {
			OpenLibraryID   string   `json:"open_library_id"`
			Title           string   `json:"title"`
			CoverURL        string   `json:"cover_url"`
			ISBN13          string   `json:"isbn13"`
			Authors         []string `json:"authors"`
			PublicationYear int      `json:"publication_year"`
			StatusSlug      string   `json:"status_slug"`
			Rating          *float64 `json:"rating"`
			ReviewText      string   `json:"review_text"`
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}
		if data.OpenLibraryID == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "open_library_id required"})
		}

		authors := strings.Join(data.Authors, ", ")
		book, err := upsertBook(app, data.OpenLibraryID, data.Title, data.CoverURL, data.ISBN13, authors, data.PublicationYear)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to save book"})
		}

		// Upsert user_books
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
				return err
			}
			ub = core.NewRecord(coll)
			ub.Set("user", user.Id)
			ub.Set("book", book.Id)
			ub.Set("date_added", time.Now().UTC().Format(time.RFC3339))
		}

		if data.Rating != nil {
			if *data.Rating != 0 && (*data.Rating < 1 || *data.Rating > 5) {
				return e.JSON(http.StatusBadRequest, map[string]any{"error": "Rating must be between 1 and 5"})
			}
			ub.Set("rating", *data.Rating)
		}
		if data.ReviewText != "" {
			ub.Set("review_text", data.ReviewText)
		}

		if err := app.Save(ub); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}

		// Set status tag if provided
		if data.StatusSlug != "" {
			setStatusTag(app, user.Id, book.Id, data.StatusSlug)
		}

		recordActivity(app, user.Id, "shelved", map[string]any{"book": book.Id})
		refreshBookStats(app, book.Id)

		return e.JSON(http.StatusOK, map[string]any{
			"book_id":         book.Id,
			"open_library_id": data.OpenLibraryID,
			"user_book_id":    ub.Id,
		})
	}
}

// UpdateBook handles PATCH /me/books/{olId}
func UpdateBook(app core.App) func(e *core.RequestEvent) error {
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

		ubs, _ := app.FindRecordsByFilter("user_books",
			"user = {:user} && book = {:book}",
			"", 1, 0,
			map[string]any{"user": user.Id, "book": book.Id},
		)
		if len(ubs) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Book not in your library"})
		}
		ub := ubs[0]

		data := struct {
			Rating                  *float64 `json:"rating"`
			ReviewText              *string  `json:"review_text"`
			Spoiler                 *bool    `json:"spoiler"`
			DateRead                *string  `json:"date_read"`
			DateDnf                 *string  `json:"date_dnf"`
			DateStarted             *string  `json:"date_started"`
			ProgressPages           *int     `json:"progress_pages"`
			ProgressPercent         *int     `json:"progress_percent"`
			StatusSlug              *string  `json:"status_slug"`
			SelectedEditionKey      *string  `json:"selected_edition_key"`
			SelectedEditionCoverURL *string  `json:"selected_edition_cover_url"`
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}

		if data.Rating != nil {
			if *data.Rating != 0 && (*data.Rating < 1 || *data.Rating > 5) {
				return e.JSON(http.StatusBadRequest, map[string]any{"error": "Rating must be between 1 and 5"})
			}
			ub.Set("rating", *data.Rating)
		}
		if data.ReviewText != nil {
			ub.Set("review_text", *data.ReviewText)
		}
		if data.Spoiler != nil {
			ub.Set("spoiler", *data.Spoiler)
		}
		if data.DateRead != nil {
			ub.Set("date_read", *data.DateRead)
		}
		if data.DateDnf != nil {
			ub.Set("date_dnf", *data.DateDnf)
		}
		if data.DateStarted != nil {
			ub.Set("date_started", *data.DateStarted)
		}
		if data.ProgressPages != nil {
			ub.Set("progress_pages", *data.ProgressPages)
		}
		if data.ProgressPercent != nil {
			ub.Set("progress_percent", *data.ProgressPercent)
		}
		if data.SelectedEditionKey != nil {
			ub.Set("selected_edition_key", *data.SelectedEditionKey)
		}
		if data.SelectedEditionCoverURL != nil {
			ub.Set("selected_edition_cover_url", *data.SelectedEditionCoverURL)
		}

		if err := app.Save(ub); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}

		if data.StatusSlug != nil {
			setStatusTag(app, user.Id, book.Id, *data.StatusSlug)
		}

		refreshBookStats(app, book.Id)

		return e.JSON(http.StatusOK, map[string]any{"message": "Book updated"})
	}
}

// DeleteBook handles DELETE /me/books/{olId}
func DeleteBook(app core.App) func(e *core.RequestEvent) error {
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

		ubs, _ := app.FindRecordsByFilter("user_books",
			"user = {:user} && book = {:book}",
			"", 1, 0,
			map[string]any{"user": user.Id, "book": book.Id},
		)
		if len(ubs) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Book not in your library"})
		}

		// Delete user_book
		if err := app.Delete(ubs[0]); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to delete"})
		}

		// Clean up tag assignments for this book
		btvs, _ := app.FindRecordsByFilter("book_tag_values",
			"user = {:user} && book = {:book}",
			"", 100, 0,
			map[string]any{"user": user.Id, "book": book.Id},
		)
		for _, btv := range btvs {
			_ = app.Delete(btv)
		}

		// Clean up collection items
		cis, _ := app.FindRecordsByFilter("collection_items",
			"user = {:user} && book = {:book}",
			"", 100, 0,
			map[string]any{"user": user.Id, "book": book.Id},
		)
		for _, ci := range cis {
			_ = app.Delete(ci)
		}

		refreshBookStats(app, book.Id)

		return e.JSON(http.StatusOK, map[string]any{"message": "Book removed"})
	}
}

// GetBookStatus handles GET /me/books/{olId}/status
func GetBookStatus(app core.App) func(e *core.RequestEvent) error {
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
			return e.JSON(http.StatusOK, map[string]any{
				"status_value_id": nil, "status_name": nil, "status_slug": nil,
				"rating": nil, "review_text": nil, "spoiler": false,
				"date_read": nil, "date_dnf": nil, "date_started": nil,
				"progress_pages": nil, "progress_percent": nil,
				"selected_edition_key": nil, "selected_edition_cover_url": nil,
			})
		}
		book := books[0]

		// Get user_book
		ubs, _ := app.FindRecordsByFilter("user_books",
			"user = {:user} && book = {:book}",
			"", 1, 0,
			map[string]any{"user": user.Id, "book": book.Id},
		)

		result := map[string]any{
			"status_value_id":            nil,
			"status_name":               nil,
			"status_slug":               nil,
			"rating":                    nil,
			"review_text":               nil,
			"spoiler":                   false,
			"date_read":                 nil,
			"date_dnf":                  nil,
			"date_started":              nil,
			"progress_pages":            nil,
			"progress_percent":          nil,
			"selected_edition_key":      nil,
			"selected_edition_cover_url": nil,
		}

		if len(ubs) > 0 {
			ub := ubs[0]
			if r := ub.GetFloat("rating"); r > 0 {
				result["rating"] = r
			}
			if rt := ub.GetString("review_text"); rt != "" {
				result["review_text"] = rt
			}
			result["spoiler"] = ub.GetBool("spoiler")
			if dr := ub.GetString("date_read"); dr != "" {
				result["date_read"] = dr
			}
			if dd := ub.GetString("date_dnf"); dd != "" {
				result["date_dnf"] = dd
			}
			if ds := ub.GetString("date_started"); ds != "" {
				result["date_started"] = ds
			}
			if pp := ub.GetInt("progress_pages"); pp > 0 {
				result["progress_pages"] = pp
			}
			if pct := ub.GetInt("progress_percent"); pct > 0 {
				result["progress_percent"] = pct
			}
			if sek := ub.GetString("selected_edition_key"); sek != "" {
				result["selected_edition_key"] = sek
			}
			if secu := ub.GetString("selected_edition_cover_url"); secu != "" {
				result["selected_edition_cover_url"] = secu
			}
		}

		// Get status tag
		type statusRow struct {
			ValueID string `db:"value_id"`
			Name    string `db:"name"`
			Slug    string `db:"slug"`
		}
		var status statusRow
		err := app.DB().NewQuery(`
			SELECT tv.id as value_id, tv.name, tv.slug
			FROM book_tag_values btv
			JOIN tag_keys tk ON btv.tag_key = tk.id
			JOIN tag_values tv ON btv.tag_value = tv.id
			WHERE btv.user = {:user} AND btv.book = {:book} AND tk.slug = 'status'
			LIMIT 1
		`).Bind(map[string]any{"user": user.Id, "book": book.Id}).One(&status)
		if err == nil {
			result["status_value_id"] = status.ValueID
			result["status_name"] = status.Name
			result["status_slug"] = status.Slug
		}

		return e.JSON(http.StatusOK, result)
	}
}

// SetBookStatus handles PUT /me/books/{olId}/status
func SetBookStatus(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		olID := e.Request.PathValue("olId")
		data := struct {
			Slug string `json:"slug"`
		}{}
		if err := e.BindBody(&data); err != nil || data.Slug == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "slug is required"})
		}

		books, _ := app.FindRecordsByFilter("books",
			"open_library_id = {:id}", "", 1, 0,
			map[string]any{"id": olID},
		)
		if len(books) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Book not found"})
		}

		setStatusTag(app, user.Id, books[0].Id, data.Slug)
		return e.JSON(http.StatusOK, map[string]any{"ok": true})
	}
}

// GetStatusMap handles GET /me/books/status-map
func GetStatusMap(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		type mapRow struct {
			OLID       string `db:"open_library_id"`
			StatusSlug string `db:"status_slug"`
		}
		var rows []mapRow
		err := app.DB().NewQuery(`
			SELECT b.open_library_id, tv.slug as status_slug
			FROM book_tag_values btv
			JOIN tag_keys tk ON btv.tag_key = tk.id
			JOIN tag_values tv ON btv.tag_value = tv.id
			JOIN books b ON btv.book = b.id
			WHERE btv.user = {:user} AND tk.slug = 'status'
		`).Bind(map[string]any{"user": user.Id}).All(&rows)
		if err != nil {
			return e.JSON(http.StatusOK, map[string]string{})
		}

		result := map[string]string{}
		for _, r := range rows {
			result[r.OLID] = r.StatusSlug
		}

		return e.JSON(http.StatusOK, result)
	}
}

// GetUserBooks handles GET /users/{username}/books
func GetUserBooks(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		username := e.Request.PathValue("username")

		users, err := app.FindRecordsByFilter("users",
			"username = {:username}", "", 1, 0,
			map[string]any{"username": username},
		)
		if err != nil || len(users) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "User not found"})
		}
		targetUser := users[0]

		viewerID := ""
		if e.Auth != nil {
			viewerID = e.Auth.Id
		}
		if !canViewProfile(app, viewerID, targetUser) {
			return e.JSON(http.StatusForbidden, map[string]any{"error": "Profile is private"})
		}

		statusFilter := e.Request.URL.Query().Get("status")
		limit, _ := strconv.Atoi(e.Request.URL.Query().Get("limit"))
		if limit <= 0 {
			limit = 8
		}

		sortParam := e.Request.URL.Query().Get("sort")
		var orderClause string
		switch sortParam {
		case "title":
			orderClause = "b.title ASC"
		case "author":
			orderClause = "b.authors ASC, b.title ASC"
		case "rating":
			orderClause = "ub.rating DESC NULLS LAST, ub.date_added DESC"
		default:
			orderClause = "ub.date_added DESC"
		}

		// If status filter provided, return flat list
		if statusFilter != "" {
			type bookRow struct {
				BookID         string   `db:"book_id" json:"book_id"`
				OLID           string   `db:"open_library_id" json:"open_library_id"`
				Title          string   `db:"title" json:"title"`
				CoverURL       *string  `db:"cover_url" json:"cover_url"`
				Authors        *string  `db:"authors" json:"authors"`
				Rating         *float64 `db:"rating" json:"rating"`
				AddedAt        string   `db:"added_at" json:"added_at"`
				SeriesPosition *int     `db:"series_position" json:"series_position"`
			}
			var books []bookRow
			if statusFilter == "_unstatused" {
				err := app.DB().NewQuery(`
					SELECT b.id as book_id, b.open_library_id, b.title,
						   COALESCE(NULLIF(ub.selected_edition_cover_url, ''), b.cover_url) as cover_url,
						   b.authors, ub.rating, ub.date_added as added_at,
						   (SELECT bs.position FROM book_series bs WHERE bs.book = b.id LIMIT 1) as series_position
					FROM user_books ub
					JOIN books b ON ub.book = b.id
					WHERE ub.user = {:user}
					AND NOT EXISTS (
						SELECT 1 FROM book_tag_values btv
						JOIN tag_keys tk ON btv.tag_key = tk.id
						WHERE btv.user = ub.user AND btv.book = ub.book AND tk.slug = 'status'
					)
					ORDER BY ` + orderClause + `
					LIMIT {:limit}
				`).Bind(map[string]any{"user": targetUser.Id, "limit": limit}).All(&books)
				if err != nil || books == nil {
					books = []bookRow{}
				}
			} else {
				err := app.DB().NewQuery(`
					SELECT b.id as book_id, b.open_library_id, b.title,
						   COALESCE(NULLIF(ub.selected_edition_cover_url, ''), b.cover_url) as cover_url,
						   b.authors, ub.rating, ub.date_added as added_at,
						   (SELECT bs.position FROM book_series bs WHERE bs.book = b.id LIMIT 1) as series_position
					FROM user_books ub
					JOIN books b ON ub.book = b.id
					JOIN book_tag_values btv ON btv.user = ub.user AND btv.book = ub.book
					JOIN tag_keys tk ON btv.tag_key = tk.id
					JOIN tag_values tv ON btv.tag_value = tv.id
					WHERE ub.user = {:user} AND tk.slug = 'status' AND tv.slug = {:status}
					ORDER BY ` + orderClause + `
					LIMIT {:limit}
				`).Bind(map[string]any{"user": targetUser.Id, "status": statusFilter, "limit": limit}).All(&books)
				if err != nil || books == nil {
					books = []bookRow{}
				}
			}
			return e.JSON(http.StatusOK, map[string]any{"books": books})
		}

		// Return grouped by status
		key, values, err := ensureStatusTagKey(app, targetUser.Id)
		if err != nil {
			return e.JSON(http.StatusOK, map[string]any{"statuses": []any{}, "unstatused_count": 0})
		}
		_ = key

		var statuses []map[string]any
		for _, v := range values {
			type bookRow struct {
				BookID          string   `db:"book_id" json:"book_id"`
				OLID            string   `db:"open_library_id" json:"open_library_id"`
				Title           string   `db:"title" json:"title"`
				CoverURL        *string  `db:"cover_url" json:"cover_url"`
				Rating          *float64 `db:"rating" json:"rating"`
				AddedAt         string   `db:"added_at" json:"added_at"`
				ProgressPages   *int     `db:"progress_pages" json:"progress_pages"`
				ProgressPercent *int     `db:"progress_percent" json:"progress_percent"`
				PageCount       *int     `db:"page_count" json:"page_count"`
				SeriesPosition  *int     `db:"series_position" json:"series_position"`
			}
			var books []bookRow
			_ = app.DB().NewQuery(`
				SELECT b.id as book_id, b.open_library_id, b.title,
					   COALESCE(NULLIF(ub.selected_edition_cover_url, ''), b.cover_url) as cover_url,
					   ub.rating, ub.date_added as added_at,
					   ub.progress_pages, ub.progress_percent, b.page_count,
					   (SELECT bs.position FROM book_series bs WHERE bs.book = b.id LIMIT 1) as series_position
				FROM book_tag_values btv
				JOIN books b ON btv.book = b.id
				LEFT JOIN user_books ub ON ub.user = btv.user AND ub.book = btv.book
				WHERE btv.user = {:user} AND btv.tag_value = {:value}
				ORDER BY ub.date_added DESC
				LIMIT {:limit}
			`).Bind(map[string]any{"user": targetUser.Id, "value": v.Id, "limit": limit}).All(&books)

			type countResult struct {
				Count int `db:"count"`
			}
			var cnt countResult
			_ = app.DB().NewQuery(`
				SELECT COUNT(*) as count FROM book_tag_values
				WHERE user = {:user} AND tag_value = {:value}
			`).Bind(map[string]any{"user": targetUser.Id, "value": v.Id}).One(&cnt)

			if books == nil {
				books = []bookRow{}
			}

			statuses = append(statuses, map[string]any{
				"name":  v.GetString("name"),
				"slug":  v.GetString("slug"),
				"count": cnt.Count,
				"books": books,
			})
		}

		if statuses == nil {
			statuses = []map[string]any{}
		}

		// Fetch unstatused books
		type unstatusedBookRow struct {
			BookID         string   `db:"book_id" json:"book_id"`
			OLID           string   `db:"open_library_id" json:"open_library_id"`
			Title          string   `db:"title" json:"title"`
			CoverURL       *string  `db:"cover_url" json:"cover_url"`
			Rating         *float64 `db:"rating" json:"rating"`
			AddedAt        string   `db:"added_at" json:"added_at"`
			SeriesPosition *int     `db:"series_position" json:"series_position"`
		}
		var unstatusedBooks []unstatusedBookRow
		_ = app.DB().NewQuery(`
			SELECT b.id as book_id, b.open_library_id, b.title,
				   COALESCE(NULLIF(ub.selected_edition_cover_url, ''), b.cover_url) as cover_url,
				   ub.rating, ub.date_added as added_at,
				   (SELECT bs.position FROM book_series bs WHERE bs.book = b.id LIMIT 1) as series_position
			FROM user_books ub
			JOIN books b ON ub.book = b.id
			WHERE ub.user = {:user}
			AND NOT EXISTS (
				SELECT 1 FROM book_tag_values btv
				JOIN tag_keys tk ON btv.tag_key = tk.id
				WHERE btv.user = ub.user AND btv.book = ub.book AND tk.slug = 'status'
			)
			ORDER BY ub.date_added DESC
			LIMIT {:limit}
		`).Bind(map[string]any{"user": targetUser.Id, "limit": limit}).All(&unstatusedBooks)
		if unstatusedBooks == nil {
			unstatusedBooks = []unstatusedBookRow{}
		}

		type unstatusedCountResult struct {
			Count int `db:"count"`
		}
		var unstatusedCnt unstatusedCountResult
		_ = app.DB().NewQuery(`
			SELECT COUNT(*) as count FROM user_books ub
			WHERE ub.user = {:user}
			AND NOT EXISTS (
				SELECT 1 FROM book_tag_values btv
				JOIN tag_keys tk ON btv.tag_key = tk.id
				WHERE btv.user = ub.user AND btv.book = ub.book AND tk.slug = 'status'
			)
		`).Bind(map[string]any{"user": targetUser.Id}).One(&unstatusedCnt)

		return e.JSON(http.StatusOK, map[string]any{
			"statuses":         statuses,
			"unstatused_count": unstatusedCnt.Count,
			"unstatused_books": unstatusedBooks,
		})
	}
}

// setStatusTag sets the status tag for a user's book.
func setStatusTag(app core.App, userID, bookID, statusSlug string) {
	if statusSlug == "" {
		return
	}

	key, values, err := ensureStatusTagKey(app, userID)
	if err != nil {
		return
	}

	// Find the matching value
	var targetValue *core.Record
	for _, v := range values {
		if v.GetString("slug") == statusSlug {
			targetValue = v
			break
		}
	}
	if targetValue == nil {
		return
	}

	// Remove existing status tag for this book (select_one mode)
	existing, _ := app.FindRecordsByFilter("book_tag_values",
		"user = {:user} && book = {:book} && tag_key = {:key}",
		"", 10, 0,
		map[string]any{"user": userID, "book": bookID, "key": key.Id},
	)
	for _, e := range existing {
		_ = app.Delete(e)
	}

	// Create new assignment
	coll, err := app.FindCollectionByNameOrId("book_tag_values")
	if err != nil {
		return
	}
	rec := core.NewRecord(coll)
	rec.Set("user", userID)
	rec.Set("book", bookID)
	rec.Set("tag_key", key.Id)
	rec.Set("tag_value", targetValue.Id)
	_ = app.Save(rec)

	// Auto-set date_started when status changes to "currently-reading"
	if statusSlug == "currently-reading" {
		ubs, _ := app.FindRecordsByFilter("user_books",
			"user = {:user} && book = {:book}",
			"", 1, 0,
			map[string]any{"user": userID, "book": bookID},
		)
		if len(ubs) > 0 && ubs[0].GetString("date_started") == "" {
			ubs[0].Set("date_started", time.Now().UTC().Format(time.RFC3339))
			_ = app.Save(ubs[0])
		}
	}
}
