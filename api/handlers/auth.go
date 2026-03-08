package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	mrand "math/rand/v2"
	"net/http"
	"strings"

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

// GoogleAuth handles POST /auth/google.
// It finds or creates a user from Google OAuth credentials.
func GoogleAuth(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		data := struct {
			GoogleID string `json:"google_id"`
			Email    string `json:"email"`
			Name     string `json:"name"`
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}

		if data.GoogleID == "" || data.Email == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "google_id and email are required"})
		}

		// Try to find existing user by google_id.
		records, _ := app.FindRecordsByFilter("users",
			"google_id = {:gid}", "", 1, 0,
			map[string]any{"gid": data.GoogleID},
		)
		if len(records) > 0 {
			return issueToken(e, records[0])
		}

		// Try to find existing user by email (link account).
		record, err := app.FindAuthRecordByEmail("users", data.Email)
		if err == nil {
			record.Set("google_id", data.GoogleID)
			if err := app.Save(record); err != nil {
				return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to link Google account"})
			}
			return issueToken(e, record)
		}

		// Create new user.
		collection, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		username := generateUsername(app, data.Email)

		newUser := core.NewRecord(collection)
		newUser.Set("username", username)
		newUser.Set("email", data.Email)
		newUser.Set("google_id", data.GoogleID)
		newUser.Set("email_verified", true)
		if data.Name != "" {
			newUser.Set("display_name", data.Name)
		}

		// PocketBase requires a password on auth records. Set a random one
		// that the user will never use (they authenticate via Google).
		randBytes := make([]byte, 32)
		if _, err := rand.Read(randBytes); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to generate credentials"})
		}
		newUser.SetPassword(hex.EncodeToString(randBytes))

		if err := app.Save(newUser); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}

		// Create default Status tag key + values
		if _, _, err := ensureStatusTagKey(app, newUser.Id); err != nil {
			_ = err
		}

		return issueToken(e, newUser)
	}
}

// issueToken generates an auth token and returns the standard auth response.
func issueToken(e *core.RequestEvent, record *core.Record) error {
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

// generateUsername creates a username from an email prefix, appending random
// digits if the base name is already taken.
func generateUsername(app core.App, email string) string {
	base := strings.Split(email, "@")[0]
	// Keep only alphanumeric and underscores.
	cleaned := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return -1
	}, base)
	if cleaned == "" {
		cleaned = "user"
	}

	candidate := cleaned
	for i := 0; i < 10; i++ {
		records, _ := app.FindRecordsByFilter("users",
			"username = {:u}", "", 1, 0,
			map[string]any{"u": candidate},
		)
		if len(records) == 0 {
			return candidate
		}
		candidate = fmt.Sprintf("%s%d", cleaned, mrand.IntN(10000))
	}
	return candidate
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
			"has_password":   !user.ValidatePassword(""),
			"email_verified": user.GetBool("email_verified"),
			"is_moderator":   user.GetBool("is_moderator"),
		})
	}
}

// ChangeEmail handles PUT /me/email.
func ChangeEmail(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		data := struct {
			NewEmail        string `json:"new_email"`
			CurrentPassword string `json:"current_password"`
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}

		if data.NewEmail == "" || data.CurrentPassword == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "New email and current password are required"})
		}

		if !strings.Contains(data.NewEmail, "@") {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid email address"})
		}

		if !user.ValidatePassword(data.CurrentPassword) {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Current password is incorrect"})
		}

		// Check if the new email is the same as the current one.
		if strings.EqualFold(user.Email(), data.NewEmail) {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "New email is the same as the current email"})
		}

		// Check if the new email is already in use.
		existing, _ := app.FindAuthRecordByEmail("users", data.NewEmail)
		if existing != nil {
			return e.JSON(http.StatusConflict, map[string]any{"error": "Email is already in use"})
		}

		user.Set("email", data.NewEmail)
		user.Set("email_verified", false)
		if err := app.Save(user); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}

		return e.JSON(http.StatusOK, map[string]any{"message": "Email updated"})
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
