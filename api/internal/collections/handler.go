package collections

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tristansaldanha/rosslib/api/internal/activity"
	"github.com/tristansaldanha/rosslib/api/internal/middleware"
	"github.com/tristansaldanha/rosslib/api/internal/privacy"
)

var multiDash = regexp.MustCompile(`-{2,}`)

// slugify converts a human-readable name into a URL slug.
// "Science Fiction" → "science-fiction"
func slugify(name string) string {
	s := strings.ToLower(name)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}
	return strings.Trim(multiDash.ReplaceAllString(b.String(), "-"), "-")
}

// tagSlugify converts a tag path into a slug-safe path, preserving "/" separators.
// "Science Fiction/Moon Landing" → "science-fiction/moon-landing"
func tagSlugify(path string) string {
	segments := strings.Split(path, "/")
	for i, seg := range segments {
		segments[i] = slugify(seg)
	}
	// Remove empty segments
	out := make([]string, 0, len(segments))
	for _, seg := range segments {
		if seg != "" {
			out = append(out, seg)
		}
	}
	return strings.Join(out, "/")
}

// EnsureShelf is a package-level helper used by the import process.
// It inserts the collection if the slug doesn't exist yet, otherwise
// returns the existing collection's ID — no error in either case.
func EnsureShelf(ctx context.Context, pool *pgxpool.Pool, userID, name, slug string, isExclusive bool, exclusiveGroup string, isPublic bool, collectionType string) (string, error) {
	if collectionType == "" {
		collectionType = "shelf"
	}
	var id string
	err := pool.QueryRow(ctx,
		`INSERT INTO collections (user_id, name, slug, is_exclusive, exclusive_group, is_public, collection_type)
		 VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, $7)
		 ON CONFLICT (user_id, slug) DO UPDATE SET name = collections.name
		 RETURNING id`,
		userID, name, slug, isExclusive, exclusiveGroup, isPublic, collectionType,
	).Scan(&id)
	return id, err
}

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// ── types ─────────────────────────────────────────────────────────────────────

type shelfBook struct {
	BookID        string  `json:"book_id"`
	OpenLibraryID string  `json:"open_library_id"`
	Title         string  `json:"title"`
	CoverURL      *string `json:"cover_url"`
}

type shelfDetailBook struct {
	BookID        string  `json:"book_id"`
	OpenLibraryID string  `json:"open_library_id"`
	Title         string  `json:"title"`
	CoverURL      *string `json:"cover_url"`
	AddedAt       string  `json:"added_at"`
	Rating        *int    `json:"rating"`
}

type shelfDetailResponse struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Slug           string            `json:"slug"`
	ExclusiveGroup string            `json:"exclusive_group"`
	CollectionType string            `json:"collection_type"`
	Books          []shelfDetailBook `json:"books"`
}

type shelfResponse struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Slug           string            `json:"slug"`
	ExclusiveGroup string            `json:"exclusive_group"`
	CollectionType string            `json:"collection_type"`
	ItemCount      int               `json:"item_count"`
	Books          []shelfDetailBook `json:"books,omitempty"`
}

type myShelfResponse struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Slug           string      `json:"slug"`
	ExclusiveGroup string      `json:"exclusive_group"`
	CollectionType string      `json:"collection_type"`
	Books          []shelfBook `json:"books"`
}

// ── handlers ──────────────────────────────────────────────────────────────────

