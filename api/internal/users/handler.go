package users

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

type profileResponse struct {
	UserID      string  `json:"user_id"`
	Username    string  `json:"username"`
	DisplayName *string `json:"display_name"`
	Bio         *string `json:"bio"`
	AvatarURL   *string `json:"avatar_url"`
	IsPrivate   bool    `json:"is_private"`
	MemberSince string  `json:"member_since"`
}

type searchResult struct {
	UserID      string  `json:"user_id"`
	Username    string  `json:"username"`
	DisplayName *string `json:"display_name"`
}

const pageSize = 20

type listResponse struct {
	Users   []searchResult `json:"users"`
	Page    int            `json:"page"`
	HasNext bool           `json:"has_next"`
}

func (h *Handler) SearchUsers(c *gin.Context) {
	q := c.Query("q")

	// No query — return paginated list of all users.
	if q == "" {
		page := 1
		if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 1 {
			page = p
		}
		offset := (page - 1) * pageSize

		rows, err := h.pool.Query(c.Request.Context(),
			`SELECT id, username, display_name
			 FROM users
			 WHERE deleted_at IS NULL
			 ORDER BY username
			 LIMIT $1 OFFSET $2`,
			pageSize+1, offset,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		defer rows.Close()

		users := []searchResult{}
		for rows.Next() {
			var r searchResult
			if err := rows.Scan(&r.UserID, &r.Username, &r.DisplayName); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
				return
			}
			users = append(users, r)
		}

		hasNext := len(users) > pageSize
		if hasNext {
			users = users[:pageSize]
		}

		c.JSON(http.StatusOK, listResponse{Users: users, Page: page, HasNext: hasNext})
		return
	}

	// Query present — search mode.
	rows, err := h.pool.Query(c.Request.Context(),
		`SELECT id, username, display_name
		 FROM users
		 WHERE deleted_at IS NULL
		   AND (username ILIKE '%' || $1 || '%' OR display_name ILIKE '%' || $1 || '%')
		 ORDER BY username
		 LIMIT 20`,
		q,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer rows.Close()

	results := []searchResult{}
	for rows.Next() {
		var r searchResult
		if err := rows.Scan(&r.UserID, &r.Username, &r.DisplayName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		results = append(results, r)
	}

	c.JSON(http.StatusOK, results)
}

func (h *Handler) GetProfile(c *gin.Context) {
	username := c.Param("username")

	var p profileResponse
	var memberSince time.Time

	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT id, username, display_name, bio, avatar_url, is_private, created_at
		 FROM users WHERE username = $1 AND deleted_at IS NULL`,
		username,
	).Scan(&p.UserID, &p.Username, &p.DisplayName, &p.Bio, &p.AvatarURL, &p.IsPrivate, &memberSince)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	p.MemberSince = memberSince.Format(time.RFC3339)
	c.JSON(http.StatusOK, p)
}
