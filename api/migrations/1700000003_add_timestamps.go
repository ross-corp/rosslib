package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Add created/updated autodate fields to activities
		activities, err := app.FindCollectionByNameOrId("activities")
		if err != nil {
			return err
		}
		activities.Fields.Add(&core.AutodateField{
			Name:     "created",
			OnCreate: true,
		})
		activities.Fields.Add(&core.AutodateField{
			Name:     "updated",
			OnCreate: true,
			OnUpdate: true,
		})
		if err := app.Save(activities); err != nil {
			return err
		}

		// Add created/updated autodate fields to book_tag_values
		btv, err := app.FindCollectionByNameOrId("book_tag_values")
		if err != nil {
			return err
		}
		btv.Fields.Add(&core.AutodateField{
			Name:     "created",
			OnCreate: true,
		})
		btv.Fields.Add(&core.AutodateField{
			Name:     "updated",
			OnCreate: true,
			OnUpdate: true,
		})
		if err := app.Save(btv); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		return nil
	})
}
