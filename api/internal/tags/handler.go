package tags

import (
	"context"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tristansaldanha/rosslib/api/internal/middleware"
	"github.com/tristansaldanha/rosslib/api/internal/privacy"
)

var multiDash = regexp.MustCompile(`-{2,}`)

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

// slugifyValue is like slugify but preserves "/" as a path separator,
// enabling nested label values like "history/engineering".
func slugifyValue(name string) string {
	parts := strings.Split(name, "/")
	segs := make([]string, 0, len(parts))
	for _, p := range parts {
		s := slugify(strings.TrimSpace(p))
		if s != "" {
			segs = append(segs, s)
		}
	}
	return strings.Join(segs, "/")
}

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// ── Response types ─────────────────────────────────────────────────────────────

type tagValueResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type tagKeyResponse struct {
	ID     string             `json:"id"`
	Name   string             `json:"name"`
	Slug   string             `json:"slug"`
	Mode   string             `json:"mode"` // "select_one" | "select_multiple"
	Values []tagValueResponse `json:"values"`
}

type bookTagResponse struct {
	KeyID     string `json:"key_id"`
	KeyName   string `json:"key_name"`
	KeySlug   string `json:"key_slug"`
	ValueID   string `json:"value_id"`
	ValueName string `json:"value_name"`
	ValueSlug string `json:"value_slug"`
}

// ── Tag key management ─────────────────────────────────────────────────────────

// ensureDefaultStatusLabel creates a "Status" tag key with the five standard
// read-status values for any user who doesn't have one yet. Idempotent.
func (h *Handler) ensureDefaultStatusLabel(ctx context.Context, userID string) error {
	return EnsureStatusLabel(ctx, h.pool, userID)
}