// GetUserShelves - GET /users/:username/shelves
// Public, but gated for private profiles.
// Optional ?include_books=N query param to include the first N books per shelf.
func (h *Handler) GetUserShelves(c *gin.Context) {
	username := c.Param("username")
	currentUserID := c.GetString(middleware.UserIDKey)

	userID, _, canView := privacy.CanViewProfile(c.Request.Context(), h.pool, username, currentUserID)
	if !canView {
		c.JSON(http.StatusOK, []shelfResponse{})
		return
	}

	if userID != "" {
		h.ensureDefaultFavorites(c.Request.Context(), userID)
	}

	rows, err := h.pool.Query(c.Request.Context(),
		`SELECT c.id, c.name, c.slug, COALESCE(c.exclusive_group, ''), c.collection_type, COUNT(ci.id) AS item_count
		 FROM collections c
		 JOIN users u ON u.id = c.user_id
		 LEFT JOIN collection_items ci ON ci.collection_id = c.id
		 WHERE u.username = $1 AND u.deleted_at IS NULL
		 GROUP BY c.id, c.name, c.slug, c.exclusive_group, c.collection_type
		 ORDER BY c.created_at`,
		username,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer rows.Close()

	shelves := []shelfResponse{}
	shelfIDs := []string{}
	idToIdx := map[string]int{}
	for rows.Next() {
		var s shelfResponse
		if err := rows.Scan(&s.ID, &s.Name, &s.Slug, &s.ExclusiveGroup, &s.CollectionType, &s.ItemCount); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		idToIdx[s.ID] = len(shelves)
		shelves = append(shelves, s)
		shelfIDs = append(shelfIDs, s.ID)
	}
	rows.Close()

	// If include_books param is set, fetch preview books per shelf
	if ibStr := c.Query("include_books"); ibStr != "" && len(shelfIDs) > 0 {
		limit, err := strconv.Atoi(ibStr)
		if err == nil && limit >= 1 && limit <= 200 {
			bookRows, err := h.pool.Query(c.Request.Context(),
				`SELECT collection_id, book_id, open_library_id, title, cover_url, added_at, rating FROM (
					SELECT ci.collection_id, b.id AS book_id, b.open_library_id, b.title, b.cover_url, ci.added_at, ci.rating,
					       ROW_NUMBER() OVER (PARTITION BY ci.collection_id ORDER BY ci.added_at DESC) AS rn
					FROM collection_items ci JOIN books b ON b.id = ci.book_id
					WHERE ci.collection_id = ANY($1)
				) sub WHERE sub.rn <= $2`,
				shelfIDs, limit,
			)
			if err == nil {
				defer bookRows.Close()
				for bookRows.Next() {
					var collID string
					var book shelfDetailBook
					var addedAt time.Time
					if err := bookRows.Scan(&collID, &book.BookID, &book.OpenLibraryID, &book.Title, &book.CoverURL, &addedAt, &book.Rating); err != nil {
						break
					}
					book.AddedAt = addedAt.Format(time.RFC3339)
					if idx, ok := idToIdx[collID]; ok {
						if shelves[idx].Books == nil {
							shelves[idx].Books = []shelfDetailBook{}
						}
						shelves[idx].Books = append(shelves[idx].Books, book)
					}
				}
			}
		}
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
			`INSERT INTO collections (user_id, name, slug, is_exclusive, exclusive_group, collection_type)
			 VALUES ($1, $2, $3, true, 'read_status', 'shelf')
			 ON CONFLICT DO NOTHING`,
			userID, s.name, s.slug,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// ensureDefaultFavorites creates the "Favorites" default tag for a user if it doesn't exist.
func (h *Handler) ensureDefaultFavorites(ctx context.Context, userID string) {
	_, _ = h.pool.Exec(ctx,
		`INSERT INTO collections (user_id, name, slug, is_exclusive, collection_type)
		 VALUES ($1, 'Favorites', 'favorites', false, 'tag')
		 ON CONFLICT (user_id, slug) DO NOTHING`,
		userID,
	)
}

// GetMyShelves - GET /me/shelves (authed)
// Returns the current user's shelves with full book lists (OL IDs included).
func (h *Handler) GetMyShelves(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	if err := h.ensureDefaultShelves(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	h.ensureDefaultFavorites(c.Request.Context(), userID)

	shelfRows, err := h.pool.Query(c.Request.Context(),
		`SELECT id, name, slug, COALESCE(exclusive_group, ''), collection_type
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
		if err := shelfRows.Scan(&s.ID, &s.Name, &s.Slug, &s.ExclusiveGroup, &s.CollectionType); err != nil {
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

	// Before removing from exclusive group, check what shelf the book is currently on
	// so we can detect started/finished transitions.
	var prevShelfSlug *string
	if exclusiveGroup != "" {
		var slug string
		lookupErr := h.pool.QueryRow(c.Request.Context(),
			`SELECT col.slug FROM collection_items ci
			 JOIN collections col ON col.id = ci.collection_id
			 WHERE col.user_id = $1
			   AND col.exclusive_group = $2
			   AND ci.book_id = $3`,
			userID, exclusiveGroup, bookID,
		).Scan(&slug)
		if lookupErr == nil {
			prevShelfSlug = &slug
		}

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

	// Look up the target shelf's slug and name for activity recording.
	var shelfName, shelfSlug string
	_ = h.pool.QueryRow(c.Request.Context(),
		`SELECT name, slug FROM collections WHERE id = $1`, shelfID,
	).Scan(&shelfName, &shelfSlug)

	// Determine activity type based on shelf transition.
	activityType := "shelved"
	if exclusiveGroup == "read_status" && (prevShelfSlug == nil || *prevShelfSlug != shelfSlug) {
		switch shelfSlug {
		case "currently-reading":
			activityType = "started_book"
		case "read":
			activityType = "finished_book"
		}
	}
	activity.Record(c.Request.Context(), h.pool, userID, activityType, &bookID, nil, &shelfID, nil,
		map[string]string{"shelf_name": shelfName})

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GetShelfBySlug - GET /users/:username/shelves/:slug
// Public, but gated for private profiles.
func (h *Handler) GetShelfBySlug(c *gin.Context) {
	username := c.Param("username")
	slug := c.Param("slug")
	currentUserID := c.GetString(middleware.UserIDKey)

	_, _, canView := privacy.CanViewProfile(c.Request.Context(), h.pool, username, currentUserID)
	if !canView {
		c.JSON(http.StatusNotFound, gin.H{"error": "shelf not found"})
		return
	}

	var shelf shelfDetailResponse
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT c.id, c.name, c.slug, COALESCE(c.exclusive_group, ''), c.collection_type
		 FROM collections c
		 JOIN users u ON u.id = c.user_id
		 WHERE u.username = $1 AND u.deleted_at IS NULL AND c.slug = $2`,
		username, slug,
	).Scan(&shelf.ID, &shelf.Name, &shelf.Slug, &shelf.ExclusiveGroup, &shelf.CollectionType)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "shelf not found"})
		return
	}

	query := `SELECT b.id, b.open_library_id, b.title, b.cover_url, ci.added_at, ci.rating
		 FROM collection_items ci
		 JOIN books b ON b.id = ci.book_id
		 WHERE ci.collection_id = $1
		 ORDER BY ci.added_at DESC`
	args := []interface{}{shelf.ID}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit >= 1 && limit <= 200 {
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

	shelf.Books = []shelfDetailBook{}
	for rows.Next() {
		var book shelfDetailBook
		var addedAt time.Time
		if err := rows.Scan(&book.BookID, &book.OpenLibraryID, &book.Title, &book.CoverURL, &addedAt, &book.Rating); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		book.AddedAt = addedAt.Format(time.RFC3339)
		shelf.Books = append(shelf.Books, book)
	}

	c.JSON(http.StatusOK, shelf)
}

// GetTagBooks - GET /users/:username/tags/*path
// Public. Returns all books tagged with the given tag path or any sub-tag.
// E.g. /users/alice/tags/sci-fi returns books tagged "sci-fi" or "sci-fi/moon" etc.
type tagBooksResponse struct {
	Path    string            `json:"path"`
	SubTags []string          `json:"sub_tags"`
	Books   []shelfDetailBook `json:"books"`
}

func (h *Handler) GetTagBooks(c *gin.Context) {
	username := c.Param("username")
	currentUserID := c.GetString(middleware.UserIDKey)

	_, _, canView := privacy.CanViewProfile(c.Request.Context(), h.pool, username, currentUserID)
	if !canView {
		c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
		return
	}

	// Gin wildcard params include a leading slash: "/sci-fi" or "/sci-fi/moon"
	rawPath := strings.TrimPrefix(c.Param("path"), "/")
	if rawPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tag path required"})
		return
	}

	// Find all matching tag collections for this user: slug = path OR slug LIKE path/%
	tagRows, err := h.pool.Query(c.Request.Context(),
		`SELECT c.id, c.slug
		 FROM collections c
		 JOIN users u ON u.id = c.user_id
		 WHERE u.username = $1
		   AND u.deleted_at IS NULL
		   AND c.collection_type = 'tag'
		   AND (c.slug = $2 OR c.slug LIKE $3)`,
		username, rawPath, rawPath+"/%",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer tagRows.Close()

	type tagEntry struct {
		id   string
		slug string
	}
	var tags []tagEntry
	for tagRows.Next() {
		var t tagEntry
		if err := tagRows.Scan(&t.id, &t.slug); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		tags = append(tags, t)
	}
	tagRows.Close()

	if len(tags) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
		return
	}

	// Collect all matching collection IDs and sub-tag slugs (direct children only)
	collectionIDs := make([]string, 0, len(tags))
	subTagSet := map[string]struct{}{}
	for _, t := range tags {
		collectionIDs = append(collectionIDs, t.id)
		// If slug is deeper than rawPath, derive its direct child segment
		if t.slug != rawPath {
			rest := strings.TrimPrefix(t.slug, rawPath+"/")
			// Take only the first extra segment
			parts := strings.SplitN(rest, "/", 2)
			subTagSet[rawPath+"/"+parts[0]] = struct{}{}
		}
	}

	subTags := make([]string, 0, len(subTagSet))
	for k := range subTagSet {
		subTags = append(subTags, k)
	}

	// Fetch books from all matching collections (deduplicated by book_id)
	bookRows, err := h.pool.Query(c.Request.Context(),
		`SELECT DISTINCT ON (b.id) b.id, b.open_library_id, b.title, b.cover_url, ci.added_at, ci.rating
		 FROM collection_items ci
		 JOIN books b ON b.id = ci.book_id
		 WHERE ci.collection_id = ANY($1)
		 ORDER BY b.id, ci.added_at DESC`,
		collectionIDs,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer bookRows.Close()

	books := []shelfDetailBook{}
	for bookRows.Next() {
		var book shelfDetailBook
		var addedAt time.Time
		if err := bookRows.Scan(&book.BookID, &book.OpenLibraryID, &book.Title, &book.CoverURL, &addedAt, &book.Rating); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		book.AddedAt = addedAt.Format(time.RFC3339)
		books = append(books, book)
	}

	c.JSON(http.StatusOK, tagBooksResponse{
		Path:    rawPath,
		SubTags: subTags,
		Books:   books,
	})
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

// ── Custom shelf management ───────────────────────────────────────────────────

// CreateShelf - POST /me/shelves (authed)
// Creates a new custom collection. For type="tag", the name may include "/"
// to express hierarchy (e.g. "sci-fi/moon"). The slug is derived from the name.
// Returns 409 if the user already has a shelf with that slug.
func (h *Handler) CreateShelf(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	var req struct {
		Name           string `json:"name"            binding:"required,max=255"`
		IsExclusive    bool   `json:"is_exclusive"`
		ExclusiveGroup string `json:"exclusive_group" binding:"max=100"`
		IsPublic       *bool  `json:"is_public"`
		Type           string `json:"type"` // "shelf" (default) or "tag"
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collectionType := req.Type
	if collectionType == "" {
		collectionType = "shelf"
	}
	if collectionType != "shelf" && collectionType != "tag" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type must be 'shelf' or 'tag'"})
		return
	}

	var slug string
	if collectionType == "tag" {
		slug = tagSlugify(req.Name)
	} else {
		slug = slugify(req.Name)
	}
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name produces an empty slug"})
		return
	}

	isPublic := true
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}

	var existing int
	_ = h.pool.QueryRow(c.Request.Context(),
		`SELECT COUNT(*) FROM collections WHERE user_id = $1 AND slug = $2`,
		userID, slug,
	).Scan(&existing)
	if existing > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "a shelf with that name already exists"})
		return
	}

	var shelf shelfResponse
	err := h.pool.QueryRow(c.Request.Context(),
		`INSERT INTO collections (user_id, name, slug, is_exclusive, exclusive_group, is_public, collection_type)
		 VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, $7)
		 RETURNING id, name, slug, COALESCE(exclusive_group, ''), collection_type`,
		userID, req.Name, slug, req.IsExclusive, req.ExclusiveGroup, isPublic, collectionType,
	).Scan(&shelf.ID, &shelf.Name, &shelf.Slug, &shelf.ExclusiveGroup, &shelf.CollectionType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusCreated, shelf)
}

