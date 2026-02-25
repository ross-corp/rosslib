package books

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tristansaldanha/rosslib/api/internal/activity"
	"github.com/tristansaldanha/rosslib/api/internal/middleware"
	"github.com/tristansaldanha/rosslib/api/internal/search"
)

type Handler struct {
	pool     *pgxpool.Pool
	search   *search.Client
	olClient *http.Client
}

func NewHandler(pool *pgxpool.Pool, searchClient *search.Client, olClient *http.Client) *Handler {
	return &Handler{pool: pool, search: searchClient, olClient: olClient}
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
	Subject          []string `json:"subject"`
	Language         []string `json:"language"`
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
	Subjects    []string        `json:"subjects"`
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

type olEdition struct {
	Key            string `json:"key"`
	Title          string `json:"title"`
	Publishers     []string `json:"publishers"`
	NumberOfPages  *int     `json:"number_of_pages"`
	PublishDate    string   `json:"publish_date"`
	ISBN13         []string `json:"isbn_13"`
	ISBN10         []string `json:"isbn_10"`
	Covers         []int    `json:"covers"`
	PhysicalFormat string   `json:"physical_format"`
	Languages      []struct {
		Key string `json:"key"`
	} `json:"languages"`
}

type olEditionsResponse struct {
	Size    int         `json:"size"`
	Entries []olEdition `json:"entries"`
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
	Subjects         []string `json:"subjects"`
}

// Edition is a single edition of a work returned to clients.
type Edition struct {
	Key            string  `json:"key"`
	Title          string  `json:"title"`
	Publisher      *string `json:"publisher"`
	PublishDate    string  `json:"publish_date"`
	PageCount      *int    `json:"page_count"`
	ISBN           *string `json:"isbn"`
	CoverURL       *string `json:"cover_url"`
	Format         string  `json:"format"`
	Language       string  `json:"language"`
}

// BookDetail is the full book detail shape returned to clients.
type BookDetail struct {
	Key               string    `json:"key"`
	Title             string    `json:"title"`
	Authors           []string  `json:"authors"`
	Description       *string   `json:"description"`
	CoverURL          *string   `json:"cover_url"`
	AverageRating     *float64  `json:"average_rating"`
	RatingCount       int       `json:"rating_count"`
	LocalReadsCount   int       `json:"local_reads_count"`
	LocalWantToRead   int       `json:"local_want_to_read_count"`
	Publisher         *string   `json:"publisher"`
	PageCount         *int      `json:"page_count"`
	FirstPublishYear  *int      `json:"first_publish_year"`
	Subjects          []string  `json:"subjects"`
	EditionCount      int       `json:"edition_count"`
	Editions          []Edition `json:"editions"`
}

// ── Constants ─────────────────────────────────────────────────────────────────

const (
	olBaseURL      = "https://openlibrary.org"
	olSearchURL    = "https://openlibrary.org/search.json"
	olCoverURL     = "https://covers.openlibrary.org/b/id/%d-L.jpg"
	olCoverMedURL  = "https://covers.openlibrary.org/b/id/%d-M.jpg"
	olSearchFields = "key,title,author_name,first_publish_year,isbn,cover_i,edition_count,ratings_average,ratings_count,already_read_count,subject,language"
	searchLimit    = 20
	maxISBNs       = 5
	maxAuthors     = 5
)

// ── Handler ───────────────────────────────────────────────────────────────────

// SearchBooks searches both Meilisearch (local catalog) and Open Library,
// returning local matches first followed by external results.
//
// GET /books/search?q=<title>[&sort=reads|rating][&year_min=N][&year_max=N][&subject=S][&language=L]
func (h *Handler) SearchBooks(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}
	sortBy := c.Query("sort")       // "reads", "rating", or "" (relevance)
	subject := c.Query("subject")   // e.g. "fiction", "fantasy"
	language := c.Query("language") // e.g. "eng", "spa"

	// Parse optional year range filters.
	var yearMin, yearMax int
	if v := c.Query("year_min"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			yearMin = n
		}
	}
	if v := c.Query("year_max"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			yearMax = n
		}
	}

	// Run Meilisearch and Open Library searches concurrently.
	type meiliResult struct {
		docs []search.BookDocument
		err  error
	}
	type olResult struct {
		resp olResponse
		err  error
	}

	meiliCh := make(chan meiliResult, 1)
	olCh := make(chan olResult, 1)

	// Meilisearch search (local catalog).
	go func() {
		if h.search == nil {
			meiliCh <- meiliResult{}
			return
		}
		docs, err := h.search.SearchBooks(q, searchLimit, yearMin, yearMax, subject)
		meiliCh <- meiliResult{docs: docs, err: err}
	}()

	// Open Library search (external discovery).
	go func() {
		apiURL := fmt.Sprintf(
			"%s?title=%s&fields=%s&limit=%d",
			olSearchURL,
			url.QueryEscape(q),
			olSearchFields,
			searchLimit,
		)
		// Add year range filter to OL query if specified.
		if yearMin > 0 || yearMax > 0 {
			lo := "*"
			hi := "*"
			if yearMin > 0 {
				lo = strconv.Itoa(yearMin)
			}
			if yearMax > 0 {
				hi = strconv.Itoa(yearMax)
			}
			apiURL += fmt.Sprintf("&first_publish_year=[%s TO %s]", lo, hi)
		}
		// Add subject filter to OL query if specified.
		if subject != "" {
			apiURL += "&subject=" + url.QueryEscape(subject)
		}
		// Add language filter to OL query if specified.
		if language != "" {
			apiURL += "&language=" + url.QueryEscape(language)
		}
		resp, err := h.olClient.Get(apiURL) //nolint:noctx
		if err != nil {
			olCh <- olResult{err: err}
			return
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			olCh <- olResult{err: err}
			return
		}
		var parsed olResponse
		olCh <- olResult{resp: parsed, err: json.Unmarshal(body, &parsed)}
	}()

	mr := <-meiliCh
	or := <-olCh

	// Build a set of OL IDs from local results to deduplicate.
	seen := map[string]bool{}
	localIdx := map[string]int{} // OL key → index in results slice
	results := make([]BookResult, 0, searchLimit)

	// Add local (Meilisearch) results first.
	if mr.err == nil {
		for _, doc := range mr.docs {
			olKey := "/works/" + doc.OpenLibraryID
			seen[olKey] = true

			var coverURL *string
			if doc.CoverURL != "" {
				coverURL = &doc.CoverURL
			}
			var pubYear *int
			if doc.PublicationYear > 0 {
				y := doc.PublicationYear
				pubYear = &y
			}
			var authors []string
			if doc.Authors != "" {
				authors = strings.Split(doc.Authors, ", ")
			}
			var isbn []string
			if doc.ISBN13 != "" {
				isbn = []string{doc.ISBN13}
			}

			localIdx[olKey] = len(results)
			results = append(results, BookResult{
				Key:         olKey,
				Title:       doc.Title,
				Authors:     authors,
				PublishYear: pubYear,
				ISBN:        isbn,
				CoverURL:    coverURL,
				Subjects:    doc.Subjects,
			})
		}
	}

	// Append Open Library results that aren't already in local results.
	// When a local result is also found in OL, enrich it with popularity data.
	if or.err == nil {
		for _, doc := range or.resp.Docs {
			if seen[doc.Key] {
				// Enrich the local result with OL popularity data.
				if idx, ok := localIdx[doc.Key]; ok {
					results[idx].EditionCount = doc.EditionCount
					results[idx].AverageRating = doc.RatingsAverage
					results[idx].RatingCount = doc.RatingsCount
					results[idx].AlreadyReadCount = doc.AlreadyReadCount
				}
				continue
			}
			if len(results) >= searchLimit {
				break
			}

			// Post-filter OL results by year range.
			if yearMin > 0 || yearMax > 0 {
				if doc.FirstPublishYear == nil {
					continue // skip books with no year data when filtering
				}
				if yearMin > 0 && *doc.FirstPublishYear < yearMin {
					continue
				}
				if yearMax > 0 && *doc.FirstPublishYear > yearMax {
					continue
				}
			}

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
			if len(doc.Subject) > 0 {
				limit := min(5, len(doc.Subject))
				b.Subjects = doc.Subject[:limit]
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
	}

	// Apply sort.
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
	default:
		// Blend search relevance with popularity to surface popular books higher.
		// Each result gets a score combining its position-based relevance (from the
		// search engine ordering) with a popularity component derived from OL signals.
		n := float64(len(results))
		if n > 1 {
			sort.SliceStable(results, func(i, j int) bool {
				si := blendedScore(i, n)
				sj := blendedScore(j, n)
				si += popularityScore(results[i])
				sj += popularityScore(results[j])
				return si > sj
			})
		}
	}

	total := len(results)
	if or.err == nil && or.resp.NumFound > total {
		total = or.resp.NumFound
	}

	c.JSON(http.StatusOK, gin.H{
		"total":   total,
		"results": results,
	})
}

