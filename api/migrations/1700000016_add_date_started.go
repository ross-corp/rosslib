package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		userBooks, err := app.FindCollectionByNameOrId("user_books")
		if err != nil {
			return err
		}

		userBooks.Fields.Add(&core.DateField{Name: "date_started"})

		return app.Save(userBooks)
	}, func(app core.App) error {
		return nil
	})
}
