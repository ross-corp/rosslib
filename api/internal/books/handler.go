package books

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tristansaldanha/rosslib/api/internal/middleware"
)

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// ── Open Library API types ────────────────────────────────────────────────────

type olDoc struct {
	Key              string   `json:"key"`
	Title            string   `json:"title"`
	AuthorName       []string `json:"author_name"`
	FirstPublishYear *int     `json:"first_publish_year"`
	ISBN             []string `json:"isbn"`
	CoverI           *int     `json:"cover_i"`
	EditionCount     int      `json:"edition_count"`
	RatingsAverage   *float64 `json:"ratings_average"`
	RatingsCount     int      `json:"ratings_count"`
	AlreadyReadCount int      `json:"already_read_count"`
}

type olResponse struct {
	NumFound int     `json:"numFound"`
	Docs     []olDoc `json:"docs"`
}

type olDescription struct {
	Value string `json:"value"`
}

type olAuthorRef struct {
	Author struct {
		Key string `json:"key"`
	} `json:"author"`
}

type olWork struct {
	Title       string          `json:"title"`
	Description json.RawMessage `json:"description"`
	Covers      []int           `json:"covers"`
	Authors     []olAuthorRef   `json:"authors"`
}

type olRatings struct {
	Summary struct {
		Average *float64 `json:"average"`
		Count   int      `json:"count"`
	} `json:"summary"`
}

type olAuthor struct {
	Name string `json:"name"`
}

// ── Response types ────────────────────────────────────────────────────────────

// BookResult is the normalized shape returned to clients.
type BookResult struct {
	// Key is the Open Library work key, e.g. "/works/OL82592W".
	// Use it to construct a canonical work URL: https://openlibrary.org<key>
	Key              string   `json:"key"`
	Title            string   `json:"title"`
	Authors          []string `json:"authors"`
	PublishYear      *int     `json:"publish_year"`
	ISBN             []string `json:"isbn"`
	CoverURL         *string  `json:"cover_url"`
	EditionCount     int      `json:"edition_count"`
	AverageRating    *float64 `json:"average_rating"`
	RatingCount      int      `json:"rating_count"`
	AlreadyReadCount int      `json:"already_read_count"`
}

// BookDetail is the full book detail shape returned to clients.
type BookDetail struct {
	Key           string   `json:"key"`
	Title         string   `json:"title"`
	Authors       []string `json:"authors"`
	Description   *string  `json:"description"`
	CoverURL      *string  `json:"cover_url"`
	AverageRating *float64 `json:"average_rating"`
	RatingCount   int      `json:"rating_count"`
}

// ── Constants ─────────────────────────────────────────────────────────────────

const (
	olBaseURL      = "https://openlibrary.org"
	olSearchURL    = "https://openlibrary.org/search.json"
	olCoverURL     = "https://covers.openlibrary.org/b/id/%d-L.jpg"
	olCoverMedURL  = "https://covers.openlibrary.org/b/id/%d-M.jpg"
	olSearchFields = "key,title,author_name,first_publish_year,isbn,cover_i,edition_count,ratings_average,ratings_count,already_read_count"
	searchLimit    = 20
	maxISBNs       = 5
	maxAuthors     = 5
)

// ── Handler ───────────────────────────────────────────────────────────────────

