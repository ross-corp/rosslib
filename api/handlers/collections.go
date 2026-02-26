package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

// GetMyShelves handles GET /me/shelves
func GetMyShelves(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		shelves, err := app.FindRecordsByFilter("collections",
			"user = {:user}", "created", 100, 0,
			map[string]any{"user": user.Id},
		)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}

		var result []map[string]any
		for _, s := range shelves {
			type countResult struct {
				Count int `db:"count"`
			}
			var cnt countResult
			_ = app.DB().NewQuery("SELECT COUNT(*) as count FROM collection_items WHERE collection = {:id}").
				Bind(map[string]any{"id": s.Id}).One(&cnt)

			result = append(result, map[string]any{
				"id":              s.Id,
				"name":            s.GetString("name"),
				"slug":            s.GetString("slug"),
				"is_exclusive":    s.GetBool("is_exclusive"),
				"exclusive_group": s.GetString("exclusive_group"),
				"is_public":       s.GetBool("is_public"),
				"collection_type": s.GetString("collection_type"),
				"item_count":      cnt.Count,
			})
		}
		if result == nil {
			result = []map[string]any{}
		}

		return e.JSON(http.StatusOK, result)
	}
}

// CreateShelf handles POST /me/shelves
func CreateShelf(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		data := struct {
			Name     string `json:"name"`
			IsPublic *bool  `json:"is_public"`
		}{}
		if err := e.BindBody(&data); err != nil || data.Name == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "name is required"})
		}

		slug := slugify(data.Name)
		isPublic := true
		if data.IsPublic != nil {
			isPublic = *data.IsPublic
		}

		coll, err := app.FindCollectionByNameOrId("collections")
		if err != nil {
			return err
		}
		rec := core.NewRecord(coll)
		rec.Set("user", user.Id)
		rec.Set("name", data.Name)
		rec.Set("slug", slug)
		rec.Set("is_public", isPublic)
		rec.Set("collection_type", "shelf")
		if err := app.Save(rec); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}

		return e.JSON(http.StatusOK, map[string]any{
			"id":   rec.Id,
			"name": data.Name,
			"slug": slug,
		})
	}
}

// UpdateShelf handles PATCH /me/shelves/{id}
func UpdateShelf(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		shelfID := e.Request.PathValue("id")

		shelf, err := app.FindRecordById("collections", shelfID)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Shelf not found"})
		}
		if shelf.GetString("user") != user.Id {
			return e.JSON(http.StatusForbidden, map[string]any{"error": "Not your shelf"})
		}

		data := struct {
			Name     *string `json:"name"`
			IsPublic *bool   `json:"is_public"`
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}

		if data.Name != nil {
			shelf.Set("name", *data.Name)
			shelf.Set("slug", slugify(*data.Name))
		}
		if data.IsPublic != nil {
			shelf.Set("is_public", *data.IsPublic)
		}

		if err := app.Save(shelf); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}

		return e.JSON(http.StatusOK, map[string]any{"message": "Shelf updated"})
	}
}

// DeleteShelf handles DELETE /me/shelves/{id}
func DeleteShelf(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		shelfID := e.Request.PathValue("id")

		shelf, err := app.FindRecordById("collections", shelfID)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Shelf not found"})
		}
		if shelf.GetString("user") != user.Id {
			return e.JSON(http.StatusForbidden, map[string]any{"error": "Not your shelf"})
		}
		if err := app.Delete(shelf); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to delete"})
		}

		return e.JSON(http.StatusOK, map[string]any{"message": "Shelf deleted"})
	}
}

