package bookstats

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Refresh recalculates and upserts the precomputed stats for a single book.
// It is designed to be called fire-and-forget from write paths.
func Refresh(ctx context.Context, pool *pgxpool.Pool, bookID string) {
	_, err := pool.Exec(ctx,
		`INSERT INTO book_stats (book_id, reads_count, want_to_read_count, rating_sum, rating_count, review_count)
		 SELECT
		   b.id,
		   COUNT(*) FILTER (WHERE tv.slug = 'finished'),
		   COUNT(*) FILTER (WHERE tv.slug = 'want-to-read'),
		   COALESCE(SUM(ub.rating) FILTER (WHERE ub.rating IS NOT NULL), 0),
		   COUNT(ub.rating),
		   COUNT(*) FILTER (WHERE ub.review_text IS NOT NULL AND ub.review_text != '')
		 FROM books b
		 LEFT JOIN user_books ub ON ub.book_id = b.id
		 LEFT JOIN users u ON u.id = ub.user_id AND u.deleted_at IS NULL
		 LEFT JOIN tag_keys tk ON tk.user_id = ub.user_id AND tk.slug = 'status'
		 LEFT JOIN book_tag_values btv ON btv.user_id = ub.user_id AND btv.book_id = ub.book_id AND btv.tag_key_id = tk.id
		 LEFT JOIN tag_values tv ON tv.id = btv.tag_value_id
		 WHERE b.id = $1
		 GROUP BY b.id
		 ON CONFLICT (book_id) DO UPDATE SET
		   reads_count        = EXCLUDED.reads_count,
		   want_to_read_count = EXCLUDED.want_to_read_count,
		   rating_sum         = EXCLUDED.rating_sum,
		   rating_count       = EXCLUDED.rating_count,
		   review_count       = EXCLUDED.review_count,
		   updated_at         = NOW()`,
		bookID,
	)
	if err != nil {
		log.Printf("bookstats.Refresh(%s): %v", bookID, err)
	}
}

// RefreshByOLID is a convenience wrapper that looks up the book ID first.
func RefreshByOLID(ctx context.Context, pool *pgxpool.Pool, olID string) {
	var bookID string
	err := pool.QueryRow(ctx,
		`SELECT id FROM books WHERE open_library_id = $1`, olID,
	).Scan(&bookID)
	if err != nil {
		return
	}
	Refresh(ctx, pool, bookID)
}

// BackfillAll populates book_stats for every book in the catalog.
// Called once at startup.
func BackfillAll(ctx context.Context, pool *pgxpool.Pool) {
	_, err := pool.Exec(ctx,
		`INSERT INTO book_stats (book_id, reads_count, want_to_read_count, rating_sum, rating_count, review_count)
		 SELECT
		   b.id,
		   COUNT(*) FILTER (WHERE tv.slug = 'finished'),
		   COUNT(*) FILTER (WHERE tv.slug = 'want-to-read'),
		   COALESCE(SUM(ub.rating) FILTER (WHERE ub.rating IS NOT NULL), 0),
		   COUNT(ub.rating),
		   COUNT(*) FILTER (WHERE ub.review_text IS NOT NULL AND ub.review_text != '')
		 FROM books b
		 LEFT JOIN user_books ub ON ub.book_id = b.id
		 LEFT JOIN users u ON u.id = ub.user_id AND u.deleted_at IS NULL
		 LEFT JOIN tag_keys tk ON tk.user_id = ub.user_id AND tk.slug = 'status'
		 LEFT JOIN book_tag_values btv ON btv.user_id = ub.user_id AND btv.book_id = ub.book_id AND btv.tag_key_id = tk.id
		 LEFT JOIN tag_values tv ON tv.id = btv.tag_value_id
		 GROUP BY b.id
		 ON CONFLICT (book_id) DO UPDATE SET
		   reads_count        = EXCLUDED.reads_count,
		   want_to_read_count = EXCLUDED.want_to_read_count,
		   rating_sum         = EXCLUDED.rating_sum,
		   rating_count       = EXCLUDED.rating_count,
		   review_count       = EXCLUDED.review_count,
		   updated_at         = NOW()`)
	if err != nil {
		log.Printf("bookstats.BackfillAll: %v", err)
	} else {
		log.Println("bookstats: backfill complete")
	}
}