// EnsureStatusLabel is the exported version so other packages (imports, userbooks)
// can call it directly. Creates a "Status" tag key with the five standard
// read-status values if the user doesn't have one. Idempotent.
func EnsureStatusLabel(ctx context.Context, pool *pgxpool.Pool, userID string) error {
	var exists bool
	if err := pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM tag_keys WHERE user_id = $1 AND slug = 'status')`,
		userID,
	).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return nil
	}

	var keyID string
	if err := pool.QueryRow(ctx,
		`INSERT INTO tag_keys (user_id, name, slug, mode)
		 VALUES ($1, 'Status', 'status', 'select_one')
		 ON CONFLICT (user_id, slug) DO UPDATE SET name = tag_keys.name
		 RETURNING id`,
		userID,
	).Scan(&keyID); err != nil {
		return err
	}

	defaults := []struct{ name, slug string }{
		{"Want to Read", "want-to-read"},
		{"Owned to Read", "owned-to-read"},
		{"Currently Reading", "currently-reading"},
		{"Finished", "finished"},
		{"DNF", "dnf"},
	}
	for _, v := range defaults {
		if _, err := pool.Exec(ctx,
			`INSERT INTO tag_values (tag_key_id, name, slug)
			 VALUES ($1, $2, $3)
			 ON CONFLICT DO NOTHING`,
			keyID, v.name, v.slug,
		); err != nil {
			return err
		}
	}
	return nil
}

// ListTagKeys - GET /me/tag-keys
func (h *Handler) ListTagKeys(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	if err := h.ensureDefaultStatusLabel(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	keyRows, err := h.pool.Query(c.Request.Context(),
		`SELECT id, name, slug, mode FROM tag_keys WHERE user_id = $1 ORDER BY created_at`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer keyRows.Close()

	keys := []tagKeyResponse{}
	keyIDs := []string{}
	idToIdx := map[string]int{}

	for keyRows.Next() {
		var k tagKeyResponse
		if err := keyRows.Scan(&k.ID, &k.Name, &k.Slug, &k.Mode); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		k.Values = []tagValueResponse{}
		idToIdx[k.ID] = len(keys)
		keyIDs = append(keyIDs, k.ID)
		keys = append(keys, k)
	}
	keyRows.Close()

	if len(keyIDs) > 0 {
		valRows, err := h.pool.Query(c.Request.Context(),
			`SELECT id, tag_key_id, name, slug FROM tag_values WHERE tag_key_id = ANY($1) ORDER BY created_at`,
			keyIDs,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		defer valRows.Close()

		for valRows.Next() {
			var v tagValueResponse
			var keyID string
			if err := valRows.Scan(&v.ID, &keyID, &v.Name, &v.Slug); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
				return
			}
			if idx, ok := idToIdx[keyID]; ok {
				keys[idx].Values = append(keys[idx].Values, v)
			}
		}
	}

	c.JSON(http.StatusOK, keys)
}

// CreateTagKey - POST /me/tag-keys
// Body: { "name": "...", "mode": "select_one" | "select_multiple" }
func (h *Handler) CreateTagKey(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	var req struct {
		Name string `json:"name" binding:"required,max=100"`
		Mode string `json:"mode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mode := req.Mode
	if mode == "" {
		mode = "select_one"
	}
	if mode != "select_one" && mode != "select_multiple" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mode must be 'select_one' or 'select_multiple'"})
		return
	}

	slug := slugify(req.Name)
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name produces an empty slug"})
		return
	}

	var key tagKeyResponse
	err := h.pool.QueryRow(c.Request.Context(),
		`INSERT INTO tag_keys (user_id, name, slug, mode) VALUES ($1, $2, $3, $4)
		 RETURNING id, name, slug, mode`,
		userID, req.Name, slug, mode,
	).Scan(&key.ID, &key.Name, &key.Slug, &key.Mode)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			c.JSON(http.StatusConflict, gin.H{"error": "a tag key with that name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	key.Values = []tagValueResponse{}

	c.JSON(http.StatusCreated, key)
}

// DeleteTagKey - DELETE /me/tag-keys/:keyId
func (h *Handler) DeleteTagKey(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	keyID := c.Param("keyId")

	result, err := h.pool.Exec(c.Request.Context(),
		`DELETE FROM tag_keys WHERE id = $1 AND user_id = $2`,
		keyID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "tag key not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// CreateTagValue - POST /me/tag-keys/:keyId/values
func (h *Handler) CreateTagValue(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	keyID := c.Param("keyId")

	var req struct {
		Name string `json:"name" binding:"required,max=100"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var exists bool
	if err := h.pool.QueryRow(c.Request.Context(),
		`SELECT EXISTS(SELECT 1 FROM tag_keys WHERE id = $1 AND user_id = $2)`,
		keyID, userID,
	).Scan(&exists); err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "tag key not found"})
		return
	}

	slug := slugifyValue(req.Name)
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name produces an empty slug"})
		return
	}

	var val tagValueResponse
	err := h.pool.QueryRow(c.Request.Context(),
		`INSERT INTO tag_values (tag_key_id, name, slug) VALUES ($1, $2, $3)
		 RETURNING id, name, slug`,
		keyID, req.Name, slug,
	).Scan(&val.ID, &val.Name, &val.Slug)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			c.JSON(http.StatusConflict, gin.H{"error": "a value with that name already exists for this key"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusCreated, val)
}

// DeleteTagValue - DELETE /me/tag-keys/:keyId/values/:valueId
func (h *Handler) DeleteTagValue(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	keyID := c.Param("keyId")
	valueID := c.Param("valueId")

	var keyOwned bool
	if err := h.pool.QueryRow(c.Request.Context(),
		`SELECT EXISTS(SELECT 1 FROM tag_keys WHERE id = $1 AND user_id = $2)`,
		keyID, userID,
	).Scan(&keyOwned); err != nil || !keyOwned {
		c.JSON(http.StatusNotFound, gin.H{"error": "tag key not found"})
		return
	}

	result, err := h.pool.Exec(c.Request.Context(),
		`DELETE FROM tag_values WHERE id = $1 AND tag_key_id = $2`,
		valueID, keyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "tag value not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ── Book tagging ───────────────────────────────────────────────────────────────

// GetBookTags - GET /me/books/:olId/tags
// Returns all key-value assignments for a book. For multi-select keys, the same
// key may appear multiple times (once per assigned value).
func (h *Handler) GetBookTags(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	olID := c.Param("olId")

	rows, err := h.pool.Query(c.Request.Context(),
		`SELECT tk.id, tk.name, tk.slug, tv.id, tv.name, tv.slug
		 FROM book_tag_values btv
		 JOIN tag_keys   tk ON tk.id = btv.tag_key_id
		 JOIN tag_values tv ON tv.id = btv.tag_value_id
		 JOIN books       b ON b.id  = btv.book_id
		 WHERE btv.user_id = $1 AND b.open_library_id = $2
		 ORDER BY tk.created_at, tv.created_at`,
		userID, olID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer rows.Close()

	result := []bookTagResponse{}
	for rows.Next() {
		var t bookTagResponse
		if err := rows.Scan(&t.KeyID, &t.KeyName, &t.KeySlug, &t.ValueID, &t.ValueName, &t.ValueSlug); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		result = append(result, t)
	}

	c.JSON(http.StatusOK, result)
}

// SetBookTag - PUT /me/books/:olId/tags/:keyId
// Assigns a label value to a book.
// Body: { "value_id": "<uuid>" } OR { "value_name": "<string>" } for free-form.
//
// select_one: replaces any existing value for this key.
// select_multiple: adds the value (idempotent; already-assigned values are ignored).
func (h *Handler) SetBookTag(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	olID := c.Param("olId")
	keyID := c.Param("keyId")

	var req struct {
		ValueID   string `json:"value_id"`
		ValueName string `json:"value_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.ValueID == "" && req.ValueName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provide value_id or value_name"})
		return
	}

	// Verify key belongs to the user; get its mode.
	var mode string
	if err := h.pool.QueryRow(c.Request.Context(),
		`SELECT mode FROM tag_keys WHERE id = $1 AND user_id = $2`,
		keyID, userID,
	).Scan(&mode); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tag key not found"})
		return
	}

	// Resolve value ID — either by the provided ID or by find-or-create from name.
	valueID := req.ValueID
	if valueID == "" {
		slug := slugifyValue(req.ValueName)
		if slug == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "value_name produces an empty slug"})
			return
		}
		if err := h.pool.QueryRow(c.Request.Context(),
			`INSERT INTO tag_values (tag_key_id, name, slug)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (tag_key_id, slug) DO UPDATE SET name = tag_values.name
			 RETURNING id`,
			keyID, req.ValueName, slug,
		).Scan(&valueID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
	} else {
		// Verify value belongs to this key.
		var ok bool
		if err := h.pool.QueryRow(c.Request.Context(),
			`SELECT EXISTS(SELECT 1 FROM tag_values WHERE id = $1 AND tag_key_id = $2)`,
			valueID, keyID,
		).Scan(&ok); err != nil || !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "tag value not found"})
			return
		}
	}

	// Resolve book_id.
	var bookID string
	if err := h.pool.QueryRow(c.Request.Context(),
		`SELECT id FROM books WHERE open_library_id = $1`,
		olID,
	).Scan(&bookID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	// When setting the Status key, auto-create user_books row if missing.
	var keySlug string
	_ = h.pool.QueryRow(c.Request.Context(),
		`SELECT slug FROM tag_keys WHERE id = $1`, keyID,
	).Scan(&keySlug)
	if keySlug == "status" {
		h.pool.Exec(c.Request.Context(), //nolint:errcheck
			`INSERT INTO user_books (user_id, book_id)
			 VALUES ($1, $2)
			 ON CONFLICT (user_id, book_id) DO NOTHING`,
			userID, bookID,
		)
	}

	// For select_one, remove any existing value for this key before inserting.
	if mode == "select_one" {
		if _, err := h.pool.Exec(c.Request.Context(),
			`DELETE FROM book_tag_values WHERE user_id = $1 AND book_id = $2 AND tag_key_id = $3`,
			userID, bookID, keyID,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
	}

	if _, err := h.pool.Exec(c.Request.Context(),
		`INSERT INTO book_tag_values (user_id, book_id, tag_key_id, tag_value_id)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT DO NOTHING`,
		userID, bookID, keyID, valueID,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// UnsetBookTag - DELETE /me/books/:olId/tags/:keyId
// Removes all label assignments for a key from a book (works for both modes).
func (h *Handler) UnsetBookTag(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	olID := c.Param("olId")
	keyID := c.Param("keyId")

	_, err := h.pool.Exec(c.Request.Context(),
		`DELETE FROM book_tag_values btv
		 USING books b
		 WHERE btv.book_id = b.id
		   AND b.open_library_id = $1
		   AND btv.user_id = $2
		   AND btv.tag_key_id = $3`,
		olID, userID, keyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// UnsetBookTagValue - DELETE /me/books/:olId/tags/:keyId/values/:valueId
// Removes a specific value assignment for a key from a book.
// Primarily for select_multiple keys where you toggle individual values.
func (h *Handler) UnsetBookTagValue(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	olID := c.Param("olId")
	valueID := c.Param("valueId")

	_, err := h.pool.Exec(c.Request.Context(),
		`DELETE FROM book_tag_values btv
		 USING books b
		 WHERE btv.book_id = b.id
		   AND b.open_library_id = $1
		   AND btv.user_id = $2
		   AND btv.tag_value_id = $3`,
		olID, userID, valueID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ── Label book listing ─────────────────────────────────────────────────────────

type labelBook struct {
	BookID        string  `json:"book_id"`
	OpenLibraryID string  `json:"open_library_id"`
	Title         string  `json:"title"`
	CoverURL      *string `json:"cover_url"`
	AddedAt       string  `json:"added_at"`
	Rating        *int    `json:"rating"`
}

type labelBooksResponse struct {
	KeySlug   string      `json:"key_slug"`
	KeyName   string      `json:"key_name"`
	ValueSlug string      `json:"value_slug"`
	ValueName string      `json:"value_name"`
	SubLabels []string    `json:"sub_labels"`
	Books     []labelBook `json:"books"`
}

// GetLabelBooks - GET /users/:username/labels/:keySlug/*valuePath
// Public, but gated for private profiles.
func (h *Handler) GetLabelBooks(c *gin.Context) {
	username := c.Param("username")
	currentUserID := c.GetString(middleware.UserIDKey)

	_, _, canView := privacy.CanViewProfile(c.Request.Context(), h.pool, username, currentUserID)
	if !canView {
		c.JSON(http.StatusNotFound, gin.H{"error": "label not found"})
		return
	}

	keySlug := c.Param("keySlug")
	// Gin wildcard params include a leading "/"
	valuePath := strings.TrimPrefix(c.Param("valuePath"), "/")
	if valuePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "value path required"})
		return
	}

	// Verify key exists for this user and get user ID + key ID + key name.
	var userID, keyID, keyName string
	if err := h.pool.QueryRow(c.Request.Context(),
		`SELECT u.id, tk.id, tk.name
		 FROM tag_keys tk
		 JOIN users u ON u.id = tk.user_id
		 WHERE u.username = $1 AND u.deleted_at IS NULL AND tk.slug = $2`,
		username, keySlug,
	).Scan(&userID, &keyID, &keyName); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "label not found"})
		return
	}

	likePattern := valuePath + "/%"

	// Verify at least one value exists for this path (exact or sub-value).
	var valueCount int
	if err := h.pool.QueryRow(c.Request.Context(),
		`SELECT COUNT(*) FROM tag_values
		 WHERE tag_key_id = $1 AND (slug = $2 OR slug LIKE $3)`,
		keyID, valuePath, likePattern,
	).Scan(&valueCount); err != nil || valueCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "label not found"})
		return
	}

	// Display name: exact value if exists, otherwise last path segment.
	var valueName string
	if err := h.pool.QueryRow(c.Request.Context(),
		`SELECT name FROM tag_values WHERE tag_key_id = $1 AND slug = $2`,
		keyID, valuePath,
	).Scan(&valueName); err != nil {
		parts := strings.Split(valuePath, "/")
		valueName = parts[len(parts)-1]
	}

	// Fetch books, deduplicating in case a book has both a parent and child value.
	rows, err := h.pool.Query(c.Request.Context(),
		`SELECT sub.id, sub.open_library_id, sub.title, sub.cover_url, sub.added_at, sub.rating
		 FROM (
		   SELECT DISTINCT ON (b.id)
		          b.id, b.open_library_id, b.title, b.cover_url, btv.created_at AS added_at,
		          (SELECT ub.rating
		           FROM user_books ub
		           WHERE ub.book_id = b.id AND ub.user_id = $1 AND ub.rating IS NOT NULL
		           LIMIT 1) AS rating
		   FROM book_tag_values btv
		   JOIN tag_values tv ON tv.id = btv.tag_value_id
		   JOIN books b ON b.id = btv.book_id
		   WHERE btv.user_id = $1
		     AND btv.tag_key_id = $2
		     AND (tv.slug = $3 OR tv.slug LIKE $4)
		   ORDER BY b.id, btv.created_at DESC
		 ) sub
		 ORDER BY sub.added_at DESC`,
		userID, keyID, valuePath, likePattern,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer rows.Close()

	books := []labelBook{}
	for rows.Next() {
		var book labelBook
		var addedAt time.Time
		if err := rows.Scan(&book.BookID, &book.OpenLibraryID, &book.Title, &book.CoverURL, &addedAt, &book.Rating); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		book.AddedAt = addedAt.Format(time.RFC3339)
		books = append(books, book)
	}

	// Derive direct sub-labels (one level deeper than valuePath).
	subRows, err := h.pool.Query(c.Request.Context(),
		`SELECT slug FROM tag_values WHERE tag_key_id = $1 AND slug LIKE $2`,
		keyID, likePattern,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer subRows.Close()

	seen := make(map[string]bool)
	subLabels := []string{}
	for subRows.Next() {
		var slug string
		if err := subRows.Scan(&slug); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		rest := strings.TrimPrefix(slug, valuePath+"/")
		childSlug := valuePath + "/" + strings.SplitN(rest, "/", 2)[0]
		if !seen[childSlug] {
			seen[childSlug] = true
			subLabels = append(subLabels, childSlug)
		}
	}
	sort.Strings(subLabels)

	c.JSON(http.StatusOK, labelBooksResponse{
		KeySlug:   keySlug,
		KeyName:   keyName,
		ValueSlug: valuePath,
		ValueName: valueName,
		SubLabels: subLabels,
		Books:     books,
	})
}
