package handlers

import (
	"net/http"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// Login handles POST /auth/login.
func Login(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		data := struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}

		record, err := app.FindAuthRecordByEmail("users", data.Email)
		if err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid credentials"})
		}
		if !record.ValidatePassword(data.Password) {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid credentials"})
		}

		token, err := record.NewAuthToken()
		if err != nil {
			return apis.NewBadRequestError("Failed to create token", err)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"token":    token,
			"user_id":  record.Id,
			"username": record.GetString("username"),
		})
	}
}

// Register handles POST /auth/register.
func Register(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		data := struct {
			Username        string `json:"username"`
			Email           string `json:"email"`
			Password        string `json:"password"`
			PasswordConfirm string `json:"passwordConfirm"`
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}

		if data.Username == "" || data.Email == "" || data.Password == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Username, email, and password are required"})
		}

		collection, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		record := core.NewRecord(collection)
		record.Set("username", data.Username)
		record.Set("email", data.Email)
		record.SetPassword(data.Password)

		if err := app.Save(record); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}

		// Create default shelves
		if err := createDefaultShelves(app, record.Id); err != nil {
			// Non-fatal: user is created but shelves failed
			_ = err
		}

		// Create default Status tag key + values
		if _, _, err := ensureStatusTagKey(app, record.Id); err != nil {
			_ = err
		}

		token, err := record.NewAuthToken()
		if err != nil {
			return apis.NewBadRequestError("Failed to create token", err)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"token":    token,
			"user_id":  record.Id,
			"username": record.GetString("username"),
		})
	}
}

// GetAccount handles GET /me/account.
func GetAccount(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		return e.JSON(http.StatusOK, map[string]any{
			"user_id":        user.Id,
			"username":       user.GetString("username"),
			"email":          user.Email(),
			"display_name":   user.GetString("display_name"),
			"has_password":   user.ValidatePassword("") == false, // always true if password is set
			"email_verified": user.GetBool("email_verified"),
			"is_moderator":   user.GetBool("is_moderator"),
		})
	}
}

// ChangePassword handles PUT /me/password.
func ChangePassword(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		data := struct {
			OldPassword string `json:"old_password"`
			NewPassword string `json:"new_password"`
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}

		if !user.ValidatePassword(data.OldPassword) {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Current password is incorrect"})
		}

		user.SetPassword(data.NewPassword)
		if err := app.Save(user); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}

		return e.JSON(http.StatusOK, map[string]any{"message": "Password updated"})
	}
}
