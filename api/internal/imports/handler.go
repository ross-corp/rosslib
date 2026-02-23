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
	"github.com/tristansaldanha/rosslib/api/internal/collections"
	"github.com/tristansaldanha/rosslib/api/internal/middleware"
)

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
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

// defaultShelfMap maps Goodreads exclusive shelf values to our slug names.
var defaultShelfMap = map[string]string{
	"read":              "read",
	"currently-reading": "currently-reading",
	"to-read":           "want-to-read",
}

// readStatusSlugs are the slugs that belong to the exclusive "read_status" group.
var readStatusSlugs = map[string]bool{
	"read": true, "currently-reading": true, "want-to-read": true,
}

// slugDisplayNames maps common Goodreads slugs to human-readable names.
var slugDisplayNames = map[string]string{
	"read":              "Read",
	"currently-reading": "Currently Reading",
	"want-to-read":      "Want to Read",
	"owned-to-read":     "Owned to Read",
	"dnf":               "Did Not Finish",
}

func slugToName(slug string) string {
	if name, ok := slugDisplayNames[slug]; ok {
		return name
	}
	parts := strings.Split(slug, "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
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
		fmt.Sscanf(record[7], "%d", &rating) // 0 if empty or unparseable

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

func searchOL(ctx context.Context, title, author string, limit int) ([]BookCandidate, error) {
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
	resp, err := http.DefaultClient.Do(req)
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
			resultCh <- indexedResult{idx, buildPreviewRow(c.Request.Context(), r)}
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

func buildPreviewRow(ctx context.Context, r goodreadsRow) PreviewRow {
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
		res, err := books.LookupBookByISBN(ctx, nil, r.ISBN13)
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
	candidates, err := searchOL(ctx, r.Title, r.Author, 5)
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

	// Ensure the 3 default shelves exist for this user.
	defaults := []struct{ name, slug string }{
		{"Want to Read", "want-to-read"},
		{"Currently Reading", "currently-reading"},
		{"Read", "read"},
	}
	for _, s := range defaults {
		if _, err := collections.EnsureShelf(c.Request.Context(), h.pool, userID, s.name, s.slug, true, "read_status", true); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
	}

	imported, failed := 0, 0
	var errs []string

	for _, row := range req.Rows {
		if err := commitRow(c.Request.Context(), h.pool, userID, row); err != nil {
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

func commitRow(ctx context.Context, pool *pgxpool.Pool, userID string, row CommitRow) error {
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

	// Determine exclusive group membership.
	excGroup := ""
	if readStatusSlugs[row.ExclusiveShelf] {
		excGroup = "read_status"
	}
	// Non-default exclusive shelves (e.g. "dnf", "owned-to-read") also get
	// read_status so they participate in the same mutual-exclusivity constraint.
	if !readStatusSlugs[row.ExclusiveShelf] && row.ExclusiveShelf != "" {
		excGroup = "read_status"
	}

	exclusiveShelfID, err := collections.EnsureShelf(
		ctx, pool, userID,
		slugToName(row.ExclusiveShelf), row.ExclusiveShelf,
		true, excGroup, true,
	)
	if err != nil {
		return fmt.Errorf("ensure exclusive shelf %q: %w", row.ExclusiveShelf, err)
	}

	// Remove book from any other shelf in the same exclusive group first.
	if excGroup != "" {
		if _, err := pool.Exec(ctx,
			`DELETE FROM collection_items ci
			 USING collections col
			 WHERE ci.collection_id = col.id
			   AND col.user_id        = $1
			   AND col.exclusive_group = $2
			   AND ci.book_id          = $3
			   AND ci.collection_id   != $4`,
			userID, excGroup, bookID, exclusiveShelfID,
		); err != nil {
			return fmt.Errorf("remove from other exclusive shelves: %w", err)
		}
	}

	// Parse date_added; fall back to now so added_at is set correctly.
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

	// Insert (or update) the item on the exclusive shelf with all metadata.
	if _, err := pool.Exec(ctx,
		`INSERT INTO collection_items
		   (collection_id, book_id, added_at, rating, review_text, spoiler, date_read, date_added)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 ON CONFLICT (collection_id, book_id) DO UPDATE
		   SET rating      = COALESCE(EXCLUDED.rating,      collection_items.rating),
		       review_text = COALESCE(EXCLUDED.review_text, collection_items.review_text),
		       spoiler     = EXCLUDED.spoiler,
		       date_read   = COALESCE(EXCLUDED.date_read,   collection_items.date_read),
		       date_added  = EXCLUDED.date_added`,
		exclusiveShelfID, bookID, addedAt,
		row.Rating, row.ReviewText, row.Spoiler, dateReadArg, addedAt,
	); err != nil {
		return fmt.Errorf("add to exclusive shelf: %w", err)
	}

	// Add to custom non-exclusive shelves.
	for _, slug := range row.CustomShelves {
		if slug == "" {
			continue
		}
		shelfID, err := collections.EnsureShelf(ctx, pool, userID, slugToName(slug), slug, false, "", true)
		if err != nil {
			continue // non-fatal; best-effort
		}
		pool.Exec(ctx, //nolint:errcheck
			`INSERT INTO collection_items (collection_id, book_id, added_at)
			 VALUES ($1, $2, $3)
			 ON CONFLICT DO NOTHING`,
			shelfID, bookID, addedAt,
		)
	}

	return nil
}
