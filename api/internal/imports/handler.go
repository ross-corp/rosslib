package imports

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tristansaldanha/rosslib/api/internal/books"
	"github.com/tristansaldanha/rosslib/api/internal/middleware"
	"github.com/tristansaldanha/rosslib/api/internal/search"
	"github.com/tristansaldanha/rosslib/api/internal/tags"
)

type Handler struct {
	pool     *pgxpool.Pool
	search   *search.Client
	olClient *http.Client
}

func NewHandler(pool *pgxpool.Pool, searchClient *search.Client, olClient *http.Client) *Handler {
	return &Handler{pool: pool, search: searchClient, olClient: olClient}
}

// ── Types ─────────────────────────────────────────────────────────────────────

// goodreadsRow holds the fields parsed from a single CSV row.
type goodreadsRow struct {
	RowID          int
	Title          string
	Author         string
	ISBN13         string
	Rating         int    // 0 = no rating
	ReviewText     string
	Spoiler        bool
	DateRead       string // "YYYY-MM-DD" or ""
	DateAdded      string // "YYYY-MM-DD" or ""
	ExclusiveShelf string // mapped slug (e.g. "want-to-read")
	CustomShelves  []string
}

// BookCandidate is a candidate OL match shown in the preview.
type BookCandidate struct {
	OLId    string   `json:"ol_id"`
	Title   string   `json:"title"`
	Authors []string `json:"authors"`
	CoverURL *string `json:"cover_url"`
	Year    *int     `json:"year"`
}

// PreviewRow is one entry in the preview response.
type PreviewRow struct {
	RowID          int             `json:"row_id"`
	Title          string          `json:"title"`
	Author         string          `json:"author"`
	ISBN13         string          `json:"isbn13"`
	Rating         *int            `json:"rating"`      // nil = no rating
	ReviewText     *string         `json:"review_text"` // nil = no review
	Spoiler        bool            `json:"spoiler"`
	DateRead       *string         `json:"date_read"`
	DateAdded      *string         `json:"date_added"`
	ExclusiveShelf string          `json:"exclusive_shelf_slug"`
	CustomShelves  []string        `json:"custom_shelves"`
	Status         string          `json:"status"` // "matched" | "ambiguous" | "unmatched"
	Match          *BookCandidate  `json:"match,omitempty"`
	Candidates     []BookCandidate `json:"candidates,omitempty"`
}

// CommitRow is one confirmed row sent in the commit request.
type CommitRow struct {
	RowID           int      `json:"row_id"`
	OLId            string   `json:"ol_id"     binding:"required"`
	Title           string   `json:"title"     binding:"required"`
	CoverURL        *string  `json:"cover_url"`
	Authors         string   `json:"authors"`
	PublicationYear *int     `json:"publication_year"`
	ISBN13          *string  `json:"isbn13"`
	Rating          *int     `json:"rating"`
	ReviewText      *string  `json:"review_text"`
	Spoiler         bool     `json:"spoiler"`
	DateRead        *string  `json:"date_read"`
	DateAdded       *string  `json:"date_added"`
	ExclusiveShelf  string   `json:"exclusive_shelf_slug" binding:"required"`
	CustomShelves   []string `json:"custom_shelves"`
}

// ── Shelf name constants ──────────────────────────────────────────────────────

// defaultShelfMap maps Goodreads exclusive shelf values to our status label slugs.
var defaultShelfMap = map[string]string{
	"read":              "finished",
	"currently-reading": "currently-reading",
	"to-read":           "want-to-read",
}


// ── CSV helpers ───────────────────────────────────────────────────────────────

// stripISBNFormula strips Goodreads' Excel formula wrapper: ="ISBN" → ISBN
func stripISBNFormula(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "=")
	s = strings.Trim(s, `"`)
	return strings.TrimSpace(s)
}

func parseDate(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	return strings.ReplaceAll(s, "/", "-")
}