// SearchBooks proxies a title query to the Open Library Search API and returns
// a normalized list of matching books.
//
// GET /books/search?q=<title>[&sort=reads|rating]
func (h *Handler) SearchBooks(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}
	sortBy := c.Query("sort") // "reads", "rating", or "" (relevance)

	apiURL := fmt.Sprintf(
		"%s?title=%s&fields=%s&limit=%d",
		olSearchURL,
		url.QueryEscape(q),
		olSearchFields,
		searchLimit,
	)

	resp, err := http.Get(apiURL) //nolint:noctx // intentional: inherits server timeout
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to reach book search service"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	var olResp olResponse
	if err := json.Unmarshal(body, &olResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	results := make([]BookResult, 0, len(olResp.Docs))
	for _, doc := range olResp.Docs {
		b := BookResult{
			Key:              doc.Key,
			Title:            doc.Title,
			Authors:          doc.AuthorName,
			PublishYear:      doc.FirstPublishYear,
			EditionCount:     doc.EditionCount,
			AverageRating:    doc.RatingsAverage,
			RatingCount:      doc.RatingsCount,
			AlreadyReadCount: doc.AlreadyReadCount,
		}

		if len(doc.ISBN) > 0 {
			b.ISBN = doc.ISBN[:min(maxISBNs, len(doc.ISBN))]
		}

		if doc.CoverI != nil {
			coverURL := fmt.Sprintf(olCoverMedURL, *doc.CoverI)
			b.CoverURL = &coverURL
		}

		results = append(results, b)
	}

	switch sortBy {
	case "reads":
		sort.Slice(results, func(i, j int) bool {
			return results[i].AlreadyReadCount > results[j].AlreadyReadCount
		})
	case "rating":
		sort.Slice(results, func(i, j int) bool {
			ai := 0.0
			if results[i].AverageRating != nil {
				ai = *results[i].AverageRating
			}
			aj := 0.0
			if results[j].AverageRating != nil {
				aj = *results[j].AverageRating
			}
			return ai > aj
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total":   olResp.NumFound,
		"results": results,
	})
}

// GetBook fetches full details for a single work from Open Library.
//
// GET /books/:workId   (workId is the bare OL ID, e.g. "OL82592W")
func (h *Handler) GetBook(c *gin.Context) {
	workID := c.Param("workId")
	if workID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workId is required"})
		return
	}

	workURL := fmt.Sprintf("%s/works/%s.json", olBaseURL, workID)
	ratingsURL := fmt.Sprintf("%s/works/%s/ratings.json", olBaseURL, workID)

	// Fetch work and ratings concurrently.
	type workResult struct {
		work olWork
		err  error
	}
	type ratingsResult struct {
		ratings olRatings
		err     error
	}

	workCh := make(chan workResult, 1)
	ratingsCh := make(chan ratingsResult, 1)

	go func() {
		resp, err := http.Get(workURL) //nolint:noctx
		if err != nil {
			workCh <- workResult{err: err}
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusNotFound {
			workCh <- workResult{err: fmt.Errorf("not found")}
			return
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			workCh <- workResult{err: err}
			return
		}
		var w olWork
		workCh <- workResult{work: w, err: json.Unmarshal(body, &w)}
	}()

	go func() {
		resp, err := http.Get(ratingsURL) //nolint:noctx
		if err != nil {
			ratingsCh <- ratingsResult{}
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var r olRatings
		_ = json.Unmarshal(body, &r)
		ratingsCh <- ratingsResult{ratings: r}
	}()

	wr := <-workCh
	if wr.err != nil {
		if wr.err.Error() == "not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		} else {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to reach book service"})
		}
		return
	}
	work := wr.work
	rr := <-ratingsCh

	// Resolve author names concurrently (up to maxAuthors).
	authorRefs := work.Authors
	if len(authorRefs) > maxAuthors {
		authorRefs = authorRefs[:maxAuthors]
	}
	authorNames := make([]string, len(authorRefs))
	var wg sync.WaitGroup
	for i, ref := range authorRefs {
		wg.Add(1)
		go func(idx int, key string) {
			defer wg.Done()
			authorURL := fmt.Sprintf("%s%s.json", olBaseURL, key)
			resp, err := http.Get(authorURL) //nolint:noctx
			if err != nil {
				return
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return
			}
			var a olAuthor
			if err := json.Unmarshal(body, &a); err == nil {
				authorNames[idx] = a.Name
			}
		}(i, ref.Author.Key)
	}
	wg.Wait()

	// Filter out empty names.
	authors := authorNames[:0]
	for _, name := range authorNames {
		if name != "" {
			authors = append(authors, name)
		}
	}

	detail := BookDetail{
		Key:         "/works/" + workID,
		Title:       work.Title,
		Authors:     authors,
		RatingCount: rr.ratings.Summary.Count,
		AverageRating: rr.ratings.Summary.Average,
	}

	// Parse description (can be a plain string or {"type":..., "value":...}).
	if len(work.Description) > 0 {
		var desc string
		if err := json.Unmarshal(work.Description, &desc); err == nil {
			detail.Description = &desc
		} else {
			var obj olDescription
			if err := json.Unmarshal(work.Description, &obj); err == nil && obj.Value != "" {
				detail.Description = &obj.Value
			}
		}
	}

	// Pick cover from work covers list.
	if len(work.Covers) > 0 && work.Covers[0] > 0 {
		coverURL := fmt.Sprintf(olCoverURL, work.Covers[0])
		detail.CoverURL = &coverURL
	}

	c.JSON(http.StatusOK, detail)
}

