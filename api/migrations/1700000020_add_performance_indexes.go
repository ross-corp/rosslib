package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Add compound index on user_books for timeline and activity queries
		userBooks, err := app.FindCollectionByNameOrId("user_books")
		if err != nil {
			return err
		}
		userBooks.AddIndex("idx_user_books_user_date_added", false, "user,date_added DESC", "")
		userBooks.AddIndex("idx_user_books_user_date_read", false, "user,date_read", "")
		if err := app.Save(userBooks); err != nil {
			return err
		}

		// Add compound index on activities for feed filtering
		// Note: saving first then adding created-based index to avoid
		// PocketBase index-on-auto-column issue
		activities, err := app.FindCollectionByNameOrId("activities")
		if err != nil {
			return err
		}
		activities.AddIndex("idx_activities_user_created", false, "user,created DESC", "")
		if err := app.Save(activities); err != nil {
			// If compound index with created fails, fall back to saving without it
			activities.RemoveIndex("idx_activities_user_created")
			_ = app.Save(activities)
		}

		return nil
	}, func(app core.App) error {
		return nil
	})
}