// blendedScore returns a position-based relevance component (0–1).
// Lower position index = higher relevance.
func blendedScore(position int, total float64) float64 {
	return (1.0 - float64(position)/total) * 0.6
}

// popularityScore computes a 0–0.4 popularity bonus from OL signals.
func popularityScore(b BookResult) float64 {
	pop := 0.0
	// Read count is the strongest signal of a book's popularity.
	if b.AlreadyReadCount > 0 {
		pop += math.Log10(float64(1+b.AlreadyReadCount)) * 0.5
	}
	// Average rating weighted by number of ratings (quality signal).
	if b.AverageRating != nil && b.RatingCount > 0 {
		pop += (*b.AverageRating / 5.0) * math.Log10(float64(1+b.RatingCount)) * 0.3
	}
	// Edition count as a proxy for cultural significance.
	if b.EditionCount > 0 {
		pop += math.Log10(float64(1+b.EditionCount)) * 0.2
	}
	// Normalize to roughly 0–1 range (log10(1M) ≈ 6, so max raw ≈ 6*0.5+1*6*0.3+0.2*6 ≈ 6).
	pop /= 6.0
	if pop > 1.0 {
		pop = 1.0
	}
	return pop * 0.4
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
	editionsURL := fmt.Sprintf("%s/works/%s/editions.json?limit=50", olBaseURL, workID)

	// Fetch work, ratings, and editions concurrently.
	type workResult struct {
		work olWork
		err  error
	}
	type ratingsResult struct {
		ratings olRatings
		err     error
	}
	type editionsResult struct {
		editions olEditionsResponse
		err      error
	}

	workCh := make(chan workResult, 1)
	ratingsCh := make(chan ratingsResult, 1)
	editionsCh := make(chan editionsResult, 1)

	go func() {
		resp, err := h.olClient.Get(workURL) //nolint:noctx
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
		resp, err := h.olClient.Get(ratingsURL) //nolint:noctx
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

	go func() {
		resp, err := h.olClient.Get(editionsURL) //nolint:noctx
		if err != nil {
			editionsCh <- editionsResult{}
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var e olEditionsResponse
		_ = json.Unmarshal(body, &e)
		editionsCh <- editionsResult{editions: e}
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
	er := <-editionsCh

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
			resp, err := h.olClient.Get(authorURL) //nolint:noctx
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

	// Include subjects from the work (cap at 10).
	if len(work.Subjects) > 0 {
		limit := min(10, len(work.Subjects))
		detail.Subjects = work.Subjects[:limit]
	}

	// Parse description (can be a plain string or {"type":..., "value":...}).
	if len(work.Description) > 0 {
		var desc string
		if json.Unmarshal(work.Description, &desc) == nil {
			detail.Description = &desc
		} else {
			var obj olDescription
			// Unused variable 'err' would be caught here, so we ignore it
			if json.Unmarshal(work.Description, &obj) == nil && obj.Value != "" {
				detail.Description = &obj.Value
			}
		}
	}

	// Pick cover from work covers list.
	if len(work.Covers) > 0 && work.Covers[0] > 0 {
		coverURL := fmt.Sprintf(olCoverURL, work.Covers[0])
		detail.CoverURL = &coverURL
	}

	// Extract edition metadata (publisher, page count) from the best edition
	// and build the full editions list.
	if er.err == nil {
		detail.EditionCount = er.editions.Size
		for _, ed := range er.editions.Entries {
			if detail.Publisher == nil && len(ed.Publishers) > 0 {
				detail.Publisher = &ed.Publishers[0]
			}
			if detail.PageCount == nil && ed.NumberOfPages != nil && *ed.NumberOfPages > 0 {
				detail.PageCount = ed.NumberOfPages
			}
		}
		detail.Editions = convertEditions(er.editions.Entries)
	}

	// Extract first publish year from the work's first edition publish date.
	// We check the local DB first; if not stored, try parsing from edition data.
	if h.pool != nil {
		var pubYear *int
		// We ignore error here since it's just enriching data
		_ = h.pool.QueryRow(c.Request.Context(),
			`SELECT publication_year FROM books WHERE open_library_id = $1`,
			workID,
		).Scan(&pubYear)
		if pubYear != nil && *pubYear > 0 {
			detail.FirstPublishYear = pubYear
		}
	}

	// Query precomputed book stats for read and want-to-read counts.
	if h.pool != nil {
		_ = h.pool.QueryRow(c.Request.Context(),
			`SELECT COALESCE(bs.reads_count, 0), COALESCE(bs.want_to_read_count, 0)
			 FROM books b
			 LEFT JOIN book_stats bs ON bs.book_id = b.id
			 WHERE b.open_library_id = $1`,
			workID,
		).Scan(&detail.LocalReadsCount, &detail.LocalWantToRead)
	}

	c.JSON(http.StatusOK, detail)
}

// ── Edition helpers ──────────────────────────────────────────────────────────

// convertEditions maps raw OL edition entries to our client-facing Edition type.
func convertEditions(entries []olEdition) []Edition {
	editions := make([]Edition, 0, len(entries))
	for _, ed := range entries {
		e := Edition{
			Key:         strings.TrimPrefix(ed.Key, "/books/"),
			Title:       ed.Title,
			PublishDate: ed.PublishDate,
			PageCount:   ed.NumberOfPages,
			Format:      ed.PhysicalFormat,
		}
		if len(ed.Publishers) > 0 {
			e.Publisher = &ed.Publishers[0]
		}
		if len(ed.ISBN13) > 0 {
			e.ISBN = &ed.ISBN13[0]
		} else if len(ed.ISBN10) > 0 {
			e.ISBN = &ed.ISBN10[0]
		}
		if len(ed.Covers) > 0 && ed.Covers[0] > 0 {
			coverURL := fmt.Sprintf(olCoverMedURL, ed.Covers[0])
			e.CoverURL = &coverURL
		}
		if len(ed.Languages) > 0 {
			e.Language = strings.TrimPrefix(ed.Languages[0].Key, "/languages/")
		}
		editions = append(editions, e)
	}
	return editions
}

// GetEditions returns all editions for a work from Open Library.
//
// GET /books/:workId/editions?limit=50&offset=0
func (h *Handler) GetEditions(c *gin.Context) {
	workID := c.Param("workId")
	if workID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workId is required"})
		return
	}

	limit := 50
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	offset := 0
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	apiURL := fmt.Sprintf("%s/works/%s/editions.json?limit=%d&offset=%d", olBaseURL, workID, limit, offset)
	resp, err := h.olClient.Get(apiURL) //nolint:noctx
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to reach book service"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "work not found"})
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	var olResp olEditionsResponse
	if err := json.Unmarshal(body, &olResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":    olResp.Size,
		"editions": convertEditions(olResp.Entries),
	})
}

// ── Genre browsing ───────────────────────────────────────────────────────────

// predefinedGenres is the canonical genre list, kept in sync with the webapp.
var predefinedGenres = []string{
	"Fiction",
	"Non-fiction",
	"Fantasy",
	"Science fiction",
	"Mystery",
	"Romance",
	"Horror",
	"Thriller",
	"Biography",
	"History",
	"Poetry",
	"Children",
}

type genreInfo struct {
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	BookCount int    `json:"book_count"`
}

// GetGenres returns the predefined genre list with book counts from the
// local catalog.
//
// GET /genres
func (h *Handler) GetGenres(c *gin.Context) {
	genres := make([]genreInfo, len(predefinedGenres))
	for i, g := range predefinedGenres {
		slug := strings.ToLower(strings.ReplaceAll(g, " ", "-"))
		genres[i] = genreInfo{Slug: slug, Name: g}
	}

	if h.pool != nil {
		// Count books per genre in a single query using FILTER clauses.
		for i, g := range predefinedGenres {
			var count int
			_ = h.pool.QueryRow(c.Request.Context(),
				`SELECT COUNT(*) FROM books WHERE subjects ILIKE '%' || $1 || '%'`,
				g,
			).Scan(&count)
			genres[i].BookCount = count
		}
	}

	c.JSON(http.StatusOK, genres)
}

// genreSlugToName converts a URL slug back to the display name.
func genreSlugToName(slug string) string {
	for _, g := range predefinedGenres {
		if strings.ToLower(strings.ReplaceAll(g, " ", "-")) == slug {
			return g
		}
	}
	return ""
}

// GetGenreBooks returns books matching a genre, browsing the local Meilisearch
// index without requiring a search query.
//
// GET /genres/:slug/books?page=1&limit=20
func (h *Handler) GetGenreBooks(c *gin.Context) {
	slug := c.Param("slug")
	genreName := genreSlugToName(slug)
	if genreName == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "unknown genre"})
		return
	}

	page := 1
	if v := c.Query("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	limit := 20
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	offset := (page - 1) * limit

	if h.search == nil {
		// Fallback: return books from DB directly.
		rows, err := h.pool.Query(c.Request.Context(),
			`SELECT open_library_id, title, COALESCE(authors, ''),
			        COALESCE(cover_url, ''), COALESCE(publication_year, 0),
			        COALESCE(isbn13, ''), COALESCE(subjects, '')
			 FROM books
			 WHERE subjects ILIKE '%' || $1 || '%'
			 ORDER BY title
			 LIMIT $2 OFFSET $3`,
			genreName, limit, offset,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		defer rows.Close()

		results := []BookResult{}
		for rows.Next() {
			var olID, title, authors, coverURL, isbn13, subjects string
			var pubYear int
			if err := rows.Scan(&olID, &title, &authors, &coverURL, &pubYear, &isbn13, &subjects); err != nil {
				continue
			}
			b := BookResult{
				Key:   "/works/" + olID,
				Title: title,
			}
			if authors != "" {
				b.Authors = strings.Split(authors, ", ")
			}
			if coverURL != "" {
				b.CoverURL = &coverURL
			}
			if pubYear > 0 {
				b.PublishYear = &pubYear
			}
			if isbn13 != "" {
				b.ISBN = []string{isbn13}
			}
			if subjects != "" {
				b.Subjects = strings.Split(subjects, ", ")
			}
			results = append(results, b)
		}

		var total int
		_ = h.pool.QueryRow(c.Request.Context(),
			`SELECT COUNT(*) FROM books WHERE subjects ILIKE '%' || $1 || '%'`,
			genreName,
		).Scan(&total)

		c.JSON(http.StatusOK, gin.H{
			"genre":   genreName,
			"total":   total,
			"page":    page,
			"results": results,
		})
		return
	}

	// Use Meilisearch for browsing.
	docs, total, err := h.search.BrowseBooks(genreName, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search error"})
		return
	}

	results := make([]BookResult, 0, len(docs))
	for _, doc := range docs {
		b := BookResult{
			Key:   "/works/" + doc.OpenLibraryID,
			Title: doc.Title,
		}
		if doc.Authors != "" {
			b.Authors = strings.Split(doc.Authors, ", ")
		}
		if doc.CoverURL != "" {
			b.CoverURL = &doc.CoverURL
		}
		if doc.PublicationYear > 0 {
			y := doc.PublicationYear
			b.PublishYear = &y
		}
		if doc.ISBN13 != "" {
			b.ISBN = []string{doc.ISBN13}
		}
		b.Subjects = doc.Subjects
		results = append(results, b)
	}

	c.JSON(http.StatusOK, gin.H{
		"genre":   genreName,
		"total":   total,
		"page":    page,
		"results": results,
	})
}

// ── ISBN lookup ───────────────────────────────────────────────────────────────

// LookupBookByISBN queries the Open Library search API for a single book by
// ISBN, upserts it into the local books table (if a pool is supplied), and
// returns the normalised BookResult. It is a package-level function so the
// import handler can call it directly without going through HTTP.
// An optional search.Client can be passed to index the book in Meilisearch.
func LookupBookByISBN(ctx context.Context, pool *pgxpool.Pool, isbn string, olClient *http.Client, searchClients ...*search.Client) (*BookResult, error) {
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

	resp, err := olClient.Get(apiURL) //nolint:noctx
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

	// Store up to 10 subjects, comma-separated.
	var subjectsStr *string
	if len(doc.Subject) > 0 {
		limit := min(10, len(doc.Subject))
		s := strings.Join(doc.Subject[:limit], ", ")
		subjectsStr = &s
	}

	var bookID string
	err = pool.QueryRow(ctx,
		`INSERT INTO books (open_library_id, title, cover_url, isbn13, authors, publication_year, subjects, publisher, page_count)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, NULL, NULL)
		 ON CONFLICT (open_library_id) DO UPDATE
		   SET title            = EXCLUDED.title,
		       cover_url        = EXCLUDED.cover_url,
		       isbn13           = COALESCE(books.isbn13, EXCLUDED.isbn13),
		       authors          = COALESCE(books.authors, EXCLUDED.authors),
		       publication_year = COALESCE(books.publication_year, EXCLUDED.publication_year),
		       subjects         = COALESCE(EXCLUDED.subjects, books.subjects)
		 RETURNING id`,
		bareID, doc.Title, coverURL, isbn13, authors, doc.FirstPublishYear, subjectsStr,
	).Scan(&bookID)
	if err != nil {
		return nil, fmt.Errorf("upsert book: %w", err)
	}

	// Index into Meilisearch if a client was provided.
	if len(searchClients) > 0 && searchClients[0] != nil {
		cv := ""
		if coverURL != nil {
			cv = *coverURL
		}
		i13 := ""
		if isbn13 != nil {
			i13 = *isbn13
		}
		py := 0
		if doc.FirstPublishYear != nil {
			py = *doc.FirstPublishYear
		}
		var subjects []string
		if len(doc.Subject) > 0 {
			limit := min(10, len(doc.Subject))
			subjects = doc.Subject[:limit]
		}
		go searchClients[0].IndexBook(search.BookDocument{
			ID:              bookID,
			OpenLibraryID:   bareID,
			Title:           doc.Title,
			Authors:         authors,
			ISBN13:          i13,
			PublicationYear: py,
			CoverURL:        cv,
			Subjects:        subjects,
		})
	}

	return result, nil
}

// GetBookStats returns precomputed aggregate stats for a book.
//
// GET /books/:workId/stats
func (h *Handler) GetBookStats(c *gin.Context) {
	workID := c.Param("workId")
	if workID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workId is required"})
		return
	}

	var readsCount, wantToReadCount, ratingCount, reviewCount int
	var ratingSum int64
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT COALESCE(bs.reads_count, 0),
		        COALESCE(bs.want_to_read_count, 0),
		        COALESCE(bs.rating_sum, 0),
		        COALESCE(bs.rating_count, 0),
		        COALESCE(bs.review_count, 0)
		 FROM books b
		 LEFT JOIN book_stats bs ON bs.book_id = b.id
		 WHERE b.open_library_id = $1`,
		workID,
	).Scan(&readsCount, &wantToReadCount, &ratingSum, &ratingCount, &reviewCount)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	var avgRating *float64
	if ratingCount > 0 {
		avg := float64(ratingSum) / float64(ratingCount)
		avgRating = &avg
	}

	c.JSON(http.StatusOK, gin.H{
		"reads_count":        readsCount,
		"want_to_read_count": wantToReadCount,
		"average_rating":     avgRating,
		"rating_count":       ratingCount,
		"review_count":       reviewCount,
	})
}

// GetBookReviews returns all community reviews for a book from the local database.
// If the caller is authenticated, reviews from followed users are sorted first.
//
// GET /books/:workId/reviews
func (h *Handler) GetBookReviews(c *gin.Context) {
	workID := c.Param("workId")
	userID := c.GetString(middleware.UserIDKey) // empty if unauthenticated

	type review struct {
		Username    string  `json:"username"`
		DisplayName *string `json:"display_name"`
		AvatarURL   *string `json:"avatar_url"`
		Rating      *int    `json:"rating"`
		ReviewText  string  `json:"review_text"`
		Spoiler     bool    `json:"spoiler"`
		DateRead    *string `json:"date_read"`
		DateDNF     *string `json:"date_dnf"`
		DateAdded   string  `json:"date_added"`
		IsFollowed  bool    `json:"is_followed"`
	}

	// Use a CTE to deduplicate (one review per user), then LEFT JOIN follows
	// so reviews from people the caller follows appear first.
	// When unauthenticated, $2 is the zero UUID and the LEFT JOIN never matches.
	followerID := userID
	if followerID == "" {
		followerID = "00000000-0000-0000-0000-000000000000"
	}

	rows, err := h.pool.Query(c.Request.Context(),
		`SELECT u.username, u.display_name, u.avatar_url,
		        ub.rating, ub.review_text, ub.spoiler, ub.date_read, ub.date_dnf, ub.date_added,
		        (f.follower_id IS NOT NULL) AS is_followed
		 FROM user_books ub
		 JOIN books b ON b.id = ub.book_id
		 JOIN users u ON u.id = ub.user_id
		 LEFT JOIN follows f ON f.follower_id = $2 AND f.followee_id = u.id
		                    AND f.status = 'active'
		 WHERE b.open_library_id = $1
		   AND u.deleted_at IS NULL
		   AND ub.review_text IS NOT NULL
		   AND ub.review_text != ''
		 ORDER BY is_followed DESC, ub.date_added DESC`,
		workID, followerID,
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
		var dateDNF *time.Time
		var dateAdded time.Time
		if err := rows.Scan(
			&r.Username, &r.DisplayName, &r.AvatarURL,
			&r.Rating, &r.ReviewText, &r.Spoiler, &dateRead, &dateDNF, &dateAdded,
			&r.IsFollowed,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		r.DateAdded = dateAdded.Format(time.RFC3339)
		if dateRead != nil {
			s := dateRead.Format(time.RFC3339)
			r.DateRead = &s
		}
		if dateDNF != nil {
			s := dateDNF.Format(time.RFC3339)
			r.DateDNF = &s
		}
		reviews = append(reviews, r)
	}

	c.JSON(http.StatusOK, reviews)
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

	result, err := LookupBookByISBN(c.Request.Context(), h.pool, isbn, h.olClient, h.search)
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

// ── Author search ─────────────────────────────────────────────────────────────

type olAuthorDoc struct {
	Key           string   `json:"key"`
	Name          string   `json:"name"`
	BirthDate     *string  `json:"birth_date"`
	DeathDate     *string  `json:"death_date"`
	TopWork       *string  `json:"top_work"`
	WorkCount     int      `json:"work_count"`
	TopSubjects   []string `json:"top_subjects"`
}

type olAuthorSearchResponse struct {
	NumFound int           `json:"numFound"`
	Docs     []olAuthorDoc `json:"docs"`
}

// AuthorResult is the normalized shape for author search results.
type AuthorResult struct {
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	BirthDate   *string  `json:"birth_date"`
	DeathDate   *string  `json:"death_date"`
	TopWork     *string  `json:"top_work"`
	WorkCount   int      `json:"work_count"`
	TopSubjects []string `json:"top_subjects"`
	PhotoURL    *string  `json:"photo_url"`
}

const (
	olAuthorSearchURL  = "https://openlibrary.org/search/authors.json"
	olAuthorPhotoURL   = "https://covers.openlibrary.org/a/olid/%s-M.jpg"
	olAuthorPhotoLgURL = "https://covers.openlibrary.org/a/olid/%s-L.jpg"
	olAuthorWorksURL   = "https://openlibrary.org/authors/%s/works.json"
)

// SearchAuthors proxies an author query to the Open Library Author Search API.
//
// GET /authors/search?q=<name>
func (h *Handler) SearchAuthors(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	apiURL := fmt.Sprintf(
		"%s?q=%s&limit=%d",
		olAuthorSearchURL,
		url.QueryEscape(q),
		searchLimit,
	)

	resp, err := h.olClient.Get(apiURL) //nolint:noctx
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to reach author search service"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	var olResp olAuthorSearchResponse
	if err := json.Unmarshal(body, &olResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	results := make([]AuthorResult, 0, len(olResp.Docs))
	for _, doc := range olResp.Docs {
		a := AuthorResult{
			Key:       doc.Key,
			Name:      doc.Name,
			BirthDate: doc.BirthDate,
			DeathDate: doc.DeathDate,
			TopWork:   doc.TopWork,
			WorkCount: doc.WorkCount,
		}

		if len(doc.TopSubjects) > 0 {
			limit := min(5, len(doc.TopSubjects))
			a.TopSubjects = doc.TopSubjects[:limit]
		}

		if doc.Key != "" {
			photoURL := fmt.Sprintf(olAuthorPhotoURL, doc.Key)
			a.PhotoURL = &photoURL
		}

		results = append(results, a)
	}

	c.JSON(http.StatusOK, gin.H{
		"total":   olResp.NumFound,
		"results": results,
	})
}

// ── Author detail ─────────────────────────────────────────────────────────────

// olAuthorDetail is the raw shape from /authors/{key}.json.
type olAuthorDetail struct {
	Name      string          `json:"name"`
	Bio       json.RawMessage `json:"bio"`
	BirthDate *string         `json:"birth_date"`
	DeathDate *string         `json:"death_date"`
	Photos    []int           `json:"photos"`
	Links     []olAuthorLink  `json:"links"`
}

type olAuthorLink struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type olAuthorWorksEntry struct {
	Key    string `json:"key"`
	Title  string `json:"title"`
	Covers []int  `json:"covers"`
}

type olAuthorWorksResponse struct {
	Entries []olAuthorWorksEntry `json:"entries"`
	Size    int                  `json:"size"`
}

// AuthorDetail is the full detail shape returned to clients.
type AuthorDetail struct {
	Key       string       `json:"key"`
	Name      string       `json:"name"`
	Bio       *string      `json:"bio"`
	BirthDate *string      `json:"birth_date"`
	DeathDate *string      `json:"death_date"`
	PhotoURL  *string      `json:"photo_url"`
	Links     []AuthorLink `json:"links"`
	WorkCount int          `json:"work_count"`
	Works     []AuthorWork `json:"works"`
}

type AuthorLink struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type AuthorWork struct {
	Key      string  `json:"key"`
	Title    string  `json:"title"`
	CoverURL *string `json:"cover_url"`
}

// GetAuthor fetches author details and a sample of works from Open Library.
//
// GET /authors/:authorKey
func (h *Handler) GetAuthor(c *gin.Context) {
	authorKey := c.Param("authorKey")
	if authorKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "authorKey is required"})
		return
	}

	authorURL := fmt.Sprintf("%s/authors/%s.json", olBaseURL, authorKey)
	worksURL := fmt.Sprintf(olAuthorWorksURL+"?limit=%d", authorKey, searchLimit)

	// Fetch author and works concurrently.
	type authorResult struct {
		author olAuthorDetail
		err    error
	}
	type worksResult struct {
		works olAuthorWorksResponse
		err   error
	}

	authorCh := make(chan authorResult, 1)
	worksCh := make(chan worksResult, 1)

	go func() {
		resp, err := h.olClient.Get(authorURL) //nolint:noctx
		if err != nil {
			authorCh <- authorResult{err: err}
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusNotFound {
			authorCh <- authorResult{err: fmt.Errorf("not found")}
			return
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			authorCh <- authorResult{err: err}
			return
		}
		var a olAuthorDetail
		authorCh <- authorResult{author: a, err: json.Unmarshal(body, &a)}
	}()

	go func() {
		resp, err := h.olClient.Get(worksURL) //nolint:noctx
		if err != nil {
			worksCh <- worksResult{}
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var w olAuthorWorksResponse
		_ = json.Unmarshal(body, &w)
		worksCh <- worksResult{works: w}
	}()

	ar := <-authorCh
	if ar.err != nil {
		if ar.err.Error() == "not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "author not found"})
		} else {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to reach author service"})
		}
		return
	}
	raw := ar.author
	wr := <-worksCh

	detail := AuthorDetail{
		Key:       authorKey,
		Name:      raw.Name,
		BirthDate: raw.BirthDate,
		DeathDate: raw.DeathDate,
		WorkCount: wr.works.Size,
	}

	// Parse bio (can be a plain string or {"type":..., "value":...}).
	if len(raw.Bio) > 0 {
		var bio string
		if json.Unmarshal(raw.Bio, &bio) == nil {
			detail.Bio = &bio
		} else {
			var obj olDescription
			// Unused variable 'err' would be caught here, so we ignore it
			if json.Unmarshal(raw.Bio, &obj) == nil && obj.Value != "" {
				detail.Bio = &obj.Value
			}
		}
	}

	// Photo URL: always construct from the author key (OL may not have photo data
	// in the photos array even when a photo exists via OLID).
	photoURL := fmt.Sprintf(olAuthorPhotoLgURL, authorKey)
	detail.PhotoURL = &photoURL

	// Links.
	for _, l := range raw.Links {
		detail.Links = append(detail.Links, AuthorLink(l))
	}

	// Works.
	for _, entry := range wr.works.Entries {
		bareKey := strings.TrimPrefix(entry.Key, "/works/")
		w := AuthorWork{
			Key:   bareKey,
			Title: entry.Title,
		}
		if len(entry.Covers) > 0 && entry.Covers[0] > 0 {
			coverURL := fmt.Sprintf(olCoverMedURL, entry.Covers[0])
			w.CoverURL = &coverURL
		}
		detail.Works = append(detail.Works, w)
	}

	c.JSON(http.StatusOK, detail)
}

// ── Author follows ──────────────────────────────────────────────────────────

// FollowAuthor creates an author follow for the current user.
//
// POST /authors/:authorKey/follow
func (h *Handler) FollowAuthor(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	authorKey := c.Param("authorKey")

	var body struct {
		AuthorName string `json:"author_name"`
	}
	_ = c.ShouldBindJSON(&body)

	_, err := h.pool.Exec(c.Request.Context(),
		`INSERT INTO author_follows (user_id, author_key, author_name)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, author_key) DO UPDATE SET author_name = EXCLUDED.author_name`,
		userID, authorKey, body.AuthorName,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	activity.Record(c.Request.Context(), h.pool, userID, "followed_author",
		nil, nil, nil, nil, map[string]string{
			"author_key":  authorKey,
			"author_name": body.AuthorName,
		})

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// UnfollowAuthor removes an author follow for the current user.
//
// DELETE /authors/:authorKey/follow
func (h *Handler) UnfollowAuthor(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	authorKey := c.Param("authorKey")

	_, err := h.pool.Exec(c.Request.Context(),
		`DELETE FROM author_follows WHERE user_id = $1 AND author_key = $2`,
		userID, authorKey,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GetAuthorFollowStatus checks whether the current user follows an author.
//
// GET /authors/:authorKey/follow
func (h *Handler) GetAuthorFollowStatus(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	authorKey := c.Param("authorKey")

	var exists bool
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT EXISTS(SELECT 1 FROM author_follows WHERE user_id = $1 AND author_key = $2)`,
		userID, authorKey,
	).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"following": exists})
}

