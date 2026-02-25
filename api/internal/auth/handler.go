package auth

import (
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var usernameRe = regexp.MustCompile(`^[a-z0-9-]{1,40}$`)

type Handler struct {
	pool      *pgxpool.Pool
	jwtSecret []byte
}

func NewHandler(pool *pgxpool.Pool, jwtSecret string) *Handler {
	return &Handler{pool: pool, jwtSecret: []byte(jwtSecret)}
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

	var userID, username string
	var isModerator bool
	err = tx.QueryRow(c.Request.Context(),
		`INSERT INTO users (username, email, password_hash)
		 VALUES ($1, $2, $3)
		 RETURNING id, username, is_moderator`,
		req.Username, strings.ToLower(strings.TrimSpace(req.Email)), string(hash),
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

	token, err := h.signToken(userID, username, isModerator)
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
	var isModerator bool
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT id, username, password_hash, is_moderator FROM users WHERE email = $1 AND deleted_at IS NULL`,
		strings.ToLower(strings.TrimSpace(req.Email)),
	).Scan(&userID, &username, &hash, &isModerator)
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

	token, err := h.signToken(userID, username, isModerator)
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
	var isModerator bool
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT id, username, is_moderator FROM users WHERE google_id = $1 AND deleted_at IS NULL`,
		req.GoogleID,
	).Scan(&userID, &username, &isModerator)

	if err == nil {
		// Existing Google user — issue token.
		token, err := h.signToken(userID, username, isModerator)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		c.JSON(http.StatusOK, authResponse{Token: token, UserID: userID, Username: username})
		return
	}

	// Check if a user with this email already exists (link Google to existing account).
	err = h.pool.QueryRow(c.Request.Context(),
		`SELECT id, username, is_moderator FROM users WHERE email = $1 AND deleted_at IS NULL`,
		req.Email,
	).Scan(&userID, &username, &isModerator)

	if err == nil {
		// Link Google ID to existing email-based account.
		_, err = h.pool.Exec(c.Request.Context(),
			`UPDATE users SET google_id = $1 WHERE id = $2`, req.GoogleID, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		token, err := h.signToken(userID, username, isModerator)
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
			`INSERT INTO users (username, email, google_id, display_name)
			 VALUES ($1, $2, $3, $4)
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

	token, err := h.signToken(userID, username, isModerator)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusCreated, authResponse{Token: token, UserID: userID, Username: username})
}

func (h *Handler) signToken(userID, username string, isModerator bool) (string, error) {
	claims := jwt.MapClaims{
		"sub":          userID,
		"username":     username,
		"is_moderator": isModerator,
		"exp":          time.Now().Add(30 * 24 * time.Hour).Unix(),
		"iat":          time.Now().Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(h.jwtSecret)
}
