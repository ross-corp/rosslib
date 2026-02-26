package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/pocketbase/pocketbase/core"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/unicode/norm"
)

var nonAlphanumRegex = regexp.MustCompile(`[^a-z0-9-]+`)
var multiDashRegex = regexp.MustCompile(`-{2,}`)

// slugify converts a string to a URL-safe slug.
func slugify(s string) string {
	// Normalize unicode
	s = norm.NFKD.String(s)
	// Remove non-ASCII
	result := make([]rune, 0, len(s))
	for _, r := range s {
		if r < unicode.MaxASCII {
			result = append(result, r)
		}
	}
	s = string(result)
	s = strings.ToLower(s)
	s = nonAlphanumRegex.ReplaceAllString(s, "-")
	s = multiDashRegex.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// tagSlugify converts a string to a tag-safe slug, preserving "/" as separator.
func tagSlugify(s string) string {
	parts := strings.Split(s, "/")
	for i, p := range parts {
		parts[i] = slugify(p)
	}
	return strings.Join(parts, "/")
}

// canViewProfile checks whether viewer can see target user's profile.
func canViewProfile(app core.App, viewerID string, targetUser *core.Record) bool {
	if !targetUser.GetBool("is_private") {
		return true
	}
	if viewerID == "" {
		return false
	}
	if viewerID == targetUser.Id {
		return true
	}
	// Check if viewer follows target with active status
	follows, err := app.FindRecordsByFilter("follows",
		"follower = {:viewer} && followee = {:target} && status = 'active'",
		"", 1, 0,
		map[string]any{"viewer": viewerID, "target": targetUser.Id},
	)
	return err == nil && len(follows) > 0
}

// defaultStatusValues are the status tag values created for every new user.
var defaultStatusValues = []struct {
	Name string
	Slug string
}{
	{"Want to Read", "want-to-read"},
	{"Currently Reading", "currently-reading"},
	{"Finished", "finished"},
	{"Did Not Finish", "dnf"},
	{"Owned", "owned"},
}

// ensureStatusTagKey creates the Status tag key and values for a user if they don't exist.
// Returns the tag key record and all value records.
func ensureStatusTagKey(app core.App, userID string) (*core.Record, []*core.Record, error) {
	// Check if Status key already exists
	existing, err := app.FindRecordsByFilter("tag_keys",
		"user = {:user} && slug = 'status'",
		"", 1, 0,
		map[string]any{"user": userID},
	)
	if err == nil && len(existing) > 0 {
		key := existing[0]
		values, _ := app.FindRecordsByFilter("tag_values",
			"tag_key = {:key}",
			"", 100, 0,
			map[string]any{"key": key.Id},
		)
		return key, values, nil
	}

	// Create the Status tag key
	tagKeysColl, err := app.FindCollectionByNameOrId("tag_keys")
	if err != nil {
		return nil, nil, err
	}
	key := core.NewRecord(tagKeysColl)
	key.Set("user", userID)
	key.Set("name", "Status")
	key.Set("slug", "status")
	key.Set("mode", "select_one")
	if err := app.Save(key); err != nil {
		return nil, nil, err
	}

	// Create default values
	tagValuesColl, err := app.FindCollectionByNameOrId("tag_values")
	if err != nil {
		return nil, nil, err
	}
	var values []*core.Record
	for _, sv := range defaultStatusValues {
		v := core.NewRecord(tagValuesColl)
		v.Set("tag_key", key.Id)
		v.Set("name", sv.Name)
		v.Set("slug", sv.Slug)
		if err := app.Save(v); err != nil {
			return nil, nil, err
		}
		values = append(values, v)
	}

	return key, values, nil
}

// titleCase converts a slug-like string to title case ("dark-fantasy" → "Dark Fantasy").
func titleCase(s string) string {
	s = strings.ReplaceAll(s, "-", " ")
	return cases.Title(language.English).String(s)
}

// ensureTagKey finds or creates a tag key by slug for a user, and ensures it has at least one value.
// Returns (tagKeyRecord, tagValueRecord, error).
func ensureTagKey(app core.App, userID, name, mode string) (*core.Record, *core.Record, error) {
	slug := tagSlugify(name)
	displayName := titleCase(name)

	// Look for existing tag key
	existing, err := app.FindRecordsByFilter("tag_keys",
		"user = {:user} && slug = {:slug}",
		"", 1, 0,
		map[string]any{"user": userID, "slug": slug},
	)
	if err == nil && len(existing) > 0 {
		key := existing[0]
		// Find existing value
		values, _ := app.FindRecordsByFilter("tag_values",
			"tag_key = {:key}",
			"", 1, 0,
			map[string]any{"key": key.Id},
		)
		if len(values) > 0 {
			return key, values[0], nil
		}
		// Create a value
		valColl, err := app.FindCollectionByNameOrId("tag_values")
		if err != nil {
			return nil, nil, err
		}
		val := core.NewRecord(valColl)
		val.Set("tag_key", key.Id)
		val.Set("name", displayName)
		val.Set("slug", slug)
		if err := app.Save(val); err != nil {
			return nil, nil, err
		}
		return key, val, nil
	}

	// Create new tag key
	keyColl, err := app.FindCollectionByNameOrId("tag_keys")
	if err != nil {
		return nil, nil, err
	}
	key := core.NewRecord(keyColl)
	key.Set("user", userID)
	key.Set("name", displayName)
	key.Set("slug", slug)
	key.Set("mode", mode)
	if err := app.Save(key); err != nil {
		return nil, nil, err
	}

	// Create a value under it
	valColl, err := app.FindCollectionByNameOrId("tag_values")
	if err != nil {
		return nil, nil, err
	}
	val := core.NewRecord(valColl)
	val.Set("tag_key", key.Id)
	val.Set("name", displayName)
	val.Set("slug", slug)
	if err := app.Save(val); err != nil {
		return nil, nil, err
	}

	return key, val, nil
}

// olClient is a simple Open Library API client.
type olClient struct {
	httpClient *http.Client
	baseURL    string
}

func newOLClient() *olClient {
	return &olClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    "https://openlibrary.org",
	}
}

