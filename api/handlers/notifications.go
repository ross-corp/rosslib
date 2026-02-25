package handlers

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

// GetNotifications handles GET /me/notifications
func GetNotifications(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		records, err := app.FindRecordsByFilter("notifications",
			"user = {:user}", "-created", 50, 0,
			map[string]any{"user": user.Id},
		)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}

		var result []map[string]any
		for _, r := range records {
			result = append(result, map[string]any{
				"id":         r.Id,
				"notif_type": r.GetString("notif_type"),
				"title":      r.GetString("title"),
				"body":       r.GetString("body"),
				"metadata":   r.Get("metadata"),
				"read":       r.GetBool("read"),
				"created_at": r.GetString("created"),
			})
		}
		if result == nil {
			result = []map[string]any{}
		}

		return e.JSON(http.StatusOK, result)
	}
}

// GetUnreadCount handles GET /me/notifications/unread-count
func GetUnreadCount(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		type countResult struct {
			Count int `db:"count"`
		}
		var cnt countResult
		_ = app.DB().NewQuery("SELECT COUNT(*) as count FROM notifications WHERE user = {:user} AND read = false").
			Bind(map[string]any{"user": user.Id}).One(&cnt)

		return e.JSON(http.StatusOK, map[string]any{"count": cnt.Count})
	}
}

// MarkNotificationRead handles POST /me/notifications/{notifId}/read
func MarkNotificationRead(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		notifID := e.Request.PathValue("notifId")

		rec, err := app.FindRecordById("notifications", notifID)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Notification not found"})
		}
		if rec.GetString("user") != user.Id {
			return e.JSON(http.StatusForbidden, map[string]any{"error": "Not your notification"})
		}

		rec.Set("read", true)
		if err := app.Save(rec); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to mark read"})
		}

		return e.JSON(http.StatusOK, map[string]any{"message": "Marked as read"})
	}
}

// MarkAllRead handles POST /me/notifications/read-all
func MarkAllRead(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		_, err := app.DB().NewQuery("UPDATE notifications SET read = true WHERE user = {:user} AND read = false").
			Bind(map[string]any{"user": user.Id}).Execute()
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to mark all read"})
		}

		return e.JSON(http.StatusOK, map[string]any{"message": "All marked as read"})
	}
}