func parseCSV(r io.Reader) ([]goodreadsRow, error) {
	cr := csv.NewReader(r)
	cr.LazyQuotes = true // Goodreads sometimes emits slightly malformed CSV

	header, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	if len(header) < 19 || header[0] != "Book Id" {
		return nil, fmt.Errorf("file does not look like a Goodreads CSV export")
	}

	var rows []goodreadsRow
	for rowIdx := 0; ; rowIdx++ {
		record, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(record) < 19 {
			continue // skip malformed rows
		}

		var rating int
		// 0 if empty or unparseable, error ignored intentionally
		_, _ = fmt.Sscanf(record[7], "%d", &rating)

		exclusiveRaw := strings.TrimSpace(record[18])
		exclusiveSlug, ok := defaultShelfMap[exclusiveRaw]
		if !ok {
			exclusiveSlug = exclusiveRaw // custom exclusive shelf — use as-is
		}

		// Collect non-exclusive custom shelves from the Bookshelves column.
		var custom []string
		for _, s := range strings.Split(record[16], ",") {
			s = strings.TrimSpace(s)
			if s == "" || s == exclusiveRaw || s == exclusiveSlug {
				continue
			}
			custom = append(custom, s)
		}
		if custom == nil {
			custom = []string{}
		}

		rows = append(rows, goodreadsRow{
			RowID:          rowIdx,
			Title:          strings.TrimSpace(record[1]),
			Author:         strings.TrimSpace(record[2]),
			ISBN13:         stripISBNFormula(record[6]),
			Rating:         rating,
			ReviewText:     strings.TrimSpace(record[19]),
			Spoiler:        strings.EqualFold(strings.TrimSpace(record[20]), "true"),
			DateRead:       parseDate(record[14]),
			DateAdded:      parseDate(record[15]),
			ExclusiveShelf: exclusiveSlug,
			CustomShelves:  custom,
		})
	}
	return rows, nil
}

// ── OL search (title + author fallback) ──────────────────────────────────────

const (
	olSearchURL    = "https://openlibrary.org/search.json"
	olSearchFields = "key,title,author_name,first_publish_year,cover_i"
	olCoverMedURL  = "https://covers.openlibrary.org/b/id/%d-M.jpg"
)

type olDoc struct {
	Key              string   `json:"key"`
	Title            string   `json:"title"`
	AuthorName       []string `json:"author_name"`
	FirstPublishYear *int     `json:"first_publish_year"`
	CoverI           *int     `json:"cover_i"`
}

type olSearchResp struct {
	Docs []olDoc `json:"docs"`
}

func searchOL(ctx context.Context, olClient *http.Client, title, author string, limit int) ([]BookCandidate, error) {
	params := url.Values{}
	params.Set("title", title)
	if author != "" {
		params.Set("author", author)
	}
	params.Set("fields", olSearchFields)
	params.Set("limit", fmt.Sprintf("%d", limit))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, olSearchURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := olClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var olResp olSearchResp
	if err := json.Unmarshal(body, &olResp); err != nil {
		return nil, err
	}

	out := make([]BookCandidate, 0, len(olResp.Docs))
	for _, doc := range olResp.Docs {
		c := BookCandidate{
			OLId:    strings.TrimPrefix(doc.Key, "/works/"),
			Title:   doc.Title,
			Authors: doc.AuthorName,
			Year:    doc.FirstPublishYear,
		}
		if doc.CoverI != nil {
			u := fmt.Sprintf(olCoverMedURL, *doc.CoverI)
			c.CoverURL = &u
		}
		out = append(out, c)
	}
	return out, nil
}

// ── Preview ───────────────────────────────────────────────────────────────────

