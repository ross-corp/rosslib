package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		prefs, err := app.FindCollectionByNameOrId("notification_preferences")
		if err != nil {
			return err
		}
		prefs.Fields.Add(&core.BoolField{Name: "new_follower"})
		return app.Save(prefs)
	}, func(app core.App) error {
		prefs, err := app.FindCollectionByNameOrId("notification_preferences")
		if err != nil {
			return nil
		}
		prefs.Fields.RemoveByName("new_follower")
		return app.Save(prefs)
	})
}
