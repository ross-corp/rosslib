package tags

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tristansaldanha/rosslib/api/internal/middleware"
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

// ListTagKeys - GET /me/tag-keys
func (h *Handler) ListTagKeys(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

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

	slug := slugify(req.Name)
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
		slug := slugify(req.ValueName)
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
	Books     []labelBook `json:"books"`
}

// GetLabelBooks - GET /users/:username/labels/:keySlug/:valueSlug (public)
// Returns all books for a user that have the given key+value label assigned.
func (h *Handler) GetLabelBooks(c *gin.Context) {
	username := c.Param("username")
	keySlug := c.Param("keySlug")
	valueSlug := c.Param("valueSlug")

	// Verify key+value exist for this user and get display names + user ID.
	var userID, keyName, valueName string
	if err := h.pool.QueryRow(c.Request.Context(),
		`SELECT u.id, tk.name, tv.name
		 FROM tag_keys tk
		 JOIN tag_values tv ON tv.tag_key_id = tk.id
		 JOIN users u ON u.id = tk.user_id
		 WHERE u.username = $1
		   AND u.deleted_at IS NULL
		   AND tk.slug = $2
		   AND tv.slug = $3`,
		username, keySlug, valueSlug,
	).Scan(&userID, &keyName, &valueName); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "label not found"})
		return
	}

	rows, err := h.pool.Query(c.Request.Context(),
		`SELECT b.id, b.open_library_id, b.title, b.cover_url, btv.created_at,
		        (SELECT ci.rating
		         FROM collection_items ci
		         JOIN collections col ON col.id = ci.collection_id
		         WHERE ci.book_id = b.id AND col.user_id = $1 AND ci.rating IS NOT NULL
		         ORDER BY ci.added_at DESC LIMIT 1) AS rating
		 FROM book_tag_values btv
		 JOIN tag_keys tk ON tk.id = btv.tag_key_id
		 JOIN tag_values tv ON tv.id = btv.tag_value_id
		 JOIN books b ON b.id = btv.book_id
		 WHERE btv.user_id = $1
		   AND tk.slug = $2
		   AND tv.slug = $3
		 ORDER BY btv.created_at DESC`,
		userID, keySlug, valueSlug,
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

	c.JSON(http.StatusOK, labelBooksResponse{
		KeySlug:   keySlug,
		KeyName:   keyName,
		ValueSlug: valueSlug,
		ValueName: valueName,
		Books:     books,
	})
}
