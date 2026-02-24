package search

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meilisearch/meilisearch-go"
)

const booksIndex = "books"

// BookDocument is the document shape stored in Meilisearch.
type BookDocument struct {
	ID              string `json:"id"`
	OpenLibraryID   string `json:"open_library_id"`
	Title           string `json:"title"`
	Authors         string `json:"authors"`
	ISBN13          string `json:"isbn13"`
	PublicationYear int    `json:"publication_year"`
	CoverURL        string `json:"cover_url"`
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
		FilterableAttributes: []string{"open_library_id", "publication_year"},
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
		        COALESCE(publication_year, 0), COALESCE(cover_url, '')
		 FROM books`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var docs []BookDocument
	for rows.Next() {
		var d BookDocument
		if err := rows.Scan(&d.ID, &d.OpenLibraryID, &d.Title,
			&d.Authors, &d.ISBN13, &d.PublicationYear, &d.CoverURL); err != nil {
			return err
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

// SearchBooks queries the Meilisearch index and returns matching documents.
func (c *Client) SearchBooks(query string, limit int) ([]BookDocument, error) {
	resp, err := c.index.Search(query, &meilisearch.SearchRequest{
		Limit: int64(limit),
	})
	if err != nil {
		return nil, err
	}

	var results []BookDocument
	if err := resp.Hits.DecodeInto(&results); err != nil {
		return nil, err
	}
	return results, nil
}
