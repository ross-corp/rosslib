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

		// OL edition key like "/books/OL123M"
		userBooks.Fields.Add(&core.TextField{Name: "selected_edition_key"})
		// Cached cover URL for the selected edition
		userBooks.Fields.Add(&core.TextField{Name: "selected_edition_cover_url"})

		return app.Save(userBooks)
	}, func(app core.App) error {
		userBooks, err := app.FindCollectionByNameOrId("user_books")
		if err != nil {
			return err
		}
		userBooks.Fields.RemoveByName("selected_edition_key")
		userBooks.Fields.RemoveByName("selected_edition_cover_url")
		return app.Save(userBooks)
	})
}
