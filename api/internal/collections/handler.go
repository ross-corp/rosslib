package collections

import (
	"context"
	"net/http"

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

// ── types ─────────────────────────────────────────────────────────────────────

type shelfBook struct {
	BookID       string  `json:"book_id"`
	OpenLibraryID string `json:"open_library_id"`
	Title        string  `json:"title"`
	CoverURL     *string `json:"cover_url"`
}

type shelfResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Slug           string `json:"slug"`
	ExclusiveGroup string `json:"exclusive_group"`
	ItemCount      int    `json:"item_count"`
}

type myShelfResponse struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Slug           string      `json:"slug"`
	ExclusiveGroup string      `json:"exclusive_group"`
	Books          []shelfBook `json:"books"`
}

// ── handlers ──────────────────────────────────────────────────────────────────

// GetUserShelves - GET /users/:username/shelves
// Public. Returns shelf metadata with item counts.
func (h *Handler) GetUserShelves(c *gin.Context) {
	username := c.Param("username")

	rows, err := h.pool.Query(c.Request.Context(),
		`SELECT c.id, c.name, c.slug, COALESCE(c.exclusive_group, ''), COUNT(ci.id) AS item_count
		 FROM collections c
		 JOIN users u ON u.id = c.user_id
		 LEFT JOIN collection_items ci ON ci.collection_id = c.id
		 WHERE u.username = $1 AND u.deleted_at IS NULL
		 GROUP BY c.id, c.name, c.slug, c.exclusive_group
		 ORDER BY c.created_at`,
		username,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer rows.Close()

	shelves := []shelfResponse{}
	for rows.Next() {
		var s shelfResponse
		if err := rows.Scan(&s.ID, &s.Name, &s.Slug, &s.ExclusiveGroup, &s.ItemCount); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		shelves = append(shelves, s)
	}

	c.JSON(http.StatusOK, shelves)
}

// ensureDefaultShelves creates the 3 default read-status shelves for a user
// if they don't have any collections yet (e.g. accounts created before migration).
func (h *Handler) ensureDefaultShelves(ctx context.Context, userID string) error {
	var count int
	err := h.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM collections WHERE user_id = $1`,
		userID,
	).Scan(&count)
	if err != nil || count > 0 {
		return err
	}

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	defaults := []struct{ name, slug string }{
		{"Want to Read", "want-to-read"},
		{"Currently Reading", "currently-reading"},
		{"Read", "read"},
	}
	for _, s := range defaults {
		_, err = tx.Exec(ctx,
			`INSERT INTO collections (user_id, name, slug, is_exclusive, exclusive_group)
			 VALUES ($1, $2, $3, true, 'read_status')
			 ON CONFLICT DO NOTHING`,
			userID, s.name, s.slug,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// GetMyShelves - GET /me/shelves (authed)
// Returns the current user's shelves with full book lists (OL IDs included).
func (h *Handler) GetMyShelves(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	if err := h.ensureDefaultShelves(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	shelfRows, err := h.pool.Query(c.Request.Context(),
		`SELECT id, name, slug, COALESCE(exclusive_group, '')
		 FROM collections
		 WHERE user_id = $1
		 ORDER BY created_at`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer shelfRows.Close()

	shelves := []myShelfResponse{}
	shelfIDs := []string{}
	idToIdx := map[string]int{}

	for shelfRows.Next() {
		var s myShelfResponse
		if err := shelfRows.Scan(&s.ID, &s.Name, &s.Slug, &s.ExclusiveGroup); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		s.Books = []shelfBook{}
		idToIdx[s.ID] = len(shelves)
		shelves = append(shelves, s)
		shelfIDs = append(shelfIDs, s.ID)
	}
	shelfRows.Close()

	if len(shelves) == 0 {
		c.JSON(http.StatusOK, shelves)
		return
	}

	bookRows, err := h.pool.Query(c.Request.Context(),
		`SELECT ci.collection_id, b.id, b.open_library_id, b.title, b.cover_url
		 FROM collection_items ci
		 JOIN books b ON b.id = ci.book_id
		 WHERE ci.collection_id = ANY($1)
		 ORDER BY ci.added_at DESC`,
		shelfIDs,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer bookRows.Close()

	for bookRows.Next() {
		var collID string
		var book shelfBook
		if err := bookRows.Scan(&collID, &book.BookID, &book.OpenLibraryID, &book.Title, &book.CoverURL); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		if idx, ok := idToIdx[collID]; ok {
			shelves[idx].Books = append(shelves[idx].Books, book)
		}
	}

	c.JSON(http.StatusOK, shelves)
}

// AddBookToShelf - POST /shelves/:shelfId/books (authed)
// Upserts the book by open_library_id, enforces mutual exclusivity within the
// shelf's exclusive_group, and adds the book to the shelf.
func (h *Handler) AddBookToShelf(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	shelfID := c.Param("shelfId")

	var req struct {
		OpenLibraryID string  `json:"open_library_id" binding:"required"`
		Title         string  `json:"title"           binding:"required"`
		CoverURL      *string `json:"cover_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var exclusiveGroup string
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT COALESCE(exclusive_group, '') FROM collections WHERE id = $1 AND user_id = $2`,
		shelfID, userID,
	).Scan(&exclusiveGroup)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "shelf not found"})
		return
	}

	var bookID string
	err = h.pool.QueryRow(c.Request.Context(),
		`INSERT INTO books (open_library_id, title, cover_url)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (open_library_id) DO UPDATE
		   SET title = EXCLUDED.title, cover_url = EXCLUDED.cover_url
		 RETURNING id`,
		req.OpenLibraryID, req.Title, req.CoverURL,
	).Scan(&bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if exclusiveGroup != "" {
		_, err = h.pool.Exec(c.Request.Context(),
			`DELETE FROM collection_items ci
			 USING collections col
			 WHERE ci.collection_id = col.id
			   AND col.user_id = $1
			   AND col.exclusive_group = $2
			   AND ci.book_id = $3
			   AND ci.collection_id != $4`,
			userID, exclusiveGroup, bookID, shelfID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
	}

	_, err = h.pool.Exec(c.Request.Context(),
		`INSERT INTO collection_items (collection_id, book_id)
		 VALUES ($1, $2)
		 ON CONFLICT DO NOTHING`,
		shelfID, bookID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// RemoveBookFromShelf - DELETE /shelves/:shelfId/books/:olId (authed)
func (h *Handler) RemoveBookFromShelf(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	shelfID := c.Param("shelfId")
	olID := c.Param("olId")

	var exists bool
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT EXISTS(SELECT 1 FROM collections WHERE id = $1 AND user_id = $2)`,
		shelfID, userID,
	).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "shelf not found"})
		return
	}

	_, err = h.pool.Exec(c.Request.Context(),
		`DELETE FROM collection_items ci
		 USING books b
		 WHERE ci.collection_id = $1
		   AND ci.book_id = b.id
		   AND b.open_library_id = $2`,
		shelfID, olID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
