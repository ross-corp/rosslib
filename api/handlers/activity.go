package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

// validActivityTypes is the set of allowed activity_type values for filtering.
var validActivityTypes = map[string]bool{
	"shelved":         true,
	"started_book":    true,
	"finished_book":   true,
	"rated":           true,
	"reviewed":        true,
	"created_thread":  true,
	"followed_user":   true,
	"followed_author": true,
	"created_link":    true,
}

// enrichActivity takes a raw activity row and enriches it with book, user, and
// metadata details to match the webapp's ActivityItem type.
func enrichActivity(app core.App, row activityRow) map[string]any {
	item := map[string]any{
		"id":         row.ID,
		"type":       row.ActivityType,
		"created_at": row.Created,
	}

	// Nested user object
	var avatarURL *string
	if row.Avatar != nil && *row.Avatar != "" {
		url := fmt.Sprintf("/api/files/users/%s/%s", row.UserID, *row.Avatar)
		avatarURL = &url
	}
	item["user"] = map[string]any{
		"user_id":      row.UserID,
		"username":     row.Username,
		"display_name": row.DisplayName,
		"avatar_url":   avatarURL,
	}

	// Nested book object (if activity has a book reference)
	if row.BookID != nil && *row.BookID != "" {
		bookObj := map[string]any{
			"open_library_id": "",
			"title":           "",
			"cover_url":       nil,
		}
		if row.BookOLID != nil {
			bookObj["open_library_id"] = *row.BookOLID
		}
		if row.BookTitle != nil {
			bookObj["title"] = *row.BookTitle
		}
		if row.BookCoverURL != nil && *row.BookCoverURL != "" {
			bookObj["cover_url"] = *row.BookCoverURL
		}
		item["book"] = bookObj
	}

	// Nested target_user object (for follow activities)
	if row.TargetUserID != nil && *row.TargetUserID != "" {
		var targetAvatarURL *string
		if row.TargetAvatar != nil && *row.TargetAvatar != "" {
			url := fmt.Sprintf("/api/files/users/%s/%s", *row.TargetUserID, *row.TargetAvatar)
			targetAvatarURL = &url
		}
		item["target_user"] = map[string]any{
			"user_id":      *row.TargetUserID,
			"username":     row.TargetUsername,
			"display_name": row.TargetDisplayName,
			"avatar_url":   targetAvatarURL,
		}
	}

	// Shelf name (for shelved activities)
	if row.ShelfName != nil && *row.ShelfName != "" {
		item["shelf_name"] = *row.ShelfName
	}

	// Thread title (for created_thread activities)
	if row.ThreadTitle != nil && *row.ThreadTitle != "" {
		item["thread_title"] = *row.ThreadTitle
	}

	// Parse metadata JSON for rating, review_snippet, etc.
	if row.Metadata != nil && *row.Metadata != "" {
		var meta map[string]any
		if err := json.Unmarshal([]byte(*row.Metadata), &meta); err == nil {
			if rating, ok := meta["rating"].(float64); ok {
				item["rating"] = int(rating)
			}
			if snippet, ok := meta["review_snippet"].(string); ok {
				item["review_snippet"] = snippet
			}
			if linkType, ok := meta["link_type"].(string); ok {
				item["link_type"] = linkType
			}
			if toBookOL, ok := meta["to_book_ol_id"].(string); ok {
				item["to_book_ol_id"] = toBookOL
			}
			if toBookTitle, ok := meta["to_book_title"].(string); ok {
				item["to_book_title"] = toBookTitle
			}
			if authorKey, ok := meta["author_key"].(string); ok {
				item["author_key"] = authorKey
			}
			if authorName, ok := meta["author_name"].(string); ok {
				item["author_name"] = authorName
			}
		}
	}

	return item
}

// activityRow holds the result of the enriched activity query.
type activityRow struct {
	ID                string  `db:"id"`
	UserID            string  `db:"user_id"`
	ActivityType      string  `db:"activity_type"`
	Created           string  `db:"created"`
	Username          string  `db:"username"`
	DisplayName       *string `db:"display_name"`
	Avatar            *string `db:"avatar"`
	BookID            *string `db:"book_id"`
	BookOLID          *string `db:"book_olid"`
	BookTitle         *string `db:"book_title"`
	BookCoverURL      *string `db:"book_cover_url"`
	TargetUserID      *string `db:"target_user_id"`
	TargetUsername    *string `db:"target_username"`
	TargetDisplayName *string `db:"target_display_name"`
	TargetAvatar      *string `db:"target_avatar"`
	ShelfName         *string `db:"shelf_name"`
	ThreadTitle       *string `db:"thread_title"`
	Metadata          *string `db:"metadata"`
}

