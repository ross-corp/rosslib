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
		books, err := app.FindCollectionByNameOrId("books")
		if err != nil {
			return err
		}

		sessions := core.NewBaseCollection("reading_sessions")
		sessions.Fields.Add(&core.RelationField{
			Name:          "user",
			CollectionId:  users.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		sessions.Fields.Add(&core.RelationField{
			Name:          "book",
			CollectionId:  books.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		sessions.Fields.Add(&core.DateField{Name: "date_started"})
		sessions.Fields.Add(&core.DateField{Name: "date_finished"})
		sessions.Fields.Add(&core.NumberField{Name: "rating"})
		sessions.Fields.Add(&core.TextField{Name: "notes"})
		sessions.AddIndex("idx_reading_sessions_user_book", false, "user,book", "")

		return app.Save(sessions)
	}, func(app core.App) error {
		coll, err := app.FindCollectionByNameOrId("reading_sessions")
		if err != nil {
			return nil
		}
		return app.Delete(coll)
	})
}
