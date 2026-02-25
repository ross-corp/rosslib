package users

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tristansaldanha/rosslib/api/internal/activity"
	"github.com/tristansaldanha/rosslib/api/internal/middleware"
	"github.com/tristansaldanha/rosslib/api/internal/privacy"
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
	UserID         string   `json:"user_id"`
	Username       string   `json:"username"`
	DisplayName    *string  `json:"display_name"`
	Bio            *string  `json:"bio"`
	AvatarURL      *string  `json:"avatar_url"`
	IsPrivate      bool     `json:"is_private"`
	MemberSince    string   `json:"member_since"`
	IsFollowing    bool     `json:"is_following"`
	FollowStatus   string   `json:"follow_status"`
	FollowersCount int      `json:"followers_count"`
	FollowingCount int      `json:"following_count"`
	FriendsCount   int      `json:"friends_count"`
	BooksRead      int      `json:"books_read"`
	ReviewsCount   int      `json:"reviews_count"`
	BooksThisYear  int      `json:"books_this_year"`
	AverageRating  *float64 `json:"average_rating"`
	IsGhost        bool     `json:"is_ghost"`
	IsRestricted   bool     `json:"is_restricted"`
	AuthorKey      *string  `json:"author_key"`
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
		`SELECT u.id, u.username, u.display_name, u.bio, u.avatar_url, u.is_private, u.is_ghost, u.created_at, u.author_key,
		        (SELECT COUNT(*) FROM follows WHERE followee_id = u.id AND status = 'active') AS followers_count,
		        (SELECT COUNT(*) FROM follows WHERE follower_id = u.id AND status = 'active') AS following_count,
		        (SELECT COUNT(*) FROM follows f1
		           JOIN follows f2 ON f1.follower_id = f2.followee_id AND f1.followee_id = f2.follower_id
		           WHERE f1.follower_id = u.id AND f1.status = 'active' AND f2.status = 'active') AS friends_count,
		        (SELECT COUNT(*) FROM book_tag_values btv
		           JOIN tag_keys tk ON tk.id = btv.tag_key_id
		           JOIN tag_values tv ON tv.id = btv.tag_value_id
		           WHERE btv.user_id = u.id AND tk.slug = 'status' AND tv.slug = 'finished') AS books_read,
		        (SELECT COUNT(*) FROM user_books ub
		           WHERE ub.user_id = u.id
		             AND ub.review_text IS NOT NULL
		             AND ub.review_text != '') AS reviews_count,
		        (SELECT COUNT(*) FROM book_tag_values btv
		           JOIN tag_keys tk ON tk.id = btv.tag_key_id
		           JOIN tag_values tv ON tv.id = btv.tag_value_id
		           JOIN user_books ub ON ub.user_id = btv.user_id AND ub.book_id = btv.book_id
		           WHERE btv.user_id = u.id AND tk.slug = 'status' AND tv.slug = 'finished'
		             AND ub.date_read >= date_trunc('year', CURRENT_DATE)) AS books_this_year,
		        (SELECT AVG(ub.rating)::double precision FROM user_books ub
		           WHERE ub.user_id = u.id AND ub.rating IS NOT NULL) AS average_rating
		 FROM users u WHERE u.username = $1 AND u.deleted_at IS NULL`,
		username,
	).Scan(&p.UserID, &p.Username, &p.DisplayName, &p.Bio, &p.AvatarURL, &p.IsPrivate, &p.IsGhost, &memberSince, &p.AuthorKey, &p.FollowersCount, &p.FollowingCount, &p.FriendsCount, &p.BooksRead, &p.ReviewsCount, &p.BooksThisYear, &p.AverageRating)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	p.MemberSince = memberSince.Format(time.RFC3339)
	p.FollowStatus = "none"

	if currentUserID != "" && currentUserID != p.UserID {
		var status string
		err := h.pool.QueryRow(c.Request.Context(),
			`SELECT status FROM follows WHERE follower_id = $1 AND followee_id = $2`,
			currentUserID, p.UserID,
		).Scan(&status)
		if err == nil {
			p.FollowStatus = status
			p.IsFollowing = status == "active"
		}
	}

	// If the profile is private and the viewer is not the owner and not an active follower,
	// zero out content stats and mark as restricted.
	if p.IsPrivate && currentUserID != p.UserID && p.FollowStatus != "active" {
		p.BooksRead = 0
		p.ReviewsCount = 0
		p.BooksThisYear = 0
		p.AverageRating = nil
		p.IsRestricted = true
	}

	c.JSON(http.StatusOK, p)
}

