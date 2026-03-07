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

		savedSearches := core.NewBaseCollection("saved_searches")
		savedSearches.Fields.Add(&core.RelationField{
			Name:          "user",
			CollectionId:  users.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		savedSearches.Fields.Add(&core.TextField{
			Name:     "name",
			Required: true,
			Max:      100,
		})
		savedSearches.Fields.Add(&core.TextField{
			Name:     "query",
			Required: true,
		})
		savedSearches.Fields.Add(&core.JSONField{
			Name: "filters",
		})

		if err := app.Save(savedSearches); err != nil {
			return err
		}

		savedSearches.AddIndex("idx_saved_searches_user", false, "user", "")

		return app.Save(savedSearches)
	}, func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("saved_searches")
		if err != nil {
			return nil
		}
		return app.Delete(col)
	})
}
