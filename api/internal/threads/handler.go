package threads

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tristansaldanha/rosslib/api/internal/activity"
	"github.com/tristansaldanha/rosslib/api/internal/middleware"
	"github.com/tristansaldanha/rosslib/api/internal/notifications"
)

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// ── Types ────────────────────────────────────────────────────────────────────

type threadResponse struct {
	ID          string  `json:"id"`
	BookID      string  `json:"book_id"`
	UserID      string  `json:"user_id"`
	Username    string  `json:"username"`
	DisplayName *string `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
	Title       string  `json:"title"`
	Body        string  `json:"body"`
	Spoiler     bool    `json:"spoiler"`
	CreatedAt   string  `json:"created_at"`
	CommentCount int    `json:"comment_count"`
}

type commentResponse struct {
	ID          string  `json:"id"`
	ThreadID    string  `json:"thread_id"`
	UserID      string  `json:"user_id"`
	Username    string  `json:"username"`
	DisplayName *string `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
	ParentID    *string `json:"parent_id"`
	Body        string  `json:"body"`
	CreatedAt   string  `json:"created_at"`
}

// ── List threads for a book ──────────────────────────────────────────────────

// GET /books/:workId/threads
func (h *Handler) ListThreads(c *gin.Context) {
	workID := c.Param("workId")

	rows, err := h.pool.Query(c.Request.Context(),
		`SELECT t.id, t.book_id, t.user_id, u.username, u.display_name, u.avatar_url,
		        t.title, t.body, t.spoiler, t.created_at,
		        (SELECT COUNT(*) FROM thread_comments tc
		         WHERE tc.thread_id = t.id AND tc.deleted_at IS NULL) AS comment_count
		 FROM threads t
		 JOIN books b ON b.id = t.book_id
		 JOIN users u ON u.id = t.user_id
		 WHERE b.open_library_id = $1
		   AND t.deleted_at IS NULL
		   AND u.deleted_at IS NULL
		 ORDER BY t.created_at DESC`,
		workID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer rows.Close()

	threads := []threadResponse{}
	for rows.Next() {
		var t threadResponse
		var createdAt time.Time
		if err := rows.Scan(
			&t.ID, &t.BookID, &t.UserID, &t.Username, &t.DisplayName, &t.AvatarURL,
			&t.Title, &t.Body, &t.Spoiler, &createdAt, &t.CommentCount,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		t.CreatedAt = createdAt.Format(time.RFC3339)
		threads = append(threads, t)
	}

	c.JSON(http.StatusOK, threads)
}

// ── Get a single thread with comments ────────────────────────────────────────

// GET /threads/:threadId
func (h *Handler) GetThread(c *gin.Context) {
	threadID := c.Param("threadId")

	// Fetch thread.
	var t threadResponse
	var createdAt time.Time
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT t.id, t.book_id, t.user_id, u.username, u.display_name, u.avatar_url,
		        t.title, t.body, t.spoiler, t.created_at,
		        (SELECT COUNT(*) FROM thread_comments tc
		         WHERE tc.thread_id = t.id AND tc.deleted_at IS NULL) AS comment_count
		 FROM threads t
		 JOIN users u ON u.id = t.user_id
		 WHERE t.id = $1
		   AND t.deleted_at IS NULL
		   AND u.deleted_at IS NULL`,
		threadID,
	).Scan(
		&t.ID, &t.BookID, &t.UserID, &t.Username, &t.DisplayName, &t.AvatarURL,
		&t.Title, &t.Body, &t.Spoiler, &createdAt, &t.CommentCount,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "thread not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	t.CreatedAt = createdAt.Format(time.RFC3339)

	// Fetch comments.
	rows, err := h.pool.Query(c.Request.Context(),
		`SELECT tc.id, tc.thread_id, tc.user_id, u.username, u.display_name, u.avatar_url,
		        tc.parent_id, tc.body, tc.created_at
		 FROM thread_comments tc
		 JOIN users u ON u.id = tc.user_id
		 WHERE tc.thread_id = $1
		   AND tc.deleted_at IS NULL
		   AND u.deleted_at IS NULL
		 ORDER BY tc.created_at ASC`,
		threadID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer rows.Close()

	comments := []commentResponse{}
	for rows.Next() {
		var cm commentResponse
		var cmCreatedAt time.Time
		if err := rows.Scan(
			&cm.ID, &cm.ThreadID, &cm.UserID, &cm.Username, &cm.DisplayName, &cm.AvatarURL,
			&cm.ParentID, &cm.Body, &cmCreatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		cm.CreatedAt = cmCreatedAt.Format(time.RFC3339)
		comments = append(comments, cm)
	}

	c.JSON(http.StatusOK, gin.H{
		"thread":   t,
		"comments": comments,
	})
}

// ── Create a thread ──────────────────────────────────────────────────────────

// POST /books/:workId/threads  (requires auth)
func (h *Handler) CreateThread(c *gin.Context) {
	workID := c.Param("workId")
	userID := c.GetString(middleware.UserIDKey)

	var body struct {
		Title   string `json:"title" binding:"required"`
		Body    string `json:"body" binding:"required"`
		Spoiler bool   `json:"spoiler"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title and body are required"})
		return
	}

	// Find the book by OL ID.
	var bookID string
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT id FROM books WHERE open_library_id = $1`, workID,
	).Scan(&bookID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	var threadID string
	var createdAt time.Time
	err = h.pool.QueryRow(c.Request.Context(),
		`INSERT INTO threads (book_id, user_id, title, body, spoiler)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at`,
		bookID, userID, body.Title, body.Body, body.Spoiler,
	).Scan(&threadID, &createdAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	activity.Record(c.Request.Context(), h.pool, userID, "created_thread", &bookID, nil, nil, &threadID,
		map[string]string{"thread_title": body.Title})

	// Notify followers of this book about the new discussion thread.
	var bookTitle string
	_ = h.pool.QueryRow(c.Request.Context(),
		`SELECT title FROM books WHERE id = $1`, bookID,
	).Scan(&bookTitle)
	go notifications.NotifyBookFollowers(c.Request.Context(), h.pool, bookID, userID,
		"book_new_thread",
		"New discussion on "+bookTitle,
		"\""+body.Title+"\"",
		map[string]string{"book_ol_id": workID, "book_title": bookTitle, "thread_title": body.Title},
	)

	c.JSON(http.StatusCreated, gin.H{
		"id":         threadID,
		"created_at": createdAt.Format(time.RFC3339),
	})
}

// ── Delete a thread (soft delete, author only) ──────────────────────────────

// DELETE /threads/:threadId  (requires auth)
func (h *Handler) DeleteThread(c *gin.Context) {
	threadID := c.Param("threadId")
	userID := c.GetString(middleware.UserIDKey)

	tag, err := h.pool.Exec(c.Request.Context(),
		`UPDATE threads SET deleted_at = NOW()
		 WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`,
		threadID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "thread not found or not yours"})
		return
	}

	c.Status(http.StatusNoContent)
}

// ── Create a comment ─────────────────────────────────────────────────────────

// POST /threads/:threadId/comments  (requires auth)
func (h *Handler) CreateComment(c *gin.Context) {
	threadID := c.Param("threadId")
	userID := c.GetString(middleware.UserIDKey)

	var body struct {
		Body     string  `json:"body" binding:"required"`
		ParentID *string `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body is required"})
		return
	}

	// Verify thread exists and is not deleted.
	var exists bool
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT EXISTS(SELECT 1 FROM threads WHERE id = $1 AND deleted_at IS NULL)`,
		threadID,
	).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "thread not found"})
		return
	}

	// If replying, enforce one level of nesting: parent must not itself have a parent.
	if body.ParentID != nil {
		var parentParentID *string
		err := h.pool.QueryRow(c.Request.Context(),
			`SELECT parent_id FROM thread_comments
			 WHERE id = $1 AND thread_id = $2 AND deleted_at IS NULL`,
			*body.ParentID, threadID,
		).Scan(&parentParentID)
		if err != nil {
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "parent comment not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		if parentParentID != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "replies can only be one level deep"})
			return
		}
	}

	var commentID string
	var createdAt time.Time
	err = h.pool.QueryRow(c.Request.Context(),
		`INSERT INTO thread_comments (thread_id, user_id, parent_id, body)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`,
		threadID, userID, body.ParentID, body.Body,
	).Scan(&commentID, &createdAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         commentID,
		"created_at": createdAt.Format(time.RFC3339),
	})
}

// ── Delete a comment (soft delete, author only) ─────────────────────────────

// DELETE /threads/:threadId/comments/:commentId  (requires auth)
func (h *Handler) DeleteComment(c *gin.Context) {
	commentID := c.Param("commentId")
	userID := c.GetString(middleware.UserIDKey)

	tag, err := h.pool.Exec(c.Request.Context(),
		`UPDATE thread_comments SET deleted_at = NOW()
		 WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`,
		commentID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found or not yours"})
		return
	}

	c.Status(http.StatusNoContent)
}
