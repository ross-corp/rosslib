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

		recommendations := core.NewBaseCollection("recommendations")
		recommendations.Fields.Add(&core.RelationField{
			Name:          "sender",
			CollectionId:  users.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		recommendations.Fields.Add(&core.RelationField{
			Name:          "recipient",
			CollectionId:  users.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		recommendations.Fields.Add(&core.RelationField{
			Name:          "book",
			CollectionId:  books.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		recommendations.Fields.Add(&core.TextField{Name: "note"})
		recommendations.Fields.Add(&core.SelectField{
			Name:      "status",
			Values:    []string{"pending", "seen", "dismissed"},
			MaxSelect: 1,
			Required:  true,
		})

		recommendations.AddIndex("idx_recommendations_recipient", false, "recipient,status", "")
		recommendations.AddIndex("idx_recommendations_sender", false, "sender", "")
		recommendations.AddIndex("idx_recommendations_unique", true, "sender,recipient,book", "")

		return app.Save(recommendations)
	}, func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("recommendations")
		if err != nil {
			return nil
		}
		return app.Delete(col)
	})
}
