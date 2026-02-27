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

		feedback := core.NewBaseCollection("feedback")
		feedback.Fields.Add(&core.RelationField{
			Name:          "user",
			CollectionId:  users.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		feedback.Fields.Add(&core.SelectField{
			Name:      "type",
			Values:    []string{"bug", "feature"},
			MaxSelect: 1,
			Required:  true,
		})
		feedback.Fields.Add(&core.TextField{Name: "title", Required: true})
		feedback.Fields.Add(&core.TextField{Name: "description", Required: true})
		feedback.Fields.Add(&core.TextField{Name: "steps_to_reproduce"})
		feedback.Fields.Add(&core.SelectField{
			Name:      "severity",
			Values:    []string{"low", "medium", "high"},
			MaxSelect: 1,
		})
		feedback.Fields.Add(&core.SelectField{
			Name:      "status",
			Values:    []string{"open", "closed"},
			MaxSelect: 1,
			Required:  true,
		})
		feedback.AddIndex("idx_feedback_user", false, "user", "")
		feedback.AddIndex("idx_feedback_status", false, "status", "")

		return app.Save(feedback)
	}, func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("feedback")
		if err != nil {
			return nil
		}
		return app.Delete(col)
	})
}
