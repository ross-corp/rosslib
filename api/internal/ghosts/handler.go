package ghosts

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tristansaldanha/rosslib/api/internal/activity"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// ── Seed — POST /admin/ghosts/seed ──────────────────────────────────────────

func (h *Handler) Seed(c *gin.Context) {
	ctx := c.Request.Context()
	created := []string{}

	ghostIDs := map[string]string{} // username → user id

	for _, p := range personas {
		hash, err := bcrypt.GenerateFromPassword([]byte("ghost-no-login"), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		var userID string
		err = h.pool.QueryRow(ctx,
			`INSERT INTO users (username, email, password_hash, display_name, is_ghost)
			 VALUES ($1, $2, $3, $4, true)
			 ON CONFLICT (username) DO UPDATE SET username = users.username
			 RETURNING id`,
			p.Username, p.Email, string(hash), p.DisplayName,
		).Scan(&userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create %s: %v", p.Username, err)})
			return
		}
		ghostIDs[p.Username] = userID

		// Create default shelves (idempotent)
		for _, s := range []struct{ name, slug string }{
			{"Want to Read", "want-to-read"},
			{"Currently Reading", "currently-reading"},
			{"Read", "read"},
		} {
			_, err = h.pool.Exec(ctx,
				`INSERT INTO collections (user_id, name, slug, is_exclusive, exclusive_group, collection_type)
				 VALUES ($1, $2, $3, true, 'read_status', 'shelf')
				 ON CONFLICT (user_id, slug) DO NOTHING`,
				userID, s.name, s.slug,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create shelves for %s: %v", p.Username, err)})
				return
			}
		}

		// Upsert seed books
		for _, b := range p.Books {
			_, err = h.pool.Exec(ctx,
				`INSERT INTO books (open_library_id, title, cover_url, authors)
				 VALUES ($1, $2, $3, $4)
				 ON CONFLICT (open_library_id) DO UPDATE
				   SET title = EXCLUDED.title, cover_url = EXCLUDED.cover_url, authors = EXCLUDED.authors`,
				b.OLID, b.Title, b.CoverURL, b.Authors,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to upsert book %s: %v", b.OLID, err)})
				return
			}
		}

		created = append(created, p.Username)
	}

	// Insert mutual follows between all ghosts (12 directed edges for 4 users)
	for _, a := range personas {
		for _, b := range personas {
			if a.Username == b.Username {
				continue
			}
			_, _ = h.pool.Exec(ctx,
				`INSERT INTO follows (follower_id, followee_id, status)
				 VALUES ($1, $2, 'active')
				 ON CONFLICT DO NOTHING`,
				ghostIDs[a.Username], ghostIDs[b.Username],
			)
		}
	}

	c.JSON(http.StatusOK, gin.H{"created": created})
}

// ── Simulate — POST /admin/ghosts/simulate ──────────────────────────────────

type actionSummary struct {
	Ghost   string   `json:"ghost"`
	Actions []string `json:"actions"`
}

func (h *Handler) Simulate(c *gin.Context) {
	ctx := c.Request.Context()
	results := []actionSummary{}

	for _, p := range personas {
		var userID string
		err := h.pool.QueryRow(ctx,
			`SELECT id FROM users WHERE username = $1 AND is_ghost = true AND deleted_at IS NULL`,
			p.Username,
		).Scan(&userID)
		if err != nil {
			continue // ghost not seeded yet
		}

		actions := simulateGhost(ctx, h.pool, userID, p)
		results = append(results, actionSummary{Ghost: p.Username, Actions: actions})
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}

func simulateGhost(ctx context.Context, pool *pgxpool.Pool, userID string, p persona) []string {
	actions := []string{}

	// Load shelf IDs
	shelves := map[string]string{} // slug → id
	rows, err := pool.Query(ctx,
		`SELECT slug, id FROM collections WHERE user_id = $1 AND exclusive_group = 'read_status'`,
		userID,
	)
	if err != nil {
		return actions
	}
	defer rows.Close()
	for rows.Next() {
		var slug, id string
		if err := rows.Scan(&slug, &id); err != nil {
			return actions
		}
		shelves[slug] = id
	}
	rows.Close()

	if len(shelves) < 3 {
		return actions
	}

	// Load current book states: which seed books are on which shelves
	type bookState struct {
		bookID string
		shelf  string
	}
	shelved := map[string]bookState{} // OLID → state
	stateRows, err := pool.Query(ctx,
		`SELECT b.open_library_id, b.id, c.slug
		 FROM collection_items ci
		 JOIN books b ON b.id = ci.book_id
		 JOIN collections c ON c.id = ci.collection_id
		 WHERE c.user_id = $1 AND c.exclusive_group = 'read_status'`,
		userID,
	)
	if err != nil {
		return actions
	}
	defer stateRows.Close()
	for stateRows.Next() {
		var olid, bookID, slug string
		if err := stateRows.Scan(&olid, &bookID, &slug); err != nil {
			break
		}
		shelved[olid] = bookState{bookID: bookID, shelf: slug}
	}
	stateRows.Close()

	// Categorize seed books
	var unshelved, wantToRead, currentlyReading []bookSeed
	for _, b := range p.Books {
		state, exists := shelved[b.OLID]
		if !exists {
			unshelved = append(unshelved, b)
		} else {
			switch state.shelf {
			case "want-to-read":
				wantToRead = append(wantToRead, b)
			case "currently-reading":
				currentlyReading = append(currentlyReading, b)
			}
		}
	}

	// Pick 1-3 random actions
	numActions := 1 + rand.IntN(3)
	for range numActions {
		roll := rand.Float64()

		switch {
		case roll < 0.30 && len(unshelved) > 0:
			// Shelve new book → want-to-read
			idx := rand.IntN(len(unshelved))
			book := unshelved[idx]
			if a := addToShelf(ctx, pool, userID, book, shelves["want-to-read"], "want-to-read"); a != "" {
				actions = append(actions, a)
				unshelved = append(unshelved[:idx], unshelved[idx+1:]...)
				wantToRead = append(wantToRead, book)
			}

		case roll < 0.55 && len(wantToRead) > 0:
			// Start reading: want-to-read → currently-reading
			idx := rand.IntN(len(wantToRead))
			book := wantToRead[idx]
			if a := moveToShelf(ctx, pool, userID, book, shelves, "currently-reading"); a != "" {
				actions = append(actions, a)
				wantToRead = append(wantToRead[:idx], wantToRead[idx+1:]...)
				currentlyReading = append(currentlyReading, book)
			}

		case roll < 0.80 && len(currentlyReading) > 0:
			// Finish book: currently-reading → read, with rating + maybe review
			idx := rand.IntN(len(currentlyReading))
			book := currentlyReading[idx]
			if a := finishBook(ctx, pool, userID, book, shelves, p); a != "" {
				actions = append(actions, a)
				currentlyReading = append(currentlyReading[:idx], currentlyReading[idx+1:]...)
			}

		default:
			// Follow a random non-ghost user
			if a := followRandomUser(ctx, pool, userID); a != "" {
				actions = append(actions, a)
			}
		}
	}

	return actions
}

func addToShelf(ctx context.Context, pool *pgxpool.Pool, userID string, book bookSeed, shelfID, shelfSlug string) string {
	var bookID string
	err := pool.QueryRow(ctx,
		`SELECT id FROM books WHERE open_library_id = $1`, book.OLID,
	).Scan(&bookID)
	if err != nil {
		return ""
	}

	_, err = pool.Exec(ctx,
		`INSERT INTO collection_items (collection_id, book_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		shelfID, bookID,
	)
	if err != nil {
		return ""
	}

	activity.Record(ctx, pool, userID, "shelved", &bookID, nil, &shelfID, nil,
		map[string]string{"shelf_name": "Want to Read"})

	return fmt.Sprintf("added '%s' to want-to-read", book.Title)
}

func moveToShelf(ctx context.Context, pool *pgxpool.Pool, userID string, book bookSeed, shelves map[string]string, targetSlug string) string {
	var bookID string
	err := pool.QueryRow(ctx,
		`SELECT id FROM books WHERE open_library_id = $1`, book.OLID,
	).Scan(&bookID)
	if err != nil {
		return ""
	}

	targetID := shelves[targetSlug]

	// Remove from other read_status shelves
	_, _ = pool.Exec(ctx,
		`DELETE FROM collection_items ci
		 USING collections col
		 WHERE ci.collection_id = col.id
		   AND col.user_id = $1
		   AND col.exclusive_group = 'read_status'
		   AND ci.book_id = $2
		   AND ci.collection_id != $3`,
		userID, bookID, targetID,
	)

	_, err = pool.Exec(ctx,
		`INSERT INTO collection_items (collection_id, book_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		targetID, bookID,
	)
	if err != nil {
		return ""
	}

	activityType := "shelved"
	shelfName := targetSlug
	if targetSlug == "currently-reading" {
		activityType = "started_book"
		shelfName = "Currently Reading"
	}
	activity.Record(ctx, pool, userID, activityType, &bookID, nil, &targetID, nil,
		map[string]string{"shelf_name": shelfName})

	return fmt.Sprintf("started reading '%s'", book.Title)
}

func finishBook(ctx context.Context, pool *pgxpool.Pool, userID string, book bookSeed, shelves map[string]string, p persona) string {
	var bookID string
	err := pool.QueryRow(ctx,
		`SELECT id FROM books WHERE open_library_id = $1`, book.OLID,
	).Scan(&bookID)
	if err != nil {
		return ""
	}

	readID := shelves["read"]

	// Remove from other read_status shelves
	_, _ = pool.Exec(ctx,
		`DELETE FROM collection_items ci
		 USING collections col
		 WHERE ci.collection_id = col.id
		   AND col.user_id = $1
		   AND col.exclusive_group = 'read_status'
		   AND ci.book_id = $2
		   AND ci.collection_id != $3`,
		userID, bookID, readID,
	)

	_, err = pool.Exec(ctx,
		`INSERT INTO collection_items (collection_id, book_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		readID, bookID,
	)
	if err != nil {
		return ""
	}

	// Set rating
	rating := p.MinRating + rand.IntN(p.MaxRating-p.MinRating+1)
	_, _ = pool.Exec(ctx,
		`UPDATE collection_items SET rating = $1 WHERE collection_id = $2 AND book_id = $3`,
		rating, readID, bookID,
	)

	activity.Record(ctx, pool, userID, "finished_book", &bookID, nil, &readID, nil,
		map[string]string{"shelf_name": "Read"})
	activity.Record(ctx, pool, userID, "rated", &bookID, nil, &readID, nil,
		map[string]string{"rating": strconv.Itoa(rating)})

	result := fmt.Sprintf("finished '%s' (%d stars)", book.Title, rating)

	// Maybe add review
	if rand.Float64() < p.ReviewRate && len(p.Reviews) > 0 {
		review := p.Reviews[rand.IntN(len(p.Reviews))]
		_, _ = pool.Exec(ctx,
			`UPDATE collection_items SET review_text = $1 WHERE collection_id = $2 AND book_id = $3`,
			review, readID, bookID,
		)
		snippet := review
		if len(snippet) > 100 {
			snippet = snippet[:100] + "..."
		}
		activity.Record(ctx, pool, userID, "reviewed", &bookID, nil, &readID, nil,
			map[string]string{"review_snippet": snippet})
		result += " + review"
	}

	return result
}

func followRandomUser(ctx context.Context, pool *pgxpool.Pool, ghostUserID string) string {
	var targetID, targetUsername string
	err := pool.QueryRow(ctx,
		`SELECT id, username FROM users
		 WHERE id != $1
		   AND is_ghost = false
		   AND deleted_at IS NULL
		   AND id NOT IN (SELECT followee_id FROM follows WHERE follower_id = $1)
		 ORDER BY RANDOM()
		 LIMIT 1`,
		ghostUserID,
	).Scan(&targetID, &targetUsername)
	if err != nil {
		return ""
	}

	// Respect private accounts
	var isPrivate bool
	_ = pool.QueryRow(ctx, `SELECT is_private FROM users WHERE id = $1`, targetID).Scan(&isPrivate)

	status := "active"
	if isPrivate {
		status = "pending"
	}

	_, err = pool.Exec(ctx,
		`INSERT INTO follows (follower_id, followee_id, status) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
		ghostUserID, targetID, status,
	)
	if err != nil {
		return ""
	}

	if status == "active" {
		activity.Record(ctx, pool, ghostUserID, "followed_user", nil, &targetID, nil, nil, nil)
	}

	return fmt.Sprintf("followed @%s", targetUsername)
}

// ── Status — GET /admin/ghosts/status ────────────────────────────────────────

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

func (h *Handler) Status(c *gin.Context) {
	ctx := c.Request.Context()
	statuses := []ghostStatus{}

	for _, p := range personas {
		var s ghostStatus
		err := h.pool.QueryRow(ctx,
			`SELECT id, username, COALESCE(display_name, '') FROM users WHERE username = $1 AND is_ghost = true AND deleted_at IS NULL`,
			p.Username,
		).Scan(&s.UserID, &s.Username, &s.DisplayName)
		if err != nil {
			continue
		}

		// Count books per shelf
		shelfRows, err := h.pool.Query(ctx,
			`SELECT c.slug, COUNT(ci.id)
			 FROM collections c
			 LEFT JOIN collection_items ci ON ci.collection_id = c.id
			 WHERE c.user_id = $1 AND c.exclusive_group = 'read_status'
			 GROUP BY c.slug`,
			s.UserID,
		)
		if err == nil {
			for shelfRows.Next() {
				var slug string
				var count int
				if err := shelfRows.Scan(&slug, &count); err != nil {
					break
				}
				switch slug {
				case "read":
					s.BooksRead = count
				case "currently-reading":
					s.CurrentlyReading = count
				case "want-to-read":
					s.WantToRead = count
				}
			}
			shelfRows.Close()
		}

		// Follow counts
		_ = h.pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM follows WHERE follower_id = $1 AND status = 'active'`, s.UserID,
		).Scan(&s.FollowingCount)
		_ = h.pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM follows WHERE followee_id = $1 AND status = 'active'`, s.UserID,
		).Scan(&s.FollowersCount)

		statuses = append(statuses, s)
	}

	c.JSON(http.StatusOK, statuses)
}
