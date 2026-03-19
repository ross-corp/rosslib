package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		reports, err := app.FindCollectionByNameOrId("reports")
		if err != nil {
			return err
		}
		// Note: created is an auto-generated column, so use fallback
		// pattern to avoid PocketBase index-on-auto-column issue
		reports.AddIndex("idx_reports_status_created", false, "status,created DESC", "")
		if err := app.Save(reports); err != nil {
			reports.RemoveIndex("idx_reports_status_created")
			_ = app.Save(reports)
		}

		return nil
	}, func(app core.App) error {
		reports, err := app.FindCollectionByNameOrId("reports")
		if err != nil {
			return err
		}
		reports.RemoveIndex("idx_reports_status_created")
		return app.Save(reports)
	})
}
