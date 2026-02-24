package userbooks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tristansaldanha/rosslib/api/internal/activity"
	"github.com/tristansaldanha/rosslib/api/internal/middleware"
	"github.com/tristansaldanha/rosslib/api/internal/privacy"
	"github.com/tristansaldanha/rosslib/api/internal/search"
	"github.com/tristansaldanha/rosslib/api/internal/tags"
)

type Handler struct {
	pool   *pgxpool.Pool
	search *search.Client
}

func NewHandler(pool *pgxpool.Pool, searchClient *search.Client) *Handler {
	return &Handler{pool: pool, search: searchClient}
}

// ── AddBook — POST /me/books ─────────────────────────────────────────────────

func (h *Handler) AddBook(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	var req struct {
		OpenLibraryID string  `json:"open_library_id" binding:"required"`
		Title         string  `json:"title"           binding:"required"`
		CoverURL      *string `json:"cover_url"`
		StatusValueID *string `json:"status_value_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Upsert book into global catalog.
	var bookID string
	err := h.pool.QueryRow(c.Request.Context(),
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

	// Index into Meilisearch (fire-and-forget).
	if h.search != nil {
		coverURL := ""
		if req.CoverURL != nil {
			coverURL = *req.CoverURL
		}
		go h.search.IndexBook(search.BookDocument{
			ID:            bookID,
			OpenLibraryID: req.OpenLibraryID,
			Title:         req.Title,
			CoverURL:      coverURL,
		})
	}

	// Insert user_books row.
	_, err = h.pool.Exec(c.Request.Context(),
		`INSERT INTO user_books (user_id, book_id)
		 VALUES ($1, $2)
		 ON CONFLICT (user_id, book_id) DO NOTHING`,
		userID, bookID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// If a status value was provided, set the Status label.
	if req.StatusValueID != nil && *req.StatusValueID != "" {
		if err := h.setStatusLabel(c.Request.Context(), userID, bookID, req.OpenLibraryID, *req.StatusValueID); err != nil {
			// Non-fatal: book is added, status just wasn't set.
			c.JSON(http.StatusOK, gin.H{"ok": true, "warning": "book added but status not set"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ── UpdateBook — PATCH /me/books/:olId ────────────────────────────────────────

func (h *Handler) UpdateBook(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	olID := c.Param("olId")

	var raw map[string]json.RawMessage
	if err := c.ShouldBindJSON(&raw); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(raw) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	// Verify user_books row exists.
	var ubID string
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT ub.id FROM user_books ub
		 JOIN books b ON b.id = ub.book_id
		 WHERE ub.user_id = $1 AND b.open_library_id = $2`,
		userID, olID,
	).Scan(&ubID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not in library"})
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

	args = append(args, ubID)
	query := fmt.Sprintf(
		"UPDATE user_books SET %s WHERE id = $%d",
		strings.Join(setClauses, ", "), idx,
	)
	if _, err := h.pool.Exec(c.Request.Context(), query, args...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// Record activity for rating/review updates.
	if _, hasRating := raw["rating"]; hasRating {
		var bookID string
		_ = h.pool.QueryRow(c.Request.Context(),
			`SELECT id FROM books WHERE open_library_id = $1`, olID,
		).Scan(&bookID)
		if bookID != "" {
			meta := map[string]string{}
			meta["rating"] = strings.Trim(string(raw["rating"]), "\"")
			activity.Record(c.Request.Context(), h.pool, userID, "rated", &bookID, nil, nil, nil, meta)
		}
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ── RemoveBook — DELETE /me/books/:olId ───────────────────────────────────────

func (h *Handler) RemoveBook(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	olID := c.Param("olId")

	// Remove status label (all tag values for Status key).
	var bookID string
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT id FROM books WHERE open_library_id = $1`, olID,
	).Scan(&bookID)
	if err == nil && bookID != "" {
		// Find the user's Status key and clear label assignments.
		var keyID string
		_ = h.pool.QueryRow(c.Request.Context(),
			`SELECT id FROM tag_keys WHERE user_id = $1 AND slug = 'status'`,
			userID,
		).Scan(&keyID)
		if keyID != "" {
			h.pool.Exec(c.Request.Context(), //nolint:errcheck
				`DELETE FROM book_tag_values
				 WHERE user_id = $1 AND book_id = $2 AND tag_key_id = $3`,
				userID, bookID, keyID,
			)
		}
	}

	// Remove user_books row.
	_, err = h.pool.Exec(c.Request.Context(),
		`DELETE FROM user_books ub
		 USING books b
		 WHERE ub.book_id = b.id
		   AND ub.user_id = $1
		   AND b.open_library_id = $2`,
		userID, olID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ── GetMyBookStatus — GET /me/books/:olId/status ──────────────────────────────

func (h *Handler) GetMyBookStatus(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	olID := c.Param("olId")

	type status struct {
		StatusValueID   *string `json:"status_value_id"`
		StatusValueName *string `json:"status_name"`
		StatusValueSlug *string `json:"status_slug"`
		Rating          *int    `json:"rating"`
		ReviewText      *string `json:"review_text"`
		Spoiler         bool    `json:"spoiler"`
		DateRead        *string `json:"date_read"`
	}

	var s status
	var dateRead *time.Time

	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT ub.rating, ub.review_text, ub.spoiler, ub.date_read,
		        tv.id, tv.name, tv.slug
		 FROM user_books ub
		 JOIN books b ON b.id = ub.book_id
		 LEFT JOIN tag_keys tk ON tk.user_id = ub.user_id AND tk.slug = 'status'
		 LEFT JOIN book_tag_values btv ON btv.user_id = ub.user_id AND btv.book_id = ub.book_id AND btv.tag_key_id = tk.id
		 LEFT JOIN tag_values tv ON tv.id = btv.tag_value_id
		 WHERE ub.user_id = $1 AND b.open_library_id = $2`,
		userID, olID,
	).Scan(&s.Rating, &s.ReviewText, &s.Spoiler, &dateRead,
		&s.StatusValueID, &s.StatusValueName, &s.StatusValueSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, nil)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if dateRead != nil {
		ds := dateRead.Format(time.RFC3339)
		s.DateRead = &ds
	}

	c.JSON(http.StatusOK, s)
}

// ── GetStatusMap — GET /me/books/status-map ───────────────────────────────────

func (h *Handler) GetStatusMap(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	rows, err := h.pool.Query(c.Request.Context(),
		`SELECT b.open_library_id, tv.id
		 FROM user_books ub
		 JOIN books b ON b.id = ub.book_id
		 JOIN tag_keys tk ON tk.user_id = ub.user_id AND tk.slug = 'status'
		 JOIN book_tag_values btv ON btv.user_id = ub.user_id AND btv.book_id = ub.book_id AND btv.tag_key_id = tk.id
		 JOIN tag_values tv ON tv.id = btv.tag_value_id
		 WHERE ub.user_id = $1`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer rows.Close()

	result := map[string]string{}
	for rows.Next() {
		var olID, valueID string
		if err := rows.Scan(&olID, &valueID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		result[olID] = valueID
	}

	c.JSON(http.StatusOK, result)
}

// ── GetUserBooks — GET /users/:username/books ─────────────────────────────────

type userBookItem struct {
	BookID        string  `json:"book_id"`
	OpenLibraryID string  `json:"open_library_id"`
	Title         string  `json:"title"`
	CoverURL      *string `json:"cover_url"`
	Authors       *string `json:"authors"`
	Rating        *int    `json:"rating"`
	AddedAt       string  `json:"added_at"`
}

type statusGroup struct {
	Name  string         `json:"name"`
	Slug  string         `json:"slug"`
	Count int            `json:"count"`
	Books []userBookItem `json:"books"`
}

func (h *Handler) GetUserBooks(c *gin.Context) {
	username := c.Param("username")
	currentUserID := c.GetString(middleware.UserIDKey)
	statusFilter := c.Query("status")

	_, _, canView := privacy.CanViewProfile(c.Request.Context(), h.pool, username, currentUserID)
	if !canView {
		c.JSON(http.StatusOK, gin.H{"statuses": []statusGroup{}, "unstatused_count": 0})
		return
	}

	// If filtering by a specific status slug
	if statusFilter != "" {
		limit := 200
		if l := c.Query("limit"); l != "" {
			if v, err := strconv.Atoi(l); err == nil && v >= 1 && v <= 200 {
				limit = v
			}
		}

		rows, err := h.pool.Query(c.Request.Context(),
			`SELECT b.id, b.open_library_id, b.title, b.cover_url, b.authors, ub.rating, ub.date_added
			 FROM user_books ub
			 JOIN books b ON b.id = ub.book_id
			 JOIN users u ON u.id = ub.user_id
			 JOIN tag_keys tk ON tk.user_id = u.id AND tk.slug = 'status'
			 JOIN book_tag_values btv ON btv.user_id = u.id AND btv.book_id = b.id AND btv.tag_key_id = tk.id
			 JOIN tag_values tv ON tv.id = btv.tag_value_id
			 WHERE u.username = $1 AND u.deleted_at IS NULL AND tv.slug = $2
			 ORDER BY ub.date_added DESC
			 LIMIT $3`,
			username, statusFilter, limit,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		defer rows.Close()

		books := []userBookItem{}
		for rows.Next() {
			var book userBookItem
			var addedAt time.Time
			if err := rows.Scan(&book.BookID, &book.OpenLibraryID, &book.Title, &book.CoverURL, &book.Authors, &book.Rating, &addedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
				return
			}
			book.AddedAt = addedAt.Format(time.RFC3339)
			books = append(books, book)
		}

		c.JSON(http.StatusOK, gin.H{"books": books})
		return
	}

	// No filter: return grouped by status
	limit := 8
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v >= 1 && v <= 200 {
			limit = v
		}
	}

	// Get user ID
	var userID string
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT id FROM users WHERE username = $1 AND deleted_at IS NULL`,
		username,
	).Scan(&userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Get status tag values for this user
	var keyID string
	_ = h.pool.QueryRow(c.Request.Context(),
		`SELECT id FROM tag_keys WHERE user_id = $1 AND slug = 'status'`,
		userID,
	).Scan(&keyID)

	statuses := []statusGroup{}

	if keyID != "" {
		valRows, err := h.pool.Query(c.Request.Context(),
			`SELECT tv.id, tv.name, tv.slug,
			        (SELECT COUNT(*) FROM book_tag_values btv2
			         WHERE btv2.user_id = $1 AND btv2.tag_key_id = $2 AND btv2.tag_value_id = tv.id) AS cnt
			 FROM tag_values tv
			 WHERE tv.tag_key_id = $2
			 ORDER BY tv.created_at`,
			userID, keyID,
		)
		if err == nil {
			defer valRows.Close()
			for valRows.Next() {
				var sg statusGroup
				var valueID string
				if err := valRows.Scan(&valueID, &sg.Name, &sg.Slug, &sg.Count); err != nil {
					continue
				}
				sg.Books = []userBookItem{}

				// Fetch preview books for this status
				bookRows, err := h.pool.Query(c.Request.Context(),
					`SELECT b.id, b.open_library_id, b.title, b.cover_url, b.authors, ub.rating, ub.date_added
					 FROM user_books ub
					 JOIN books b ON b.id = ub.book_id
					 JOIN book_tag_values btv ON btv.user_id = ub.user_id AND btv.book_id = b.id AND btv.tag_value_id = $1
					 WHERE ub.user_id = $2
					 ORDER BY ub.date_added DESC
					 LIMIT $3`,
					valueID, userID, limit,
				)
				if err == nil {
					for bookRows.Next() {
						var book userBookItem
						var addedAt time.Time
						if err := bookRows.Scan(&book.BookID, &book.OpenLibraryID, &book.Title, &book.CoverURL, &book.Authors, &book.Rating, &addedAt); err != nil {
							break
						}
						book.AddedAt = addedAt.Format(time.RFC3339)
						sg.Books = append(sg.Books, book)
					}
					bookRows.Close()
				}

				statuses = append(statuses, sg)
			}
			valRows.Close()
		}
	}

	// Count books without a status label
	var unstatusedCount int
	h.pool.QueryRow(c.Request.Context(),
		`SELECT COUNT(*) FROM user_books ub
		 WHERE ub.user_id = $1
		   AND NOT EXISTS (
		     SELECT 1 FROM book_tag_values btv
		     JOIN tag_keys tk ON tk.id = btv.tag_key_id
		     WHERE btv.user_id = ub.user_id AND btv.book_id = ub.book_id AND tk.slug = 'status'
		   )`,
		userID,
	).Scan(&unstatusedCount) //nolint:errcheck

	c.JSON(http.StatusOK, gin.H{"statuses": statuses, "unstatused_count": unstatusedCount})
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (h *Handler) setStatusLabel(ctx context.Context, userID, bookID, olID, statusValueID string) error {
	// Ensure Status key exists.
	if err := tags.EnsureStatusLabel(ctx, h.pool, userID); err != nil {
		return err
	}

	var keyID string
	if err := h.pool.QueryRow(ctx,
		`SELECT id FROM tag_keys WHERE user_id = $1 AND slug = 'status'`,
		userID,
	).Scan(&keyID); err != nil {
		return err
	}

	// Check what the previous status was (for activity recording).
	var prevSlug *string
	_ = h.pool.QueryRow(ctx,
		`SELECT tv.slug FROM book_tag_values btv
		 JOIN tag_values tv ON tv.id = btv.tag_value_id
		 WHERE btv.user_id = $1 AND btv.book_id = $2 AND btv.tag_key_id = $3`,
		userID, bookID, keyID,
	).Scan(&prevSlug)

	// Remove old status value (select_one).
	h.pool.Exec(ctx, //nolint:errcheck
		`DELETE FROM book_tag_values WHERE user_id = $1 AND book_id = $2 AND tag_key_id = $3`,
		userID, bookID, keyID,
	)

	// Insert new status value.
	_, err := h.pool.Exec(ctx,
		`INSERT INTO book_tag_values (user_id, book_id, tag_key_id, tag_value_id)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT DO NOTHING`,
		userID, bookID, keyID, statusValueID,
	)
	if err != nil {
		return err
	}

	// Look up the new slug for activity recording.
	var newSlug string
	_ = h.pool.QueryRow(ctx,
		`SELECT slug FROM tag_values WHERE id = $1`, statusValueID,
	).Scan(&newSlug)

	// Record activity for status transitions.
	if prevSlug == nil || *prevSlug != newSlug {
		activityType := "shelved"
		switch newSlug {
		case "currently-reading":
			activityType = "started_book"
		case "finished":
			activityType = "finished_book"
		}
		var statusName string
		_ = h.pool.QueryRow(ctx,
			`SELECT name FROM tag_values WHERE id = $1`, statusValueID,
		).Scan(&statusName)
		activity.Record(ctx, h.pool, userID, activityType, &bookID, nil, nil, nil,
			map[string]string{"status_name": statusName})
	}

	return nil
}
