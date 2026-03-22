package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		bookLinks, err := app.FindCollectionByNameOrId("book_links")
		if err != nil {
			return err
		}
		bookLinks.AddIndex("idx_book_links_from_book_deleted_at", false, "from_book,deleted_at", "")
		if err := app.Save(bookLinks); err != nil {
			return err
		}
		return nil
	}, func(app core.App) error {
		bookLinks, err := app.FindCollectionByNameOrId("book_links")
		if err != nil {
			return err
		}
		bookLinks.RemoveIndex("idx_book_links_from_book_deleted_at")
		return app.Save(bookLinks)
	})
}