const activitySelectClause = `
	SELECT a.id, a.user as user_id, a.activity_type, a.created, a.metadata,
		   u.username, u.display_name, u.avatar,
		   a.book as book_id, b.open_library_id as book_olid, b.title as book_title, b.cover_url as book_cover_url,
		   a.target_user as target_user_id, tu.username as target_username, tu.display_name as target_display_name, tu.avatar as target_avatar,
		   c.name as shelf_name,
		   t.title as thread_title
	FROM activities a
	JOIN users u ON a.user = u.id
	LEFT JOIN books b ON a.book = b.id
	LEFT JOIN users tu ON a.target_user = tu.id
	LEFT JOIN collections c ON a.collection_ref = c.id
	LEFT JOIN threads t ON a.thread = t.id
`

// GetFeed handles GET /me/feed
func GetFeed(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		cursor := e.Request.URL.Query().Get("cursor")
		typeFilter := e.Request.URL.Query().Get("type")
		limit := 30

		// Check if user follows anyone
		type countResult struct {
			Count int `db:"count"`
		}
		var cnt countResult
		_ = app.DB().NewQuery(`
			SELECT COUNT(*) as count FROM follows
			WHERE follower = {:user} AND status = 'active'
		`).Bind(map[string]any{"user": user.Id}).One(&cnt)

		if cnt.Count == 0 {
			return e.JSON(http.StatusOK, map[string]any{
				"activities":  []any{},
				"next_cursor": nil,
			})
		}

		query := activitySelectClause + `
			WHERE a.user IN (SELECT followee FROM follows WHERE follower = {:user} AND status = 'active')
			AND a.user NOT IN (SELECT blocked FROM blocks WHERE blocker = {:user})
			AND a.user NOT IN (SELECT blocker FROM blocks WHERE blocked = {:user})
		`
		params := map[string]any{"user": user.Id}

		if typeFilter != "" {
			types := strings.Split(typeFilter, ",")
			var filtered []string
			for _, t := range types {
				t = strings.TrimSpace(t)
				if validActivityTypes[t] {
					filtered = append(filtered, t)
				}
			}
			if len(filtered) > 0 {
				placeholders := make([]string, len(filtered))
				for i, t := range filtered {
					key := fmt.Sprintf("type%d", i)
					placeholders[i] = "{:" + key + "}"
					params[key] = t
				}
				query += " AND a.activity_type IN (" + strings.Join(placeholders, ", ") + ")"
			}
		}

		if cursor != "" {
			query += " AND a.created < {:cursor}"
			params["cursor"] = cursor
		}
		query += " ORDER BY a.created DESC LIMIT {:limit}"
		params["limit"] = limit

		var rows []activityRow
		err := app.DB().NewQuery(query).Bind(params).All(&rows)
		if err != nil {
			return e.JSON(http.StatusOK, map[string]any{
				"activities":  []any{},
				"next_cursor": nil,
			})
		}

		result := make([]map[string]any, 0, len(rows))
		for _, row := range rows {
			result = append(result, enrichActivity(app, row))
		}

		var nextCursor any
		if len(rows) == limit {
			nextCursor = rows[len(rows)-1].Created
		}

		return e.JSON(http.StatusOK, map[string]any{
			"activities":  result,
			"next_cursor": nextCursor,
		})
	}
}

// GetUserActivity handles GET /users/{username}/activity
func GetUserActivity(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		username := e.Request.PathValue("username")

		users, err := app.FindRecordsByFilter("users",
			"username = {:username}", "", 1, 0,
			map[string]any{"username": username},
		)
		if err != nil || len(users) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "User not found"})
		}
		targetUser := users[0]

		viewerID := ""
		if e.Auth != nil {
			viewerID = e.Auth.Id
		}
		if !canViewProfile(app, viewerID, targetUser) {
			return e.JSON(http.StatusForbidden, map[string]any{"error": "Profile is private"})
		}

		query := activitySelectClause + `
			WHERE a.user = {:user}
			ORDER BY a.created DESC
			LIMIT 30
		`

		var rows []activityRow
		err = app.DB().NewQuery(query).Bind(map[string]any{"user": targetUser.Id}).All(&rows)
		if err != nil {
			return e.JSON(http.StatusOK, map[string]any{"activities": []any{}})
		}

		result := make([]map[string]any, 0, len(rows))
		for _, row := range rows {
			result = append(result, enrichActivity(app, row))
		}

		return e.JSON(http.StatusOK, map[string]any{"activities": result})
	}
}
