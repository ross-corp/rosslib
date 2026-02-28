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

		bookQuotes := core.NewBaseCollection("book_quotes")
		bookQuotes.Fields.Add(&core.RelationField{
			Name:          "user",
			CollectionId:  users.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		bookQuotes.Fields.Add(&core.RelationField{
			Name:          "book",
			CollectionId:  books.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		bookQuotes.Fields.Add(&core.TextField{
			Name:     "text",
			Required: true,
			Max:      2000,
		})
		bookQuotes.Fields.Add(&core.NumberField{
			Name: "page_number",
		})
		bookQuotes.Fields.Add(&core.TextField{
			Name: "note",
			Max:  500,
		})
		bookQuotes.Fields.Add(&core.BoolField{
			Name: "is_public",
		})

		if err := app.Save(bookQuotes); err != nil {
			return err
		}

		bookQuotes.AddIndex("idx_book_quotes_book", false, "book", "")
		bookQuotes.AddIndex("idx_book_quotes_user_book", false, "user, book", "")

		return app.Save(bookQuotes)
	}, func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("book_quotes")
		if err != nil {
			return nil
		}
		return app.Delete(col)
	})
}
