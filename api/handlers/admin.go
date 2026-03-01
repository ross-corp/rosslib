package handlers

import (
	"net/http"
	"strconv"

	"github.com/pocketbase/pocketbase/core"
)

// GetAdminUsers handles GET /admin/users?q=<query>&page=<n>
func GetAdminUsers(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		q := e.Request.URL.Query().Get("q")
		page, _ := strconv.Atoi(e.Request.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		perPage := 20
		offset := (page - 1) * perPage

		filter := "1=1"
		params := map[string]any{}
		if q != "" {
			filter = "username LIKE {:q} || display_name LIKE {:q} || email LIKE {:q}"
			params["q"] = "%" + q + "%"
		}

		records, err := app.FindRecordsByFilter("users", filter, "-created", perPage+1, offset, params)
		if err != nil {
			return e.JSON(http.StatusOK, map[string]any{
				"users":    []any{},
				"page":     page,
				"has_next": false,
			})
		}

		hasNext := len(records) > perPage
		if hasNext {
			records = records[:perPage]
		}

		users := make([]map[string]any, 0, len(records))
		for _, r := range records {
			users = append(users, map[string]any{
				"user_id":      r.Id,
				"username":     r.GetString("username"),
				"display_name": nilIfEmpty(r.GetString("display_name")),
				"email":        r.GetString("email"),
				"is_moderator": r.GetBool("is_moderator"),
				"author_key":   nilIfEmpty(r.GetString("author_key")),
			})
		}

		return e.JSON(http.StatusOK, map[string]any{
			"users":    users,
			"page":     page,
			"has_next": hasNext,
		})
	}
}

// SetModerator handles PUT /admin/users/{userId}/moderator
func SetModerator(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		userId := e.Request.PathValue("userId")

		var body struct {
			IsModerator bool `json:"is_moderator"`
		}
		if err := e.BindBody(&body); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "invalid request body"})
		}

		record, err := app.FindRecordById("users", userId)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "user not found"})
		}

		record.Set("is_moderator", body.IsModerator)
		if err := app.Save(record); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "failed to update user"})
		}

		return e.JSON(http.StatusOK, map[string]any{
			"ok":           true,
			"is_moderator": body.IsModerator,
		})
	}
}

// SetAuthorKey handles PUT /admin/users/{userId}/author
func SetAuthorKey(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		userId := e.Request.PathValue("userId")

		var body struct {
			AuthorKey *string `json:"author_key"`
		}
		if err := e.BindBody(&body); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "invalid request body"})
		}

		record, err := app.FindRecordById("users", userId)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "user not found"})
		}

		authorKey := ""
		if body.AuthorKey != nil {
			authorKey = *body.AuthorKey
		}
		record.Set("author_key", authorKey)
		if err := app.Save(record); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "failed to update user"})
		}

		return e.JSON(http.StatusOK, map[string]any{
			"ok":         true,
			"author_key": nilIfEmpty(authorKey),
		})
	}
}

func nilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
