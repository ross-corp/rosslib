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

		// selected_edition_key stores the Open Library edition key (e.g. "OL123M")
		// When set, the frontend displays this edition's cover instead of the default work cover.
		userBooks.Fields.Add(&core.TextField{Name: "selected_edition_key"})

		// selected_edition_cover_url caches the edition's cover URL so the frontend
		// doesn't need an extra API call to resolve the cover.
		userBooks.Fields.Add(&core.TextField{Name: "selected_edition_cover_url"})

		return app.Save(userBooks)
	}, func(app core.App) error {
		return nil
	})
}
