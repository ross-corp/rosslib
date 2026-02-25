package links

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

// ── Delete a link (soft delete, author or moderator) ─────────────────────────

// DELETE /links/:linkId  (requires auth)
func (h *Handler) DeleteLink(c *gin.Context) {
	linkID := c.Param("linkId")
	userID := c.GetString(middleware.UserIDKey)
	isMod, _ := c.Get(middleware.IsModeratorKey)

	var tag pgconn.CommandTag
	var err error

	if isMod == true {
		tag, err = h.pool.Exec(c.Request.Context(),
			`UPDATE book_links SET deleted_at = NOW()
			 WHERE id = $1 AND deleted_at IS NULL`,
			linkID,
		)
	} else {
		tag, err = h.pool.Exec(c.Request.Context(),
			`UPDATE book_links SET deleted_at = NOW()
			 WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`,
			linkID, userID,
		)
	}
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

// ── Edit queue ──────────────────────────────────────────────────────────────

type editResponse struct {
	ID              string  `json:"id"`
	BookLinkID      string  `json:"book_link_id"`
	Username        string  `json:"username"`
	DisplayName     *string `json:"display_name"`
	ProposedType    *string `json:"proposed_type"`
	ProposedNote    *string `json:"proposed_note"`
	CurrentType     string  `json:"current_type"`
	CurrentNote     *string `json:"current_note"`
	FromBookOLID    string  `json:"from_book_ol_id"`
	FromBookTitle   string  `json:"from_book_title"`
	ToBookOLID      string  `json:"to_book_ol_id"`
	ToBookTitle     string  `json:"to_book_title"`
	Status          string  `json:"status"`
	ReviewerName    *string `json:"reviewer_name"`
	ReviewerComment *string `json:"reviewer_comment"`
	CreatedAt       string  `json:"created_at"`
	ReviewedAt      *string `json:"reviewed_at"`
}

// POST /links/:linkId/edits  (requires auth)
func (h *Handler) ProposeEdit(c *gin.Context) {
	linkID := c.Param("linkId")
	userID := c.GetString(middleware.UserIDKey)

	var body struct {
		ProposedType *string `json:"proposed_type"`
		ProposedNote *string `json:"proposed_note"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if body.ProposedType == nil && body.ProposedNote == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one of proposed_type or proposed_note is required"})
		return
	}

	if body.ProposedType != nil && !validLinkTypes[*body.ProposedType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid proposed_type"})
		return
	}

	// Verify the link exists and isn't deleted.
	var exists bool
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT EXISTS(SELECT 1 FROM book_links WHERE id = $1 AND deleted_at IS NULL)`,
		linkID,
	).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}

	// Check for existing pending edit by same user on same link.
	var hasPending bool
	err = h.pool.QueryRow(c.Request.Context(),
		`SELECT EXISTS(SELECT 1 FROM book_link_edits WHERE book_link_id = $1 AND user_id = $2 AND status = 'pending')`,
		linkID, userID,
	).Scan(&hasPending)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if hasPending {
		c.JSON(http.StatusConflict, gin.H{"error": "you already have a pending edit for this link"})
		return
	}

	var editID string
	var createdAt time.Time
	err = h.pool.QueryRow(c.Request.Context(),
		`INSERT INTO book_link_edits (book_link_id, user_id, proposed_type, proposed_note)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`,
		linkID, userID, body.ProposedType, body.ProposedNote,
	).Scan(&editID, &createdAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         editID,
		"created_at": createdAt.Format(time.RFC3339),
	})
}

// GET /admin/link-edits?status=pending  (moderator only)
func (h *Handler) ListEdits(c *gin.Context) {
	statusFilter := c.DefaultQuery("status", "pending")
	if statusFilter != "pending" && statusFilter != "approved" && statusFilter != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status must be pending, approved, or rejected"})
		return
	}

	rows, err := h.pool.Query(c.Request.Context(),
		`SELECT e.id, e.book_link_id,
		        u.username, u.display_name,
		        e.proposed_type, e.proposed_note,
		        bl.link_type AS current_type, bl.note AS current_note,
		        fb.open_library_id AS from_ol_id, fb.title AS from_title,
		        tb.open_library_id AS to_ol_id, tb.title AS to_title,
		        e.status,
		        ru.username AS reviewer_name,
		        e.reviewer_comment,
		        e.created_at, e.reviewed_at
		 FROM book_link_edits e
		 JOIN book_links bl ON bl.id = e.book_link_id
		 JOIN books fb ON fb.id = bl.from_book_id
		 JOIN books tb ON tb.id = bl.to_book_id
		 JOIN users u ON u.id = e.user_id
		 LEFT JOIN users ru ON ru.id = e.reviewer_id
		 WHERE e.status = $1
		 ORDER BY e.created_at ASC`,
		statusFilter,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer rows.Close()

	edits := []editResponse{}
	for rows.Next() {
		var e editResponse
		var createdAt time.Time
		var reviewedAt *time.Time
		if err := rows.Scan(
			&e.ID, &e.BookLinkID,
			&e.Username, &e.DisplayName,
			&e.ProposedType, &e.ProposedNote,
			&e.CurrentType, &e.CurrentNote,
			&e.FromBookOLID, &e.FromBookTitle,
			&e.ToBookOLID, &e.ToBookTitle,
			&e.Status,
			&e.ReviewerName,
			&e.ReviewerComment,
			&createdAt, &reviewedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		e.CreatedAt = createdAt.Format(time.RFC3339)
		if reviewedAt != nil {
			formatted := reviewedAt.Format(time.RFC3339)
			e.ReviewedAt = &formatted
		}
		edits = append(edits, e)
	}

	c.JSON(http.StatusOK, edits)
}

// PUT /admin/link-edits/:editId  (moderator only)
func (h *Handler) ReviewEdit(c *gin.Context) {
	editID := c.Param("editId")
	reviewerID := c.GetString(middleware.UserIDKey)

	var body struct {
		Action  string  `json:"action" binding:"required"`
		Comment *string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "action is required (approve or reject)"})
		return
	}

	if body.Action != "approve" && body.Action != "reject" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "action must be approve or reject"})
		return
	}

	ctx := c.Request.Context()
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer tx.Rollback(ctx)

	// Fetch the pending edit.
	var bookLinkID string
	var proposedType *string
	var proposedNote *string
	err = tx.QueryRow(ctx,
		`SELECT book_link_id, proposed_type, proposed_note
		 FROM book_link_edits
		 WHERE id = $1 AND status = 'pending'`,
		editID,
	).Scan(&bookLinkID, &proposedType, &proposedNote)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "pending edit not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	newStatus := "rejected"
	if body.Action == "approve" {
		newStatus = "approved"

		// Apply the proposed changes to the link.
		if proposedType != nil {
			_, err = tx.Exec(ctx,
				`UPDATE book_links SET link_type = $1 WHERE id = $2`,
				*proposedType, bookLinkID,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
				return
			}
		}
		if proposedNote != nil {
			_, err = tx.Exec(ctx,
				`UPDATE book_links SET note = $1 WHERE id = $2`,
				*proposedNote, bookLinkID,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
				return
			}
		}
	}

	// Mark edit as reviewed.
	_, err = tx.Exec(ctx,
		`UPDATE book_link_edits
		 SET status = $1, reviewer_id = $2, reviewer_comment = $3, reviewed_at = NOW()
		 WHERE id = $4`,
		newStatus, reviewerID, body.Comment, editID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "status": newStatus})
}