// GetUserShelves handles GET /users/{username}/shelves
func GetUserShelves(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		username := e.Request.PathValue("username")
		includeBooks, _ := strconv.Atoi(e.Request.URL.Query().Get("include_books"))

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

		shelves, err := app.FindRecordsByFilter("collections",
			"user = {:user} && is_public = true", "created", 100, 0,
			map[string]any{"user": targetUser.Id},
		)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}

		var result []map[string]any
		for _, s := range shelves {
			type countResult struct {
				Count int `db:"count"`
			}
			var cnt countResult
			_ = app.DB().NewQuery("SELECT COUNT(*) as count FROM collection_items WHERE collection = {:id}").
				Bind(map[string]any{"id": s.Id}).One(&cnt)

			entry := map[string]any{
				"id":              s.Id,
				"name":            s.GetString("name"),
				"slug":            s.GetString("slug"),
				"exclusive_group": s.GetString("exclusive_group"),
				"collection_type": s.GetString("collection_type"),
				"item_count":      cnt.Count,
			}

			if includeBooks > 0 {
				type bookRow struct {
					BookID   string   `db:"book_id" json:"book_id"`
					OLID     string   `db:"open_library_id" json:"open_library_id"`
					Title    string   `db:"title" json:"title"`
					CoverURL *string  `db:"cover_url" json:"cover_url"`
					Rating   *float64 `db:"rating" json:"rating"`
					AddedAt  string   `db:"added_at" json:"added_at"`
				}
				var books []bookRow
				_ = app.DB().NewQuery(`
					SELECT b.id as book_id, b.open_library_id, b.title, b.cover_url,
						   ci.rating, ci.created as added_at
					FROM collection_items ci
					JOIN books b ON ci.book = b.id
					WHERE ci.collection = {:coll}
					ORDER BY ci.created DESC
					LIMIT {:limit}
				`).Bind(map[string]any{"coll": s.Id, "limit": includeBooks}).All(&books)
				if books == nil {
					books = []bookRow{}
				}
				entry["books"] = books
			}

			result = append(result, entry)
		}
		if result == nil {
			result = []map[string]any{}
		}

		return e.JSON(http.StatusOK, result)
	}
}

// GetShelfDetail handles GET /users/{username}/shelves/{slug}
func GetShelfDetail(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		username := e.Request.PathValue("username")
		shelfSlug := e.Request.PathValue("slug")

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

		shelves, err := app.FindRecordsByFilter("collections",
			"user = {:user} && slug = {:slug}", "", 1, 0,
			map[string]any{"user": targetUser.Id, "slug": shelfSlug},
		)
		if err != nil || len(shelves) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Shelf not found"})
		}
		shelf := shelves[0]

		type bookRow struct {
			BookID   string   `db:"book_id" json:"book_id"`
			OLID     string   `db:"open_library_id" json:"open_library_id"`
			Title    string   `db:"title" json:"title"`
			CoverURL *string  `db:"cover_url" json:"cover_url"`
			AddedAt  string   `db:"added_at" json:"added_at"`
			Rating   *float64 `db:"rating" json:"rating"`
		}
		var books []bookRow
		_ = app.DB().NewQuery(`
			SELECT b.id as book_id, b.open_library_id, b.title,
				   COALESCE(NULLIF(ub.selected_edition_cover_url, ''), b.cover_url) as cover_url,
				   ci.created as added_at, ci.rating
			FROM collection_items ci
			JOIN books b ON ci.book = b.id
			LEFT JOIN user_books ub ON ub.user = {:user} AND ub.book = b.id
			WHERE ci.collection = {:coll}
			ORDER BY ci.created DESC
		`).Bind(map[string]any{"coll": shelf.Id, "user": targetUser.Id}).All(&books)
		if books == nil {
			books = []bookRow{}
		}

		return e.JSON(http.StatusOK, map[string]any{
			"id":              shelf.Id,
			"name":            shelf.GetString("name"),
			"slug":            shelf.GetString("slug"),
			"exclusive_group": shelf.GetString("exclusive_group"),
			"books":           books,
		})
	}
}

