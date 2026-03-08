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

		apiTokens := core.NewBaseCollection("api_tokens")
		apiTokens.Fields.Add(&core.RelationField{
			Name:          "user",
			CollectionId:  users.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		apiTokens.Fields.Add(&core.TextField{
			Name:     "name",
			Required: true,
		})
		apiTokens.Fields.Add(&core.TextField{
			Name:     "token_hash",
			Required: true,
		})
		apiTokens.Fields.Add(&core.DateField{Name: "last_used_at"})

		apiTokens.AddIndex("idx_api_tokens_user", false, "user", "")
		apiTokens.AddIndex("idx_api_tokens_hash", true, "token_hash", "")

		return app.Save(apiTokens)
	}, func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("api_tokens")
		if err != nil {
			return nil
		}
		return app.Delete(col)
	})
}
