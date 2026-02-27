package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		prefs := core.NewBaseCollection("notification_preferences")
		prefs.Fields.Add(&core.RelationField{
			Name:          "user",
			CollectionId:  users.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		prefs.Fields.Add(&core.BoolField{Name: "new_publication"})
		prefs.Fields.Add(&core.BoolField{Name: "book_new_thread"})
		prefs.Fields.Add(&core.BoolField{Name: "book_new_link"})
		prefs.Fields.Add(&core.BoolField{Name: "book_new_review"})
		prefs.Fields.Add(&core.BoolField{Name: "review_liked"})
		prefs.Fields.Add(&core.BoolField{Name: "thread_mention"})
		prefs.Fields.Add(&core.BoolField{Name: "book_recommendation"})

		if err := app.Save(prefs); err != nil {
			return err
		}

		// Add unique index on user after saving (to avoid issues with auto columns)
		prefs.AddIndex("idx_notif_prefs_user", true, "user", "")
		return app.Save(prefs)
	}, func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("notification_preferences")
		if err != nil {
			return nil
		}
		return app.Delete(col)
	})
}