// ── ISBN lookup ───────────────────────────────────────────────────────────────

// LookupBookByISBN queries the Open Library search API for a single book by
// ISBN, upserts it into the local books table (if a pool is supplied), and
// returns the normalised BookResult. It is a package-level function so the
// import handler can call it directly without going through HTTP.
func LookupBookByISBN(ctx context.Context, pool *pgxpool.Pool, isbn string) (*BookResult, error) {
	// Strip everything that isn't a digit or trailing X (ISBN-10 check digit).
	clean := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' || r == 'X' || r == 'x' {
			return r
		}
		return -1
	}, isbn)
	if clean == "" {
		return nil, fmt.Errorf("invalid ISBN")
	}

	apiURL := fmt.Sprintf(
		"%s?isbn=%s&fields=%s&limit=1",
		olSearchURL,
		url.QueryEscape(clean),
		olSearchFields,
	)

	resp, err := http.Get(apiURL) //nolint:noctx
	if err != nil {
		return nil, fmt.Errorf("reach OL: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read OL response: %w", err)
	}

	var olResp olResponse
	if err := json.Unmarshal(body, &olResp); err != nil {
		return nil, fmt.Errorf("parse OL response: %w", err)
	}
	if len(olResp.Docs) == 0 {
		return nil, nil // not found
	}

	doc := olResp.Docs[0]

	result := &BookResult{
		Key:              doc.Key,
		Title:            doc.Title,
		Authors:          doc.AuthorName,
		PublishYear:      doc.FirstPublishYear,
		EditionCount:     doc.EditionCount,
		AverageRating:    doc.RatingsAverage,
		RatingCount:      doc.RatingsCount,
		AlreadyReadCount: doc.AlreadyReadCount,
	}

	if len(doc.ISBN) > 0 {
		result.ISBN = doc.ISBN[:min(maxISBNs, len(doc.ISBN))]
	}

	var coverURL *string
	if doc.CoverI != nil {
		u := fmt.Sprintf(olCoverMedURL, *doc.CoverI)
		coverURL = &u
		result.CoverURL = coverURL
	}

	if pool == nil {
		return result, nil
	}

	// Strip the "/works/" prefix — the DB stores bare OL IDs (e.g. "OL82592W").
	bareID := strings.TrimPrefix(doc.Key, "/works/")

	// Pick the best ISBN-13 to store (prefer the one we searched with if it's 13 digits).
	var isbn13 *string
	if len(clean) == 13 {
		isbn13 = &clean
	} else {
		for _, i := range doc.ISBN {
			if len(i) == 13 {
				isbn13 = &i
				break
			}
		}
	}

	authors := strings.Join(doc.AuthorName, ", ")

	_, err = pool.Exec(ctx,
		`INSERT INTO books (open_library_id, title, cover_url, isbn13, authors, publication_year)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (open_library_id) DO UPDATE
		   SET title            = EXCLUDED.title,
		       cover_url        = EXCLUDED.cover_url,
		       isbn13           = COALESCE(books.isbn13, EXCLUDED.isbn13),
		       authors          = COALESCE(books.authors, EXCLUDED.authors),
		       publication_year = COALESCE(books.publication_year, EXCLUDED.publication_year)`,
		bareID, doc.Title, coverURL, isbn13, authors, doc.FirstPublishYear,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert book: %w", err)
	}

	return result, nil
}

