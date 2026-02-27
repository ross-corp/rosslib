package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collections, err := app.FindCollectionByNameOrId("collections")
		if err != nil {
			return err
		}

		// Add computed list fields
		collections.Fields.Add(&core.SelectField{
			Name:      "operation_type",
			Values:    []string{"union", "intersection", "difference"},
			MaxSelect: 1,
		})
		collections.Fields.Add(&core.BoolField{Name: "is_continuous"})
		collections.Fields.Add(&core.RelationField{
			Name:         "source_collection_a",
			CollectionId: collections.Id,
			MaxSelect:    1,
		})
		collections.Fields.Add(&core.RelationField{
			Name:         "source_collection_b",
			CollectionId: collections.Id,
			MaxSelect:    1,
		})
		collections.Fields.Add(&core.DateField{Name: "last_computed_at"})

		// Update collection_type to also allow "computed"
		collections.Fields.Add(&core.SelectField{
			Name:      "collection_type",
			Values:    []string{"shelf", "list", "tag", "computed"},
			MaxSelect: 1,
		})

		if err := app.Save(collections); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		// Downgrade: remove computed fields
		collections, err := app.FindCollectionByNameOrId("collections")
		if err != nil {
			return err
		}

		collections.Fields.RemoveByName("operation_type")
		collections.Fields.RemoveByName("is_continuous")
		collections.Fields.RemoveByName("source_collection_a")
		collections.Fields.RemoveByName("source_collection_b")
		collections.Fields.RemoveByName("last_computed_at")

		// Restore original collection_type values
		collections.Fields.Add(&core.SelectField{
			Name:      "collection_type",
			Values:    []string{"shelf", "list"},
			MaxSelect: 1,
		})

		return app.Save(collections)
	})
}