func (h *Handler) UpdateMe(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	var req struct {
		DisplayName string `json:"display_name"`
		Bio         string `json:"bio"`
		IsPrivate   *bool  `json:"is_private"`
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

	if req.IsPrivate != nil {
		_, err := h.pool.Exec(c.Request.Context(),
			`UPDATE users SET display_name = $1, bio = $2, is_private = $3 WHERE id = $4`,
			displayName, bio, *req.IsPrivate, userID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
	} else {
		_, err := h.pool.Exec(c.Request.Context(),
			`UPDATE users SET display_name = $1, bio = $2 WHERE id = $3`,
			displayName, bio, userID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) Follow(c *gin.Context) {
	followerID := c.GetString(middleware.UserIDKey)
	username := c.Param("username")

	var followeeID string
	var isPrivate bool
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT id, is_private FROM users WHERE username = $1 AND deleted_at IS NULL`,
		username,
	).Scan(&followeeID, &isPrivate)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if followerID == followeeID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot follow yourself"})
		return
	}

	status := "active"
	if isPrivate {
		status = "pending"
	}

	_, err = h.pool.Exec(c.Request.Context(),
		`INSERT INTO follows (follower_id, followee_id, status) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
		followerID, followeeID, status,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// Only record activity for active follows, not pending requests
	if status == "active" {
		activity.Record(c.Request.Context(), h.pool, followerID, "followed_user", nil, &followeeID, nil, nil, nil)
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "status": status})
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

// ── Follow request management ─────────────────────────────────────────────────

type followRequestUser struct {
	UserID      string  `json:"user_id"`
	Username    string  `json:"username"`
	DisplayName *string `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
	CreatedAt   string  `json:"created_at"`
}

// GetFollowRequests - GET /me/follow-requests (authed)
// Lists pending follow requests for the current user.
func (h *Handler) GetFollowRequests(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	rows, err := h.pool.Query(c.Request.Context(),
		`SELECT u.id, u.username, u.display_name, u.avatar_url, f.created_at
		 FROM follows f
		 JOIN users u ON u.id = f.follower_id
		 WHERE f.followee_id = $1 AND f.status = 'pending' AND u.deleted_at IS NULL
		 ORDER BY f.created_at DESC`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer rows.Close()

	requests := []followRequestUser{}
	for rows.Next() {
		var r followRequestUser
		var createdAt time.Time
		if err := rows.Scan(&r.UserID, &r.Username, &r.DisplayName, &r.AvatarURL, &createdAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		r.CreatedAt = createdAt.Format(time.RFC3339)
		requests = append(requests, r)
	}

	c.JSON(http.StatusOK, requests)
}

// AcceptFollowRequest - POST /me/follow-requests/:userId/accept (authed)
func (h *Handler) AcceptFollowRequest(c *gin.Context) {
	currentUserID := c.GetString(middleware.UserIDKey)
	followerID := c.Param("userId")

	result, err := h.pool.Exec(c.Request.Context(),
		`UPDATE follows SET status = 'active' WHERE follower_id = $1 AND followee_id = $2 AND status = 'pending'`,
		followerID, currentUserID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "follow request not found"})
		return
	}

	// Record the followed_user activity now that the follow is accepted
	activity.Record(c.Request.Context(), h.pool, followerID, "followed_user", nil, &currentUserID, nil, nil, nil)

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// RejectFollowRequest - DELETE /me/follow-requests/:userId/reject (authed)
func (h *Handler) RejectFollowRequest(c *gin.Context) {
	currentUserID := c.GetString(middleware.UserIDKey)
	followerID := c.Param("userId")

	result, err := h.pool.Exec(c.Request.Context(),
		`DELETE FROM follows WHERE follower_id = $1 AND followee_id = $2 AND status = 'pending'`,
		followerID, currentUserID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "follow request not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GetUserReviews - GET /users/:username/reviews
// Public, but gated for private profiles. Returns all collection items for the user that have review text.
func (h *Handler) GetUserReviews(c *gin.Context) {
	username := c.Param("username")
	currentUserID := c.GetString(middleware.UserIDKey)

	_, _, canView := privacy.CanViewProfile(c.Request.Context(), h.pool, username, currentUserID)
	if !canView {
		c.JSON(http.StatusOK, []struct{}{})
		return
	}

	query := `SELECT b.id, b.open_library_id, b.title, b.cover_url, b.authors,
		        ub.rating, ub.review_text, ub.spoiler, ub.date_read, ub.date_dnf, ub.date_added
		 FROM user_books ub
		 JOIN books b  ON b.id  = ub.book_id
		 JOIN users u  ON u.id  = ub.user_id
		 WHERE u.username = $1
		   AND u.deleted_at IS NULL
		   AND ub.review_text IS NOT NULL
		   AND ub.review_text != ''
		 ORDER BY ub.date_added DESC`
	args := []interface{}{username}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit >= 1 && limit <= 100 {
			query += " LIMIT $2"
			args = append(args, limit)
		}
	}

	rows, err := h.pool.Query(c.Request.Context(), query, args...)
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
		DateDNF       *string `json:"date_dnf"`
		DateAdded     string  `json:"date_added"`
	}

	reviews := []reviewItem{}
	for rows.Next() {
		var r reviewItem
		var dateRead *time.Time
		var dateDNF *time.Time
		var dateAdded time.Time
		if err := rows.Scan(
			&r.BookID, &r.OpenLibraryID, &r.Title, &r.CoverURL, &r.Authors,
			&r.Rating, &r.ReviewText, &r.Spoiler, &dateRead, &dateDNF, &dateAdded,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		r.DateAdded = dateAdded.Format(time.RFC3339)
		if dateRead != nil {
			s := dateRead.Format(time.RFC3339)
			r.DateRead = &s
		}
		if dateDNF != nil {
			s := dateDNF.Format(time.RFC3339)
			r.DateDNF = &s
		}
		reviews = append(reviews, r)
	}

	c.JSON(http.StatusOK, reviews)
}

// ── Admin ─────────────────────────────────────────────────────────────────────

type adminUserRow struct {
	UserID      string  `json:"user_id"`
	Username    string  `json:"username"`
	DisplayName *string `json:"display_name"`
	Email       string  `json:"email"`
	IsModerator bool    `json:"is_moderator"`
	AuthorKey   *string `json:"author_key"`
}

// ListAllUsers - GET /admin/users (moderator only)
func (h *Handler) ListAllUsers(c *gin.Context) {
	q := c.Query("q")
	page := 1
	if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 1 {
		page = p
	}
	offset := (page - 1) * pageSize

	var query string
	var args []interface{}
	if q != "" {
		query = `SELECT id, username, display_name, email, is_moderator, author_key
			 FROM users
			 WHERE deleted_at IS NULL
			   AND (username ILIKE '%' || $1 || '%' OR display_name ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%')
			 ORDER BY username
			 LIMIT $2 OFFSET $3`
		args = []interface{}{q, pageSize + 1, offset}
	} else {
		query = `SELECT id, username, display_name, email, is_moderator, author_key
			 FROM users
			 WHERE deleted_at IS NULL
			 ORDER BY username
			 LIMIT $1 OFFSET $2`
		args = []interface{}{pageSize + 1, offset}
	}

	rows, err := h.pool.Query(c.Request.Context(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer rows.Close()

	users := []adminUserRow{}
	for rows.Next() {
		var u adminUserRow
		if err := rows.Scan(&u.UserID, &u.Username, &u.DisplayName, &u.Email, &u.IsModerator, &u.AuthorKey); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		users = append(users, u)
	}

	hasNext := len(users) > pageSize
	if hasNext {
		users = users[:pageSize]
	}

	c.JSON(http.StatusOK, gin.H{"users": users, "page": page, "has_next": hasNext})
}

// SetModerator - PUT /admin/users/:userId/moderator (moderator only)
func (h *Handler) SetModerator(c *gin.Context) {
	targetUserID := c.Param("userId")

	var req struct {
		IsModerator bool `json:"is_moderator"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.pool.Exec(c.Request.Context(),
		`UPDATE users SET is_moderator = $1 WHERE id = $2 AND deleted_at IS NULL`,
		req.IsModerator, targetUserID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "is_moderator": req.IsModerator})
}

// SetAuthor - PUT /admin/users/:userId/author (moderator only)
func (h *Handler) SetAuthor(c *gin.Context) {
	targetUserID := c.Param("userId")

	var req struct {
		AuthorKey *string `json:"author_key"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Normalize empty string to nil (clears the author key)
	if req.AuthorKey != nil && *req.AuthorKey == "" {
		req.AuthorKey = nil
	}

	result, err := h.pool.Exec(c.Request.Context(),
		`UPDATE users SET author_key = $1 WHERE id = $2 AND deleted_at IS NULL`,
		req.AuthorKey, targetUserID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "author_key": req.AuthorKey})
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
