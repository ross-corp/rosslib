package handlers

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

// GetBookLinks handles GET /books/{workId}/links
func GetBookLinks(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		workID := e.Request.PathValue("workId")

		books, _ := app.FindRecordsByFilter("books",
			"open_library_id = {:id}", "", 1, 0,
			map[string]any{"id": workID},
		)
		if len(books) == 0 {
			return e.JSON(http.StatusOK, []any{})
		}

		viewerID := ""
		if e.Auth != nil {
			viewerID = e.Auth.Id
		}

		type linkRow struct {
			ID       string  `db:"id" json:"id"`
			FromBook string  `db:"from_book" json:"from_book"`
			ToBook   string  `db:"to_book" json:"to_book"`
			ToOLID   string  `db:"to_olid" json:"to_open_library_id"`
			ToTitle  string  `db:"to_title" json:"to_title"`
			UserID   string  `db:"user_id" json:"user_id"`
			LinkType string  `db:"link_type" json:"link_type"`
			Note     *string `db:"note" json:"note"`
		}

		var links []linkRow
		err := app.DB().NewQuery(`
			SELECT bl.id, bl.from_book, bl.to_book, b2.open_library_id as to_olid,
				   b2.title as to_title, bl.user as user_id, bl.link_type, bl.note
			FROM book_links bl
			JOIN books b2 ON bl.to_book = b2.id
			WHERE bl.from_book = {:book} AND (bl.deleted_at IS NULL OR bl.deleted_at = '')
			ORDER BY bl.created DESC
		`).Bind(map[string]any{"book": books[0].Id}).All(&links)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}

		var result []map[string]any
		for _, l := range links {
			// Count votes
			type countResult struct {
				Count int `db:"count"`
			}
			var cnt countResult
			_ = app.DB().NewQuery("SELECT COUNT(*) as count FROM book_link_votes WHERE book_link = {:link}").
				Bind(map[string]any{"link": l.ID}).One(&cnt)

			voted := false
			if viewerID != "" {
				votes, _ := app.FindRecordsByFilter("book_link_votes",
					"user = {:user} && book_link = {:link}",
					"", 1, 0,
					map[string]any{"user": viewerID, "link": l.ID},
				)
				voted = len(votes) > 0
			}

			result = append(result, map[string]any{
				"id":                   l.ID,
				"from_book":            l.FromBook,
				"to_book":              l.ToBook,
				"to_open_library_id":   l.ToOLID,
				"to_title":             l.ToTitle,
				"user_id":              l.UserID,
				"link_type":            l.LinkType,
				"note":                 l.Note,
				"vote_count":           cnt.Count,
				"voted":                voted,
			})
		}
		if result == nil {
			result = []map[string]any{}
		}

		return e.JSON(http.StatusOK, result)
	}
}

// CreateBookLink handles POST /books/{workId}/links
func CreateBookLink(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		workID := e.Request.PathValue("workId")

		fromBooks, _ := app.FindRecordsByFilter("books",
			"open_library_id = {:id}", "", 1, 0,
			map[string]any{"id": workID},
		)
		if len(fromBooks) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Book not found"})
		}

		data := struct {
			ToOpenLibraryID string `json:"to_open_library_id"`
			LinkType        string `json:"link_type"`
			Note            string `json:"note"`
		}{}
		if err := e.BindBody(&data); err != nil || data.ToOpenLibraryID == "" || data.LinkType == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "to_open_library_id and link_type required"})
		}

		toBooks, _ := app.FindRecordsByFilter("books",
			"open_library_id = {:id}", "", 1, 0,
			map[string]any{"id": data.ToOpenLibraryID},
		)
		if len(toBooks) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Target book not found"})
		}

		coll, err := app.FindCollectionByNameOrId("book_links")
		if err != nil {
			return err
		}
		rec := core.NewRecord(coll)
		rec.Set("from_book", fromBooks[0].Id)
		rec.Set("to_book", toBooks[0].Id)
		rec.Set("user", user.Id)
		rec.Set("link_type", data.LinkType)
		rec.Set("note", data.Note)
		if err := app.Save(rec); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}

		return e.JSON(http.StatusOK, map[string]any{
			"id":        rec.Id,
			"link_type": data.LinkType,
		})
	}
}