// Preview - POST /me/import/goodreads/preview (authed)
// Accepts a multipart file upload, parses the Goodreads CSV, attempts to match
// each book against Open Library, and returns a categorised preview. No DB
// writes are performed.
func (h *Handler) Preview(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "multipart field 'file' is required"})
		return
	}
	defer file.Close()

	rows, err := parseCSV(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid CSV: %v", err)})
		return
	}
	if len(rows) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV contains no books"})
		return
	}

	// Process rows concurrently with a bounded worker pool (5 workers).
	type indexedResult struct {
		idx int
		row PreviewRow
	}
	resultCh := make(chan indexedResult, len(rows))
	sem := make(chan struct{}, 5)
	var wg sync.WaitGroup

	for i, row := range rows {
		wg.Add(1)
		go func(idx int, r goodreadsRow) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			resultCh <- indexedResult{idx, buildPreviewRow(c.Request.Context(), h.olClient, r)}
		}(i, row)
	}
	wg.Wait()
	close(resultCh)

	previewRows := make([]PreviewRow, len(rows))
	matched, ambiguous, unmatched := 0, 0, 0
	for res := range resultCh {
		previewRows[res.idx] = res.row
		switch res.row.Status {
		case "matched":
			matched++
		case "ambiguous":
			ambiguous++
		default:
			unmatched++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total":     len(rows),
		"matched":   matched,
		"ambiguous": ambiguous,
		"unmatched": unmatched,
		"rows":      previewRows,
	})
}

func buildPreviewRow(ctx context.Context, olClient *http.Client, r goodreadsRow) PreviewRow {
	pr := PreviewRow{
		RowID:          r.RowID,
		Title:          r.Title,
		Author:         r.Author,
		ISBN13:         r.ISBN13,
		Spoiler:        r.Spoiler,
		ExclusiveShelf: r.ExclusiveShelf,
		CustomShelves:  r.CustomShelves,
	}
	if r.Rating > 0 {
		v := r.Rating
		pr.Rating = &v
	}
	if r.ReviewText != "" {
		pr.ReviewText = &r.ReviewText
	}
	if r.DateRead != "" {
		pr.DateRead = &r.DateRead
	}
	if r.DateAdded != "" {
		pr.DateAdded = &r.DateAdded
	}

	// 1. Try ISBN13 lookup via Open Library (no DB write: nil pool).
	if r.ISBN13 != "" {
		res, err := books.LookupBookByISBN(ctx, nil, r.ISBN13, olClient)
		if err == nil && res != nil {
			c := BookCandidate{
				OLId:     strings.TrimPrefix(res.Key, "/works/"),
				Title:    res.Title,
				Authors:  res.Authors,
				Year:     res.PublishYear,
				CoverURL: res.CoverURL,
			}
			pr.Status = "matched"
			pr.Match = &c
			return pr
		}
	}

	// 2. Fallback: title + author search.
	candidates, err := searchOL(ctx, olClient, r.Title, r.Author, 5)
	if err != nil || len(candidates) == 0 {
		pr.Status = "unmatched"
		return pr
	}
	if len(candidates) == 1 {
		pr.Status = "matched"
		pr.Match = &candidates[0]
		return pr
	}
	pr.Status = "ambiguous"
	pr.Candidates = candidates
	return pr
}

// ── Commit ────────────────────────────────────────────────────────────────────

