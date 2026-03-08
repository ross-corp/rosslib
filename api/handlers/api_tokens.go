package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

const maxTokensPerUser = 5

// generateAPIToken creates a cryptographically random 32-byte token and returns it hex-encoded.
func generateAPIToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// hashToken returns the SHA-256 hex digest of a raw token string.
func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

// GetAPITokens handles GET /me/api-tokens.
func GetAPITokens(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		tokens, err := app.FindRecordsByFilter("api_tokens",
			"user = {:user}",
			"-created", 100, 0,
			map[string]any{"user": user.Id},
		)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to fetch tokens"})
		}

		result := make([]map[string]any, 0, len(tokens))
		for _, t := range tokens {
			item := map[string]any{
				"id":         t.Id,
				"name":       t.GetString("name"),
				"created":    t.GetString("created"),
				"last_used_at": t.GetString("last_used_at"),
			}
			result = append(result, item)
		}

		return e.JSON(http.StatusOK, map[string]any{"tokens": result})
	}
}

// CreateAPIToken handles POST /me/api-tokens.
func CreateAPIToken(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		data := struct {
			Name string `json:"name"`
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}
		if data.Name == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Token name is required"})
		}
		if len(data.Name) > 100 {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Token name must be 100 characters or less"})
		}

		// Check token limit
		existing, err := app.FindRecordsByFilter("api_tokens",
			"user = {:user}",
			"", 100, 0,
			map[string]any{"user": user.Id},
		)
		if err == nil && len(existing) >= maxTokensPerUser {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Maximum of 5 API tokens reached. Delete an existing token first."})
		}

		rawToken, err := generateAPIToken()
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to generate token"})
		}

		coll, err := app.FindCollectionByNameOrId("api_tokens")
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to find collection"})
		}

		record := core.NewRecord(coll)
		record.Set("user", user.Id)
		record.Set("name", data.Name)
		record.Set("token_hash", hashToken(rawToken))

		if err := app.Save(record); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to save token"})
		}

		return e.JSON(http.StatusOK, map[string]any{
			"id":    record.Id,
			"name":  data.Name,
			"token": rawToken,
		})
	}
}

// DeleteAPIToken handles DELETE /me/api-tokens/:id.
func DeleteAPIToken(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		tokenId := e.Request.PathValue("tokenId")
		if tokenId == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Token ID is required"})
		}

		record, err := app.FindRecordById("api_tokens", tokenId)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Token not found"})
		}

		if record.GetString("user") != user.Id {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Token not found"})
		}

		if err := app.Delete(record); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to delete token"})
		}

		return e.JSON(http.StatusOK, map[string]any{"message": "Token deleted"})
	}
}

// authenticateByAPIToken checks the Authorization header for an API token,
// hashes it, and looks it up in the api_tokens collection. If found, it sets
// e.Auth to the owning user and updates last_used_at.
func authenticateByAPIToken(app core.App, e *core.RequestEvent) bool {
	header := e.Request.Header.Get("Authorization")
	if len(header) <= 7 || header[:7] != "Bearer " {
		return false
	}
	raw := header[7:]

	h := hashToken(raw)

	records, err := app.FindRecordsByFilter("api_tokens",
		"token_hash = {:hash}",
		"", 1, 0,
		map[string]any{"hash": h},
	)
	if err != nil || len(records) == 0 {
		return false
	}

	tokenRecord := records[0]
	userID := tokenRecord.GetString("user")

	user, err := app.FindRecordById("users", userID)
	if err != nil {
		return false
	}

	e.Auth = user

	// Update last_used_at in the background
	go func() {
		tokenRecord.Set("last_used_at", time.Now().UTC().Format("2006-01-02 15:04:05.000Z"))
		_ = app.Save(tokenRecord)
	}()

	return true
}
