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

		readingGoals := core.NewBaseCollection("reading_goals")
		readingGoals.Fields.Add(&core.RelationField{
			Name:          "user",
			CollectionId:  users.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		readingGoals.Fields.Add(&core.NumberField{
			Name:     "year",
			Required: true,
		})
		readingGoals.Fields.Add(&core.NumberField{
			Name:     "target",
			Required: true,
		})
		readingGoals.AddIndex("idx_reading_goals_user_year", true, "user,year", "")

		return app.Save(readingGoals)
	}, func(app core.App) error {
		coll, err := app.FindCollectionByNameOrId("reading_goals")
		if err != nil {
			return nil
		}
		return app.Delete(coll)
	})
}
