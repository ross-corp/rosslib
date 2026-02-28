package handlers

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

// GetSavedSearches handles GET /me/saved-searches
func GetSavedSearches(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		records, err := app.FindRecordsByFilter("saved_searches",
			"user = {:user}", "-created", 20, 0,
			map[string]any{"user": user.Id},
		)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}

		var result []map[string]any
		for _, r := range records {
			result = append(result, map[string]any{
				"id":         r.Id,
				"name":       r.GetString("name"),
				"query":      r.GetString("query"),
				"filters":    r.Get("filters"),
				"created_at": r.GetString("created"),
			})
		}
		if result == nil {
			result = []map[string]any{}
		}

		return e.JSON(http.StatusOK, result)
	}
}

// CreateSavedSearch handles POST /me/saved-searches
func CreateSavedSearch(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		data := struct {
			Name    string         `json:"name"`
			Query   string         `json:"query"`
			Filters map[string]any `json:"filters"`
		}{}
		if err := e.BindBody(&data); err != nil || data.Name == "" || data.Query == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "name and query are required"})
		}

		if len(data.Name) > 100 {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "name must be 100 characters or fewer"})
		}

		// Check limit: max 20 saved searches per user
		type countResult struct {
			Count int `db:"count"`
		}
		var cnt countResult
		_ = app.DB().NewQuery("SELECT COUNT(*) as count FROM saved_searches WHERE user = {:user}").
			Bind(map[string]any{"user": user.Id}).One(&cnt)
		if cnt.Count >= 20 {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "maximum of 20 saved searches reached"})
		}

		coll, err := app.FindCollectionByNameOrId("saved_searches")
		if err != nil {
			return err
		}
		rec := core.NewRecord(coll)
		rec.Set("user", user.Id)
		rec.Set("name", data.Name)
		rec.Set("query", data.Query)
		if data.Filters != nil {
			rec.Set("filters", data.Filters)
		}
		if err := app.Save(rec); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "failed to save search"})
		}

		return e.JSON(http.StatusCreated, map[string]any{
			"id":         rec.Id,
			"name":       rec.GetString("name"),
			"query":      rec.GetString("query"),
			"filters":    rec.Get("filters"),
			"created_at": rec.GetString("created"),
		})
	}
}

// DeleteSavedSearch handles DELETE /me/saved-searches/:id
func DeleteSavedSearch(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		searchID := e.Request.PathValue("id")
		rec, err := app.FindRecordById("saved_searches", searchID)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Saved search not found"})
		}
		if rec.GetString("user") != user.Id {
			return e.JSON(http.StatusForbidden, map[string]any{"error": "Not your saved search"})
		}

		if err := app.Delete(rec); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "failed to delete"})
		}

		return e.JSON(http.StatusOK, map[string]any{"ok": true})
	}
}
