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

		reviewComments := core.NewBaseCollection("review_comments")
		reviewComments.Fields.Add(&core.RelationField{
			Name:          "user",
			CollectionId:  users.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		reviewComments.Fields.Add(&core.RelationField{
			Name:          "book",
			CollectionId:  books.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		reviewComments.Fields.Add(&core.RelationField{
			Name:          "review_user",
			CollectionId:  users.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		reviewComments.Fields.Add(&core.TextField{
			Name:     "body",
			Required: true,
			Max:      2000,
		})
		reviewComments.Fields.Add(&core.DateField{Name: "deleted_at"})

		if err := app.Save(reviewComments); err != nil {
			return err
		}

		reviewComments.AddIndex("idx_review_comments_review", false, "book, review_user", "")
		reviewComments.AddIndex("idx_review_comments_user", false, "user", "")

		if err := app.Save(reviewComments); err != nil {
			return err
		}

		// Add review_comment preference to notification_preferences
		prefs, err := app.FindCollectionByNameOrId("notification_preferences")
		if err != nil {
			return err
		}
		prefs.Fields.Add(&core.BoolField{Name: "review_comment"})
		return app.Save(prefs)
	}, func(app core.App) error {
		// Remove review_comment preference
		prefs, err := app.FindCollectionByNameOrId("notification_preferences")
		if err == nil {
			prefs.Fields.RemoveByName("review_comment")
			_ = app.Save(prefs)
		}

		col, err := app.FindCollectionByNameOrId("review_comments")
		if err != nil {
			return nil
		}
		return app.Delete(col)
	})
}
