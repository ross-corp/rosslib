package handlers

import (
	"fmt"
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

// CreateFeedback handles POST /feedback
func CreateFeedback(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		data := struct {
			Type             string `json:"type"`
			Title            string `json:"title"`
			Description      string `json:"description"`
			StepsToReproduce string `json:"steps_to_reproduce"`
			Severity         string `json:"severity"`
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}
		if data.Title == "" || data.Description == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "title and description are required"})
		}
		if data.Type != "bug" && data.Type != "feature" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "type must be bug or feature"})
		}
		if data.Type == "bug" && data.Severity != "" && data.Severity != "low" && data.Severity != "medium" && data.Severity != "high" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "severity must be low, medium, or high"})
		}

		coll, err := app.FindCollectionByNameOrId("feedback")
		if err != nil {
			return err
		}
		rec := core.NewRecord(coll)
		rec.Set("user", user.Id)
		rec.Set("type", data.Type)
		rec.Set("title", data.Title)
		rec.Set("description", data.Description)
		rec.Set("status", "open")
		if data.Type == "bug" {
			rec.Set("steps_to_reproduce", data.StepsToReproduce)
			if data.Severity != "" {
				rec.Set("severity", data.Severity)
			}
		}
		if err := app.Save(rec); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to save feedback"})
		}

		return e.JSON(http.StatusCreated, map[string]any{
			"id":         rec.Id,
			"created_at": rec.GetString("created"),
		})
	}
}

// GetFeedback handles GET /admin/feedback
func GetFeedback(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		status := e.Request.URL.Query().Get("status")

		type feedbackRow struct {
			ID               string  `db:"id" json:"id"`
			UserID           string  `db:"user_id" json:"user_id"`
			Username         string  `db:"username" json:"username"`
			DisplayName      *string `db:"display_name" json:"display_name"`
			Type             string  `db:"type" json:"type"`
			Title            string  `db:"title" json:"title"`
			Description      string  `db:"description" json:"description"`
			StepsToReproduce *string `db:"steps_to_reproduce" json:"steps_to_reproduce"`
			Severity         *string `db:"severity" json:"severity"`
			Status           string  `db:"status" json:"status"`
			CreatedAt        string  `db:"created_at" json:"created_at"`
		}

		query := `
			SELECT f.id, f.user as user_id, u.username, u.display_name,
				   f.type, f.title, f.description, f.steps_to_reproduce,
				   f.severity, f.status, f.created as created_at
			FROM feedback f
			JOIN users u ON f.user = u.id
		`
		binds := map[string]any{}
		if status != "" {
			query += " WHERE f.status = {:status}"
			binds["status"] = status
		}
		query += " ORDER BY f.created DESC"

		var rows []feedbackRow
		err := app.DB().NewQuery(query).Bind(binds).All(&rows)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}

		result := make([]map[string]any, 0, len(rows))
		for _, r := range rows {
			result = append(result, map[string]any{
				"id":                  r.ID,
				"user_id":            r.UserID,
				"username":           r.Username,
				"display_name":       r.DisplayName,
				"type":               r.Type,
				"title":              r.Title,
				"description":        r.Description,
				"steps_to_reproduce": r.StepsToReproduce,
				"severity":           r.Severity,
				"status":             r.Status,
				"created_at":         r.CreatedAt,
			})
		}

		return e.JSON(http.StatusOK, result)
	}
}

// DeleteFeedback handles DELETE /admin/feedback/{feedbackId}
func DeleteFeedback(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		feedbackID := e.Request.PathValue("feedbackId")

		rec, err := app.FindRecordById("feedback", feedbackID)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Feedback not found"})
		}

		if err := app.Delete(rec); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to delete feedback"})
		}

		return e.JSON(http.StatusOK, map[string]any{"ok": true})
	}
}

// UpdateFeedbackStatus handles PATCH /admin/feedback/{feedbackId}
func UpdateFeedbackStatus(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		feedbackID := e.Request.PathValue("feedbackId")

		rec, err := app.FindRecordById("feedback", feedbackID)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Feedback not found"})
		}

		data := struct {
			Status string `json:"status"`
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}
		if data.Status != "open" && data.Status != "closed" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "status must be open or closed"})
		}

		rec.Set("status", data.Status)
		if err := app.Save(rec); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": fmt.Sprintf("Failed to update: %v", err)})
		}

		return e.JSON(http.StatusOK, map[string]any{"ok": true, "status": data.Status})
	}
}