// AddBookToShelf handles POST /shelves/{shelfId}/books
func AddBookToShelf(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		shelfID := e.Request.PathValue("shelfId")

		shelf, err := app.FindRecordById("collections", shelfID)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Shelf not found"})
		}
		if shelf.GetString("user") != user.Id {
			return e.JSON(http.StatusForbidden, map[string]any{"error": "Not your shelf"})
		}

		data := struct {
			OpenLibraryID   string   `json:"open_library_id"`
			Title           string   `json:"title"`
			CoverURL        string   `json:"cover_url"`
			ISBN13          string   `json:"isbn13"`
			Authors         []string `json:"authors"`
			PublicationYear int      `json:"publication_year"`
		}{}
		if err := e.BindBody(&data); err != nil || data.OpenLibraryID == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "open_library_id required"})
		}

		authors := strings.Join(data.Authors, ", ")
		book, err := upsertBook(app, data.OpenLibraryID, data.Title, data.CoverURL, data.ISBN13, authors, data.PublicationYear)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to save book"})
		}

		// Enforce mutual exclusivity
		if eg := shelf.GetString("exclusive_group"); eg != "" {
			// Find all shelves in the same exclusive group
			groupShelves, _ := app.FindRecordsByFilter("collections",
				"user = {:user} && exclusive_group = {:eg} && id != {:id}",
				"", 100, 0,
				map[string]any{"user": user.Id, "eg": eg, "id": shelfID},
			)
			for _, gs := range groupShelves {
				items, _ := app.FindRecordsByFilter("collection_items",
					"collection = {:coll} && book = {:book}",
					"", 1, 0,
					map[string]any{"coll": gs.Id, "book": book.Id},
				)
				for _, item := range items {
					_ = app.Delete(item)
				}
			}
		}

		// Check if already on this shelf
		existing, _ := app.FindRecordsByFilter("collection_items",
			"collection = {:coll} && book = {:book}",
			"", 1, 0,
			map[string]any{"coll": shelfID, "book": book.Id},
		)
		if len(existing) > 0 {
			return e.JSON(http.StatusOK, map[string]any{"message": "Already on shelf"})
		}

		ciColl, err := app.FindCollectionByNameOrId("collection_items")
		if err != nil {
			return err
		}
		ci := core.NewRecord(ciColl)
		ci.Set("collection", shelfID)
		ci.Set("book", book.Id)
		ci.Set("user", user.Id)
		if err := app.Save(ci); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}

		recordActivity(app, user.Id, "shelved", map[string]any{
			"book":           book.Id,
			"collection_ref": shelfID,
		})

		return e.JSON(http.StatusOK, map[string]any{
			"message":  "Book added to shelf",
			"item_id":  ci.Id,
			"book_id":  book.Id,
			"shelf_id": shelfID,
		})
	}
}

// UpdateShelfBook handles PATCH /shelves/{shelfId}/books/{olId}
func UpdateShelfBook(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		shelfID := e.Request.PathValue("shelfId")
		olID := e.Request.PathValue("olId")

		shelf, err := app.FindRecordById("collections", shelfID)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Shelf not found"})
		}
		if shelf.GetString("user") != user.Id {
			return e.JSON(http.StatusForbidden, map[string]any{"error": "Not your shelf"})
		}

		books, _ := app.FindRecordsByFilter("books",
			"open_library_id = {:id}", "", 1, 0,
			map[string]any{"id": olID},
		)
		if len(books) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Book not found"})
		}

		items, _ := app.FindRecordsByFilter("collection_items",
			"collection = {:coll} && book = {:book}",
			"", 1, 0,
			map[string]any{"coll": shelfID, "book": books[0].Id},
		)
		if len(items) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Book not on shelf"})
		}

		item := items[0]
		data := struct {
			Rating     *float64 `json:"rating"`
			ReviewText *string  `json:"review_text"`
			Spoiler    *bool    `json:"spoiler"`
			DateRead   *string  `json:"date_read"`
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}

		if data.Rating != nil {
			item.Set("rating", *data.Rating)
		}
		if data.ReviewText != nil {
			item.Set("review_text", *data.ReviewText)
		}
		if data.Spoiler != nil {
			item.Set("spoiler", *data.Spoiler)
		}
		if data.DateRead != nil {
			item.Set("date_read", *data.DateRead)
		}

		if err := app.Save(item); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}

		return e.JSON(http.StatusOK, map[string]any{"message": "Updated"})
	}
}

