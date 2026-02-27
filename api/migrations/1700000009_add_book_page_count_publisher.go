package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		books, err := app.FindCollectionByNameOrId("books")
		if err != nil {
			return err
		}

		books.Fields.Add(&core.NumberField{Name: "page_count"})
		books.Fields.Add(&core.TextField{Name: "publisher"})

		return app.Save(books)
	}, func(app core.App) error {
		return nil
	})
}