// ── Book follows ────────────────────────────────────────────────────────────

// FollowBook creates a book follow for the current user.
//
// POST /books/:workId/follow
func (h *Handler) FollowBook(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	workID := c.Param("workId")

	var bookID string
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT id FROM books WHERE open_library_id = $1`, workID,
	).Scan(&bookID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found in local catalog"})
		return
	}

	_, err = h.pool.Exec(c.Request.Context(),
		`INSERT INTO book_follows (user_id, book_id)
		 VALUES ($1, $2)
		 ON CONFLICT (user_id, book_id) DO NOTHING`,
		userID, bookID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	var bookTitle string
	_ = h.pool.QueryRow(c.Request.Context(),
		`SELECT title FROM books WHERE id = $1`, bookID,
	).Scan(&bookTitle)

	activity.Record(c.Request.Context(), h.pool, userID, "followed_book",
		&bookID, nil, nil, nil, map[string]string{
			"book_ol_id": workID,
			"book_title": bookTitle,
		})

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// UnfollowBook removes a book follow for the current user.
//
// DELETE /books/:workId/follow
func (h *Handler) UnfollowBook(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	workID := c.Param("workId")

	_, err := h.pool.Exec(c.Request.Context(),
		`DELETE FROM book_follows
		 USING books b
		 WHERE book_follows.book_id = b.id
		   AND book_follows.user_id = $1
		   AND b.open_library_id = $2`,
		userID, workID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GetBookFollowStatus checks whether the current user follows a book.
//
// GET /books/:workId/follow
func (h *Handler) GetBookFollowStatus(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	workID := c.Param("workId")

	var exists bool
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT EXISTS(
			SELECT 1 FROM book_follows bf
			JOIN books b ON b.id = bf.book_id
			WHERE bf.user_id = $1 AND b.open_library_id = $2
		)`,
		userID, workID,
	).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"following": exists})
}