// DeleteBookLink handles DELETE /links/{linkId}
func DeleteBookLink(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		linkID := e.Request.PathValue("linkId")

		link, err := app.FindRecordById("book_links", linkID)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Link not found"})
		}
		if link.GetString("user") != user.Id {
			return e.JSON(http.StatusForbidden, map[string]any{"error": "Not your link"})
		}

		if err := app.Delete(link); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to delete"})
		}

		return e.JSON(http.StatusOK, map[string]any{"message": "Link deleted"})
	}
}

// VoteLink handles POST /links/{linkId}/vote
func VoteLink(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		linkID := e.Request.PathValue("linkId")

		existing, _ := app.FindRecordsByFilter("book_link_votes",
			"user = {:user} && book_link = {:link}",
			"", 1, 0,
			map[string]any{"user": user.Id, "link": linkID},
		)
		if len(existing) > 0 {
			return e.JSON(http.StatusOK, map[string]any{"message": "Already voted"})
		}

		coll, err := app.FindCollectionByNameOrId("book_link_votes")
		if err != nil {
			return err
		}
		rec := core.NewRecord(coll)
		rec.Set("user", user.Id)
		rec.Set("book_link", linkID)
		if err := app.Save(rec); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}

		return e.JSON(http.StatusOK, map[string]any{"message": "Voted"})
	}
}

// UnvoteLink handles DELETE /links/{linkId}/vote
func UnvoteLink(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		linkID := e.Request.PathValue("linkId")

		existing, _ := app.FindRecordsByFilter("book_link_votes",
			"user = {:user} && book_link = {:link}",
			"", 1, 0,
			map[string]any{"user": user.Id, "link": linkID},
		)
		if len(existing) == 0 {
			return e.JSON(http.StatusOK, map[string]any{"message": "Not voted"})
		}

		_ = app.Delete(existing[0])
		return e.JSON(http.StatusOK, map[string]any{"message": "Vote removed"})
	}
}

// ProposeLinkEdit handles POST /links/{linkId}/edits
func ProposeLinkEdit(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		linkID := e.Request.PathValue("linkId")

		data := struct {
			ProposedType string `json:"proposed_type"`
			ProposedNote string `json:"proposed_note"`
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}

		coll, err := app.FindCollectionByNameOrId("book_link_edits")
		if err != nil {
			return err
		}
		rec := core.NewRecord(coll)
		rec.Set("book_link", linkID)
		rec.Set("user", user.Id)
		rec.Set("proposed_type", data.ProposedType)
		rec.Set("proposed_note", data.ProposedNote)
		rec.Set("status", "pending")
		if err := app.Save(rec); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}

		return e.JSON(http.StatusOK, map[string]any{"id": rec.Id})
	}
}

// GetPendingLinkEdits handles GET /admin/link-edits
func GetPendingLinkEdits(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		edits, err := app.FindRecordsByFilter("book_link_edits",
			"status = 'pending'", "-created", 50, 0, nil,
		)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}

		var result []map[string]any
		for _, edit := range edits {
			result = append(result, map[string]any{
				"id":            edit.Id,
				"book_link":     edit.GetString("book_link"),
				"user":          edit.GetString("user"),
				"proposed_type": edit.GetString("proposed_type"),
				"proposed_note": edit.GetString("proposed_note"),
				"status":        edit.GetString("status"),
				"created_at":    edit.GetString("created"),
			})
		}
		if result == nil {
			result = []map[string]any{}
		}

		return e.JSON(http.StatusOK, result)
	}
}

// ReviewLinkEdit handles PUT /admin/link-edits/{editId}
func ReviewLinkEdit(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		editID := e.Request.PathValue("editId")

		edit, err := app.FindRecordById("book_link_edits", editID)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Edit not found"})
		}

		data := struct {
			Status string `json:"status"` // "approved" or "rejected"
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}
		if data.Status != "approved" && data.Status != "rejected" {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "status must be approved or rejected"})
		}

		edit.Set("status", data.Status)
		edit.Set("reviewer", user.Id)
		if err := app.Save(edit); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to update"})
		}

		// If approved, apply the edit to the link
		if data.Status == "approved" {
			linkID := edit.GetString("book_link")
			link, err := app.FindRecordById("book_links", linkID)
			if err == nil {
				if pt := edit.GetString("proposed_type"); pt != "" {
					link.Set("link_type", pt)
				}
				if pn := edit.GetString("proposed_note"); pn != "" {
					link.Set("note", pn)
				}
				_ = app.Save(link)
			}
		}

		return e.JSON(http.StatusOK, map[string]any{"message": "Edit " + data.Status})
	}
}
