package bookstats

import (
	"log"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// BackfillAll recalculates book_stats for every book in the database.
// It returns the number of rows updated and any error encountered.
func BackfillAll(app core.App) (int, error) {
	type bookRow struct {
		ID string `db:"id"`
	}

	var books []bookRow
	err := app.DB().NewQuery("SELECT id FROM books").All(&books)
	if err != nil {
		return 0, err
	}

	updated := 0
	for _, b := range books {
		if err := refreshStats(app, b.ID); err != nil {
			log.Printf("[BookStats] error refreshing stats for book %s: %v", b.ID, err)
			continue
		}
		updated++
	}

	return updated, nil
}

// refreshStats recalculates and upserts book_stats for a single book.
func refreshStats(app core.App, bookID string) error {
	type statsResult struct {
		RatingSum   float64 `db:"rating_sum"`
		RatingCount int     `db:"rating_count"`
		ReviewCount int     `db:"review_count"`
	}
	var stats statsResult
	err := app.DB().NewQuery(`
		SELECT
			COALESCE(SUM(CASE WHEN rating > 0 THEN rating ELSE 0 END), 0) as rating_sum,
			COALESCE(SUM(CASE WHEN rating > 0 THEN 1 ELSE 0 END), 0) as rating_count,
			COALESCE(SUM(CASE WHEN review_text != '' AND review_text IS NOT NULL THEN 1 ELSE 0 END), 0) as review_count
		FROM user_books WHERE book = {:book}
	`).Bind(map[string]any{"book": bookID}).One(&stats)
	if err != nil {
		return err
	}

	type countResult struct {
		Count int `db:"count"`
	}

	var readsCount countResult
	_ = app.DB().NewQuery(`
		SELECT COUNT(DISTINCT btv.user) as count
		FROM book_tag_values btv
		JOIN tag_values tv ON btv.tag_value = tv.id
		WHERE btv.book = {:book} AND tv.slug = 'finished'
	`).Bind(map[string]any{"book": bookID}).One(&readsCount)

	var wtrCount countResult
	_ = app.DB().NewQuery(`
		SELECT COUNT(DISTINCT btv.user) as count
		FROM book_tag_values btv
		JOIN tag_values tv ON btv.tag_value = tv.id
		WHERE btv.book = {:book} AND tv.slug = 'want-to-read'
	`).Bind(map[string]any{"book": bookID}).One(&wtrCount)

	existing, err := app.FindRecordsByFilter("book_stats",
		"book = {:book}", "", 1, 0,
		map[string]any{"book": bookID},
	)
	var rec *core.Record
	if err == nil && len(existing) > 0 {
		rec = existing[0]
	} else {
		coll, err := app.FindCollectionByNameOrId("book_stats")
		if err != nil {
			return err
		}
		rec = core.NewRecord(coll)
		rec.Set("book", bookID)
	}
	rec.Set("reads_count", readsCount.Count)
	rec.Set("want_to_read_count", wtrCount.Count)
	rec.Set("rating_sum", stats.RatingSum)
	rec.Set("rating_count", stats.RatingCount)
	rec.Set("review_count", stats.ReviewCount)

	return app.Save(rec)
}

// StartPoller launches a background goroutine that runs BackfillAll
// once immediately on startup, then every 24 hours thereafter.
func StartPoller(app core.App) {
	go func() {
		// Run once on startup.
		start := time.Now()
		updated, err := BackfillAll(app)
		if err != nil {
			log.Printf("[BookStats] startup backfill error: %v", err)
		} else {
			log.Printf("[BookStats] startup backfill complete: %d books updated in %s", updated, time.Since(start).Round(time.Millisecond))
		}

		// Then run every 24 hours.
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			start := time.Now()
			updated, err := BackfillAll(app)
			if err != nil {
				log.Printf("[BookStats] periodic backfill error: %v", err)
			} else {
				log.Printf("[BookStats] periodic backfill complete: %d books updated in %s", updated, time.Since(start).Round(time.Millisecond))
			}
		}
	}()
}
