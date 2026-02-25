package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// 1. Update 'users' collection
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}
		users.Fields.Add(&core.TextField{Name: "bio"})
		users.Fields.Add(&core.BoolField{Name: "is_private"})
		if err := app.Save(users); err != nil {
			return err
		}

		// 2. Create 'books' collection
		books := core.NewBaseCollection("books")
		books.Fields.Add(&core.TextField{Name: "open_library_id", Required: true})
		books.Fields.Add(&core.TextField{Name: "title", Required: true})
		books.Fields.Add(&core.TextField{Name: "cover_url"})
		books.Fields.Add(&core.TextField{Name: "isbn13"})
		books.Fields.Add(&core.TextField{Name: "authors"})
		books.Fields.Add(&core.NumberField{Name: "publication_year"})
		// Indexes
		books.AddIndex("idx_books_olid", false, "open_library_id", "")
		if err := app.Save(books); err != nil {
			return err
		}

		// 3. Create 'collections' collection
		collections := core.NewBaseCollection("collections")
		collections.Fields.Add(&core.RelationField{
			Name: "user",
			CollectionId: "users",
			CascadeDelete: true,
			MaxSelect: 1,
			Required: true,
		})
		collections.Fields.Add(&core.TextField{Name: "name", Required: true})
		collections.Fields.Add(&core.TextField{Name: "slug", Required: true})
		collections.Fields.Add(&core.BoolField{Name: "is_exclusive"})
		collections.Fields.Add(&core.TextField{Name: "exclusive_group"})
		collections.Fields.Add(&core.BoolField{Name: "is_public"})
		collections.Fields.Add(&core.SelectField{
			Name: "collection_type",
			Values: []string{"shelf", "list"},
			MaxSelect: 1,
		})
		collections.AddIndex("idx_collections_user_slug", true, "user,slug", "")
		if err := app.Save(collections); err != nil {
			return err
		}

		// 4. Create 'collection_items' collection
		collectionItems := core.NewBaseCollection("collection_items")
		collectionItems.Fields.Add(&core.RelationField{
			Name: "collection",
			CollectionId: "collections",
			CascadeDelete: true,
			MaxSelect: 1,
			Required: true,
		})
		collectionItems.Fields.Add(&core.RelationField{
			Name: "book",
			CollectionId: "books",
			CascadeDelete: true, // Maybe restrict? But usually cascade is fine.
			MaxSelect: 1,
			Required: true,
		})
		collectionItems.Fields.Add(&core.RelationField{
			Name: "user",
			CollectionId: "users",
			CascadeDelete: true,
			MaxSelect: 1,
			Required: true,
		})
		collectionItems.Fields.Add(&core.NumberField{Name: "rating"})
		collectionItems.Fields.Add(&core.TextField{Name: "review_text"})
		collectionItems.Fields.Add(&core.BoolField{Name: "spoiler"})
		collectionItems.Fields.Add(&core.DateField{Name: "date_read"})
		collectionItems.AddIndex("idx_ci_collection_book", true, "collection,book", "")
		if err := app.Save(collectionItems); err != nil {
			return err
		}

		// 5. Create 'follows' collection
		follows := core.NewBaseCollection("follows")
		follows.Fields.Add(&core.RelationField{
			Name: "follower",
			CollectionId: "users",
			CascadeDelete: true,
			MaxSelect: 1,
			Required: true,
		})
		follows.Fields.Add(&core.RelationField{
			Name: "followee",
			CollectionId: "users",
			CascadeDelete: true,
			MaxSelect: 1,
			Required: true,
		})
		follows.Fields.Add(&core.SelectField{
			Name: "status",
			Values: []string{"active", "pending"},
			MaxSelect: 1,
		})
		follows.AddIndex("idx_follows_unique", true, "follower,followee", "")
		if err := app.Save(follows); err != nil {
			return err
		}

        // 6. Tag Keys
        tagKeys := core.NewBaseCollection("tag_keys")
        tagKeys.Fields.Add(&core.RelationField{
            Name: "user",
            CollectionId: "users",
            CascadeDelete: true,
            MaxSelect: 1,
            Required: true,
        })
        tagKeys.Fields.Add(&core.TextField{Name: "name", Required: true})
        tagKeys.Fields.Add(&core.TextField{Name: "slug", Required: true})
        tagKeys.Fields.Add(&core.SelectField{
            Name: "mode",
            Values: []string{"select_one", "select_multiple"},
            MaxSelect: 1,
        })
        tagKeys.AddIndex("idx_tag_keys_user_slug", true, "user,slug", "")
        if err := app.Save(tagKeys); err != nil {
            return err
        }

        // 7. Tag Values
        tagValues := core.NewBaseCollection("tag_values")
        tagValues.Fields.Add(&core.RelationField{
            Name: "tag_key",
            CollectionId: "tag_keys",
            CascadeDelete: true,
            MaxSelect: 1,
            Required: true,
        })
        tagValues.Fields.Add(&core.TextField{Name: "name", Required: true})
        tagValues.Fields.Add(&core.TextField{Name: "slug", Required: true})
        tagValues.AddIndex("idx_tag_values_key_slug", true, "tag_key,slug", "")
        if err := app.Save(tagValues); err != nil {
            return err
        }

        // 8. Book Tag Values
        bookTagValues := core.NewBaseCollection("book_tag_values")
        bookTagValues.Fields.Add(&core.RelationField{
            Name: "user",
            CollectionId: "users",
            CascadeDelete: true,
            MaxSelect: 1,
            Required: true,
        })
        bookTagValues.Fields.Add(&core.RelationField{
            Name: "book",
            CollectionId: "books",
            CascadeDelete: true,
            MaxSelect: 1,
            Required: true,
        })
        bookTagValues.Fields.Add(&core.RelationField{
            Name: "tag_key",
            CollectionId: "tag_keys",
            CascadeDelete: true,
            MaxSelect: 1,
            Required: true,
        })
        bookTagValues.Fields.Add(&core.RelationField{
            Name: "tag_value",
            CollectionId: "tag_values",
            CascadeDelete: true,
            MaxSelect: 1,
            Required: true,
        })
        bookTagValues.AddIndex("idx_btv_user_book_key", true, "user,book,tag_key", "")
        if err := app.Save(bookTagValues); err != nil {
            return err
        }

		return nil
	}, func(app core.App) error {
		// Downgrade logic (optional for this task, mostly ensuring Create)
		return nil
	})
}
