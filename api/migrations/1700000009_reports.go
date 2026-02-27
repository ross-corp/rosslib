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

		reports := core.NewBaseCollection("reports")
		reports.Fields.Add(&core.RelationField{
			Name:          "reporter",
			CollectionId:  users.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		reports.Fields.Add(&core.SelectField{
			Name:      "content_type",
			Values:    []string{"review", "thread", "comment", "link"},
			MaxSelect: 1,
			Required:  true,
		})
		reports.Fields.Add(&core.TextField{Name: "content_id", Required: true})
		reports.Fields.Add(&core.SelectField{
			Name:      "reason",
			Values:    []string{"spam", "harassment", "inappropriate", "other"},
			MaxSelect: 1,
			Required:  true,
		})
		reports.Fields.Add(&core.TextField{Name: "details"})
		reports.Fields.Add(&core.SelectField{
			Name:      "status",
			Values:    []string{"pending", "reviewed", "dismissed"},
			MaxSelect: 1,
			Required:  true,
		})
		reports.Fields.Add(&core.RelationField{
			Name:         "reviewer",
			CollectionId: users.Id,
			MaxSelect:    1,
		})
		reports.AddIndex("idx_reports_status", false, "status", "")
		reports.AddIndex("idx_reports_reporter_content", true, "reporter, content_type, content_id", "")

		return app.Save(reports)
	}, func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("reports")
		if err != nil {
			return nil
		}
		return app.Delete(col)
	})
}
