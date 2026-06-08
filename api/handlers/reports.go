package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

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
		if len(data.Details) > 2000 {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "details must be 2000 characters or fewer"})
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

// GetReports handles GET /admin/reports?status=<s>&page=<n>&perPage=<n>
func GetReports(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		status := e.Request.URL.Query().Get("status")
		page, _ := strconv.Atoi(e.Request.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		perPage, _ := strconv.Atoi(e.Request.URL.Query().Get("perPage"))
		if perPage < 1 || perPage > 100 {
			perPage = 20
		}
		offset := (page - 1) * perPage

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
		query += fmt.Sprintf(" ORDER BY r.created DESC LIMIT %d OFFSET %d", perPage+1, offset)

		var rows []reportRow
		err := app.DB().NewQuery(query).Bind(binds).All(&rows)
		if err != nil {
			return e.JSON(http.StatusOK, map[string]any{
				"reports":  []any{},
				"page":     page,
				"per_page": perPage,
				"has_next": false,
			})
		}

		hasNext := len(rows) > perPage
		if hasNext {
			rows = rows[:perPage]
		}

		// Batch-fetch content previews grouped by type to avoid N+1 queries.
		previews := batchFetchContentPreviews(app, rows)

		result := make([]map[string]any, 0, len(rows))
		for _, r := range rows {
			preview := previews[r.ContentType+":"+r.ContentID]
			if preview == "" {
				preview = "(content not found)"
			}

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

		return e.JSON(http.StatusOK, map[string]any{
			"reports":  result,
			"page":     page,
			"per_page": perPage,
			"has_next": hasNext,
		})
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

// batchFetchContentPreviews batch-fetches content previews for all reports,
// returning a map keyed by "content_type:content_id" → preview string.
// This reduces DB round-trips from N+1 to at most 4 (one per content type).
func batchFetchContentPreviews(app core.App, rows []reportRow) map[string]string {
	// Group content IDs by type.
	idsByType := map[string][]string{}
	for _, r := range rows {
		idsByType[r.ContentType] = append(idsByType[r.ContentType], r.ContentID)
	}

	previews := map[string]string{}

	// Reviews: fetch from user_books
	if ids := idsByType["review"]; len(ids) > 0 {
		placeholders := make([]string, len(ids))
		binds := map[string]any{}
		for i, id := range ids {
			key := fmt.Sprintf("id%d", i)
			placeholders[i] = "{:" + key + "}"
			binds[key] = id
		}
		type reviewRow struct {
			ID         string  `db:"id"`
			ReviewText *string `db:"review_text"`
		}
		var results []reviewRow
		err := app.DB().NewQuery(
			"SELECT id, review_text FROM user_books WHERE id IN ("+strings.Join(placeholders, ",")+")").
			Bind(binds).All(&results)
		if err == nil {
			for _, r := range results {
				text := ""
				if r.ReviewText != nil {
					text = *r.ReviewText
				}
				if len(text) > 200 {
					text = text[:200] + "..."
				}
				if text == "" {
					text = "(rating only, no text)"
				}
				previews["review:"+r.ID] = text
			}
		}
	}

	// Threads: fetch from threads
	if ids := idsByType["thread"]; len(ids) > 0 {
		placeholders := make([]string, len(ids))
		binds := map[string]any{}
		for i, id := range ids {
			key := fmt.Sprintf("id%d", i)
			placeholders[i] = "{:" + key + "}"
			binds[key] = id
		}
		type threadRow struct {
			ID    string `db:"id"`
			Title string `db:"title"`
			Body  string `db:"body"`
		}
		var results []threadRow
		err := app.DB().NewQuery(
			"SELECT id, title, body FROM threads WHERE id IN ("+strings.Join(placeholders, ",")+")").
			Bind(binds).All(&results)
		if err == nil {
			for _, r := range results {
				body := r.Body
				if len(body) > 150 {
					body = body[:150] + "..."
				}
				previews["thread:"+r.ID] = r.Title + ": " + body
			}
		}
	}

	// Comments: fetch from thread_comments
	if ids := idsByType["comment"]; len(ids) > 0 {
		placeholders := make([]string, len(ids))
		binds := map[string]any{}
		for i, id := range ids {
			key := fmt.Sprintf("id%d", i)
			placeholders[i] = "{:" + key + "}"
			binds[key] = id
		}
		type commentRow struct {
			ID   string `db:"id"`
			Body string `db:"body"`
		}
		var results []commentRow
		err := app.DB().NewQuery(
			"SELECT id, body FROM thread_comments WHERE id IN ("+strings.Join(placeholders, ",")+")").
			Bind(binds).All(&results)
		if err == nil {
			for _, r := range results {
				body := r.Body
				if len(body) > 200 {
					body = body[:200] + "..."
				}
				previews["comment:"+r.ID] = body
			}
		}
	}

	// Links: fetch from book_links
	if ids := idsByType["link"]; len(ids) > 0 {
		placeholders := make([]string, len(ids))
		binds := map[string]any{}
		for i, id := range ids {
			key := fmt.Sprintf("id%d", i)
			placeholders[i] = "{:" + key + "}"
			binds[key] = id
		}
		type linkRow struct {
			ID       string  `db:"id"`
			LinkType string  `db:"link_type"`
			Note     *string `db:"note"`
		}
		var results []linkRow
		err := app.DB().NewQuery(
			"SELECT id, link_type, note FROM book_links WHERE id IN ("+strings.Join(placeholders, ",")+")").
			Bind(binds).All(&results)
		if err == nil {
			for _, r := range results {
				if r.Note != nil && *r.Note != "" {
					note := *r.Note
					if len(note) > 150 {
						note = note[:150] + "..."
					}
					previews["link:"+r.ID] = r.LinkType + ": " + note
				} else {
					previews["link:"+r.ID] = r.LinkType + " link"
				}
			}
		}
	}

	return previews
}
