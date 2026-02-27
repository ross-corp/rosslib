package handlers

import (
	"fmt"
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

// CreateReport handles POST /reports
func CreateReport(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		data := struct {
			ContentType string `json:"content_type"`
			ContentID   string `json:"content_id"`
			Reason      string `json:"reason"`
			Details     string `json:"details"`
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}

		// Validate content_type
		validTypes := map[string]bool{"review": true, "thread": true, "comment": true, "link": true}
		if !validTypes[data.ContentType] {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "content_type must be review, thread, comment, or link"})
		}

		// Validate reason
		validReasons := map[string]bool{"spam": true, "harassment": true, "inappropriate": true, "other": true}
		if !validReasons[data.Reason] {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "reason must be spam, harassment, inappropriate, or other"})
		}

		if data.ContentID == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "content_id is required"})
		}

		// Check for duplicate report from same user on same content
		existing, err := app.FindRecordsByFilter("reports",
			"reporter = {:reporter} && content_type = {:ct} && content_id = {:cid}",
			"", 1, 0,
			map[string]any{"reporter": user.Id, "ct": data.ContentType, "cid": data.ContentID},
		)
		if err == nil && len(existing) > 0 {
			return e.JSON(http.StatusConflict, map[string]any{"error": "You have already reported this content"})
		}

		coll, err := app.FindCollectionByNameOrId("reports")
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to create report"})
		}
		rec := core.NewRecord(coll)
		rec.Set("reporter", user.Id)
		rec.Set("content_type", data.ContentType)
		rec.Set("content_id", data.ContentID)
		rec.Set("reason", data.Reason)
		rec.Set("details", data.Details)
		rec.Set("status", "pending")
		if err := app.Save(rec); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to save report"})
		}

		return e.JSON(http.StatusCreated, map[string]any{
			"id":         rec.Id,
			"created_at": rec.GetString("created"),
		})
	}
}

// GetReports handles GET /admin/reports
func GetReports(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		status := e.Request.URL.Query().Get("status")

		type reportRow struct {
			ID              string  `db:"id" json:"id"`
			ReporterID      string  `db:"reporter_id" json:"reporter_id"`
			ReporterName    string  `db:"reporter_username" json:"reporter_username"`
			ReporterDisplay *string `db:"reporter_display" json:"reporter_display_name"`
			ContentType     string  `db:"content_type" json:"content_type"`
			ContentID       string  `db:"content_id" json:"content_id"`
			Reason          string  `db:"reason" json:"reason"`
			Details         *string `db:"details" json:"details"`
			Status          string  `db:"status" json:"status"`
			ReviewerID      *string `db:"reviewer_id" json:"reviewer_id"`
			ReviewerName    *string `db:"reviewer_username" json:"reviewer_username"`
			CreatedAt       string  `db:"created_at" json:"created_at"`
		}

		query := `
			SELECT r.id, r.reporter as reporter_id, u.username as reporter_username,
				   u.display_name as reporter_display,
				   r.content_type, r.content_id, r.reason, r.details, r.status,
				   r.reviewer as reviewer_id, rv.username as reviewer_username,
				   r.created as created_at
			FROM reports r
			JOIN users u ON r.reporter = u.id
			LEFT JOIN users rv ON r.reviewer = rv.id
		`
		binds := map[string]any{}
		if status != "" {
			query += " WHERE r.status = {:status}"
			binds["status"] = status
		}
		query += " ORDER BY r.created DESC"

		var rows []reportRow
		err := app.DB().NewQuery(query).Bind(binds).All(&rows)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}

		result := make([]map[string]any, 0, len(rows))
		for _, r := range rows {
			// Fetch a content preview based on type
			preview := fetchContentPreview(app, r.ContentType, r.ContentID)

			result = append(result, map[string]any{
				"id":                    r.ID,
				"reporter_id":          r.ReporterID,
				"reporter_username":    r.ReporterName,
				"reporter_display_name": r.ReporterDisplay,
				"content_type":         r.ContentType,
				"content_id":           r.ContentID,
				"reason":               r.Reason,
				"details":              r.Details,
				"status":               r.Status,
				"reviewer_id":          r.ReviewerID,
				"reviewer_username":    r.ReviewerName,
				"created_at":           r.CreatedAt,
				"content_preview":      preview,
			})
		}

		return e.JSON(http.StatusOK, result)
	}
}

// UpdateReportStatus handles PATCH /admin/reports/{reportId}
func UpdateReportStatus(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		reportID := e.Request.PathValue("reportId")

		rec, err := app.FindRecordById("reports", reportID)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Report not found"})
		}

		data := struct {
			Status string `json:"status"`
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}
		if data.Status != "reviewed" && data.Status != "dismissed" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "status must be reviewed or dismissed"})
		}

		rec.Set("status", data.Status)
		rec.Set("reviewer", e.Auth.Id)
		if err := app.Save(rec); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": fmt.Sprintf("Failed to update: %v", err)})
		}

		return e.JSON(http.StatusOK, map[string]any{"ok": true, "status": data.Status})
	}
}

// fetchContentPreview returns a short preview string for reported content.
func fetchContentPreview(app core.App, contentType, contentID string) string {
	switch contentType {
	case "review":
		// contentID is the user_books record ID
		rec, err := app.FindRecordById("user_books", contentID)
		if err != nil {
			return "(content not found)"
		}
		text := rec.GetString("review_text")
		if len(text) > 200 {
			text = text[:200] + "..."
		}
		if text == "" {
			text = "(rating only, no text)"
		}
		return text

	case "thread":
		rec, err := app.FindRecordById("threads", contentID)
		if err != nil {
			return "(content not found)"
		}
		title := rec.GetString("title")
		body := rec.GetString("body")
		if len(body) > 150 {
			body = body[:150] + "..."
		}
		return title + ": " + body

	case "comment":
		rec, err := app.FindRecordById("thread_comments", contentID)
		if err != nil {
			return "(content not found)"
		}
		body := rec.GetString("body")
		if len(body) > 200 {
			body = body[:200] + "..."
		}
		return body

	case "link":
		rec, err := app.FindRecordById("book_links", contentID)
		if err != nil {
			return "(content not found)"
		}
		note := rec.GetString("note")
		linkType := rec.GetString("link_type")
		if note != "" {
			if len(note) > 150 {
				note = note[:150] + "..."
			}
			return linkType + ": " + note
		}
		return linkType + " link"

	default:
		return "(unknown content type)"
	}
}