// GetBookReviews returns all community reviews for a book from the local database.
//
// GET /books/:workId/reviews
func (h *Handler) GetBookReviews(c *gin.Context) {
	workID := c.Param("workId")

	type review struct {
		Username    string  `json:"username"`
		DisplayName *string `json:"display_name"`
		AvatarURL   *string `json:"avatar_url"`
		Rating      *int    `json:"rating"`
		ReviewText  string  `json:"review_text"`
		Spoiler     bool    `json:"spoiler"`
		DateRead    *string `json:"date_read"`
		DateAdded   string  `json:"date_added"`
	}

	rows, err := h.pool.Query(c.Request.Context(),
		`SELECT DISTINCT ON (u.id) u.username, u.display_name, u.avatar_url,
		        ci.rating, ci.review_text, ci.spoiler, ci.date_read, ci.date_added
		 FROM collection_items ci
		 JOIN books b       ON b.id  = ci.book_id
		 JOIN collections c ON c.id  = ci.collection_id
		 JOIN users u       ON u.id  = c.user_id
		 WHERE b.open_library_id = $1
		   AND u.deleted_at IS NULL
		   AND ci.review_text IS NOT NULL
		   AND ci.review_text != ''
		 ORDER BY u.id, ci.date_added DESC`,
		workID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer rows.Close()

	reviews := []review{}
	for rows.Next() {
		var r review
		var dateRead *time.Time
		var dateAdded time.Time
		if err := rows.Scan(
			&r.Username, &r.DisplayName, &r.AvatarURL,
			&r.Rating, &r.ReviewText, &r.Spoiler, &dateRead, &dateAdded,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		r.DateAdded = dateAdded.Format(time.RFC3339)
		if dateRead != nil {
			s := dateRead.Format(time.RFC3339)
			r.DateRead = &s
		}
		reviews = append(reviews, r)
	}

	c.JSON(http.StatusOK, reviews)
}

// GetMyBookStatus returns the current user's shelf placement and review for a book.
//
// GET /me/books/:olId/status  (requires auth)
func (h *Handler) GetMyBookStatus(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	olID := c.Param("olId")

	type status struct {
		ShelfID    string  `json:"shelf_id"`
		ShelfName  string  `json:"shelf_name"`
		ShelfSlug  string  `json:"shelf_slug"`
		Rating     *int    `json:"rating"`
		ReviewText *string `json:"review_text"`
		Spoiler    bool    `json:"spoiler"`
		DateRead   *string `json:"date_read"`
	}

	var s status
	var dateRead *time.Time

	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT c.id, c.name, c.slug,
		        ci.rating, ci.review_text, ci.spoiler, ci.date_read
		 FROM collection_items ci
		 JOIN collections c ON c.id = ci.collection_id
		 JOIN books b        ON b.id = ci.book_id
		 WHERE c.user_id = $1
		   AND b.open_library_id = $2
		 ORDER BY (c.exclusive_group = 'read_status') DESC, ci.date_added DESC
		 LIMIT 1`,
		userID, olID,
	).Scan(&s.ShelfID, &s.ShelfName, &s.ShelfSlug,
		&s.Rating, &s.ReviewText, &s.Spoiler, &dateRead)
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

// LookupBook handles the ISBN lookup HTTP endpoint.
//
// GET /books/lookup?isbn=<isbn>
func (h *Handler) LookupBook(c *gin.Context) {
	isbn := c.Query("isbn")
	if isbn == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'isbn' is required"})
		return
	}

	result, err := LookupBookByISBN(c.Request.Context(), h.pool, isbn)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to reach book service"})
		return
	}
	if result == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	c.JSON(http.StatusOK, result)
}
