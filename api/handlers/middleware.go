package handlers

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

// optionalAuthMiddleware is a custom implementation that doesn't reject unauthenticated requests.
func optionalAuthMiddleware(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		token := e.Request.Header.Get("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
			record, err := app.FindAuthRecordByToken(token, core.TokenTypeAuth)
			if err == nil && record != nil {
				e.Auth = record
			}
		}
		return e.Next()
	}
}

// OptionalAuthFunc returns a hook function for optional auth (use with .BindFunc).
func OptionalAuthFunc(app core.App) func(e *core.RequestEvent) error {
	return optionalAuthMiddleware(app)
}

// RequireModerator checks that the authenticated user has is_moderator = true.
func RequireModerator(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if e.Auth == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		if !e.Auth.GetBool("is_moderator") {
			return e.JSON(http.StatusForbidden, map[string]any{"error": "Moderator access required"})
		}
		return e.Next()
	}
}
