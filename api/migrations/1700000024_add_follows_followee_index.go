package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		follows, err := app.FindCollectionByNameOrId("follows")
		if err != nil {
			return err
		}
		follows.AddIndex("idx_follows_followee", false, "followee", "")
		if err := app.Save(follows); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		follows, err := app.FindCollectionByNameOrId("follows")
		if err != nil {
			return err
		}
		follows.RemoveIndex("idx_follows_followee")
		return app.Save(follows)
	})
}
