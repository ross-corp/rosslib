package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

var ghostPersonas = []struct {
	Username    string
	DisplayName string
	Bio         string
}{
	{"ghost_bookworm", "Avid Reader", "I read everything I can get my hands on."},
	{"ghost_scifi", "Sci-Fi Fan", "Science fiction is my escape from reality."},
	{"ghost_mystery", "Mystery Maven", "Always looking for the next great mystery."},
	{"ghost_romance", "Romance Reader", "Love stories are my weakness."},
	{"ghost_history", "History Buff", "I learn from the past through great books."},
	{"ghost_fantasy", "Fantasy Lover", "Lost in worlds of magic and wonder."},
	{"ghost_nonfiction", "Non-Fiction Nerd", "Give me facts and real stories."},
	{"ghost_literary", "Literary Lion", "Classic and contemporary literary fiction."},
	{"ghost_thriller", "Thriller Seeker", "I love a good page-turner."},
	{"ghost_ya", "YA Enthusiast", "Young adult fiction speaks to everyone."},
}

// SeedGhosts handles POST /admin/ghosts/seed
func SeedGhosts(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		usersColl, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		var createdNames []string
		for _, persona := range ghostPersonas {
			// Check if already exists
			existing, _ := app.FindRecordsByFilter("users",
				"username = {:username}", "", 1, 0,
				map[string]any{"username": persona.Username},
			)
			if len(existing) > 0 {
				continue
			}

			rec := core.NewRecord(usersColl)
			rec.Set("username", persona.Username)
			rec.Set("email", persona.Username+"@ghost.rosslib.local")
			rec.Set("display_name", persona.DisplayName)
			rec.Set("bio", persona.Bio)
			rec.Set("is_ghost", true)
			rec.SetPassword("ghost_password_" + persona.Username)
			if err := app.Save(rec); err != nil {
				continue
			}

			// Create default shelves and status tags
			_ = createDefaultShelves(app, rec.Id)
			_, _, _ = ensureStatusTagKey(app, rec.Id)

			createdNames = append(createdNames, persona.Username)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"message": fmt.Sprintf("Created %d ghost users", len(createdNames)),
			"created": createdNames,
		})
	}
}

// SimulateGhosts handles POST /admin/ghosts/simulate
func SimulateGhosts(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		// Find all ghost users
		ghosts, err := app.FindRecordsByFilter("users",
			"is_ghost = true", "", 100, 0, nil,
		)
		if err != nil || len(ghosts) == 0 {
			return e.JSON(http.StatusOK, map[string]any{
				"results": []any{},
				"error":   "No ghost users found. Seed ghosts first.",
			})
		}

		// Get some books to work with
		books, _ := app.FindRecordsByFilter("books", "1=1", "-created", 200, 0, nil)
		if len(books) == 0 {
			return e.JSON(http.StatusOK, map[string]any{
				"results": []any{},
				"error":   "No books in the database. Search for and add some books first.",
			})
		}

		rng := rand.New(rand.NewSource(time.Now().UnixNano()))

		type ghostResult struct {
			Ghost   string   `json:"ghost"`
			Actions []string `json:"actions"`
		}
		var results []ghostResult

		for _, ghost := range ghosts {
			gr := ghostResult{
				Ghost:   ghost.GetString("username"),
				Actions: []string{},
			}

			// Find books this ghost hasn't rated yet
			ratedBooks, _ := app.FindRecordsByFilter("user_books",
				"user = {:user}", "", 1000, 0,
				map[string]any{"user": ghost.Id},
			)
			ratedSet := make(map[string]bool)
			for _, rb := range ratedBooks {
				ratedSet[rb.GetString("book")] = true
			}

			var available []*core.Record
			for _, b := range books {
				if !ratedSet[b.Id] {
					available = append(available, b)
				}
			}

			if len(available) == 0 {
				gr.Actions = append(gr.Actions, "Already rated all available books")
				results = append(results, gr)
				continue
			}

			// Shuffle available books
			rng.Shuffle(len(available), func(i, j int) {
				available[i], available[j] = available[j], available[i]
			})

			numBooks := rng.Intn(5) + 1
			if numBooks > len(available) {
				numBooks = len(available)
			}

			for i := 0; i < numBooks; i++ {
				book := available[i]

				coll, _ := app.FindCollectionByNameOrId("user_books")
				ub := core.NewRecord(coll)
				ub.Set("user", ghost.Id)
				ub.Set("book", book.Id)
				rating := rng.Intn(5) + 1
				ub.Set("rating", rating)
				ub.Set("date_added", time.Now().UTC().Format(time.RFC3339))
				if err := app.Save(ub); err != nil {
					gr.Actions = append(gr.Actions, fmt.Sprintf("Failed to rate \"%s\": %v", book.GetString("title"), err))
					continue
				}

				setStatusTag(app, ghost.Id, book.Id, "finished")
				refreshBookStats(app, book.Id)
				gr.Actions = append(gr.Actions, fmt.Sprintf("Rated \"%s\" %d/5", book.GetString("title"), rating))
			}

			results = append(results, gr)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"results": results,
		})
	}
}

