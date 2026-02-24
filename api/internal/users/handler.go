package users

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tristansaldanha/rosslib/api/internal/middleware"
	"github.com/tristansaldanha/rosslib/api/internal/storage"
)

type Handler struct {
	pool  *pgxpool.Pool
	store *storage.Client
}

func NewHandler(pool *pgxpool.Pool, store *storage.Client) *Handler {
	return &Handler{pool: pool, store: store}
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

// GetUserReviews - GET /users/:username/reviews
// Public. Returns all collection items for the user that have review text.
func (h *Handler) GetUserReviews(c *gin.Context) {
	username := c.Param("username")

	rows, err := h.pool.Query(c.Request.Context(),
		`SELECT b.id, b.open_library_id, b.title, b.cover_url, b.authors,
		        ci.rating, ci.review_text, ci.spoiler, ci.date_read, ci.date_added
		 FROM collection_items ci
		 JOIN books b       ON b.id  = ci.book_id
		 JOIN collections c ON c.id  = ci.collection_id
		 JOIN users u       ON u.id  = c.user_id
		 WHERE u.username = $1
		   AND u.deleted_at IS NULL
		   AND ci.review_text IS NOT NULL
		   AND ci.review_text != ''
		 ORDER BY ci.date_added DESC`,
		username,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer rows.Close()

	type reviewItem struct {
		BookID        string  `json:"book_id"`
		OpenLibraryID string  `json:"open_library_id"`
		Title         string  `json:"title"`
		CoverURL      *string `json:"cover_url"`
		Authors       *string `json:"authors"`
		Rating        *int    `json:"rating"`
		ReviewText    string  `json:"review_text"`
		Spoiler       bool    `json:"spoiler"`
		DateRead      *string `json:"date_read"`
		DateAdded     string  `json:"date_added"`
	}

	reviews := []reviewItem{}
	for rows.Next() {
		var r reviewItem
		var dateRead *time.Time
		var dateAdded time.Time
		if err := rows.Scan(
			&r.BookID, &r.OpenLibraryID, &r.Title, &r.CoverURL, &r.Authors,
			&r.Rating, &r.ReviewText, &r.Spoiler, &dateRead, &dateAdded,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		r.DateAdded = dateAdded.Format(time.RFC3339)
		if dateRead != nil {
			s := dateRead.Format(time.RFC3339)
			r.DateRead = &s
		}
		reviews = append(reviews, r)
	}

	c.JSON(http.StatusOK, reviews)
}

func (h *Handler) UploadAvatar(c *gin.Context) {
	if h.store == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "storage not configured"})
		return
	}

	userID := c.GetString(middleware.UserIDKey)

	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not parse form"})
		return
	}

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing avatar file"})
		return
	}
	defer file.Close()

	if header.Size > 5<<20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large (max 5 MB)"})
		return
	}

	url, err := h.store.UploadAvatar(c.Request.Context(), userID, file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = h.pool.Exec(c.Request.Context(),
		`UPDATE users SET avatar_url = $1 WHERE id = $2`,
		url, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"avatar_url": url})
}
