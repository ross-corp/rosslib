package handlers

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

// GetFeed handles GET /me/feed
func GetFeed(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		cursor := e.Request.URL.Query().Get("cursor")
		limit := 30

		// Get IDs of users the current user follows
		type followRow struct {
			Followee string `db:"followee"`
		}
		var follows []followRow
		_ = app.DB().NewQuery(`
			SELECT followee FROM follows
			WHERE follower = {:user} AND status = 'active'
		`).Bind(map[string]any{"user": user.Id}).All(&follows)

		if len(follows) == 0 {
			return e.JSON(http.StatusOK, map[string]any{
				"activities":  []any{},
				"next_cursor": nil,
			})
		}

		// Build IN clause
		followeeIDs := make([]any, len(follows))
		placeholders := ""
		for i, f := range follows {
			followeeIDs[i] = f.Followee
			if i > 0 {
				placeholders += ","
			}
			placeholders += "{:f" + string(rune('0'+i)) + "}"
		}

		// Use simple filter approach instead of raw IN clause
		query := `
			SELECT a.id, a.user, a.activity_type, a.book, a.target_user,
				   a.collection_ref, a.thread, a.metadata, a.created,
				   u.username, u.display_name, u.avatar
			FROM activities a
			JOIN users u ON a.user = u.id
			WHERE a.user IN (SELECT followee FROM follows WHERE follower = {:user} AND status = 'active')
		`
		params := map[string]any{"user": user.Id}

		if cursor != "" {
			query += " AND a.created < {:cursor}"
			params["cursor"] = cursor
		}
		query += " ORDER BY a.created DESC LIMIT {:limit}"
		params["limit"] = limit

		type actRow struct {
			ID            string  `db:"id" json:"id"`
			UserID        string  `db:"user" json:"user_id"`
			ActivityType  string  `db:"activity_type" json:"activity_type"`
			Book          *string `db:"book" json:"book"`
			TargetUser    *string `db:"target_user" json:"target_user"`
			CollectionRef *string `db:"collection_ref" json:"collection_ref"`
			Thread        *string `db:"thread" json:"thread"`
			Metadata      *string `db:"metadata" json:"metadata"`
			Created       string  `db:"created" json:"created_at"`
			Username      string  `db:"username" json:"username"`
			DisplayName   *string `db:"display_name" json:"display_name"`
			Avatar        *string `db:"avatar" json:"avatar"`
		}

		var activities []actRow
		err := app.DB().NewQuery(query).Bind(params).All(&activities)
		if err != nil {
			return e.JSON(http.StatusOK, map[string]any{
				"activities":  []any{},
				"next_cursor": nil,
			})
		}

		var result []map[string]any
		for _, a := range activities {
			result = append(result, map[string]any{
				"id":             a.ID,
				"user_id":        a.UserID,
				"username":       a.Username,
				"display_name":   a.DisplayName,
				"activity_type":  a.ActivityType,
				"book":           a.Book,
				"target_user":    a.TargetUser,
				"collection_ref": a.CollectionRef,
				"thread":         a.Thread,
				"metadata":       a.Metadata,
				"created_at":     a.Created,
			})
		}
		if result == nil {
			result = []map[string]any{}
		}

		var nextCursor any
		if len(activities) == limit {
			nextCursor = activities[len(activities)-1].Created
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

		activities, err := app.FindRecordsByFilter("activities",
			"user = {:user}", "-created", 30, 0,
			map[string]any{"user": targetUser.Id},
		)
		if err != nil {
			return e.JSON(http.StatusOK, map[string]any{"activities": []any{}})
		}

		var result []map[string]any
		for _, a := range activities {
			result = append(result, map[string]any{
				"id":             a.Id,
				"activity_type":  a.GetString("activity_type"),
				"book":           a.GetString("book"),
				"target_user":    a.GetString("target_user"),
				"collection_ref": a.GetString("collection_ref"),
				"thread":         a.GetString("thread"),
				"created_at":     a.GetString("created"),
			})
		}
		if result == nil {
			result = []map[string]any{}
		}

		return e.JSON(http.StatusOK, map[string]any{"activities": result})
	}
}
