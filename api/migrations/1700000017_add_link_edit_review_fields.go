package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		bookLinkEdits, err := app.FindCollectionByNameOrId("book_link_edits")
		if err != nil {
			return err
		}

		bookLinkEdits.Fields.Add(&core.TextField{Name: "reviewer_comment"})
		bookLinkEdits.Fields.Add(&core.DateField{Name: "reviewed_at"})

		return app.Save(bookLinkEdits)
	}, func(app core.App) error {
		return nil
	})
}
