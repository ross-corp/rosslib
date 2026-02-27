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

		blocks := core.NewBaseCollection("blocks")
		blocks.Fields.Add(&core.RelationField{
			Name:          "blocker",
			CollectionId:  users.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		blocks.Fields.Add(&core.RelationField{
			Name:          "blocked",
			CollectionId:  users.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		blocks.AddIndex("idx_blocks_blocker_blocked", true, "blocker, blocked", "")
		blocks.AddIndex("idx_blocks_blocked", false, "blocked", "")

		return app.Save(blocks)
	}, func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("blocks")
		if err != nil {
			return nil
		}
		return app.Delete(col)
	})
}
