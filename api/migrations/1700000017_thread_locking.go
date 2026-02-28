package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		threads, err := app.FindCollectionByNameOrId("threads")
		if err != nil {
			return err
		}

		threads.Fields.Add(&core.DateField{Name: "locked_at"})

		return app.Save(threads)
	}, func(app core.App) error {
		threads, err := app.FindCollectionByNameOrId("threads")
		if err != nil {
			return nil
		}

		threads.Fields.RemoveByName("locked_at")

		return app.Save(threads)
	})
}
