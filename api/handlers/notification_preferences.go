package handlers

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

// notifPrefFields lists all notification preference fields and their defaults (all true).
var notifPrefFields = []string{
	"new_publication",
	"book_new_thread",
	"book_new_link",
	"book_new_review",
	"review_liked",
	"thread_mention",
	"book_recommendation",
}

// GetNotificationPreferences handles GET /me/notification-preferences
func GetNotificationPreferences(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		existing, err := app.FindRecordsByFilter("notification_preferences",
			"user = {:user}", "", 1, 0,
			map[string]any{"user": user.Id},
		)

		result := map[string]any{}
		if err == nil && len(existing) > 0 {
			rec := existing[0]
			for _, f := range notifPrefFields {
				result[f] = rec.GetBool(f)
			}
		} else {
			// No row exists — return all defaults as true
			for _, f := range notifPrefFields {
				result[f] = true
			}
		}

		return e.JSON(http.StatusOK, result)
	}
}

// UpdateNotificationPreferences handles PUT /me/notification-preferences
func UpdateNotificationPreferences(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		var body map[string]any
		if err := e.BindBody(&body); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}

		// Find or create the preferences record
		existing, err := app.FindRecordsByFilter("notification_preferences",
			"user = {:user}", "", 1, 0,
			map[string]any{"user": user.Id},
		)

		var rec *core.Record
		if err == nil && len(existing) > 0 {
			rec = existing[0]
		} else {
			coll, err := app.FindCollectionByNameOrId("notification_preferences")
			if err != nil {
				return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to find collection"})
			}
			rec = core.NewRecord(coll)
			rec.Set("user", user.Id)
			// Initialize all fields to true (default)
			for _, f := range notifPrefFields {
				rec.Set(f, true)
			}
		}

		// Update only the fields that were provided in the body
		for _, f := range notifPrefFields {
			if val, ok := body[f]; ok {
				if boolVal, ok := val.(bool); ok {
					rec.Set(f, boolVal)
				}
			}
		}

		if err := app.Save(rec); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to save preferences"})
		}

		// Return the full current state
		result := map[string]any{}
		for _, f := range notifPrefFields {
			result[f] = rec.GetBool(f)
		}

		return e.JSON(http.StatusOK, result)
	}
}

// ShouldNotify checks whether a notification of the given type should be sent to the user.
// Returns true if the user hasn't set preferences (all default to enabled) or if the
// specific notification type is enabled. This is meant to be called before creating
// a notification record.
func ShouldNotify(app core.App, userID, notifType string) bool {
	existing, err := app.FindRecordsByFilter("notification_preferences",
		"user = {:user}", "", 1, 0,
		map[string]any{"user": userID},
	)
	if err != nil || len(existing) == 0 {
		// No preferences row — all types default to enabled
		return true
	}

	rec := existing[0]

	// Map notification types to preference field names
	fieldMap := map[string]string{
		"new_publication":    "new_publication",
		"book_new_thread":    "book_new_thread",
		"book_new_link":      "book_new_link",
		"book_new_review":    "book_new_review",
		"review_liked":       "review_liked",
		"thread_mention":     "thread_mention",
		"book_recommendation": "book_recommendation",
	}

	field, ok := fieldMap[notifType]
	if !ok {
		// Unknown notification type — allow by default
		return true
	}

	return rec.GetBool(field)
}
