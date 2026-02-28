package handlers

import (
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// GetYearInReview handles GET /users/{username}/year-in-review?year=2025
func GetYearInReview(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		username := e.Request.PathValue("username")

		users, err := app.FindRecordsByFilter("users",
			"username = {:username}", "", 1, 0,
			map[string]any{"username": username},
		)
		if err != nil || len(users) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "User not found"})
		}
		user := users[0]

		viewerID := ""
		if e.Auth != nil {
			viewerID = e.Auth.Id
		}
		if !canViewProfile(app, viewerID, user) {
			return e.JSON(http.StatusForbidden, map[string]any{"error": "Profile is private"})
		}

		// Parse year parameter, default to current year
		year, _ := strconv.Atoi(e.Request.URL.Query().Get("year"))
		if year <= 0 {
			year = time.Now().Year()
		}

		uid := user.Id
		yearStart := strconv.Itoa(year) + "-01-01 00:00:00.000Z"
		yearEnd := strconv.Itoa(year+1) + "-01-01 00:00:00.000Z"

		// Query finished books with date_read in the requested year
		type bookRow struct {
			BookID    string   `db:"book_id"`
			OLID      string   `db:"open_library_id"`
			Title     string   `db:"title"`
			CoverURL  *string  `db:"cover_url"`
			Authors   *string  `db:"authors"`
			PageCount *int     `db:"page_count"`
			Subjects  *string  `db:"subjects"`
			Rating    *float64 `db:"rating"`
			DateRead  string   `db:"date_read"`
		}

		var books []bookRow
		err = app.DB().NewQuery(`
			SELECT b.id as book_id, b.open_library_id, b.title,
				   COALESCE(NULLIF(ub.selected_edition_cover_url, ''), b.cover_url) as cover_url,
				   b.authors, b.page_count, b.subjects,
				   ub.rating, ub.date_read
			FROM user_books ub
			JOIN books b ON ub.book = b.id
			JOIN book_tag_values btv ON btv.user = ub.user AND btv.book = ub.book
			JOIN tag_keys tk ON btv.tag_key = tk.id
			JOIN tag_values tv ON btv.tag_value = tv.id
			WHERE ub.user = {:user}
			  AND tk.slug = 'status' AND tv.slug = 'finished'
			  AND ub.date_read >= {:yearStart}
			  AND ub.date_read < {:yearEnd}
			ORDER BY ub.date_read ASC
		`).Bind(map[string]any{
			"user":      uid,
			"yearStart": yearStart,
			"yearEnd":   yearEnd,
		}).All(&books)
		if err != nil {
			books = []bookRow{}
		}

		// Compute stats
		totalBooks := len(books)
		totalPages := 0
		var ratingSum float64
		ratingCount := 0
		var highestRatedBook map[string]any
		var highestRating float64
		var longestBook map[string]any
		var longestPages int
		var shortestBook map[string]any
		shortestPages := math.MaxInt32

		// Genre counting
		genreCounts := map[string]int{}

		// Books per month
		monthCounts := map[int]int{}
		monthBooks := map[int][]map[string]any{}

		for _, b := range books {
			if b.PageCount != nil && *b.PageCount > 0 {
				totalPages += *b.PageCount
			}

			if b.Rating != nil && *b.Rating > 0 {
				ratingSum += *b.Rating
				ratingCount++
				if *b.Rating > highestRating {
					highestRating = *b.Rating
					highestRatedBook = map[string]any{
						"open_library_id": b.OLID,
						"title":           b.Title,
						"cover_url":       b.CoverURL,
						"rating":          *b.Rating,
					}
				}
			}

			if b.PageCount != nil && *b.PageCount > 0 {
				if *b.PageCount > longestPages {
					longestPages = *b.PageCount
					longestBook = map[string]any{
						"open_library_id": b.OLID,
						"title":           b.Title,
						"cover_url":       b.CoverURL,
						"page_count":      *b.PageCount,
					}
				}
				if *b.PageCount < shortestPages {
					shortestPages = *b.PageCount
					shortestBook = map[string]any{
						"open_library_id": b.OLID,
						"title":           b.Title,
						"cover_url":       b.CoverURL,
						"page_count":      *b.PageCount,
					}
				}
			}

			// Count genres from subjects
			if b.Subjects != nil && *b.Subjects != "" {
				for _, subj := range strings.Split(*b.Subjects, ",") {
					subj = strings.TrimSpace(subj)
					if subj != "" {
						genreCounts[subj]++
					}
				}
			}

			// Parse date_read month
			t, parseErr := time.Parse(time.RFC3339Nano, b.DateRead)
			if parseErr != nil {
				t, parseErr = time.Parse("2006-01-02 15:04:05.000Z", b.DateRead)
				if parseErr != nil {
					t, parseErr = time.Parse("2006-01-02", b.DateRead)
				}
			}
			if parseErr == nil {
				m := int(t.Month())
				monthCounts[m]++
				monthBooks[m] = append(monthBooks[m], map[string]any{
					"open_library_id": b.OLID,
					"title":           b.Title,
					"cover_url":       b.CoverURL,
					"rating":          b.Rating,
				})
			}
		}

		// Average rating
		var avgRating *float64
		if ratingCount > 0 {
			avg := ratingSum / float64(ratingCount)
			avgRating = &avg
		}

		// Top genres (top 5)
		type genreEntry struct {
			Name  string `json:"name"`
			Count int    `json:"count"`
		}
		var genres []genreEntry
		for name, count := range genreCounts {
			genres = append(genres, genreEntry{Name: name, Count: count})
		}
		sort.Slice(genres, func(i, j int) bool {
			return genres[i].Count > genres[j].Count
		})
		if len(genres) > 5 {
			genres = genres[:5]
		}
		if genres == nil {
			genres = []genreEntry{}
		}

		// Build books_by_month array
		type monthGroup struct {
			Month int              `json:"month"`
			Count int              `json:"count"`
			Books []map[string]any `json:"books"`
		}
		var byMonth []monthGroup
		for m := 1; m <= 12; m++ {
			if c, ok := monthCounts[m]; ok {
				byMonth = append(byMonth, monthGroup{
					Month: m,
					Count: c,
					Books: monthBooks[m],
				})
			}
		}
		if byMonth == nil {
			byMonth = []monthGroup{}
		}

		// Collect available years for navigation
		type yearRow struct {
			Year int `db:"year"`
		}
		var availableYears []int
		var yearRows []yearRow
		_ = app.DB().NewQuery(`
			SELECT DISTINCT CAST(strftime('%Y', ub.date_read) AS INTEGER) as year
			FROM user_books ub
			JOIN book_tag_values btv ON btv.user = ub.user AND btv.book = ub.book
			JOIN tag_keys tk ON btv.tag_key = tk.id
			JOIN tag_values tv ON btv.tag_value = tv.id
			WHERE ub.user = {:user}
			  AND tk.slug = 'status' AND tv.slug = 'finished'
			  AND ub.date_read IS NOT NULL AND ub.date_read != ''
			ORDER BY year DESC
		`).Bind(map[string]any{"user": uid}).All(&yearRows)
		for _, yr := range yearRows {
			if yr.Year > 0 {
				availableYears = append(availableYears, yr.Year)
			}
		}
		if availableYears == nil {
			availableYears = []int{}
		}

		result := map[string]any{
			"year":             year,
			"total_books":      totalBooks,
			"total_pages":      totalPages,
			"average_rating":   avgRating,
			"highest_rated":    highestRatedBook,
			"longest_book":     longestBook,
			"shortest_book":    shortestBook,
			"top_genres":       genres,
			"books_by_month":   byMonth,
			"available_years":  availableYears,
		}

		return e.JSON(http.StatusOK, result)
	}
}
