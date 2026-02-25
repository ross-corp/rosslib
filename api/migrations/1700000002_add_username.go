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
		users.Fields.Add(&core.TextField{Name: "username"})
		if err := app.Save(users); err != nil {
			return err
		}

		// Backfill existing users that have no username.
		records, err := app.FindRecordsByFilter("users", "username = ''", "", 0, 0)
		if err != nil {
			return nil // no users to backfill
		}
		for _, r := range records {
			// Use the email prefix as a fallback username.
			email := r.Email()
			if email != "" {
				parts := []rune(email)
				at := 0
				for i, c := range parts {
					if c == '@' {
						at = i
						break
					}
				}
				if at > 0 {
					r.Set("username", string(parts[:at]))
					_ = app.Save(r)
				}
			}
		}

		return nil
	}, func(app core.App) error {
		return nil
	})
}
