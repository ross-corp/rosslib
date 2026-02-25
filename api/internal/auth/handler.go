package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tristansaldanha/rosslib/api/internal/email"
	"golang.org/x/crypto/bcrypt"
)

var usernameRe = regexp.MustCompile(`^[a-z0-9-]{1,40}$`)

type Handler struct {
	pool        *pgxpool.Pool
	jwtSecret   []byte
	emailClient *email.Client
	webappURL   string
}

func NewHandler(pool *pgxpool.Pool, jwtSecret string, emailClient *email.Client, webappURL string) *Handler {
	return &Handler{pool: pool, jwtSecret: []byte(jwtSecret), emailClient: emailClient, webappURL: webappURL}
}

type registerRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type loginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type authResponse struct {
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Username = strings.ToLower(strings.TrimSpace(req.Username))
	if !usernameRe.MatchString(req.Username) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username may only contain lowercase letters, numbers, and hyphens"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	tx, err := h.pool.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer tx.Rollback(c.Request.Context()) //nolint:errcheck

	emailAddr := strings.ToLower(strings.TrimSpace(req.Email))

	var userID, username string
	var isModerator bool
	err = tx.QueryRow(c.Request.Context(),
		`INSERT INTO users (username, email, password_hash, email_verified)
		 VALUES ($1, $2, $3, false)
		 RETURNING id, username, is_moderator`,
		req.Username, emailAddr, string(hash),
	).Scan(&userID, &username, &isModerator)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			c.JSON(http.StatusConflict, gin.H{"error": "email or username already taken"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	defaultShelves := []struct{ name, slug string }{
		{"Want to Read", "want-to-read"},
		{"Currently Reading", "currently-reading"},
		{"Read", "read"},
	}
	for _, s := range defaultShelves {
		_, err = tx.Exec(c.Request.Context(),
			`INSERT INTO collections (user_id, name, slug, is_exclusive, exclusive_group)
			 VALUES ($1, $2, $3, true, 'read_status')`,
			userID, s.name, s.slug,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
	}

	if err = tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// Send verification email (fire-and-forget).
	go h.sendVerificationEmail(userID, emailAddr)

	token, err := h.signToken(userID, username, isModerator, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusCreated, authResponse{Token: token, UserID: userID, Username: username})
}

func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userID, username string
	var hash sql.NullString
	var isModerator, emailVerified bool
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT id, username, password_hash, is_moderator, email_verified FROM users WHERE email = $1 AND deleted_at IS NULL`,
		strings.ToLower(strings.TrimSpace(req.Email)),
	).Scan(&userID, &username, &hash, &isModerator, &emailVerified)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	if !hash.Valid || hash.String == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "this account uses Google sign-in"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash.String), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	token, err := h.signToken(userID, username, isModerator, emailVerified)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, authResponse{Token: token, UserID: userID, Username: username})
}

type googleLoginRequest struct {
	GoogleID string `json:"google_id" binding:"required"`
	Email    string `json:"email"     binding:"required,email"`
	Name     string `json:"name"`
}

func (h *Handler) GoogleLogin(c *gin.Context) {
	var req googleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Try to find existing user by google_id.
	var userID, username string
	var isModerator, emailVerified bool
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT id, username, is_moderator, email_verified FROM users WHERE google_id = $1 AND deleted_at IS NULL`,
		req.GoogleID,
	).Scan(&userID, &username, &isModerator, &emailVerified)

	if err == nil {
		// Existing Google user — issue token.
		token, err := h.signToken(userID, username, isModerator, emailVerified)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		c.JSON(http.StatusOK, authResponse{Token: token, UserID: userID, Username: username})
		return
	}

	// Check if a user with this email already exists (link Google to existing account).
	err = h.pool.QueryRow(c.Request.Context(),
		`SELECT id, username, is_moderator, email_verified FROM users WHERE email = $1 AND deleted_at IS NULL`,
		req.Email,
	).Scan(&userID, &username, &isModerator, &emailVerified)

	if err == nil {
		// Link Google ID to existing email-based account and mark email as verified.
		_, err = h.pool.Exec(c.Request.Context(),
			`UPDATE users SET google_id = $1, email_verified = true WHERE id = $2`, req.GoogleID, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		token, err := h.signToken(userID, username, isModerator, true)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		c.JSON(http.StatusOK, authResponse{Token: token, UserID: userID, Username: username})
		return
	}

	// New user — create account. Derive username from email prefix.
	baseUsername := strings.ToLower(strings.Split(req.Email, "@")[0])
	baseUsername = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(baseUsername, "")
	if len(baseUsername) > 30 {
		baseUsername = baseUsername[:30]
	}
	if baseUsername == "" {
		baseUsername = "user"
	}

	tx, err := h.pool.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer tx.Rollback(c.Request.Context()) //nolint:errcheck

	// Try the base username, then append numbers if taken.
	candidateUsername := baseUsername
	for i := 1; ; i++ {
		err = tx.QueryRow(c.Request.Context(),
			`INSERT INTO users (username, email, google_id, display_name, email_verified)
			 VALUES ($1, $2, $3, $4, true)
			 RETURNING id, username, is_moderator`,
			candidateUsername, req.Email, req.GoogleID, req.Name,
		).Scan(&userID, &username, &isModerator)
		if err == nil {
			break
		}
		if !strings.Contains(err.Error(), "unique") {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		// Username or email conflict — if email conflict, the email is already used.
		if strings.Contains(err.Error(), "users_email_key") {
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}
		candidateUsername = fmt.Sprintf("%s-%d", baseUsername, i)
		if i > 100 {
			c.JSON(http.StatusConflict, gin.H{"error": "could not generate unique username"})
			return
		}
	}

	// Create default shelves.
	defaultShelves := []struct{ name, slug string }{
		{"Want to Read", "want-to-read"},
		{"Currently Reading", "currently-reading"},
		{"Read", "read"},
	}
	for _, s := range defaultShelves {
		_, err = tx.Exec(c.Request.Context(),
			`INSERT INTO collections (user_id, name, slug, is_exclusive, exclusive_group)
			 VALUES ($1, $2, $3, true, 'read_status')`,
			userID, s.name, s.slug,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
	}

	if err = tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	token, err := h.signToken(userID, username, isModerator, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusCreated, authResponse{Token: token, UserID: userID, Username: username})
}

// GetAccountInfo - GET /me/account (authed)
// Returns whether the user has a password and/or Google linked.
func (h *Handler) GetAccountInfo(c *gin.Context) {
	userID := c.GetString("user_id")

	var hasPassword bool
	var hasGoogle bool
	var hash sql.NullString
	var googleID sql.NullString
	var emailVerified bool
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT password_hash, google_id, email_verified FROM users WHERE id = $1 AND deleted_at IS NULL`,
		userID,
	).Scan(&hash, &googleID, &emailVerified)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	hasPassword = hash.Valid && hash.String != ""
	hasGoogle = googleID.Valid && googleID.String != ""

	c.JSON(http.StatusOK, gin.H{
		"has_password":   hasPassword,
		"has_google":     hasGoogle,
		"email_verified": emailVerified,
	})
}

// SetPassword - PUT /me/password (authed)
// Sets or changes the user's password.
// If the user already has a password, current_password is required.
// If the user is Google-only (no password), only new_password is required.
func (h *Handler) SetPassword(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "new password must be at least 8 characters"})
		return
	}

	var hash sql.NullString
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT password_hash FROM users WHERE id = $1 AND deleted_at IS NULL`,
		userID,
	).Scan(&hash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// If user already has a password, verify the current one.
	if hash.Valid && hash.String != "" {
		if req.CurrentPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "current password is required"})
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(hash.String), []byte(req.CurrentPassword)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "current password is incorrect"})
			return
		}
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	_, err = h.pool.Exec(c.Request.Context(),
		`UPDATE users SET password_hash = $1 WHERE id = $2`,
		string(newHash), userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ForgotPassword - POST /auth/forgot-password
// Generates a password reset token and emails it to the user.
// Always returns 200 to avoid leaking whether an email exists.
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valid email is required"})
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Always return success to avoid leaking whether an email exists.
	okMsg := gin.H{"ok": true, "message": "If an account with that email exists, a password reset link has been sent."}

	if h.emailClient == nil {
		log.Printf("password reset requested for %s but SMTP is not configured", req.Email)
		c.JSON(http.StatusOK, okMsg)
		return
	}

	// Look up user by email.
	var userID string
	var hasPassword bool
	var hash sql.NullString
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT id, password_hash FROM users WHERE email = $1 AND deleted_at IS NULL`,
		req.Email,
	).Scan(&userID, &hash)
	if err != nil {
		// User not found — return success anyway.
		c.JSON(http.StatusOK, okMsg)
		return
	}
	hasPassword = hash.Valid && hash.String != ""
	_ = hasPassword // Reset works for all accounts (including Google-only, to set a password).

	// Invalidate any existing unused tokens for this user.
	_, _ = h.pool.Exec(c.Request.Context(),
		`UPDATE password_reset_tokens SET used = true WHERE user_id = $1 AND used = false`,
		userID,
	)

	// Generate a random token.
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	rawToken := hex.EncodeToString(tokenBytes)

	// Store the hash of the token (not the raw token).
	tokenHash := sha256.Sum256([]byte(rawToken))
	tokenHashHex := hex.EncodeToString(tokenHash[:])

	expiresAt := time.Now().Add(1 * time.Hour)
	_, err = h.pool.Exec(c.Request.Context(),
		`INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		userID, tokenHashHex, expiresAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", h.webappURL, rawToken)
	if err := h.emailClient.SendPasswordReset(req.Email, resetURL); err != nil {
		log.Printf("failed to send password reset email to %s: %v", req.Email, err)
		// Still return success to avoid leaking information.
	}

	c.JSON(http.StatusOK, okMsg)
}

// ResetPassword - POST /auth/reset-password
// Validates the reset token and sets a new password.
func (h *Handler) ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token"        binding:"required"`
		NewPassword string `json:"new_password"  binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token and new password (min 8 chars) are required"})
		return
	}

	// Hash the provided token to look it up.
	tokenHash := sha256.Sum256([]byte(req.Token))
	tokenHashHex := hex.EncodeToString(tokenHash[:])

	var tokenID, userID string
	var expiresAt time.Time
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT id, user_id, expires_at FROM password_reset_tokens
		 WHERE token_hash = $1 AND used = false`,
		tokenHashHex,
	).Scan(&tokenID, &userID, &expiresAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired reset token"})
		return
	}

	if time.Now().After(expiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired reset token"})
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// Update password and mark token as used in a transaction.
	tx, err := h.pool.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer tx.Rollback(c.Request.Context()) //nolint:errcheck

	_, err = tx.Exec(c.Request.Context(),
		`UPDATE users SET password_hash = $1 WHERE id = $2`,
		string(newHash), userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	_, err = tx.Exec(c.Request.Context(),
		`UPDATE password_reset_tokens SET used = true WHERE id = $1`,
		tokenID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if err = tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) signToken(userID, username string, isModerator, emailVerified bool) (string, error) {
	claims := jwt.MapClaims{
		"sub":            userID,
		"username":       username,
		"is_moderator":   isModerator,
		"email_verified": emailVerified,
		"exp":            time.Now().Add(30 * 24 * time.Hour).Unix(),
		"iat":            time.Now().Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(h.jwtSecret)
}

// sendVerificationEmail generates a verification token and sends the email.
// Intended to be called as a goroutine (fire-and-forget).
func (h *Handler) sendVerificationEmail(userID, emailAddr string) {
	if h.emailClient == nil {
		log.Printf("verification email for %s skipped — SMTP not configured", emailAddr)
		return
	}

	ctx := context.Background()

	// Invalidate any existing unused tokens.
	_, _ = h.pool.Exec(ctx,
		`UPDATE email_verification_tokens SET used = true WHERE user_id = $1 AND used = false`, userID)

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		log.Printf("failed to generate verification token for %s: %v", emailAddr, err)
		return
	}
	rawToken := hex.EncodeToString(tokenBytes)

	tokenHash := sha256.Sum256([]byte(rawToken))
	tokenHashHex := hex.EncodeToString(tokenHash[:])

	expiresAt := time.Now().Add(24 * time.Hour)
	_, err := h.pool.Exec(ctx,
		`INSERT INTO email_verification_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		userID, tokenHashHex, expiresAt,
	)
	if err != nil {
		log.Printf("failed to store verification token for %s: %v", emailAddr, err)
		return
	}

	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", h.webappURL, rawToken)
	if err := h.emailClient.SendVerification(emailAddr, verifyURL); err != nil {
		log.Printf("failed to send verification email to %s: %v", emailAddr, err)
	}
}

// VerifyEmail - POST /auth/verify-email
// Validates the verification token and marks the user's email as verified.
func (h *Handler) VerifyEmail(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	tokenHash := sha256.Sum256([]byte(req.Token))
	tokenHashHex := hex.EncodeToString(tokenHash[:])

	var tokenID, userID string
	var expiresAt time.Time
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT id, user_id, expires_at FROM email_verification_tokens
		 WHERE token_hash = $1 AND used = false`,
		tokenHashHex,
	).Scan(&tokenID, &userID, &expiresAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired verification link"})
		return
	}

	if time.Now().After(expiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired verification link"})
		return
	}

	tx, err := h.pool.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer tx.Rollback(c.Request.Context()) //nolint:errcheck

	_, err = tx.Exec(c.Request.Context(),
		`UPDATE users SET email_verified = true WHERE id = $1`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	_, err = tx.Exec(c.Request.Context(),
		`UPDATE email_verification_tokens SET used = true WHERE id = $1`, tokenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if err = tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// Issue a fresh token with email_verified = true.
	var username string
	var isModerator bool
	err = h.pool.QueryRow(c.Request.Context(),
		`SELECT username, is_moderator FROM users WHERE id = $1`, userID,
	).Scan(&username, &isModerator)
	if err != nil {
		// Verified, but can't issue token — they'll get one on next login.
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	token, err := h.signToken(userID, username, isModerator, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	c.JSON(http.StatusOK, authResponse{Token: token, UserID: userID, Username: username})
}

// ResendVerification - POST /auth/resend-verification (authed)
// Generates a new verification token and sends the email.
func (h *Handler) ResendVerification(c *gin.Context) {
	userID := c.GetString("user_id")

	var emailAddr string
	var emailVerified bool
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT email, email_verified FROM users WHERE id = $1 AND deleted_at IS NULL`,
		userID,
	).Scan(&emailAddr, &emailVerified)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if emailVerified {
		c.JSON(http.StatusOK, gin.H{"ok": true, "message": "Email is already verified."})
		return
	}

	go h.sendVerificationEmail(userID, emailAddr)

	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "Verification email sent."})
}
