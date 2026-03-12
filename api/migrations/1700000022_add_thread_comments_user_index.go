package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		threadComments, err := app.FindCollectionByNameOrId("thread_comments")
		if err != nil {
			return err
		}
		threadComments.AddIndex("idx_thread_comments_user", false, "user", "")
		if err := app.Save(threadComments); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		threadComments, err := app.FindCollectionByNameOrId("thread_comments")
		if err != nil {
			return err
		}
		threadComments.RemoveIndex("idx_thread_comments_user")
		return app.Save(threadComments)
	})
}