// UpdateShelf - PATCH /me/shelves/:id (authed)
// Allows renaming a shelf or toggling its visibility.
func (h *Handler) UpdateShelf(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	shelfID := c.Param("id")

	var req struct {
		Name     *string `json:"name"      binding:"omitempty,max=255"`
		IsPublic *bool   `json:"is_public"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Name == nil && req.IsPublic == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provide at least one of: name, is_public"})
		return
	}

	var curName string
	var curIsPublic bool
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT name, is_public FROM collections WHERE id = $1 AND user_id = $2`,
		shelfID, userID,
	).Scan(&curName, &curIsPublic)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "shelf not found"})
		return
	}

	newName := curName
	if req.Name != nil {
		newName = *req.Name
	}
	newIsPublic := curIsPublic
	if req.IsPublic != nil {
		newIsPublic = *req.IsPublic
	}

	var shelf shelfResponse
	err = h.pool.QueryRow(c.Request.Context(),
		`UPDATE collections SET name = $1, is_public = $2
		 WHERE id = $3
		 RETURNING id, name, slug, COALESCE(exclusive_group, ''), collection_type`,
		newName, newIsPublic, shelfID,
	).Scan(&shelf.ID, &shelf.Name, &shelf.Slug, &shelf.ExclusiveGroup, &shelf.CollectionType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, shelf)
}

