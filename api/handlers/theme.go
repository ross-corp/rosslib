package handlers

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

var validThemes = map[string]bool{
	"light":  true,
	"dark":   true,
	"system": true,
}

// GetTheme handles GET /me/theme
func GetTheme(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		theme := user.GetString("theme")
		if theme == "" {
			theme = "system"
		}

		return e.JSON(http.StatusOK, map[string]any{"theme": theme})
	}
}

// UpdateTheme handles PUT /me/theme
func UpdateTheme(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		data := struct {
			Theme string `json:"theme"`
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}

		if !validThemes[data.Theme] {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Theme must be light, dark, or system"})
		}

		user.Set("theme", data.Theme)
		if err := app.Save(user); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to save theme"})
		}

		return e.JSON(http.StatusOK, map[string]any{"theme": data.Theme})
	}
}
