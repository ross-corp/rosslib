package activity

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tristansaldanha/rosslib/api/internal/middleware"
	"github.com/tristansaldanha/rosslib/api/internal/privacy"
)

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// Record inserts an activity row. Callers pass nil for unused foreign keys.
// This is fire-and-forget: errors are silently ignored so activity tracking
// never breaks a primary operation.
func Record(ctx context.Context, pool *pgxpool.Pool, userID, activityType string, bookID, targetUserID, collectionID, threadID *string, metadata map[string]string) {
	var meta interface{}
	if len(metadata) > 0 {
		meta = metadata
	}
	_, _ = pool.Exec(ctx,
		`INSERT INTO activities (user_id, activity_type, book_id, target_user_id, collection_id, thread_id, metadata)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		userID, activityType, bookID, targetUserID, collectionID, threadID, meta,
	)
}

// ── Types ────────────────────────────────────────────────────────────────────

type activityUser struct {
	UserID      string  `json:"user_id"`
	Username    string  `json:"username"`
	DisplayName *string `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
}

type activityBook struct {
	OpenLibraryID string  `json:"open_library_id"`
	Title         string  `json:"title"`
	CoverURL      *string `json:"cover_url"`
}

type activityItem struct {
	ID           string        `json:"id"`
	Type         string        `json:"type"`
	CreatedAt    string        `json:"created_at"`
	User         activityUser  `json:"user"`
	Book         *activityBook `json:"book,omitempty"`
	TargetUser   *activityUser `json:"target_user,omitempty"`
	ShelfName    *string       `json:"shelf_name,omitempty"`
	Rating       *int          `json:"rating,omitempty"`
	ReviewText   *string       `json:"review_snippet,omitempty"`
	ThreadTitle  *string       `json:"thread_title,omitempty"`
	LinkType     *string       `json:"link_type,omitempty"`
	ToBookOLID   *string       `json:"to_book_ol_id,omitempty"`
	ToBookTitle  *string       `json:"to_book_title,omitempty"`
	AuthorKey    *string       `json:"author_key,omitempty"`
	AuthorName   *string       `json:"author_name,omitempty"`
}

type feedResponse struct {
	Activities []activityItem `json:"activities"`
	NextCursor string         `json:"next_cursor,omitempty"`
}

// ── GetFeed — GET /me/feed (authed) ─────────────────────────────────────────

func (h *Handler) GetFeed(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	cursor := c.Query("cursor")
	const limit = 30

	query := `
		SELECT a.id, a.activity_type, a.created_at,
		       u.id, u.username, u.display_name, u.avatar_url,
		       b.open_library_id, b.title, b.cover_url,
		       tu.id, tu.username, tu.display_name, tu.avatar_url,
		       COALESCE(a.metadata->>'status_name', col.name),
		       a.metadata
		FROM activities a
		JOIN users u ON u.id = a.user_id
		LEFT JOIN books b ON b.id = a.book_id
		LEFT JOIN users tu ON tu.id = a.target_user_id
		LEFT JOIN collections col ON col.id = a.collection_id
		WHERE a.user_id IN (
			SELECT followee_id FROM follows WHERE follower_id = $1 AND status = 'active'
		)
		AND u.deleted_at IS NULL`

	args := []interface{}{userID}
	idx := 2

	if cursor != "" {
		t, err := time.Parse(time.RFC3339Nano, cursor)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cursor"})
			return
		}
		query += fmt.Sprintf(` AND a.created_at < $%d`, idx)
		args = append(args, t)
		idx++
	}

	query += fmt.Sprintf(` ORDER BY a.created_at DESC LIMIT $%d`, idx)
	args = append(args, limit+1)

	rows, err := h.pool.Query(c.Request.Context(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer rows.Close()

	activities := []activityItem{}
	for rows.Next() {
		var item activityItem
		var createdAt time.Time
		var bookOLID, bookTitle *string
		var bookCover *string
		var tuID, tuUsername *string
		var tuDisplayName, tuAvatarURL *string
		var shelfName *string
		var meta map[string]string

		if err := rows.Scan(
			&item.ID, &item.Type, &createdAt,
			&item.User.UserID, &item.User.Username, &item.User.DisplayName, &item.User.AvatarURL,
			&bookOLID, &bookTitle, &bookCover,
			&tuID, &tuUsername, &tuDisplayName, &tuAvatarURL,
			&shelfName,
			&meta,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		item.CreatedAt = createdAt.Format(time.RFC3339Nano)

		if bookOLID != nil && bookTitle != nil {
			item.Book = &activityBook{
				OpenLibraryID: *bookOLID,
				Title:         *bookTitle,
				CoverURL:      bookCover,
			}
		}

		if tuID != nil && tuUsername != nil {
			item.TargetUser = &activityUser{
				UserID:      *tuID,
				Username:    *tuUsername,
				DisplayName: tuDisplayName,
				AvatarURL:   tuAvatarURL,
			}
		}

		item.ShelfName = shelfName

		if meta != nil {
			if v, ok := meta["rating"]; ok {
				r := atoiPtr(v)
				item.Rating = r
			}
			if v, ok := meta["review_snippet"]; ok {
				item.ReviewText = &v
			}
			if v, ok := meta["thread_title"]; ok {
				item.ThreadTitle = &v
			}
			if v, ok := meta["link_type"]; ok {
				item.LinkType = &v
			}
			if v, ok := meta["to_book_ol_id"]; ok {
				item.ToBookOLID = &v
			}
			if v, ok := meta["to_book_title"]; ok {
				item.ToBookTitle = &v
			}
			if v, ok := meta["author_key"]; ok {
				item.AuthorKey = &v
			}
			if v, ok := meta["author_name"]; ok {
				item.AuthorName = &v
			}
		}

		activities = append(activities, item)
	}

	resp := feedResponse{Activities: activities}
	if len(activities) > limit {
		resp.Activities = activities[:limit]
		resp.NextCursor = activities[limit-1].CreatedAt
	}

	c.JSON(http.StatusOK, resp)
}

// ── GetUserActivity — GET /users/:username/activity ─────────────────────────

func (h *Handler) GetUserActivity(c *gin.Context) {
	username := c.Param("username")
	currentUserID := c.GetString(middleware.UserIDKey)

	_, _, canView := privacy.CanViewProfile(c.Request.Context(), h.pool, username, currentUserID)
	if !canView {
		c.JSON(http.StatusOK, feedResponse{Activities: []activityItem{}})
		return
	}

	cursor := c.Query("cursor")
	limit := 30
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v >= 1 && v <= 30 {
			limit = v
		}
	}

	query := `
		SELECT a.id, a.activity_type, a.created_at,
		       u.id, u.username, u.display_name, u.avatar_url,
		       b.open_library_id, b.title, b.cover_url,
		       tu.id, tu.username, tu.display_name, tu.avatar_url,
		       COALESCE(a.metadata->>'status_name', col.name),
		       a.metadata
		FROM activities a
		JOIN users u ON u.id = a.user_id
		LEFT JOIN books b ON b.id = a.book_id
		LEFT JOIN users tu ON tu.id = a.target_user_id
		LEFT JOIN collections col ON col.id = a.collection_id
		WHERE u.username = $1
		AND u.deleted_at IS NULL`

	args := []interface{}{username}
	idx := 2

	if cursor != "" {
		t, err := time.Parse(time.RFC3339Nano, cursor)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cursor"})
			return
		}
		query += fmt.Sprintf(` AND a.created_at < $%d`, idx)
		args = append(args, t)
		idx++
	}

	query += fmt.Sprintf(` ORDER BY a.created_at DESC LIMIT $%d`, idx)
	args = append(args, limit+1)

	rows, err := h.pool.Query(c.Request.Context(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer rows.Close()

	activities := []activityItem{}
	for rows.Next() {
		var item activityItem
		var createdAt time.Time
		var bookOLID, bookTitle *string
		var bookCover *string
		var tuID, tuUsername *string
		var tuDisplayName, tuAvatarURL *string
		var shelfName *string
		var meta map[string]string

		if err := rows.Scan(
			&item.ID, &item.Type, &createdAt,
			&item.User.UserID, &item.User.Username, &item.User.DisplayName, &item.User.AvatarURL,
			&bookOLID, &bookTitle, &bookCover,
			&tuID, &tuUsername, &tuDisplayName, &tuAvatarURL,
			&shelfName,
			&meta,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		item.CreatedAt = createdAt.Format(time.RFC3339Nano)

		if bookOLID != nil && bookTitle != nil {
			item.Book = &activityBook{
				OpenLibraryID: *bookOLID,
				Title:         *bookTitle,
				CoverURL:      bookCover,
			}
		}

		if tuID != nil && tuUsername != nil {
			item.TargetUser = &activityUser{
				UserID:      *tuID,
				Username:    *tuUsername,
				DisplayName: tuDisplayName,
				AvatarURL:   tuAvatarURL,
			}
		}

		item.ShelfName = shelfName

		if meta != nil {
			if v, ok := meta["rating"]; ok {
				r := atoiPtr(v)
				item.Rating = r
			}
			if v, ok := meta["review_snippet"]; ok {
				item.ReviewText = &v
			}
			if v, ok := meta["thread_title"]; ok {
				item.ThreadTitle = &v
			}
			if v, ok := meta["link_type"]; ok {
				item.LinkType = &v
			}
			if v, ok := meta["to_book_ol_id"]; ok {
				item.ToBookOLID = &v
			}
			if v, ok := meta["to_book_title"]; ok {
				item.ToBookTitle = &v
			}
			if v, ok := meta["author_key"]; ok {
				item.AuthorKey = &v
			}
			if v, ok := meta["author_name"]; ok {
				item.AuthorName = &v
			}
		}

		activities = append(activities, item)
	}

	resp := feedResponse{Activities: activities}
	if len(activities) > limit {
		resp.Activities = activities[:limit]
		resp.NextCursor = activities[limit-1].CreatedAt
	}

	c.JSON(http.StatusOK, resp)
}

// ── helpers ──────────────────────────────────────────────────────────────────

func atoiPtr(s string) *int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &v
}
