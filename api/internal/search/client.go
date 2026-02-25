package search

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meilisearch/meilisearch-go"
)

const booksIndex = "books"

// BookDocument is the document shape stored in Meilisearch.
type BookDocument struct {
	ID              string   `json:"id"`
	OpenLibraryID   string   `json:"open_library_id"`
	Title           string   `json:"title"`
	Authors         string   `json:"authors"`
	ISBN13          string   `json:"isbn13"`
	PublicationYear int      `json:"publication_year"`
	CoverURL        string   `json:"cover_url"`
	Subjects        []string `json:"subjects"`
}

// Client wraps a Meilisearch index for book search.
type Client struct {
	meili meilisearch.ServiceManager
	index meilisearch.IndexManager
}

// NewClient creates a Meilisearch client and ensures the books index exists
// with correct settings. Returns nil (no error) if url is empty, allowing
// the app to run without Meilisearch.
func NewClient(url, apiKey string) (*Client, error) {
	if url == "" {
		return nil, nil
	}

	ms := meilisearch.New(url, meilisearch.WithAPIKey(apiKey))

	idx := ms.Index(booksIndex)

	// Create index if it doesn't exist (idempotent).
	_, err := ms.CreateIndex(&meilisearch.IndexConfig{
		Uid:        booksIndex,
		PrimaryKey: "id",
	})
	if err != nil {
		// Ignore "index already exists" errors.
		if msErr, ok := err.(*meilisearch.Error); ok && msErr.MeilisearchApiError.Code == "index_already_exists" {
			// fine
		} else {
			return nil, err
		}
	}

	// Configure searchable attributes.
	_, err = idx.UpdateSettings(&meilisearch.Settings{
		SearchableAttributes: []string{"title", "authors", "isbn13"},
		FilterableAttributes: []string{"open_library_id", "publication_year", "subjects"},
		SortableAttributes:   []string{"publication_year"},
	})
	if err != nil {
		return nil, err
	}

	return &Client{meili: ms, index: idx}, nil
}

// SyncBooks loads all books from the database and indexes them into Meilisearch.
// Called once at startup to ensure the index is populated.
func (c *Client) SyncBooks(ctx context.Context, pool *pgxpool.Pool) error {
	rows, err := pool.Query(ctx,
		`SELECT id, open_library_id, title,
		        COALESCE(authors, ''), COALESCE(isbn13, ''),
		        COALESCE(publication_year, 0), COALESCE(cover_url, ''),
		        COALESCE(subjects, '')
		 FROM books`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var docs []BookDocument
	for rows.Next() {
		var d BookDocument
		var subjectsRaw string
		if err := rows.Scan(&d.ID, &d.OpenLibraryID, &d.Title,
			&d.Authors, &d.ISBN13, &d.PublicationYear, &d.CoverURL,
			&subjectsRaw); err != nil {
			return err
		}
		if subjectsRaw != "" {
			d.Subjects = strings.Split(subjectsRaw, ", ")
		}
		docs = append(docs, d)
	}

	if len(docs) == 0 {
		return nil
	}

	// Index in batches of 1000.
	const batchSize = 1000
	for i := 0; i < len(docs); i += batchSize {
		end := i + batchSize
		if end > len(docs) {
			end = len(docs)
		}
		batch := docs[i:end]
		if _, err := c.index.AddDocuments(batch, nil); err != nil {
			return err
		}
	}

	log.Printf("meilisearch: synced %d books to index", len(docs))
	return nil
}

// IndexBook adds or updates a single book in the Meilisearch index.
func (c *Client) IndexBook(doc BookDocument) {
	if _, err := c.index.AddDocuments([]BookDocument{doc}, nil); err != nil {
		log.Printf("meilisearch: failed to index book %s: %v", doc.OpenLibraryID, err)
	}
}

// BrowseBooks returns documents matching a subject filter without requiring a
// search query. Used by genre pages to list books in a genre.
func (c *Client) BrowseBooks(subject string, limit, offset int) ([]BookDocument, int64, error) {
	req := &meilisearch.SearchRequest{
		Limit:  int64(limit),
		Offset: int64(offset),
		Filter: fmt.Sprintf(`subjects = "%s"`, subject),
	}

	resp, err := c.index.Search("", req)
	if err != nil {
		return nil, 0, err
	}

	var results []BookDocument
	if err := resp.Hits.DecodeInto(&results); err != nil {
		return nil, 0, err
	}
	return results, resp.EstimatedTotalHits, nil
}

// SearchBooks queries the Meilisearch index and returns matching documents.
// Optional yearMin/yearMax apply a publication_year range filter.
// Optional subject filters results to books with a matching subject.
func (c *Client) SearchBooks(query string, limit int, yearMin, yearMax int, subject string) ([]BookDocument, error) {
	req := &meilisearch.SearchRequest{
		Limit: int64(limit),
	}

	// Build Meilisearch filter for year range and subject.
	var filters []string
	if yearMin > 0 {
		filters = append(filters, fmt.Sprintf("publication_year >= %d", yearMin))
	}
	if yearMax > 0 {
		filters = append(filters, fmt.Sprintf("publication_year <= %d", yearMax))
	}
	if yearMin > 0 || yearMax > 0 {
		// Exclude books with no year data (publication_year == 0) when filtering.
		filters = append(filters, "publication_year > 0")
	}
	if subject != "" {
		filters = append(filters, fmt.Sprintf(`subjects = "%s"`, subject))
	}
	if len(filters) > 0 {
		req.Filter = strings.Join(filters, " AND ")
	}

	resp, err := c.index.Search(query, req)
	if err != nil {
		return nil, err
	}

	var results []BookDocument
	if err := resp.Hits.DecodeInto(&results); err != nil {
		return nil, err
	}
	return results, nil
}
