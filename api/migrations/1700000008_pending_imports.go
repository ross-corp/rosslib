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

		pendingImports := core.NewBaseCollection("pending_imports")
		pendingImports.Fields.Add(&core.RelationField{
			Name:          "user",
			CollectionId:  users.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		pendingImports.Fields.Add(&core.TextField{Name: "source", Required: true})
		pendingImports.Fields.Add(&core.TextField{Name: "title", Required: true})
		pendingImports.Fields.Add(&core.TextField{Name: "author"})
		pendingImports.Fields.Add(&core.TextField{Name: "isbn13"})
		pendingImports.Fields.Add(&core.TextField{Name: "exclusive_shelf"})
		pendingImports.Fields.Add(&core.JSONField{Name: "custom_shelves"})
		pendingImports.Fields.Add(&core.NumberField{Name: "rating"})
		pendingImports.Fields.Add(&core.TextField{Name: "review_text"})
		pendingImports.Fields.Add(&core.TextField{Name: "date_read"})
		pendingImports.Fields.Add(&core.TextField{Name: "date_added"})
		pendingImports.Fields.Add(&core.SelectField{
			Name:      "status",
			Values:    []string{"unmatched", "resolved"},
			MaxSelect: 1,
			Required:  true,
		})
		pendingImports.AddIndex("idx_pending_imports_user_status", false, "user,status", "")

		return app.Save(pendingImports)
	}, func(app core.App) error {
		coll, err := app.FindCollectionByNameOrId("pending_imports")
		if err != nil {
			return nil
		}
		return app.Delete(coll)
	})
}
