package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		threads, err := app.FindCollectionByNameOrId("threads")
		if err != nil {
			return err
		}
		// Note: created is an auto-generated column, so use fallback
		// pattern to avoid PocketBase index-on-auto-column issue
		threads.AddIndex("idx_threads_book_created", false, "book,created DESC", "")
		if err := app.Save(threads); err != nil {
			threads.RemoveIndex("idx_threads_book_created")
			_ = app.Save(threads)
		}

		return nil
	}, func(app core.App) error {
		threads, err := app.FindCollectionByNameOrId("threads")
		if err != nil {
			return err
		}
		threads.RemoveIndex("idx_threads_book_created")
		return app.Save(threads)
	})
}
