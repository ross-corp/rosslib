package links

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tristansaldanha/rosslib/api/internal/middleware"
)

var validLinkTypes = map[string]bool{
	"sequel":       true,
	"prequel":      true,
	"companion":    true,
	"mentioned_in": true,
	"similar":      true,
	"adaptation":   true,
}

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// ── Types ────────────────────────────────────────────────────────────────────

type linkResponse struct {
	ID            string  `json:"id"`
	FromBookOLID  string  `json:"from_book_ol_id"`
	ToBookOLID    string  `json:"to_book_ol_id"`
	ToBookTitle   string  `json:"to_book_title"`
	ToBookAuthors *string `json:"to_book_authors"`
	ToBookCover   *string `json:"to_book_cover_url"`
	LinkType      string  `json:"link_type"`
	Note          *string `json:"note"`
	Username      string  `json:"username"`
	DisplayName   *string `json:"display_name"`
	Votes         int     `json:"votes"`
	UserVoted     bool    `json:"user_voted"`
	CreatedAt     string  `json:"created_at"`
}

// ── List links for a book ────────────────────────────────────────────────────

// GET /books/:workId/links
func (h *Handler) ListLinks(c *gin.Context) {
	workID := c.Param("workId")
	userID := c.GetString(middleware.UserIDKey) // empty if unauthenticated

	voterID := userID
	if voterID == "" {
		voterID = "00000000-0000-0000-0000-000000000000"
	}

	rows, err := h.pool.Query(c.Request.Context(),
		`SELECT bl.id,
		        from_b.open_library_id AS from_ol_id,
		        to_b.open_library_id   AS to_ol_id,
		        to_b.title             AS to_title,
		        to_b.authors           AS to_authors,
		        to_b.cover_url         AS to_cover,
		        bl.link_type,
		        bl.note,
		        u.username,
		        u.display_name,
		        (SELECT COUNT(*) FROM book_link_votes v WHERE v.book_link_id = bl.id) AS votes,
		        EXISTS(SELECT 1 FROM book_link_votes v WHERE v.book_link_id = bl.id AND v.user_id = $2) AS user_voted,
		        bl.created_at
		 FROM book_links bl
		 JOIN books from_b ON from_b.id = bl.from_book_id
		 JOIN books to_b   ON to_b.id   = bl.to_book_id
		 JOIN users u      ON u.id      = bl.user_id
		 WHERE from_b.open_library_id = $1
		   AND bl.deleted_at IS NULL
		   AND u.deleted_at IS NULL
		 ORDER BY votes DESC, bl.created_at ASC`,
		workID, voterID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer rows.Close()

	links := []linkResponse{}
	for rows.Next() {
		var l linkResponse
		var createdAt time.Time
		if err := rows.Scan(
			&l.ID, &l.FromBookOLID, &l.ToBookOLID,
			&l.ToBookTitle, &l.ToBookAuthors, &l.ToBookCover,
			&l.LinkType, &l.Note,
			&l.Username, &l.DisplayName,
			&l.Votes, &l.UserVoted, &createdAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		l.CreatedAt = createdAt.Format(time.RFC3339)
		links = append(links, l)
	}

	c.JSON(http.StatusOK, links)
}

// ── Create a link ────────────────────────────────────────────────────────────

// POST /books/:workId/links  (requires auth)
func (h *Handler) CreateLink(c *gin.Context) {
	workID := c.Param("workId")
	userID := c.GetString(middleware.UserIDKey)

	var body struct {
		ToWorkID string  `json:"to_work_id" binding:"required"`
		LinkType string  `json:"link_type" binding:"required"`
		Note     *string `json:"note"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "to_work_id and link_type are required"})
		return
	}

	if !validLinkTypes[body.LinkType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid link_type; must be one of: sequel, prequel, companion, mentioned_in, similar, adaptation"})
		return
	}

	if workID == body.ToWorkID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot link a book to itself"})
		return
	}

	// Look up both books by OL ID.
	var fromBookID string
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT id FROM books WHERE open_library_id = $1`, workID,
	).Scan(&fromBookID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "source book not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	var toBookID string
	err = h.pool.QueryRow(c.Request.Context(),
		`SELECT id FROM books WHERE open_library_id = $1`, body.ToWorkID,
	).Scan(&toBookID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "target book not found in local catalog; add it to a shelf first"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	var linkID string
	var createdAt time.Time
	err = h.pool.QueryRow(c.Request.Context(),
		`INSERT INTO book_links (from_book_id, to_book_id, user_id, link_type, note)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (from_book_id, to_book_id, link_type, user_id) DO NOTHING
		 RETURNING id, created_at`,
		fromBookID, toBookID, userID, body.LinkType, body.Note,
	).Scan(&linkID, &createdAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusConflict, gin.H{"error": "you already submitted this link"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// Auto-upvote by the creator.
	_, _ = h.pool.Exec(c.Request.Context(),
		`INSERT INTO book_link_votes (user_id, book_link_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, linkID,
	)

	c.JSON(http.StatusCreated, gin.H{
		"id":         linkID,
		"created_at": createdAt.Format(time.RFC3339),
	})
}

// ── Delete a link (soft delete, author only) ─────────────────────────────────

// DELETE /links/:linkId  (requires auth)
func (h *Handler) DeleteLink(c *gin.Context) {
	linkID := c.Param("linkId")
	userID := c.GetString(middleware.UserIDKey)

	tag, err := h.pool.Exec(c.Request.Context(),
		`UPDATE book_links SET deleted_at = NOW()
		 WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`,
		linkID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found or not yours"})
		return
	}

	c.Status(http.StatusNoContent)
}

// ── Vote / unvote a link ─────────────────────────────────────────────────────

// POST /links/:linkId/vote  (requires auth)
func (h *Handler) Vote(c *gin.Context) {
	linkID := c.Param("linkId")
	userID := c.GetString(middleware.UserIDKey)

	// Verify link exists.
	var exists bool
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT EXISTS(SELECT 1 FROM book_links WHERE id = $1 AND deleted_at IS NULL)`,
		linkID,
	).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}

	_, err = h.pool.Exec(c.Request.Context(),
		`INSERT INTO book_link_votes (user_id, book_link_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, linkID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.Status(http.StatusNoContent)
}

// DELETE /links/:linkId/vote  (requires auth)
func (h *Handler) Unvote(c *gin.Context) {
	linkID := c.Param("linkId")
	userID := c.GetString(middleware.UserIDKey)

	_, err := h.pool.Exec(c.Request.Context(),
		`DELETE FROM book_link_votes WHERE user_id = $1 AND book_link_id = $2`,
		userID, linkID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.Status(http.StatusNoContent)
}
