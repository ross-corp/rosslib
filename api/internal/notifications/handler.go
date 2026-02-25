package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tristansaldanha/rosslib/api/internal/middleware"
)

// NotifyBookFollowers sends a notification to all users who follow a book,
// excluding the actor who triggered the event. This is fire-and-forget:
// errors are silently ignored so notification creation never blocks a
// primary operation.
func NotifyBookFollowers(ctx context.Context, pool *pgxpool.Pool, bookID, actorID, notifType, title, body string, metadata map[string]string) {
	rows, err := pool.Query(ctx,
		`SELECT user_id FROM book_follows WHERE book_id = $1 AND user_id != $2`,
		bookID, actorID,
	)
	if err != nil {
		return
	}
	defer rows.Close()

	var meta interface{}
	if len(metadata) > 0 {
		b, err := json.Marshal(metadata)
		if err == nil {
			meta = b
		}
	}

	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			continue
		}
		_, _ = pool.Exec(ctx,
			`INSERT INTO notifications (user_id, notif_type, title, body, metadata)
			 VALUES ($1, $2, $3, $4, $5)`,
			userID, notifType, title, body, meta,
		)
	}
}

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

type notificationItem struct {
	ID        string            `json:"id"`
	Type      string            `json:"notif_type"`
	Title     string            `json:"title"`
	Body      *string           `json:"body"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Read      bool              `json:"read"`
	CreatedAt string            `json:"created_at"`
}

// GetNotifications returns paginated notifications for the current user.
//
// GET /me/notifications?cursor=<RFC3339Nano>&limit=<int>
func (h *Handler) GetNotifications(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	cursor := c.Query("cursor")
	const limit = 30

	query := `
		SELECT id, notif_type, title, body, metadata, read, created_at
		FROM notifications
		WHERE user_id = $1`

	args := []interface{}{userID}
	idx := 2

	if cursor != "" {
		t, err := time.Parse(time.RFC3339Nano, cursor)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cursor"})
			return
		}
		query += fmt.Sprintf(` AND created_at < $%d`, idx)
		args = append(args, t)
		idx++
	}

	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d`, idx)
	args = append(args, limit+1)

	rows, err := h.pool.Query(c.Request.Context(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer rows.Close()

	items := []notificationItem{}
	for rows.Next() {
		var item notificationItem
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.Type, &item.Title, &item.Body, &item.Metadata, &item.Read, &createdAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		item.CreatedAt = createdAt.Format(time.RFC3339Nano)
		items = append(items, item)
	}

	resp := gin.H{"notifications": items}
	if len(items) > limit {
		items = items[:limit]
		resp["notifications"] = items
		resp["next_cursor"] = items[limit-1].CreatedAt
	}

	c.JSON(http.StatusOK, resp)
}

// GetUnreadCount returns the number of unread notifications.
//
// GET /me/notifications/unread-count
func (h *Handler) GetUnreadCount(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	var count int
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read = false`,
		userID,
	).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

// MarkRead marks a single notification as read.
//
// POST /me/notifications/:notifId/read
func (h *Handler) MarkRead(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	notifID := c.Param("notifId")

	tag, err := h.pool.Exec(c.Request.Context(),
		`UPDATE notifications SET read = true WHERE id = $1 AND user_id = $2`,
		notifID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// MarkAllRead marks all unread notifications as read for the current user.
//
// POST /me/notifications/read-all
func (h *Handler) MarkAllRead(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	_, err := h.pool.Exec(c.Request.Context(),
		`UPDATE notifications SET read = true WHERE user_id = $1 AND read = false`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
