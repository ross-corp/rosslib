package bookstats

import (
	"log"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// bookStats holds the aggregated stats for a single book.
type bookStats struct {
	RatingSum       float64
	RatingCount     int
	ReviewCount     int
	ReadsCount      int
	WantToReadCount int
}

// BackfillAll recalculates book_stats for every book in the database
// using batch aggregation queries instead of per-book queries.
// It returns the number of rows upserted and any error encountered.
func BackfillAll(app core.App) (int, error) {
	statsMap := make(map[string]*bookStats)

	// Ensure every book gets an entry (even if all counts are zero).
	type bookRow struct {
		ID string `db:"id"`
	}
	var books []bookRow
	if err := app.DB().NewQuery("SELECT id FROM books").All(&books); err != nil {
		return 0, err
	}
	for _, b := range books {
		statsMap[b.ID] = &bookStats{}
	}

	// 1. Batch query: rating/review stats from user_books.
	type ratingRow struct {
		Book        string  `db:"book"`
		RatingSum   float64 `db:"rating_sum"`
		RatingCount int     `db:"rating_count"`
		ReviewCount int     `db:"review_count"`
	}
	var ratingRows []ratingRow
	err := app.DB().NewQuery(`
		SELECT
			book,
			COALESCE(SUM(CASE WHEN rating > 0 THEN rating ELSE 0 END), 0) as rating_sum,
			COALESCE(SUM(CASE WHEN rating > 0 THEN 1 ELSE 0 END), 0) as rating_count,
			COALESCE(SUM(CASE WHEN review_text != '' AND review_text IS NOT NULL THEN 1 ELSE 0 END), 0) as review_count
		FROM user_books
		GROUP BY book
	`).All(&ratingRows)
	if err != nil {
		return 0, err
	}
	for _, r := range ratingRows {
		s := statsMap[r.Book]
		if s == nil {
			s = &bookStats{}
			statsMap[r.Book] = s
		}
		s.RatingSum = r.RatingSum
		s.RatingCount = r.RatingCount
		s.ReviewCount = r.ReviewCount
	}

	// 2. Batch query: reads count (tag slug = 'finished').
	type tagCountRow struct {
		Book  string `db:"book"`
		Count int    `db:"count"`
	}
	var readRows []tagCountRow
	err = app.DB().NewQuery(`
		SELECT btv.book, COUNT(DISTINCT btv.user) as count
		FROM book_tag_values btv
		JOIN tag_values tv ON btv.tag_value = tv.id
		WHERE tv.slug = 'finished'
		GROUP BY btv.book
	`).All(&readRows)
	if err != nil {
		log.Printf("[BookStats] warning: reads count query failed: %v", err)
	}
	for _, r := range readRows {
		s := statsMap[r.Book]
		if s == nil {
			s = &bookStats{}
			statsMap[r.Book] = s
		}
		s.ReadsCount = r.Count
	}

	// 3. Batch query: want-to-read count (tag slug = 'want-to-read').
	var wtrRows []tagCountRow
	err = app.DB().NewQuery(`
		SELECT btv.book, COUNT(DISTINCT btv.user) as count
		FROM book_tag_values btv
		JOIN tag_values tv ON btv.tag_value = tv.id
		WHERE tv.slug = 'want-to-read'
		GROUP BY btv.book
	`).All(&wtrRows)
	if err != nil {
		log.Printf("[BookStats] warning: want-to-read count query failed: %v", err)
	}
	for _, r := range wtrRows {
		s := statsMap[r.Book]
		if s == nil {
			s = &bookStats{}
			statsMap[r.Book] = s
		}
		s.WantToReadCount = r.Count
	}

	// 4. Load existing book_stats records into a map for fast lookup.
	existingMap := make(map[string]*core.Record)
	existingRecords, err := app.FindAllRecords("book_stats")
	if err == nil {
		for _, rec := range existingRecords {
			existingMap[rec.GetString("book")] = rec
		}
	}

	coll, err := app.FindCollectionByNameOrId("book_stats")
	if err != nil {
		return 0, err
	}

	// 5. Upsert all stats.
	updated := 0
	for bookID, s := range statsMap {
		rec, ok := existingMap[bookID]
		if !ok {
			rec = core.NewRecord(coll)
			rec.Set("book", bookID)
		}
		rec.Set("reads_count", s.ReadsCount)
		rec.Set("want_to_read_count", s.WantToReadCount)
		rec.Set("rating_sum", s.RatingSum)
		rec.Set("rating_count", s.RatingCount)
		rec.Set("review_count", s.ReviewCount)

		if err := app.Save(rec); err != nil {
			log.Printf("[BookStats] error saving stats for book %s: %v", bookID, err)
			continue
		}
		updated++
	}

	return updated, nil
}

// StartPoller runs BackfillAll once immediately, then every 24 hours.
// It blocks forever, so it should be called from a goroutine.
func StartPoller(app core.App) {
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
}
