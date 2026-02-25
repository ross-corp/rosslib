package handlers

import (
	"net/http"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

// GetTagKeys handles GET /me/tag-keys
func GetTagKeys(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		// Ensure Status key exists
		_, _, _ = ensureStatusTagKey(app, user.Id)

		keys, err := app.FindRecordsByFilter("tag_keys",
			"user = {:user}", "created", 100, 0,
			map[string]any{"user": user.Id},
		)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}

		var result []map[string]any
		for _, k := range keys {
			values, _ := app.FindRecordsByFilter("tag_values",
				"tag_key = {:key}", "created", 100, 0,
				map[string]any{"key": k.Id},
			)
			var vals []map[string]any
			for _, v := range values {
				vals = append(vals, map[string]any{
					"id":   v.Id,
					"name": v.GetString("name"),
					"slug": v.GetString("slug"),
				})
			}
			if vals == nil {
				vals = []map[string]any{}
			}
			result = append(result, map[string]any{
				"id":     k.Id,
				"name":   k.GetString("name"),
				"slug":   k.GetString("slug"),
				"mode":   k.GetString("mode"),
				"values": vals,
			})
		}
		if result == nil {
			result = []map[string]any{}
		}

		return e.JSON(http.StatusOK, result)
	}
}

// CreateTagKey handles POST /me/tag-keys
func CreateTagKey(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		data := struct {
			Name string `json:"name"`
			Mode string `json:"mode"`
		}{}
		if err := e.BindBody(&data); err != nil || data.Name == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "name is required"})
		}
		if data.Mode == "" {
			data.Mode = "select_one"
		}

		slug := tagSlugify(data.Name)

		coll, err := app.FindCollectionByNameOrId("tag_keys")
		if err != nil {
			return err
		}
		rec := core.NewRecord(coll)
		rec.Set("user", user.Id)
		rec.Set("name", data.Name)
		rec.Set("slug", slug)
		rec.Set("mode", data.Mode)
		if err := app.Save(rec); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}

		return e.JSON(http.StatusOK, map[string]any{
			"id":   rec.Id,
			"name": data.Name,
			"slug": slug,
			"mode": data.Mode,
		})
	}
}

// DeleteTagKey handles DELETE /me/tag-keys/{keyId}
func DeleteTagKey(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		keyID := e.Request.PathValue("keyId")

		key, err := app.FindRecordById("tag_keys", keyID)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Tag key not found"})
		}
		if key.GetString("user") != user.Id {
			return e.JSON(http.StatusForbidden, map[string]any{"error": "Not your tag key"})
		}
		// Prevent deleting Status key
		if key.GetString("slug") == "status" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Cannot delete Status tag"})
		}

		// Cascade: delete values and book_tag_values (handled by CascadeDelete on relations)
		if err := app.Delete(key); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to delete"})
		}

		return e.JSON(http.StatusOK, map[string]any{"message": "Tag key deleted"})
	}
}

// CreateTagValue handles POST /me/tag-keys/{keyId}/values
func CreateTagValue(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		keyID := e.Request.PathValue("keyId")

		key, err := app.FindRecordById("tag_keys", keyID)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Tag key not found"})
		}
		if key.GetString("user") != user.Id {
			return e.JSON(http.StatusForbidden, map[string]any{"error": "Not your tag key"})
		}

		data := struct {
			Name string `json:"name"`
		}{}
		if err := e.BindBody(&data); err != nil || data.Name == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "name is required"})
		}

		slug := tagSlugify(data.Name)

		coll, err := app.FindCollectionByNameOrId("tag_values")
		if err != nil {
			return err
		}
		rec := core.NewRecord(coll)
		rec.Set("tag_key", keyID)
		rec.Set("name", data.Name)
		rec.Set("slug", slug)
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

// DeleteTagValue handles DELETE /me/tag-keys/{keyId}/values/{valueId}
func DeleteTagValue(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		keyID := e.Request.PathValue("keyId")
		valueID := e.Request.PathValue("valueId")

		key, err := app.FindRecordById("tag_keys", keyID)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Tag key not found"})
		}
		if key.GetString("user") != user.Id {
			return e.JSON(http.StatusForbidden, map[string]any{"error": "Not your tag key"})
		}

		value, err := app.FindRecordById("tag_values", valueID)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Tag value not found"})
		}
		if value.GetString("tag_key") != keyID {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Value does not belong to this key"})
		}

		if err := app.Delete(value); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to delete"})
		}

		return e.JSON(http.StatusOK, map[string]any{"message": "Tag value deleted"})
	}
}

