package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Delete collection_items belonging to default read_status shelves
		_, err := app.DB().NewQuery(`
			DELETE FROM collection_items
			WHERE collection IN (
				SELECT id FROM collections WHERE exclusive_group = 'read_status'
			)
		`).Execute()
		if err != nil {
			return err
		}

		// Delete the default read_status shelves themselves
		_, err = app.DB().NewQuery(`
			DELETE FROM collections WHERE exclusive_group = 'read_status'
		`).Execute()
		return err
	}, func(app core.App) error {
		// No rollback — shelves can be recreated manually if needed
		return nil
	})
}
