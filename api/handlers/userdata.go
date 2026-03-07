package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

// DeleteAccount handles DELETE /me/account.
// Removes all user-owned data and then deletes the user account itself.
func DeleteAccount(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		userID := user.Id

		// Delete all user-owned data first (same logic as DeleteAllData)
		userOwned := []string{
			"book_tag_values",
			"collection_items",
			"genre_ratings",
			"user_books",
			"pending_imports",
			"notifications",
			"author_follows",
			"book_follows",
			"activities",
			"book_link_votes",
			"book_link_edits",
			"book_links",
			"thread_comments",
			"threads",
			"tag_values",
			"tag_keys",
			"collections",
			"recommendations",
		}

		for _, coll := range userOwned {
			if err := deleteUserRecords(app, coll, "user", userID); err != nil {
				log.Printf("DeleteAccount: error deleting from %s: %v", coll, err)
			}
		}

		// Follows have follower/followee instead of user
		if err := deleteUserRecords(app, "follows", "follower", userID); err != nil {
			log.Printf("DeleteAccount: error deleting follows (follower): %v", err)
		}
		if err := deleteUserRecords(app, "follows", "followee", userID); err != nil {
			log.Printf("DeleteAccount: error deleting follows (followee): %v", err)
		}

		// book_link_edits also has a "reviewer" field
		if err := deleteUserRecords(app, "book_link_edits", "reviewer", userID); err != nil {
			log.Printf("DeleteAccount: error deleting book_link_edits (reviewer): %v", err)
		}

		// activities also has a "target_user" field
		if err := deleteUserRecords(app, "activities", "target_user", userID); err != nil {
			log.Printf("DeleteAccount: error deleting activities (target_user): %v", err)
		}

		// recommendations where user is the recipient
		if err := deleteUserRecords(app, "recommendations", "recipient", userID); err != nil {
			log.Printf("DeleteAccount: error deleting recommendations (recipient): %v", err)
		}

		// Delete the user account itself
		if err := app.Delete(user); err != nil {
			log.Printf("DeleteAccount: error deleting user record: %v", err)
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to delete account"})
		}

		return e.JSON(http.StatusOK, map[string]any{"message": "Account deleted"})
	}
}

// DeleteAllData handles DELETE /me/account/data.
// Removes all user-owned data but keeps the user account itself.
func DeleteAllData(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		userID := user.Id

		// Collections to purge where user is the owner field "user"
		userOwned := []string{
			"book_tag_values",
			"collection_items",
			"genre_ratings",
			"user_books",
			"pending_imports",
			"notifications",
			"author_follows",
			"book_follows",
			"activities",
			"book_link_votes",
			"book_link_edits",
			"book_links",
			"thread_comments",
			"threads",
			"tag_values",
			"tag_keys",
			"collections",
		}

		for _, coll := range userOwned {
			if err := deleteUserRecords(app, coll, "user", userID); err != nil {
				log.Printf("DeleteAllData: error deleting from %s: %v", coll, err)
			}
		}

		// Follows have follower/followee instead of user
		if err := deleteUserRecords(app, "follows", "follower", userID); err != nil {
			log.Printf("DeleteAllData: error deleting follows (follower): %v", err)
		}
		if err := deleteUserRecords(app, "follows", "followee", userID); err != nil {
			log.Printf("DeleteAllData: error deleting follows (followee): %v", err)
		}

		// book_link_edits also has a "reviewer" field
		if err := deleteUserRecords(app, "book_link_edits", "reviewer", userID); err != nil {
			log.Printf("DeleteAllData: error deleting book_link_edits (reviewer): %v", err)
		}

		// activities also has a "target_user" field
		if err := deleteUserRecords(app, "activities", "target_user", userID); err != nil {
			log.Printf("DeleteAllData: error deleting activities (target_user): %v", err)
		}

		return e.JSON(http.StatusOK, map[string]any{"message": "All data deleted"})
	}
}

// deleteUserRecords deletes all records in a collection where field = userID.
// Processes in batches to handle large datasets.
func deleteUserRecords(app core.App, collection, field, userID string) error {
	filter := fmt.Sprintf("%s = {:uid}", field)
	params := map[string]any{"uid": userID}

	for {
		records, err := app.FindRecordsByFilter(collection, filter, "", 200, 0, params)
		if err != nil {
			return err
		}
		if len(records) == 0 {
			return nil
		}
		for _, r := range records {
			if err := app.Delete(r); err != nil {
				return fmt.Errorf("delete %s/%s: %w", collection, r.Id, err)
			}
		}
	}
}