// GetBookTags handles GET /me/books/{olId}/tags
func GetBookTags(app core.App) func(e *core.RequestEvent) error {
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
			return e.JSON(http.StatusOK, []any{})
		}

		btvs, _ := app.FindRecordsByFilter("book_tag_values",
			"user = {:user} && book = {:book}",
			"", 100, 0,
			map[string]any{"user": user.Id, "book": books[0].Id},
		)

		var result []map[string]any
		for _, btv := range btvs {
			result = append(result, map[string]any{
				"key_id":   btv.GetString("tag_key"),
				"value_id": btv.GetString("tag_value"),
			})
		}
		if result == nil {
			result = []map[string]any{}
		}

		return e.JSON(http.StatusOK, result)
	}
}

// SetBookTag handles PUT /me/books/{olId}/tags/{keyId}
func SetBookTag(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		olID := e.Request.PathValue("olId")
		keyID := e.Request.PathValue("keyId")

		books, _ := app.FindRecordsByFilter("books",
			"open_library_id = {:id}", "", 1, 0,
			map[string]any{"id": olID},
		)
		if len(books) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Book not found"})
		}
		book := books[0]

		key, err := app.FindRecordById("tag_keys", keyID)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Tag key not found"})
		}
		if key.GetString("user") != user.Id {
			return e.JSON(http.StatusForbidden, map[string]any{"error": "Not your tag key"})
		}

		data := struct {
			ValueID string `json:"value_id"`
		}{}
		if err := e.BindBody(&data); err != nil || data.ValueID == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "value_id is required"})
		}

		mode := key.GetString("mode")
		if mode == "select_one" {
			// Remove existing assignments for this key
			existing, _ := app.FindRecordsByFilter("book_tag_values",
				"user = {:user} && book = {:book} && tag_key = {:key}",
				"", 10, 0,
				map[string]any{"user": user.Id, "book": book.Id, "key": keyID},
			)
			for _, e := range existing {
				_ = app.Delete(e)
			}
		}

		coll, err := app.FindCollectionByNameOrId("book_tag_values")
		if err != nil {
			return err
		}
		rec := core.NewRecord(coll)
		rec.Set("user", user.Id)
		rec.Set("book", book.Id)
		rec.Set("tag_key", keyID)
		rec.Set("tag_value", data.ValueID)
		if err := app.Save(rec); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}

		// If this is the status tag, refresh book stats
		if key.GetString("slug") == "status" {
			refreshBookStats(app, book.Id)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"key_id":   keyID,
			"value_id": data.ValueID,
		})
	}
}

// UnsetBookTag handles DELETE /me/books/{olId}/tags/{keyId}
func UnsetBookTag(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		olID := e.Request.PathValue("olId")
		keyID := e.Request.PathValue("keyId")

		books, _ := app.FindRecordsByFilter("books",
			"open_library_id = {:id}", "", 1, 0,
			map[string]any{"id": olID},
		)
		if len(books) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Book not found"})
		}

		existing, _ := app.FindRecordsByFilter("book_tag_values",
			"user = {:user} && book = {:book} && tag_key = {:key}",
			"", 10, 0,
			map[string]any{"user": user.Id, "book": books[0].Id, "key": keyID},
		)
		for _, e := range existing {
			_ = app.Delete(e)
		}

		return e.JSON(http.StatusOK, map[string]any{"message": "Tag unset"})
	}
}

// UnsetBookTagValue handles DELETE /me/books/{olId}/tags/{keyId}/values/{valueId}
func UnsetBookTagValue(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		olID := e.Request.PathValue("olId")
		keyID := e.Request.PathValue("keyId")
		valueID := e.Request.PathValue("valueId")

		books, _ := app.FindRecordsByFilter("books",
			"open_library_id = {:id}", "", 1, 0,
			map[string]any{"id": olID},
		)
		if len(books) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Book not found"})
		}

		existing, _ := app.FindRecordsByFilter("book_tag_values",
			"user = {:user} && book = {:book} && tag_key = {:key} && tag_value = {:value}",
			"", 1, 0,
			map[string]any{"user": user.Id, "book": books[0].Id, "key": keyID, "value": valueID},
		)
		if len(existing) > 0 {
			_ = app.Delete(existing[0])
		}

		return e.JSON(http.StatusOK, map[string]any{"message": "Tag value unset"})
	}
}

// GetUserTagKeys handles GET /users/{username}/tag-keys
func GetUserTagKeys(app core.App) func(e *core.RequestEvent) error {
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

		keys, err := app.FindRecordsByFilter("tag_keys",
			"user = {:user} && slug != 'status'", "created", 100, 0,
			map[string]any{"user": targetUser.Id},
		)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}

		var result []map[string]any
		for _, k := range keys {
			values, _ := app.FindRecordsByFilter("tag_values",
				"tag_key = {:key}", "created", 100, 0,
				map[string]any{"key": k.Id},
			)
			var vals []map[string]any
			for _, v := range values {
				vals = append(vals, map[string]any{
					"id":   v.Id,
					"name": v.GetString("name"),
					"slug": v.GetString("slug"),
				})
			}
			if vals == nil {
				vals = []map[string]any{}
			}
			result = append(result, map[string]any{
				"id":     k.Id,
				"name":   k.GetString("name"),
				"slug":   k.GetString("slug"),
				"mode":   k.GetString("mode"),
				"values": vals,
			})
		}
		if result == nil {
			result = []map[string]any{}
		}

		return e.JSON(http.StatusOK, result)
	}
}