// GetGhostStatus handles GET /admin/ghosts/status
func GetGhostStatus(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		ghosts, err := app.FindRecordsByFilter("users", "is_ghost = true", "", 100, 0, nil)
		if err != nil || len(ghosts) == 0 {
			return e.JSON(http.StatusOK, []any{})
		}

		type ghostStatus struct {
			Username       string `json:"username"`
			DisplayName    string `json:"display_name"`
			UserID         string `json:"user_id"`
			BooksRead      int    `json:"books_read"`
			CurrentlyReading int  `json:"currently_reading"`
			WantToRead     int    `json:"want_to_read"`
			FollowingCount int    `json:"following_count"`
			FollowersCount int    `json:"followers_count"`
		}

		type countResult struct {
			Count int `db:"count"`
		}

		var result []ghostStatus
		for _, g := range ghosts {
			gs := ghostStatus{
				Username:    g.GetString("username"),
				DisplayName: g.GetString("display_name"),
				UserID:      g.Id,
			}

			var c countResult

			// Books read (status tag = "finished")
			_ = app.DB().NewQuery(`
				SELECT COUNT(*) as count FROM book_tag_values btv
				JOIN tag_values tv ON btv.tag_value = tv.id
				JOIN tag_keys tk ON tv.tag_key = tk.id
				WHERE tk.user = {:uid} AND tk.name = 'Status' AND tv.value = 'finished'
			`).Bind(map[string]any{"uid": g.Id}).One(&c)
			gs.BooksRead = c.Count

			// Currently reading
			c.Count = 0
			_ = app.DB().NewQuery(`
				SELECT COUNT(*) as count FROM book_tag_values btv
				JOIN tag_values tv ON btv.tag_value = tv.id
				JOIN tag_keys tk ON tv.tag_key = tk.id
				WHERE tk.user = {:uid} AND tk.name = 'Status' AND tv.value = 'reading'
			`).Bind(map[string]any{"uid": g.Id}).One(&c)
			gs.CurrentlyReading = c.Count

			// Want to read
			c.Count = 0
			_ = app.DB().NewQuery(`
				SELECT COUNT(*) as count FROM book_tag_values btv
				JOIN tag_values tv ON btv.tag_value = tv.id
				JOIN tag_keys tk ON tv.tag_key = tk.id
				WHERE tk.user = {:uid} AND tk.name = 'Status' AND tv.value = 'to-read'
			`).Bind(map[string]any{"uid": g.Id}).One(&c)
			gs.WantToRead = c.Count

			// Following count
			c.Count = 0
			_ = app.DB().NewQuery(`
				SELECT COUNT(*) as count FROM follows WHERE follower = {:uid} AND status = 'active'
			`).Bind(map[string]any{"uid": g.Id}).One(&c)
			gs.FollowingCount = c.Count

			// Followers count
			c.Count = 0
			_ = app.DB().NewQuery(`
				SELECT COUNT(*) as count FROM follows WHERE following = {:uid} AND status = 'active'
			`).Bind(map[string]any{"uid": g.Id}).One(&c)
			gs.FollowersCount = c.Count

			result = append(result, gs)
		}

		return e.JSON(http.StatusOK, result)
	}
}
