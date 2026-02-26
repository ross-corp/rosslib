package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

// GetPendingImports handles GET /me/imports/pending
func GetPendingImports(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		records, err := app.FindRecordsByFilter(
			"pending_imports",
			"user = {:user} && status = 'unmatched'",
			"-created",
			200, 0,
			map[string]any{"user": user.Id},
		)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to fetch pending imports"})
		}

		items := make([]map[string]any, 0, len(records))
		for _, r := range records {
			var customShelves []string
			raw := r.Get("custom_shelves")
			if raw != nil {
				if b, err := json.Marshal(raw); err == nil {
					_ = json.Unmarshal(b, &customShelves)
				}
			}
			if customShelves == nil {
				customShelves = []string{}
			}

			items = append(items, map[string]any{
				"id":              r.Id,
				"title":           r.GetString("title"),
				"author":          r.GetString("author"),
				"isbn13":          r.GetString("isbn13"),
				"exclusive_shelf": r.GetString("exclusive_shelf"),
				"custom_shelves":  customShelves,
				"rating":          r.Get("rating"),
				"review_text":     r.GetString("review_text"),
				"date_read":       r.GetString("date_read"),
				"date_added":      r.GetString("date_added"),
				"created":         r.GetString("created"),
			})
		}

		return e.JSON(http.StatusOK, items)
	}
}

// ResolvePendingImport handles PATCH /me/imports/pending/:id
// Resolves a pending import by matching it to an OL work ID, or dismisses it.
func ResolvePendingImport(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		id := e.Request.PathValue("id")
		record, err := app.FindRecordById("pending_imports", id)
		if err != nil || record.GetString("user") != user.Id {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Pending import not found"})
		}
		if record.GetString("status") != "unmatched" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Already resolved"})
		}

		data := struct {
			Action string `json:"action"` // "resolve" or "dismiss"
			OLID   string `json:"ol_id"`  // required if action=resolve
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}

		switch data.Action {
		case "dismiss":
			record.Set("status", "resolved")
			if err := app.Save(record); err != nil {
				return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to update"})
			}
			return e.JSON(http.StatusOK, map[string]any{"ok": true})

		case "resolve":
			if data.OLID == "" {
				return e.JSON(http.StatusBadRequest, map[string]any{"error": "ol_id is required for resolve action"})
			}

			// Look up the book from OL and upsert
			book, err := upsertBook(app, data.OLID, record.GetString("title"), "", record.GetString("isbn13"), record.GetString("author"), 0)
			if err != nil {
				return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to save book"})
			}

			// Create user_book
			existing, _ := app.FindRecordsByFilter("user_books",
				"user = {:user} && book = {:book}",
				"", 1, 0,
				map[string]any{"user": user.Id, "book": book.Id},
			)
			var ub *core.Record
			if len(existing) > 0 {
				ub = existing[0]
			} else {
				coll, err := app.FindCollectionByNameOrId("user_books")
				if err != nil {
					return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to create user_book"})
				}
				ub = core.NewRecord(coll)
				ub.Set("user", user.Id)
				ub.Set("book", book.Id)
			}

			rating := record.Get("rating")
			if rating != nil && rating != 0.0 {
				ub.Set("rating", rating)
			}
			if review := record.GetString("review_text"); review != "" {
				ub.Set("review_text", review)
			}
			if dr := record.GetString("date_read"); dr != "" {
				ub.Set("date_read", dr)
			}
			if da := record.GetString("date_added"); da != "" {
				ub.Set("date_added", da)
			}

			if err := app.Save(ub); err != nil {
				return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to save user_book"})
			}

			// Map exclusive shelf to status tag
			statusSlug := mapGoodreadsShelf(record.GetString("exclusive_shelf"))
			if statusSlug != "" {
				setStatusTag(app, user.Id, book.Id, statusSlug)
			}

			// Mark resolved
			record.Set("status", "resolved")
			if err := app.Save(record); err != nil {
				return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to update"})
			}

			refreshBookStats(app, book.Id)

			return e.JSON(http.StatusOK, map[string]any{"ok": true, "book_id": book.Id})

		default:
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "action must be 'resolve' or 'dismiss'"})
		}
	}
}

// DeletePendingImport handles DELETE /me/imports/pending/:id
func DeletePendingImport(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		id := e.Request.PathValue("id")
		record, err := app.FindRecordById("pending_imports", id)
		if err != nil || record.GetString("user") != user.Id {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Pending import not found"})
		}

		if err := app.Delete(record); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to delete"})
		}

		return e.JSON(http.StatusOK, map[string]any{"ok": true})
	}
}
