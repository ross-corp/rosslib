package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Add selected_edition_key and selected_edition_cover_url to user_books
		ub, err := app.FindCollectionByNameOrId("user_books")
		if err != nil {
			return err
		}
		ub.Fields.Add(&core.TextField{
			Name: "selected_edition_key",
		})
		ub.Fields.Add(&core.TextField{
			Name: "selected_edition_cover_url",
		})
		return app.Save(ub)
	}, func(app core.App) error {
		return nil
	})
}
