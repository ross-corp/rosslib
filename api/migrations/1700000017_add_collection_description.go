package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collections, err := app.FindCollectionByNameOrId("collections")
		if err != nil {
			return err
		}

		collections.Fields.Add(&core.TextField{
			Name:    "description",
			Max:     1000,
		})

		return app.Save(collections)
	}, func(app core.App) error {
		return nil
	})
}
