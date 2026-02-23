package users

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tristansaldanha/rosslib/api/internal/middleware"
)

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// ── types ────────────────────────────────────────────────────────────────────

type profileResponse struct {
	UserID         string  `json:"user_id"`
	Username       string  `json:"username"`
	DisplayName    *string `json:"display_name"`
	Bio            *string `json:"bio"`
	AvatarURL      *string `json:"avatar_url"`
	IsPrivate      bool    `json:"is_private"`
	MemberSince    string  `json:"member_since"`
	IsFollowing    bool    `json:"is_following"`
	FollowersCount int     `json:"followers_count"`
	FollowingCount int     `json:"following_count"`
	FriendsCount   int     `json:"friends_count"`
	BooksRead      int     `json:"books_read"`
}

type searchResult struct {
	UserID      string  `json:"user_id"`
	Username    string  `json:"username"`
	DisplayName *string `json:"display_name"`
}

type listResponse struct {
	Users   []searchResult `json:"users"`
	Page    int            `json:"page"`
	HasNext bool           `json:"has_next"`
}

// ── handlers ─────────────────────────────────────────────────────────────────

const pageSize = 20

func (h *Handler) SearchUsers(c *gin.Context) {
	q := c.Query("q")

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
	currentUserID := c.GetString(middleware.UserIDKey)

	var p profileResponse
	var memberSince time.Time

	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT u.id, u.username, u.display_name, u.bio, u.avatar_url, u.is_private, u.created_at,
		        (SELECT COUNT(*) FROM follows WHERE followee_id = u.id) AS followers_count,
		        (SELECT COUNT(*) FROM follows WHERE follower_id = u.id) AS following_count,
		        (SELECT COUNT(*) FROM follows f1
		           JOIN follows f2 ON f1.follower_id = f2.followee_id AND f1.followee_id = f2.follower_id
		           WHERE f1.follower_id = u.id) AS friends_count,
		        (SELECT COUNT(*) FROM collection_items ci
		           JOIN collections c ON c.id = ci.collection_id
		           WHERE c.user_id = u.id AND c.slug = 'read') AS books_read
		 FROM users u WHERE u.username = $1 AND u.deleted_at IS NULL`,
		username,
	).Scan(&p.UserID, &p.Username, &p.DisplayName, &p.Bio, &p.AvatarURL, &p.IsPrivate, &memberSince, &p.FollowersCount, &p.FollowingCount, &p.FriendsCount, &p.BooksRead)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	p.MemberSince = memberSince.Format(time.RFC3339)

	if currentUserID != "" && currentUserID != p.UserID {
		var isFollowing bool
		_ = h.pool.QueryRow(c.Request.Context(),
			`SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND followee_id = $2)`,
			currentUserID, p.UserID,
		).Scan(&isFollowing)
		p.IsFollowing = isFollowing
	}

	c.JSON(http.StatusOK, p)
}

func (h *Handler) UpdateMe(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	var req struct {
		DisplayName string `json:"display_name"`
		Bio         string `json:"bio"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var displayName, bio *string
	if v := strings.TrimSpace(req.DisplayName); v != "" {
		displayName = &v
	}
	if v := strings.TrimSpace(req.Bio); v != "" {
		bio = &v
	}

	_, err := h.pool.Exec(c.Request.Context(),
		`UPDATE users SET display_name = $1, bio = $2 WHERE id = $3`,
		displayName, bio, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) Follow(c *gin.Context) {
	followerID := c.GetString(middleware.UserIDKey)
	username := c.Param("username")

	var followeeID string
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT id FROM users WHERE username = $1 AND deleted_at IS NULL`,
		username,
	).Scan(&followeeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if followerID == followeeID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot follow yourself"})
		return
	}

	_, err = h.pool.Exec(c.Request.Context(),
		`INSERT INTO follows (follower_id, followee_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		followerID, followeeID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) Unfollow(c *gin.Context) {
	followerID := c.GetString(middleware.UserIDKey)
	username := c.Param("username")

	var followeeID string
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT id FROM users WHERE username = $1 AND deleted_at IS NULL`,
		username,
	).Scan(&followeeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	_, err = h.pool.Exec(c.Request.Context(),
		`DELETE FROM follows WHERE follower_id = $1 AND followee_id = $2`,
		followerID, followeeID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