// GetUserTagBooks handles GET /users/{username}/tags/{path...}
func GetUserTagBooks(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		username := e.Request.PathValue("username")
		tagPath := e.Request.PathValue("path")

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

		// Parse path: could be "key-slug" or "key-slug/value-slug"
		parts := strings.SplitN(tagPath, "/", 2)
		keySlug := parts[0]
		valueSlug := ""
		if len(parts) > 1 {
			valueSlug = parts[1]
		}

		// Find tag key
		keys, err := app.FindRecordsByFilter("tag_keys",
			"user = {:user} && slug = {:slug}",
			"", 1, 0,
			map[string]any{"user": targetUser.Id, "slug": keySlug},
		)
		if err != nil || len(keys) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Tag key not found"})
		}
		key := keys[0]

		filter := "btv.user = {:user} AND btv.tag_key = {:key}"
		params := map[string]any{"user": targetUser.Id, "key": key.Id}

		if valueSlug != "" {
			// Find tag value
			values, err := app.FindRecordsByFilter("tag_values",
				"tag_key = {:key} && slug = {:slug}",
				"", 1, 0,
				map[string]any{"key": key.Id, "slug": valueSlug},
			)
			if err != nil || len(values) == 0 {
				return e.JSON(http.StatusNotFound, map[string]any{"error": "Tag value not found"})
			}
			filter += " AND btv.tag_value = {:value}"
			params["value"] = values[0].Id
		}

		type bookRow struct {
			BookID   string   `db:"book_id" json:"book_id"`
			OLID     string   `db:"open_library_id" json:"open_library_id"`
			Title    string   `db:"title" json:"title"`
			CoverURL *string  `db:"cover_url" json:"cover_url"`
			Rating   *float64 `db:"rating" json:"rating"`
		}
		var books []bookRow
		err = app.DB().NewQuery(`
			SELECT b.id as book_id, b.open_library_id, b.title, b.cover_url, ub.rating
			FROM book_tag_values btv
			JOIN books b ON btv.book = b.id
			LEFT JOIN user_books ub ON ub.user = btv.user AND ub.book = btv.book
			WHERE ` + filter + `
			ORDER BY btv.created DESC
		`).Bind(params).All(&books)
		if err != nil || books == nil {
			books = []bookRow{}
		}

		return e.JSON(http.StatusOK, map[string]any{"books": books})
	}
}

// GetUserLabelBooks handles GET /users/{username}/labels/{keySlug}/{valuePath...}
func GetUserLabelBooks(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		username := e.Request.PathValue("username")
		keySlug := e.Request.PathValue("keySlug")
		valuePath := e.Request.PathValue("valuePath")

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

		// Find tag key
		keys, err := app.FindRecordsByFilter("tag_keys",
			"user = {:user} && slug = {:slug}",
			"", 1, 0,
			map[string]any{"user": targetUser.Id, "slug": keySlug},
		)
		if err != nil || len(keys) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Tag key not found"})
		}
		key := keys[0]

		// Find tag value
		values, err := app.FindRecordsByFilter("tag_values",
			"tag_key = {:key} && slug = {:slug}",
			"", 1, 0,
			map[string]any{"key": key.Id, "slug": valuePath},
		)
		if err != nil || len(values) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Tag value not found"})
		}

		type bookRow struct {
			BookID   string   `db:"book_id" json:"book_id"`
			OLID     string   `db:"open_library_id" json:"open_library_id"`
			Title    string   `db:"title" json:"title"`
			CoverURL *string  `db:"cover_url" json:"cover_url"`
			Rating   *float64 `db:"rating" json:"rating"`
		}
		var books []bookRow
		err = app.DB().NewQuery(`
			SELECT b.id as book_id, b.open_library_id, b.title, b.cover_url, ub.rating
			FROM book_tag_values btv
			JOIN books b ON btv.book = b.id
			LEFT JOIN user_books ub ON ub.user = btv.user AND ub.book = btv.book
			WHERE btv.user = {:user} AND btv.tag_value = {:value}
			ORDER BY btv.created DESC
		`).Bind(map[string]any{"user": targetUser.Id, "value": values[0].Id}).All(&books)
		if err != nil || books == nil {
			books = []bookRow{}
		}

		return e.JSON(http.StatusOK, map[string]any{
			"key_name":   key.GetString("name"),
			"value_name": values[0].GetString("name"),
			"books":      books,
		})
	}
}