// Commit - POST /me/import/goodreads/commit (authed)
// Accepts the confirmed rows from the preview and writes them to the database.
func (h *Handler) Commit(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	var req struct {
		Rows []CommitRow `json:"rows" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.Rows) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no rows to import"})
		return
	}

	// Ensure Status label key exists for this user.
	if err := tags.EnsureStatusLabel(c.Request.Context(), h.pool, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	imported, failed := 0, 0
	var errs []string

	for _, row := range req.Rows {
		if err := commitRow(c.Request.Context(), h.pool, h.search, userID, row); err != nil {
			failed++
			errs = append(errs, fmt.Sprintf("row %d (%q): %v", row.RowID, row.Title, err))
		} else {
			imported++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"imported": imported,
		"failed":   failed,
		"errors":   errs,
	})
}

func commitRow(ctx context.Context, pool *pgxpool.Pool, searchClient *search.Client, userID string, row CommitRow) error {
	// Upsert book into the global catalog.
	var isbn13, publicationYear interface{}
	if row.ISBN13 != nil && *row.ISBN13 != "" {
		isbn13 = *row.ISBN13
	}
	if row.PublicationYear != nil {
		publicationYear = *row.PublicationYear
	}

	var bookID string
	err := pool.QueryRow(ctx,
		`INSERT INTO books (open_library_id, title, cover_url, isbn13, authors, publication_year)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (open_library_id) DO UPDATE
		   SET title            = EXCLUDED.title,
		       cover_url        = EXCLUDED.cover_url,
		       isbn13           = COALESCE(books.isbn13, EXCLUDED.isbn13),
		       authors          = COALESCE(books.authors, EXCLUDED.authors),
		       publication_year = COALESCE(books.publication_year, EXCLUDED.publication_year)
		 RETURNING id`,
		row.OLId, row.Title, row.CoverURL, isbn13, row.Authors, publicationYear,
	).Scan(&bookID)
	if err != nil {
		return fmt.Errorf("upsert book: %w", err)
	}

	// Index into Meilisearch (fire-and-forget).
	if searchClient != nil {
		cv := ""
		if row.CoverURL != nil {
			cv = *row.CoverURL
		}
		i13 := ""
		if row.ISBN13 != nil {
			i13 = *row.ISBN13
		}
		py := 0
		if row.PublicationYear != nil {
			py = *row.PublicationYear
		}
		go searchClient.IndexBook(search.BookDocument{
			ID:              bookID,
			OpenLibraryID:   row.OLId,
			Title:           row.Title,
			Authors:         row.Authors,
			ISBN13:          i13,
			PublicationYear: py,
			CoverURL:        cv,
		})
	}

	// Parse date_added; fall back to now.
	addedAt := time.Now()
	if row.DateAdded != nil && *row.DateAdded != "" {
		if t, err := time.Parse("2006-01-02", *row.DateAdded); err == nil {
			addedAt = t
		}
	}

	var dateReadArg interface{}
	if row.DateRead != nil && *row.DateRead != "" {
		if t, err := time.Parse("2006-01-02", *row.DateRead); err == nil {
			dateReadArg = t
		}
	}

	// Insert (or update) user_books row with all metadata.
	if _, err := pool.Exec(ctx,
		`INSERT INTO user_books
		   (user_id, book_id, rating, review_text, spoiler, date_read, date_added)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (user_id, book_id) DO UPDATE
		   SET rating      = COALESCE(EXCLUDED.rating,      user_books.rating),
		       review_text = COALESCE(EXCLUDED.review_text, user_books.review_text),
		       spoiler     = EXCLUDED.spoiler,
		       date_read   = COALESCE(EXCLUDED.date_read,   user_books.date_read),
		       date_added  = EXCLUDED.date_added`,
		userID, bookID,
		row.Rating, row.ReviewText, row.Spoiler, dateReadArg, addedAt,
	); err != nil {
		return fmt.Errorf("add to user_books: %w", err)
	}

	// Set Status label for this book.
	statusSlug := row.ExclusiveShelf
	if statusSlug != "" {
		// Look up the user's Status key and the matching value.
		var keyID string
		if err := pool.QueryRow(ctx,
			`SELECT id FROM tag_keys WHERE user_id = $1 AND slug = 'status'`,
			userID,
		).Scan(&keyID); err != nil {
			return fmt.Errorf("find status key: %w", err)
		}

		var valueID string
		if err := pool.QueryRow(ctx,
			`SELECT id FROM tag_values WHERE tag_key_id = $1 AND slug = $2`,
			keyID, statusSlug,
		).Scan(&valueID); err != nil {
			// If value doesn't exist, skip status assignment (non-fatal).
			return nil //nolint:nilerr
		}

		// Remove any existing status value (select_one).
		pool.Exec(ctx, //nolint:errcheck
			`DELETE FROM book_tag_values WHERE user_id = $1 AND book_id = $2 AND tag_key_id = $3`,
			userID, bookID, keyID,
		)

		// Set the status label.
		if _, err := pool.Exec(ctx,
			`INSERT INTO book_tag_values (user_id, book_id, tag_key_id, tag_value_id)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT DO NOTHING`,
			userID, bookID, keyID, valueID,
		); err != nil {
			return fmt.Errorf("set status label: %w", err)
		}
	}

	return nil
}
