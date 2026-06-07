package handlers

import (
	"net/http"
	"strconv"

	"github.com/pocketbase/pocketbase/core"
)

// GetNotifications handles GET /me/notifications?page=<n>&perPage=<n>
func GetNotifications(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		page, _ := strconv.Atoi(e.Request.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		perPage, _ := strconv.Atoi(e.Request.URL.Query().Get("perPage"))
		if perPage < 1 || perPage > 100 {
			perPage = 20
		}
		offset := (page - 1) * perPage

		// Get total count
		type countResult struct {
			Count int `db:"count"`
		}
		var total countResult
		_ = app.DB().NewQuery("SELECT COUNT(*) as count FROM notifications WHERE user = {:user}").
			Bind(map[string]any{"user": user.Id}).One(&total)

		records, err := app.FindRecordsByFilter("notifications",
			"user = {:user}", "-created", perPage, offset,
			map[string]any{"user": user.Id},
		)
		if err != nil {
			return e.JSON(http.StatusOK, map[string]any{
				"notifications": []any{},
				"total":         0,
				"page":          page,
				"perPage":       perPage,
			})
		}

		result := make([]map[string]any, 0, len(records))
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

		return e.JSON(http.StatusOK, map[string]any{
			"notifications": result,
			"total":         total.Count,
			"page":          page,
			"perPage":       perPage,
		})
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

// DeleteNotification handles DELETE /me/notifications/{notifId}
func DeleteNotification(app core.App) func(e *core.RequestEvent) error {
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

		if err := app.Delete(rec); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to delete notification"})
		}

		return e.JSON(http.StatusOK, map[string]any{"ok": true})
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
