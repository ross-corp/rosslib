package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		books, err := app.FindCollectionByNameOrId("books")
		if err != nil {
			return err
		}

		// Create series collection
		series := core.NewBaseCollection("series")
		series.Fields.Add(&core.TextField{Name: "name", Required: true})
		series.Fields.Add(&core.TextField{Name: "open_library_id"})
		series.Fields.Add(&core.TextField{Name: "description"})
		series.AddIndex("idx_series_name", false, "name", "")

		if err := app.Save(series); err != nil {
			return err
		}

		// Create book_series collection
		bookSeries := core.NewBaseCollection("book_series")
		bookSeries.Fields.Add(&core.RelationField{
			Name:          "book",
			CollectionId:  books.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		bookSeries.Fields.Add(&core.RelationField{
			Name:          "series",
			CollectionId:  series.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		bookSeries.Fields.Add(&core.NumberField{Name: "position"})
		bookSeries.AddIndex("idx_book_series_unique", true, "book,series", "")
		bookSeries.AddIndex("idx_book_series_series", false, "series", "")

		return app.Save(bookSeries)
	}, func(app core.App) error {
		bs, err := app.FindCollectionByNameOrId("book_series")
		if err == nil {
			_ = app.Delete(bs)
		}
		s, err := app.FindCollectionByNameOrId("series")
		if err == nil {
			_ = app.Delete(s)
		}
		return nil
	})
}
