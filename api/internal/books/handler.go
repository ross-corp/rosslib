package books

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// ── Open Library API types ────────────────────────────────────────────────────

type olDoc struct {
	Key             string   `json:"key"`
	Title           string   `json:"title"`
	AuthorName      []string `json:"author_name"`
	FirstPublishYear *int    `json:"first_publish_year"`
	ISBN            []string `json:"isbn"`
	CoverI          *int     `json:"cover_i"`
	EditionCount    int      `json:"edition_count"`
}

type olResponse struct {
	NumFound int     `json:"numFound"`
	Docs     []olDoc `json:"docs"`
}

// ── Response types ────────────────────────────────────────────────────────────

// BookResult is the normalized shape returned to clients.
type BookResult struct {
	// Key is the Open Library work key, e.g. "/works/OL82592W".
	// Use it to construct a canonical work URL: https://openlibrary.org<key>
	Key          string   `json:"key"`
	Title        string   `json:"title"`
	Authors      []string `json:"authors"`
	PublishYear  *int     `json:"publish_year"`
	ISBN         []string `json:"isbn"`
	CoverURL     *string  `json:"cover_url"`
	EditionCount int      `json:"edition_count"`
}

// ── Constants ─────────────────────────────────────────────────────────────────

const (
	olSearchURL  = "https://openlibrary.org/search.json"
	olCoverURL   = "https://covers.openlibrary.org/b/id/%d-M.jpg"
	olSearchFields = "key,title,author_name,first_publish_year,isbn,cover_i,edition_count"
	searchLimit  = 20
	maxISBNs     = 5
)

// ── Handler ───────────────────────────────────────────────────────────────────

// SearchBooks proxies a title query to the Open Library Search API and returns
// a normalized list of matching books.
//
// GET /books/search?q=<title>
func (h *Handler) SearchBooks(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

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
			Key:          doc.Key,
			Title:        doc.Title,
			Authors:      doc.AuthorName,
			PublishYear:  doc.FirstPublishYear,
			EditionCount: doc.EditionCount,
		}

		if len(doc.ISBN) > 0 {
			b.ISBN = doc.ISBN[:min(maxISBNs, len(doc.ISBN))]
		}

		if doc.CoverI != nil {
			coverURL := fmt.Sprintf(olCoverURL, *doc.CoverI)
			b.CoverURL = &coverURL
		}

		results = append(results, b)
	}

	c.JSON(http.StatusOK, gin.H{
		"total":   olResp.NumFound,
		"results": results,
	})
}