// DeleteShelf - DELETE /me/shelves/:id (authed)
// Deletes a custom collection and all its items.
// The 3 default shelves (exclusive_group = 'read_status') cannot be deleted.
func (h *Handler) DeleteShelf(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	shelfID := c.Param("id")

	var exclusiveGroup string
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT COALESCE(exclusive_group, '') FROM collections WHERE id = $1 AND user_id = $2`,
		shelfID, userID,
	).Scan(&exclusiveGroup)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "shelf not found"})
		return
	}
	if exclusiveGroup == "read_status" {
		c.JSON(http.StatusForbidden, gin.H{"error": "default shelves cannot be deleted"})
		return
	}

	tx, err := h.pool.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer tx.Rollback(c.Request.Context()) //nolint:errcheck

	_, err = tx.Exec(c.Request.Context(),
		`DELETE FROM collection_items WHERE collection_id = $1`, shelfID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	_, err = tx.Exec(c.Request.Context(),
		`DELETE FROM collections WHERE id = $1`, shelfID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ── Review / rating ───────────────────────────────────────────────────────────

// UpdateBookInShelf - PATCH /shelves/:shelfId/books/:olId (authed)
// Updates review metadata on a collection_item: rating, review_text, spoiler,
// date_read. Only fields present in the JSON body are written; absent fields
// are left unchanged. Send null to clear a field.
//
// Accepted fields:
//
//	rating      int | null   (1–5; null clears it)
//	review_text string | null
//	spoiler     bool
//	date_read   string | null  (RFC3339 or "YYYY-MM-DD"; null clears it)
func (h *Handler) UpdateBookInShelf(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	shelfID := c.Param("shelfId")
	olID := c.Param("olId")

	// Decode to a raw map so we know which keys the caller actually sent.
	var raw map[string]json.RawMessage
	if err := c.ShouldBindJSON(&raw); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(raw) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	// Verify the book is in this shelf and the shelf belongs to the user.
	var itemID string
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT ci.id
		 FROM collection_items ci
		 JOIN books b       ON b.id  = ci.book_id
		 JOIN collections c ON c.id  = ci.collection_id
		 WHERE ci.collection_id = $1
		   AND b.open_library_id = $2
		   AND c.user_id         = $3`,
		shelfID, olID, userID,
	).Scan(&itemID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found in shelf"})
		return
	}

	setClauses := []string{}
	args := []interface{}{}
	idx := 1

	if v, ok := raw["rating"]; ok {
		var rating *int
		if err := json.Unmarshal(v, &rating); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "rating must be an integer or null"})
			return
		}
		if rating != nil && (*rating < 1 || *rating > 5) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "rating must be between 1 and 5"})
			return
		}
		setClauses = append(setClauses, fmt.Sprintf("rating = $%d", idx))
		args = append(args, rating)
		idx++
	}

	if v, ok := raw["review_text"]; ok {
		var text *string
		if err := json.Unmarshal(v, &text); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "review_text must be a string or null"})
			return
		}
		setClauses = append(setClauses, fmt.Sprintf("review_text = $%d", idx))
		args = append(args, text)
		idx++
	}

	if v, ok := raw["spoiler"]; ok {
		var spoiler bool
		if err := json.Unmarshal(v, &spoiler); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "spoiler must be a boolean"})
			return
		}
		setClauses = append(setClauses, fmt.Sprintf("spoiler = $%d", idx))
		args = append(args, spoiler)
		idx++
	}

	if v, ok := raw["date_read"]; ok {
		var dateStr *string
		if err := json.Unmarshal(v, &dateStr); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "date_read must be a string or null"})
			return
		}
		var dateVal interface{}
		if dateStr != nil && *dateStr != "" {
			t, err := time.Parse(time.RFC3339, *dateStr)
			if err != nil {
				t, err = time.Parse("2006-01-02", *dateStr)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "date_read must be RFC3339 or YYYY-MM-DD"})
					return
				}
			}
			dateVal = t
		}
		setClauses = append(setClauses, fmt.Sprintf("date_read = $%d", idx))
		args = append(args, dateVal)
		idx++
	}

	if len(setClauses) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no recognised fields to update"})
		return
	}

	args = append(args, itemID)
	query := fmt.Sprintf(
		"UPDATE collection_items SET %s WHERE id = $%d",
		strings.Join(setClauses, ", "), idx,
	)
	if _, err := h.pool.Exec(c.Request.Context(), query, args...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// Record activity for rating or review updates.
	if _, hasRating := raw["rating"]; hasRating {
		var bookID string
		_ = h.pool.QueryRow(c.Request.Context(),
			`SELECT b.id FROM books b WHERE b.open_library_id = $1`, olID,
		).Scan(&bookID)
		if bookID != "" {
			meta := map[string]string{}
			if v, ok := raw["rating"]; ok {
				meta["rating"] = strings.Trim(string(v), "\"")
			}
			activity.Record(c.Request.Context(), h.pool, userID, "rated", &bookID, nil, &shelfID, nil, meta)
		}
	}
	if _, hasReview := raw["review_text"]; hasReview {
		var bookID string
		_ = h.pool.QueryRow(c.Request.Context(),
			`SELECT b.id FROM books b WHERE b.open_library_id = $1`, olID,
		).Scan(&bookID)
		if bookID != "" {
			meta := map[string]string{}
			var snippet string
			var text *string
			_ = json.Unmarshal(raw["review_text"], &text)
			if text != nil && len(*text) > 100 {
				snippet = (*text)[:100] + "..."
			} else if text != nil {
				snippet = *text
			}
			if snippet != "" {
				meta["review_snippet"] = snippet
			}
			activity.Record(c.Request.Context(), h.pool, userID, "reviewed", &bookID, nil, &shelfID, nil, meta)
		}
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ── Export ─────────────────────────────────────────────────────────────────────

// ExportCSV - GET /me/export/csv (authed)
// Exports the user's library as a CSV file. Optionally filter to a single shelf
// with ?shelf=<id>. Columns: Title, Author, ISBN13, Collection, Rating,
// Review, Date Added, Date Read.
func (h *Handler) ExportCSV(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	shelfFilter := c.Query("shelf")

	query := `SELECT c.name, b.title, b.authors, b.isbn13,
	                  ci.rating, ci.review_text, ci.date_added, ci.date_read
	           FROM collection_items ci
	           JOIN books b ON b.id = ci.book_id
	           JOIN collections c ON c.id = ci.collection_id
	           WHERE c.user_id = $1`
	args := []interface{}{userID}

	if shelfFilter != "" {
		query += ` AND c.id = $2`
		args = append(args, shelfFilter)
	}
	query += ` ORDER BY c.name, ci.added_at DESC`

	rows, err := h.pool.Query(c.Request.Context(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer rows.Close()

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", `attachment; filename="rosslib-export.csv"`)

	w := csv.NewWriter(c.Writer)
	w.Write([]string{"Title", "Author", "ISBN13", "Collection", "Rating", "Review", "Date Added", "Date Read"})

	for rows.Next() {
		var (
			collName   string
			title      string
			authors    *string
			isbn13     *string
			rating     *int
			reviewText *string
			dateAdded  *time.Time
			dateRead   *time.Time
		)
		if err := rows.Scan(&collName, &title, &authors, &isbn13, &rating, &reviewText, &dateAdded, &dateRead); err != nil {
			break
		}

		record := []string{
			title,
			derefStr(authors),
			derefStr(isbn13),
			collName,
			formatRating(rating),
			derefStr(reviewText),
			formatDate(dateAdded),
			formatDate(dateRead),
		}
		w.Write(record)
	}

	w.Flush()
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func formatRating(r *int) string {
	if r == nil {
		return ""
	}
	return fmt.Sprintf("%d", *r)
}

func formatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}
