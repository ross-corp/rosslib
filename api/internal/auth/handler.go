package auth

import (
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

	var userID, username string
	err = h.pool.QueryRow(c.Request.Context(),
		`INSERT INTO users (username, email, password_hash)
		 VALUES ($1, $2, $3)
		 RETURNING id, username`,
		req.Username, strings.ToLower(strings.TrimSpace(req.Email)), string(hash),
	).Scan(&userID, &username)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			c.JSON(http.StatusConflict, gin.H{"error": "email or username already taken"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	token, err := h.signToken(userID, username)
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

	var userID, username, hash string
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT id, username, password_hash FROM users WHERE email = $1 AND deleted_at IS NULL`,
		strings.ToLower(strings.TrimSpace(req.Email)),
	).Scan(&userID, &username, &hash)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	token, err := h.signToken(userID, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, authResponse{Token: token, UserID: userID, Username: username})
}

func (h *Handler) signToken(userID, username string) (string, error) {
	claims := jwt.MapClaims{
		"sub":      userID,
		"username": username,
		"exp":      time.Now().Add(30 * 24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(h.jwtSecret)
}