type followedBook struct {
	OpenLibraryID string  `json:"open_library_id"`
	Title         string  `json:"title"`
	CoverURL      *string `json:"cover_url"`
	Authors       *string `json:"authors"`
	CreatedAt     string  `json:"created_at"`
}

// GetFollowedBooks returns the list of books followed by the current user.
//
// GET /me/followed-books
func (h *Handler) GetFollowedBooks(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	rows, err := h.pool.Query(c.Request.Context(),
		`SELECT b.open_library_id, b.title, b.cover_url, b.authors, bf.created_at
		 FROM book_follows bf
		 JOIN books b ON b.id = bf.book_id
		 WHERE bf.user_id = $1
		 ORDER BY bf.created_at DESC`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer rows.Close()

	books := []followedBook{}
	for rows.Next() {
		var fb followedBook
		var t time.Time
		if err := rows.Scan(&fb.OpenLibraryID, &fb.Title, &fb.CoverURL, &fb.Authors, &t); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		fb.CreatedAt = t.Format(time.RFC3339)
		books = append(books, fb)
	}

	c.JSON(http.StatusOK, gin.H{"books": books})
}

type followedAuthor struct {
	AuthorKey  string `json:"author_key"`
	AuthorName string `json:"author_name"`
	CreatedAt  string `json:"created_at"`
}

// GetFollowedAuthors returns the list of authors followed by the current user.
//
// GET /me/followed-authors
func (h *Handler) GetFollowedAuthors(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	rows, err := h.pool.Query(c.Request.Context(),
		`SELECT author_key, author_name, created_at
		 FROM author_follows
		 WHERE user_id = $1
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer rows.Close()

	authors := []followedAuthor{}
	for rows.Next() {
		var a followedAuthor
		var t time.Time
		if err := rows.Scan(&a.AuthorKey, &a.AuthorName, &t); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		a.CreatedAt = t.Format(time.RFC3339)
		authors = append(authors, a)
	}

	c.JSON(http.StatusOK, gin.H{"authors": authors})
}
