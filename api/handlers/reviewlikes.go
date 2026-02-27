package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

// ToggleReviewLike handles POST /books/{workId}/reviews/{userId}/like
// Toggles a like on a review — likes if not liked, unlikes if already liked.
func ToggleReviewLike(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		workID := e.Request.PathValue("workId")
		reviewUserID := e.Request.PathValue("userId")

		// Prevent self-likes
		if user.Id == reviewUserID {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Cannot like your own review"})
		}

		// Find the book
		books, _ := app.FindRecordsByFilter("books",
			"open_library_id = {:id}", "", 1, 0,
			map[string]any{"id": workID},
		)
		if len(books) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Book not found"})
		}
		book := books[0]

		// Verify the review exists (review_user has a review on this book)
		type reviewCheck struct {
			Count int `db:"count"`
		}
		var rc reviewCheck
		_ = app.DB().NewQuery(`
			SELECT COUNT(*) as count FROM user_books
			WHERE user = {:review_user} AND book = {:book}
			AND review_text != '' AND review_text IS NOT NULL
		`).Bind(map[string]any{"review_user": reviewUserID, "book": book.Id}).One(&rc)
		if rc.Count == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Review not found"})
		}

		// Check if already liked
		existing, _ := app.FindRecordsByFilter("review_likes",
			"user = {:user} && book = {:book} && review_user = {:review_user}",
			"", 1, 0,
			map[string]any{"user": user.Id, "book": book.Id, "review_user": reviewUserID},
		)

		if len(existing) > 0 {
			// Unlike
			if err := app.Delete(existing[0]); err != nil {
				return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to unlike"})
			}
			return e.JSON(http.StatusOK, map[string]any{"liked": false})
		}

		// Like
		coll, err := app.FindCollectionByNameOrId("review_likes")
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Collection not found"})
		}
		rec := core.NewRecord(coll)
		rec.Set("user", user.Id)
		rec.Set("book", book.Id)
		rec.Set("review_user", reviewUserID)
		if err := app.Save(rec); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to like"})
		}

		// Record activity
		metadata, _ := json.Marshal(map[string]any{
			"book_ol_id":     workID,
			"review_user_id": reviewUserID,
		})
		recordActivity(app, user.Id, "liked_review", map[string]any{
			"book":     book.Id,
			"metadata": string(metadata),
		})

		// Send notification to review author
		go func() {
			// Get review author's username for the notification
			reviewUser, err := app.FindRecordById("users", reviewUserID)
			if err != nil {
				return
			}
			_ = reviewUser // we only need the review author ID as recipient

			likerUsername := user.GetString("username")
			bookTitle := book.GetString("title")

			notifColl, err := app.FindCollectionByNameOrId("notifications")
			if err != nil {
				return
			}
			notif := core.NewRecord(notifColl)
			notif.Set("user", reviewUserID)
			notif.Set("notif_type", "review_liked")
			notif.Set("title", fmt.Sprintf("%s liked your review of %s", likerUsername, bookTitle))
			notif.Set("metadata", map[string]any{
				"book_ol_id":     workID,
				"liker_username": likerUsername,
			})
			notif.Set("read", false)
			_ = app.Save(notif)
		}()

		return e.JSON(http.StatusOK, map[string]any{"liked": true})
	}
}

// GetReviewLikeStatus handles GET /books/{workId}/reviews/{userId}/like
// Returns whether the current user has liked the specified review.
func GetReviewLikeStatus(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		workID := e.Request.PathValue("workId")
		reviewUserID := e.Request.PathValue("userId")

		books, _ := app.FindRecordsByFilter("books",
			"open_library_id = {:id}", "", 1, 0,
			map[string]any{"id": workID},
		)
		if len(books) == 0 {
			return e.JSON(http.StatusOK, map[string]any{"liked": false})
		}

		existing, _ := app.FindRecordsByFilter("review_likes",
			"user = {:user} && book = {:book} && review_user = {:review_user}",
			"", 1, 0,
			map[string]any{"user": user.Id, "book": books[0].Id, "review_user": reviewUserID},
		)

		return e.JSON(http.StatusOK, map[string]any{"liked": len(existing) > 0})
	}
}
