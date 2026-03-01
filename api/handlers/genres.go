package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

// ListGenres handles GET /genres
// Returns a list of genres with book counts, derived from the "subjects" field on books.
func ListGenres(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		// Get all books that have subjects
		books, err := app.FindRecordsByFilter("books",
			"subjects != ''", "", 0, 0, nil,
		)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}

		// Count books per genre (first subject = primary genre)
		genreCounts := map[string]int{}    // slug -> count
		genreNames := map[string]string{}  // slug -> display name

		for _, b := range books {
			subj := b.GetString("subjects")
			if subj == "" {
				continue
			}
			// Use all subjects as genres
			for _, part := range strings.Split(subj, ",") {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				slug := slugify(part)
				genreCounts[slug]++
				if _, ok := genreNames[slug]; !ok {
					genreNames[slug] = part
				}
			}
		}

		// Build sorted result (by count descending)
		type genreInfo struct {
			Slug      string `json:"slug"`
			Name      string `json:"name"`
			BookCount int    `json:"book_count"`
		}

		var results []genreInfo
		for slug, count := range genreCounts {
			results = append(results, genreInfo{
				Slug:      slug,
				Name:      genreNames[slug],
				BookCount: count,
			})
		}

		// Sort by book_count descending
		for i := 0; i < len(results); i++ {
			for j := i + 1; j < len(results); j++ {
				if results[j].BookCount > results[i].BookCount {
					results[i], results[j] = results[j], results[i]
				}
			}
		}

		if results == nil {
			results = []genreInfo{}
		}

		return e.JSON(http.StatusOK, results)
	}
}

// GetGenreBooks handles GET /genres/{slug}/books?page=1&limit=20&sort=title|rating|year
func GetGenreBooks(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		slug := e.Request.PathValue("slug")

		page := 1
		if p := e.Request.URL.Query().Get("page"); p != "" {
			if n, err := strconv.Atoi(p); err == nil && n > 0 {
				page = n
			}
		}
		limit := 20
		if l := e.Request.URL.Query().Get("limit"); l != "" {
			if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
				limit = n
			}
		}
		sortParam := e.Request.URL.Query().Get("sort")

		// Find all books with subjects (we need to filter in Go since subjects is comma-separated)
		allBooks, err := app.FindRecordsByFilter("books",
			"subjects != ''", "", 0, 0, nil,
		)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "genre not found"})
		}

		// Filter books whose subjects contain this genre slug
		var matchedBooks []*core.Record
		genreName := ""
		for _, b := range allBooks {
			subj := b.GetString("subjects")
			for _, part := range strings.Split(subj, ",") {
				part = strings.TrimSpace(part)
				if slugify(part) == slug {
					matchedBooks = append(matchedBooks, b)
					if genreName == "" {
						genreName = part
					}
					break
				}
			}
		}

		if len(matchedBooks) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "genre not found"})
		}

		// Batch-fetch book_stats for ratings/counts
		statsMap := map[string]*core.Record{}
		if len(matchedBooks) > 0 {
			var bookIDs []any
			for _, b := range matchedBooks {
				bookIDs = append(bookIDs, b.Id)
			}
			allStats, _ := app.FindRecordsByFilter("book_stats",
				"book IN {:ids}", "", 0, 0,
				map[string]any{"ids": bookIDs},
			)
			for _, s := range allStats {
				statsMap[s.GetString("book")] = s
			}
		}

		// Sort matched books
		switch sortParam {
		case "title":
			for i := 0; i < len(matchedBooks); i++ {
				for j := i + 1; j < len(matchedBooks); j++ {
					if strings.ToLower(matchedBooks[i].GetString("title")) > strings.ToLower(matchedBooks[j].GetString("title")) {
						matchedBooks[i], matchedBooks[j] = matchedBooks[j], matchedBooks[i]
					}
				}
			}
		case "rating":
			// Sort by average rating descending (books without ratings go last)
			for i := 0; i < len(matchedBooks); i++ {
				for j := i + 1; j < len(matchedBooks); j++ {
					ri := avgRatingForBook(statsMap, matchedBooks[i].Id)
					rj := avgRatingForBook(statsMap, matchedBooks[j].Id)
					if rj > ri {
						matchedBooks[i], matchedBooks[j] = matchedBooks[j], matchedBooks[i]
					}
				}
			}
		case "year":
			// Sort by publication year descending (newest first)
			for i := 0; i < len(matchedBooks); i++ {
				for j := i + 1; j < len(matchedBooks); j++ {
					yi := matchedBooks[i].GetInt("publication_year")
					yj := matchedBooks[j].GetInt("publication_year")
					if yj > yi {
						matchedBooks[i], matchedBooks[j] = matchedBooks[j], matchedBooks[i]
					}
				}
			}
		default:
			// Default: no explicit sort (order as stored)
		}

		// Paginate
		total := len(matchedBooks)
		offset := (page - 1) * limit
		end := offset + limit
		if offset >= total {
			return e.JSON(http.StatusOK, map[string]any{
				"genre":   genreName,
				"total":   total,
				"page":    page,
				"results": []any{},
			})
		}
		if end > total {
			end = total
		}
		pageBooks := matchedBooks[offset:end]

		// Build results
		var results []map[string]any
		for _, b := range pageBooks {
			olid := b.GetString("open_library_id")

			var avgRating *float64
			var ratingCount, alreadyReadCount int
			if s, ok := statsMap[b.Id]; ok {
				if rc := s.GetInt("rating_count"); rc > 0 {
					avg := s.GetFloat("rating_sum") / float64(rc)
					avgRating = &avg
				}
				ratingCount = s.GetInt("rating_count")
				alreadyReadCount = s.GetInt("reads_count")
			}

			var subjects []string
			if s := b.GetString("subjects"); s != "" {
				for _, part := range strings.Split(s, ",") {
					part = strings.TrimSpace(part)
					if part != "" {
						subjects = append(subjects, part)
						if len(subjects) >= 3 {
							break
						}
					}
				}
			}

			results = append(results, map[string]any{
				"key":                olid,
				"title":              b.GetString("title"),
				"authors":            splitAuthors(b.GetString("authors")),
				"publish_year":       b.GetInt("publication_year"),
				"cover_url":          b.GetString("cover_url"),
				"edition_count":      0,
				"average_rating":     avgRating,
				"rating_count":       ratingCount,
				"already_read_count": alreadyReadCount,
				"subjects":           subjects,
			})
		}

		if results == nil {
			results = []map[string]any{}
		}

		return e.JSON(http.StatusOK, map[string]any{
			"genre":   genreName,
			"total":   total,
			"page":    page,
			"results": results,
		})
	}
}

// avgRatingForBook returns the average rating for a book, or 0 if no ratings.
func avgRatingForBook(statsMap map[string]*core.Record, bookID string) float64 {
	if s, ok := statsMap[bookID]; ok {
		if rc := s.GetInt("rating_count"); rc > 0 {
			return s.GetFloat("rating_sum") / float64(rc)
		}
	}
	return 0
}
