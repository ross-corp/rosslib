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
		bookLinks.AddIndex("idx_book_links_user", false, "user", "")
		if err := app.Save(bookLinks); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		bookLinks, err := app.FindCollectionByNameOrId("book_links")
		if err != nil {
			return err
		}
		bookLinks.RemoveIndex("idx_book_links_user")
		return app.Save(bookLinks)
	})
}
