package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		reviewComments, err := app.FindCollectionByNameOrId("review_comments")
		if err != nil {
			return err
		}
		reviewComments.AddIndex("idx_review_comments_book_review_user_deleted_at", false, "book,review_user,deleted_at", "")
		if err := app.Save(reviewComments); err != nil {
			return err
		}
		return nil
	}, func(app core.App) error {
		reviewComments, err := app.FindCollectionByNameOrId("review_comments")
		if err != nil {
			return err
		}
		reviewComments.RemoveIndex("idx_review_comments_book_review_user_deleted_at")
		return app.Save(reviewComments)
	})
}
