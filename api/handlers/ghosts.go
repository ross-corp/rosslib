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

		created := 0
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

			created++
		}

		return e.JSON(http.StatusOK, map[string]any{
			"message": fmt.Sprintf("Created %d ghost users", created),
			"created": created,
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
			return e.JSON(http.StatusOK, map[string]any{"message": "No ghost users found"})
		}

		// Get some books to work with
		books, _ := app.FindRecordsByFilter("books", "1=1", "-created", 50, 0, nil)
		if len(books) == 0 {
			return e.JSON(http.StatusOK, map[string]any{"message": "No books in database"})
		}

		actions := 0
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))

		for _, ghost := range ghosts {
			// Each ghost interacts with a few random books
			numBooks := rng.Intn(5) + 1
			for i := 0; i < numBooks && i < len(books); i++ {
				book := books[rng.Intn(len(books))]

				// Create user_book with random rating
				existing, _ := app.FindRecordsByFilter("user_books",
					"user = {:user} && book = {:book}",
					"", 1, 0,
					map[string]any{"user": ghost.Id, "book": book.Id},
				)
				if len(existing) > 0 {
					continue
				}

				coll, _ := app.FindCollectionByNameOrId("user_books")
				ub := core.NewRecord(coll)
				ub.Set("user", ghost.Id)
				ub.Set("book", book.Id)
				ub.Set("rating", rng.Intn(5)+1)
				ub.Set("date_added", time.Now().UTC().Format(time.RFC3339))
				if err := app.Save(ub); err != nil {
					continue
				}

				// Set status tag to "finished"
				setStatusTag(app, ghost.Id, book.Id, "finished")
				refreshBookStats(app, book.Id)
				actions++
			}
		}

		return e.JSON(http.StatusOK, map[string]any{
			"message": fmt.Sprintf("Simulated %d actions", actions),
			"actions": actions,
		})
	}
}

// GetGhostStatus handles GET /admin/ghosts/status
func GetGhostStatus(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		type countResult struct {
			Count int `db:"count"`
		}
		var ghostCount countResult
		_ = app.DB().NewQuery("SELECT COUNT(*) as count FROM users WHERE is_ghost = true").One(&ghostCount)

		var bookCount countResult
		_ = app.DB().NewQuery(`
			SELECT COUNT(DISTINCT book) as count FROM user_books ub
			JOIN users u ON ub.user = u.id
			WHERE u.is_ghost = true
		`).One(&bookCount)

		return e.JSON(http.StatusOK, map[string]any{
			"ghost_users":  ghostCount.Count,
			"books_rated":  bookCount.Count,
			"total_ghosts": len(ghostPersonas),
		})
	}
}