// RemoveBookFromShelf handles DELETE /shelves/{shelfId}/books/{olId}
func RemoveBookFromShelf(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		shelfID := e.Request.PathValue("shelfId")
		olID := e.Request.PathValue("olId")

		shelf, err := app.FindRecordById("collections", shelfID)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Shelf not found"})
		}
		if shelf.GetString("user") != user.Id {
			return e.JSON(http.StatusForbidden, map[string]any{"error": "Not your shelf"})
		}

		books, _ := app.FindRecordsByFilter("books",
			"open_library_id = {:id}", "", 1, 0,
			map[string]any{"id": olID},
		)
		if len(books) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Book not found"})
		}

		items, _ := app.FindRecordsByFilter("collection_items",
			"collection = {:coll} && book = {:book}",
			"", 1, 0,
			map[string]any{"coll": shelfID, "book": books[0].Id},
		)
		if len(items) == 0 {
			return e.JSON(http.StatusOK, map[string]any{"message": "Not on shelf"})
		}

		if err := app.Delete(items[0]); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to remove"})
		}

		return e.JSON(http.StatusOK, map[string]any{"message": "Book removed from shelf"})
	}
}

// ExportCSV handles GET /me/export/csv
func ExportCSV(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		query := `
			SELECT b.open_library_id, b.title, b.authors, b.isbn13,
				   ub.rating, ub.review_text, ub.date_read, ub.date_added as date_added,
				   COALESCE(tv.slug, '') as status
			FROM user_books ub
			JOIN books b ON ub.book = b.id
			LEFT JOIN book_tag_values btv ON btv.user = ub.user AND btv.book = ub.book
			LEFT JOIN tag_keys tk ON btv.tag_key = tk.id AND tk.slug = 'status'
			LEFT JOIN tag_values tv ON btv.tag_value = tv.id AND tk.id IS NOT NULL
			WHERE ub.user = {:user}
			ORDER BY ub.date_added DESC
		`
		params := map[string]any{"user": user.Id}

		type row struct {
			OLID       string   `db:"open_library_id"`
			Title      string   `db:"title"`
			Authors    *string  `db:"authors"`
			ISBN13     *string  `db:"isbn13"`
			Rating     *float64 `db:"rating"`
			ReviewText *string  `db:"review_text"`
			DateRead   *string  `db:"date_read"`
			DateAdded  string   `db:"date_added"`
			Status     string   `db:"status"`
		}
		var rows []row
		err := app.DB().NewQuery(query).Bind(params).All(&rows)
		if err != nil {
			rows = []row{}
		}

		// Build CSV
		var sb strings.Builder
		sb.WriteString("Open Library ID,Title,Authors,ISBN13,Rating,Review,Date Read,Date Added,Status\n")
		for _, r := range rows {
			sb.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
				csvEscape(r.OLID),
				csvEscape(r.Title),
				csvEscape(ptrStr(r.Authors)),
				csvEscape(ptrStr(r.ISBN13)),
				csvFloat(r.Rating),
				csvEscape(ptrStr(r.ReviewText)),
				csvEscape(ptrStr(r.DateRead)),
				csvEscape(r.DateAdded),
				csvEscape(r.Status),
			))
		}

		e.Response.Header().Set("Content-Type", "text/csv")
		e.Response.Header().Set("Content-Disposition", `attachment; filename="rosslib-export.csv"`)
		e.Response.WriteHeader(http.StatusOK)
		_, _ = e.Response.Write([]byte(sb.String()))
		return nil
	}
}

func csvEscape(s string) string {
	if strings.ContainsAny(s, ",\"\n") {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return s
}

func ptrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func csvFloat(f *float64) string {
	if f == nil {
		return ""
	}
	return fmt.Sprintf("%.1f", *f)
}
