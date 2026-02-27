package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// GetReadingTimeline handles GET /users/{username}/timeline?year=2026
func GetReadingTimeline(app core.App) func(e *core.RequestEvent) error {
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

		// Query finished books with date_read in the requested year
		type bookRow struct {
			BookID   string   `db:"book_id" json:"book_id"`
			OLID     string   `db:"open_library_id" json:"open_library_id"`
			Title    string   `db:"title" json:"title"`
			CoverURL *string  `db:"cover_url" json:"cover_url"`
			Rating   *float64 `db:"rating" json:"rating"`
			DateRead string   `db:"date_read" json:"date_read"`
		}

		yearStart := strconv.Itoa(year) + "-01-01 00:00:00.000Z"
		yearEnd := strconv.Itoa(year+1) + "-01-01 00:00:00.000Z"

		var books []bookRow
		err = app.DB().NewQuery(`
			SELECT b.id as book_id, b.open_library_id, b.title,
				   COALESCE(NULLIF(ub.selected_edition_cover_url, ''), b.cover_url) as cover_url,
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
			"user":      user.Id,
			"yearStart": yearStart,
			"yearEnd":   yearEnd,
		}).All(&books)
		if err != nil {
			books = []bookRow{}
		}

		// Group books by month
		type monthGroup struct {
			Month int       `json:"month"`
			Books []bookRow `json:"books"`
		}

		monthMap := map[int][]bookRow{}
		for _, b := range books {
			t, err := time.Parse(time.RFC3339Nano, b.DateRead)
			if err != nil {
				// Try alternate formats PocketBase may use
				t, err = time.Parse("2006-01-02 15:04:05.000Z", b.DateRead)
				if err != nil {
					t, err = time.Parse("2006-01-02", b.DateRead)
					if err != nil {
						continue
					}
				}
			}
			m := int(t.Month())
			monthMap[m] = append(monthMap[m], b)
		}

		var months []monthGroup
		for m := 1; m <= 12; m++ {
			if bks, ok := monthMap[m]; ok {
				months = append(months, monthGroup{Month: m, Books: bks})
			}
		}
		if months == nil {
			months = []monthGroup{}
		}

		return e.JSON(http.StatusOK, map[string]any{
			"year":   year,
			"months": months,
		})
	}
}
