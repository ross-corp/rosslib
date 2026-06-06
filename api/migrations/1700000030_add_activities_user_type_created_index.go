package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		activities, err := app.FindCollectionByNameOrId("activities")
		if err != nil {
			return err
		}
		activities.AddIndex("idx_activities_user_type_created", false, "user,activity_type,created DESC", "")
		if err := app.Save(activities); err != nil {
			// If compound index with created fails, fall back to saving without it
			activities.RemoveIndex("idx_activities_user_type_created")
			_ = app.Save(activities)
		}
		return nil
	}, func(app core.App) error {
		activities, err := app.FindCollectionByNameOrId("activities")
		if err != nil {
			return nil
		}
		activities.RemoveIndex("idx_activities_user_type_created")
		return app.Save(activities)
	})
}
