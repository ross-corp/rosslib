// Package notifications provides a background poller that detects new publications
// by followed authors via the Open Library API, and creates per-user notifications.
package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PollInterval is how often the poller checks for new works.
const PollInterval = 6 * time.Hour

type olWorksResponse struct {
	Size    int `json:"size"`
	Entries []struct {
		Key    string `json:"key"`
		Title  string `json:"title"`
		Covers []int  `json:"covers"`
	} `json:"entries"`
}

// StartPoller launches a background goroutine that periodically checks
// Open Library for new works by authors that any user follows.
// It respects the provided context for graceful shutdown.
func StartPoller(ctx context.Context, pool *pgxpool.Pool, olClient *http.Client) {
	// Run once at startup (after a short delay to let the server warm up).
	time.AfterFunc(30*time.Second, func() {
		pollOnce(ctx, pool, olClient)
	})

	go func() {
		ticker := time.NewTicker(PollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				pollOnce(ctx, pool, olClient)
			}
		}
	}()
}

// pollOnce runs a single pass: find all distinct followed authors, check
// their work counts against our snapshot, and create notifications for any
// that have increased.
func pollOnce(ctx context.Context, pool *pgxpool.Pool, olClient *http.Client) {
	log.Println("notifications: polling for new author works")

	// Get distinct author keys that at least one user follows.
	rows, err := pool.Query(ctx,
		`SELECT DISTINCT author_key, author_name FROM author_follows`)
	if err != nil {
		log.Printf("notifications: failed to query followed authors: %v", err)
		return
	}
	defer rows.Close()

	type authorInfo struct {
		key  string
		name string
	}
	var authors []authorInfo
	for rows.Next() {
		var a authorInfo
		if err := rows.Scan(&a.key, &a.name); err != nil {
			log.Printf("notifications: scan error: %v", err)
			return
		}
		authors = append(authors, a)
	}

	if len(authors) == 0 {
		log.Println("notifications: no followed authors, skipping")
		return
	}

	for _, author := range authors {
		select {
		case <-ctx.Done():
			return
		default:
		}
		checkAuthor(ctx, pool, olClient, author.key, author.name)
	}

	log.Printf("notifications: poll complete, checked %d authors", len(authors))
}

// checkAuthor fetches the current work count from OL for a single author
// and compares it against our stored snapshot.
func checkAuthor(ctx context.Context, pool *pgxpool.Pool, olClient *http.Client, authorKey, authorName string) {
	worksURL := fmt.Sprintf("https://openlibrary.org/authors/%s/works.json?limit=5", authorKey)

	req, err := http.NewRequestWithContext(ctx, "GET", worksURL, nil)
	if err != nil {
		return
	}
	resp, err := olClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var works olWorksResponse
	if err := json.Unmarshal(body, &works); err != nil {
		return
	}

	currentCount := works.Size

	// Get our stored snapshot.
	var prevCount int
	var exists bool
	err = pool.QueryRow(ctx,
		`SELECT work_count FROM author_works_snapshot WHERE author_key = $1`,
		authorKey,
	).Scan(&prevCount)
	if err != nil {
		// No snapshot yet — store current count and move on (no notification
		// on first-time snapshot creation to avoid flooding users).
		exists = false
	} else {
		exists = true
	}

	// Update (or insert) the snapshot.
	if exists {
		_, _ = pool.Exec(ctx,
			`UPDATE author_works_snapshot SET work_count = $1, checked_at = NOW() WHERE author_key = $2`,
			currentCount, authorKey,
		)
	} else {
		_, _ = pool.Exec(ctx,
			`INSERT INTO author_works_snapshot (author_key, work_count)
			 VALUES ($1, $2)
			 ON CONFLICT (author_key) DO UPDATE SET work_count = EXCLUDED.work_count, checked_at = NOW()`,
			authorKey, currentCount,
		)
	}

	// If the work count has increased, create notifications for all followers.
	if exists && currentCount > prevCount {
		newCount := currentCount - prevCount

		// Try to get titles of the newest works.
		var newTitles []string
		for i, entry := range works.Entries {
			if i >= newCount {
				break
			}
			newTitles = append(newTitles, entry.Title)
		}

		title := fmt.Sprintf("New work by %s", authorName)
		var bodyText string
		if len(newTitles) == 1 {
			bodyText = fmt.Sprintf("%s published a new work: %s", authorName, newTitles[0])
		} else if len(newTitles) > 1 {
			bodyText = fmt.Sprintf("%s published %d new works: %s", authorName, newCount, strings.Join(newTitles, ", "))
		} else {
			bodyText = fmt.Sprintf("%s published %d new work(s)", authorName, newCount)
		}

		metadata := map[string]string{
			"author_key":  authorKey,
			"author_name": authorName,
			"new_count":   fmt.Sprintf("%d", newCount),
		}
		if len(newTitles) > 0 {
			metadata["new_titles"] = strings.Join(newTitles, "; ")
		}

		// Fan out to all followers of this author.
		followerRows, err := pool.Query(ctx,
			`SELECT user_id FROM author_follows WHERE author_key = $1`,
			authorKey,
		)
		if err != nil {
			log.Printf("notifications: failed to get followers for %s: %v", authorKey, err)
			return
		}
		defer followerRows.Close()

		for followerRows.Next() {
			var userID string
			if err := followerRows.Scan(&userID); err != nil {
				continue
			}
			_, _ = pool.Exec(ctx,
				`INSERT INTO notifications (user_id, notif_type, title, body, metadata)
				 VALUES ($1, $2, $3, $4, $5)`,
				userID, "new_publication", title, bodyText, metadata,
			)
		}

		log.Printf("notifications: %s (%s) has %d new work(s)", authorName, authorKey, newCount)
	}
}