func (c *olClient) get(path string) (map[string]any, error) {
	resp, err := c.httpClient.Get(c.baseURL + path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OL API returned %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *olClient) getRaw(path string) ([]byte, error) {
	resp, err := c.httpClient.Get(c.baseURL + path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OL API returned %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// upsertBook finds or creates a book record by open_library_id.
func upsertBook(app core.App, olID, title, coverURL, isbn13, authors string, pubYear int) (*core.Record, error) {
	existing, err := app.FindRecordsByFilter("books",
		"open_library_id = {:id}",
		"", 1, 0,
		map[string]any{"id": olID},
	)
	if err == nil && len(existing) > 0 {
		rec := existing[0]
		// Update fields if they were empty
		changed := false
		if rec.GetString("title") == "" && title != "" {
			rec.Set("title", title)
			changed = true
		}
		if rec.GetString("cover_url") == "" && coverURL != "" {
			rec.Set("cover_url", coverURL)
			changed = true
		}
		if rec.GetString("isbn13") == "" && isbn13 != "" {
			rec.Set("isbn13", isbn13)
			changed = true
		}
		if rec.GetString("authors") == "" && authors != "" {
			rec.Set("authors", authors)
			changed = true
		}
		if rec.GetInt("publication_year") == 0 && pubYear != 0 {
			rec.Set("publication_year", pubYear)
			changed = true
		}
		if changed {
			if err := app.Save(rec); err != nil {
				return nil, err
			}
		}
		return rec, nil
	}

	booksColl, err := app.FindCollectionByNameOrId("books")
	if err != nil {
		return nil, err
	}
	rec := core.NewRecord(booksColl)
	rec.Set("open_library_id", olID)
	rec.Set("title", title)
	rec.Set("cover_url", coverURL)
	rec.Set("isbn13", isbn13)
	rec.Set("authors", authors)
	rec.Set("publication_year", pubYear)
	if err := app.Save(rec); err != nil {
		return nil, err
	}
	return rec, nil
}

// recordActivity creates an activity record in a fire-and-forget goroutine.
func recordActivity(app core.App, userID, activityType string, opts map[string]any) {
	go func() {
		coll, err := app.FindCollectionByNameOrId("activities")
		if err != nil {
			return
		}
		rec := core.NewRecord(coll)
		rec.Set("user", userID)
		rec.Set("activity_type", activityType)
		rec.Set("created", time.Now().UTC().Format("2006-01-02 15:04:05.000Z"))
		if v, ok := opts["book"]; ok {
			rec.Set("book", v)
		}
		if v, ok := opts["target_user"]; ok {
			rec.Set("target_user", v)
		}
		if v, ok := opts["collection_ref"]; ok {
			rec.Set("collection_ref", v)
		}
		if v, ok := opts["thread"]; ok {
			rec.Set("thread", v)
		}
		if v, ok := opts["metadata"]; ok {
			rec.Set("metadata", v)
		}
		_ = app.Save(rec)
	}()
}

// refreshBookStats recalculates the book_stats for a given book.
func refreshBookStats(app core.App, bookID string) {
	go func() {
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
			return
		}

		// Count reads and want_to_read from book_tag_values + tag_values
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

		// Upsert book_stats
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
				return
			}
			rec = core.NewRecord(coll)
			rec.Set("book", bookID)
		}
		rec.Set("reads_count", readsCount.Count)
		rec.Set("want_to_read_count", wtrCount.Count)
		rec.Set("rating_sum", stats.RatingSum)
		rec.Set("rating_count", stats.RatingCount)
		rec.Set("review_count", stats.ReviewCount)
		_ = app.Save(rec)
	}()
}
