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

		users.Fields.Add(&core.FileField{
			Name:      "banner",
			MaxSelect: 1,
			MaxSize:   10 * 1024 * 1024, // 10MB
			MimeTypes: []string{"image/jpeg", "image/png", "image/gif", "image/webp"},
		})

		return app.Save(users)
	}, func(app core.App) error {
		return nil
	})
}
