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

		reviewLikes := core.NewBaseCollection("review_likes")
		reviewLikes.Fields.Add(&core.RelationField{
			Name:          "user",
			CollectionId:  users.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		reviewLikes.Fields.Add(&core.RelationField{
			Name:          "book",
			CollectionId:  books.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		reviewLikes.Fields.Add(&core.RelationField{
			Name:          "review_user",
			CollectionId:  users.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})

		if err := app.Save(reviewLikes); err != nil {
			return err
		}

		reviewLikes.AddIndex("idx_review_likes_unique", true, "user, book, review_user", "")
		reviewLikes.AddIndex("idx_review_likes_review", false, "book, review_user", "")

		return app.Save(reviewLikes)
	}, func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("review_likes")
		if err != nil {
			return nil
		}
		return app.Delete(col)
	})
}
