package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// --- Look up collections created in previous migration ---
		books, err := app.FindCollectionByNameOrId("books")
		if err != nil {
			return err
		}
		collectionsCol, err := app.FindCollectionByNameOrId("collections")
		if err != nil {
			return err
		}

		// --- Update existing collections ---

		// Update users: add display_name, is_moderator, is_ghost, google_id, author_key, email_verified, avatar
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}
		users.Fields.Add(&core.TextField{Name: "display_name"})
		users.Fields.Add(&core.BoolField{Name: "is_moderator"})
		users.Fields.Add(&core.BoolField{Name: "is_ghost"})
		users.Fields.Add(&core.TextField{Name: "google_id"})
		users.Fields.Add(&core.TextField{Name: "author_key"})
		users.Fields.Add(&core.BoolField{Name: "email_verified"})
		users.Fields.Add(&core.FileField{
			Name:      "avatar",
			MaxSelect: 1,
			MaxSize:   5 * 1024 * 1024, // 5MB
			MimeTypes: []string{"image/jpeg", "image/png", "image/gif", "image/webp"},
		})
		if err := app.Save(users); err != nil {
			return err
		}

		// Fix book_tag_values index: change from (user,book,tag_key) to (user,book,tag_value)
		btv, err := app.FindCollectionByNameOrId("book_tag_values")
		if err != nil {
			return err
		}
		btv.RemoveIndex("idx_btv_user_book_key")
		btv.AddIndex("idx_btv_user_book_value", true, "user,book,tag_value", "")
		if err := app.Save(btv); err != nil {
			return err
		}

		// --- Create new collections ---

		// user_books
		userBooks := core.NewBaseCollection("user_books")
		userBooks.Fields.Add(&core.RelationField{
			Name:          "user",
			CollectionId:  users.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		userBooks.Fields.Add(&core.RelationField{
			Name:          "book",
			CollectionId:  books.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		userBooks.Fields.Add(&core.NumberField{Name: "rating"})
		userBooks.Fields.Add(&core.TextField{Name: "review_text"})
		userBooks.Fields.Add(&core.BoolField{Name: "spoiler"})
		userBooks.Fields.Add(&core.DateField{Name: "date_read"})
		userBooks.Fields.Add(&core.DateField{Name: "date_dnf"})
		userBooks.Fields.Add(&core.DateField{Name: "date_added"})
		userBooks.Fields.Add(&core.NumberField{Name: "progress_pages"})
		userBooks.Fields.Add(&core.NumberField{Name: "progress_percent"})
		userBooks.AddIndex("idx_user_books_unique", true, "user,book", "")
		if err := app.Save(userBooks); err != nil {
			return err
		}

		// threads
		threads := core.NewBaseCollection("threads")
		threads.Fields.Add(&core.RelationField{
			Name:          "book",
			CollectionId:  books.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		threads.Fields.Add(&core.RelationField{
			Name:          "user",
			CollectionId:  users.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		threads.Fields.Add(&core.TextField{Name: "title", Required: true})
		threads.Fields.Add(&core.TextField{Name: "body", Required: true})
		threads.Fields.Add(&core.BoolField{Name: "spoiler"})
		threads.Fields.Add(&core.DateField{Name: "deleted_at"})
		threads.AddIndex("idx_threads_book", false, "book", "")
		if err := app.Save(threads); err != nil {
			return err
		}

		// thread_comments
		threadComments := core.NewBaseCollection("thread_comments")
		threadComments.Fields.Add(&core.RelationField{
			Name:          "thread",
			CollectionId:  threads.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		threadComments.Fields.Add(&core.RelationField{
			Name:          "user",
			CollectionId:  users.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		threadComments.Fields.Add(&core.TextField{Name: "parent"}) // parent comment ID (nullable)
		threadComments.Fields.Add(&core.TextField{Name: "body", Required: true})
		threadComments.Fields.Add(&core.DateField{Name: "deleted_at"})
		threadComments.AddIndex("idx_thread_comments_thread", false, "thread", "")
		if err := app.Save(threadComments); err != nil {
			return err
		}

		// activities
		activities := core.NewBaseCollection("activities")
		activities.Fields.Add(&core.RelationField{
			Name:          "user",
			CollectionId:  users.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		activities.Fields.Add(&core.TextField{Name: "activity_type", Required: true})
		activities.Fields.Add(&core.RelationField{
			Name:         "book",
			CollectionId: books.Id,
			MaxSelect:    1,
		})
		activities.Fields.Add(&core.RelationField{
			Name:         "target_user",
			CollectionId: users.Id,
			MaxSelect:    1,
		})
		activities.Fields.Add(&core.RelationField{
			Name:         "collection_ref",
			CollectionId: collectionsCol.Id,
			MaxSelect:    1,
		})
		activities.Fields.Add(&core.RelationField{
			Name:         "thread",
			CollectionId: threads.Id,
			MaxSelect:    1,
		})
		activities.Fields.Add(&core.JSONField{Name: "metadata"})
		activities.AddIndex("idx_activities_user", false, "user", "")
		if err := app.Save(activities); err != nil {
			return err
		}

		// book_stats
		bookStats := core.NewBaseCollection("book_stats")
		bookStats.Fields.Add(&core.RelationField{
			Name:          "book",
			CollectionId:  books.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		bookStats.Fields.Add(&core.NumberField{Name: "reads_count"})
		bookStats.Fields.Add(&core.NumberField{Name: "want_to_read_count"})
		bookStats.Fields.Add(&core.NumberField{Name: "rating_sum"})
		bookStats.Fields.Add(&core.NumberField{Name: "rating_count"})
		bookStats.Fields.Add(&core.NumberField{Name: "review_count"})
		bookStats.AddIndex("idx_book_stats_book", true, "book", "")
		if err := app.Save(bookStats); err != nil {
			return err
		}

		// author_follows
		authorFollows := core.NewBaseCollection("author_follows")
		authorFollows.Fields.Add(&core.RelationField{
			Name:          "user",
			CollectionId:  users.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		authorFollows.Fields.Add(&core.TextField{Name: "author_key", Required: true})
		authorFollows.Fields.Add(&core.TextField{Name: "author_name"})
		authorFollows.AddIndex("idx_author_follows_unique", true, "user,author_key", "")
		if err := app.Save(authorFollows); err != nil {
			return err
		}

		// book_follows
		bookFollows := core.NewBaseCollection("book_follows")
		bookFollows.Fields.Add(&core.RelationField{
			Name:          "user",
			CollectionId:  users.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		bookFollows.Fields.Add(&core.RelationField{
			Name:          "book",
			CollectionId:  books.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		bookFollows.AddIndex("idx_book_follows_unique", true, "user,book", "")
		if err := app.Save(bookFollows); err != nil {
			return err
		}

		// notifications
		notifications := core.NewBaseCollection("notifications")
		notifications.Fields.Add(&core.RelationField{
			Name:          "user",
			CollectionId:  users.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		notifications.Fields.Add(&core.TextField{Name: "notif_type", Required: true})
		notifications.Fields.Add(&core.TextField{Name: "title"})
		notifications.Fields.Add(&core.TextField{Name: "body"})
		notifications.Fields.Add(&core.JSONField{Name: "metadata"})
		notifications.Fields.Add(&core.BoolField{Name: "read"})
		notifications.AddIndex("idx_notifications_user_read", false, "user,read", "")
		if err := app.Save(notifications); err != nil {
			return err
		}

		// genre_ratings
		genreRatings := core.NewBaseCollection("genre_ratings")
		genreRatings.Fields.Add(&core.RelationField{
			Name:          "user",
			CollectionId:  users.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		genreRatings.Fields.Add(&core.RelationField{
			Name:          "book",
			CollectionId:  books.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		genreRatings.Fields.Add(&core.TextField{Name: "genre", Required: true})
		genreRatings.Fields.Add(&core.NumberField{Name: "rating", Required: true})
		genreRatings.AddIndex("idx_genre_ratings_unique", true, "user,book,genre", "")
		if err := app.Save(genreRatings); err != nil {
			return err
		}

		// book_links
		bookLinks := core.NewBaseCollection("book_links")
		bookLinks.Fields.Add(&core.RelationField{
			Name:          "from_book",
			CollectionId:  books.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		bookLinks.Fields.Add(&core.RelationField{
			Name:          "to_book",
			CollectionId:  books.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		bookLinks.Fields.Add(&core.RelationField{
			Name:          "user",
			CollectionId:  users.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		bookLinks.Fields.Add(&core.SelectField{
			Name:      "link_type",
			Values:    []string{"sequel", "prequel", "companion", "adaptation", "related", "inspired_by"},
			MaxSelect: 1,
			Required:  true,
		})
		bookLinks.Fields.Add(&core.TextField{Name: "note"})
		bookLinks.Fields.Add(&core.DateField{Name: "deleted_at"})
		bookLinks.AddIndex("idx_book_links_unique", true, "from_book,to_book,link_type,user", "")
		if err := app.Save(bookLinks); err != nil {
			return err
		}

		// book_link_votes
		bookLinkVotes := core.NewBaseCollection("book_link_votes")
		bookLinkVotes.Fields.Add(&core.RelationField{
			Name:          "user",
			CollectionId:  users.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		bookLinkVotes.Fields.Add(&core.RelationField{
			Name:          "book_link",
			CollectionId:  bookLinks.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		bookLinkVotes.AddIndex("idx_book_link_votes_unique", true, "user,book_link", "")
		if err := app.Save(bookLinkVotes); err != nil {
			return err
		}

		// book_link_edits
		bookLinkEdits := core.NewBaseCollection("book_link_edits")
		bookLinkEdits.Fields.Add(&core.RelationField{
			Name:          "book_link",
			CollectionId:  bookLinks.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		bookLinkEdits.Fields.Add(&core.RelationField{
			Name:          "user",
			CollectionId:  users.Id,
			CascadeDelete: true,
			MaxSelect:     1,
			Required:      true,
		})
		bookLinkEdits.Fields.Add(&core.SelectField{
			Name:      "proposed_type",
			Values:    []string{"sequel", "prequel", "companion", "adaptation", "related", "inspired_by"},
			MaxSelect: 1,
		})
		bookLinkEdits.Fields.Add(&core.TextField{Name: "proposed_note"})
		bookLinkEdits.Fields.Add(&core.SelectField{
			Name:      "status",
			Values:    []string{"pending", "approved", "rejected"},
			MaxSelect: 1,
			Required:  true,
		})
		bookLinkEdits.Fields.Add(&core.RelationField{
			Name:         "reviewer",
			CollectionId: users.Id,
			MaxSelect:    1,
		})
		bookLinkEdits.AddIndex("idx_book_link_edits_status", false, "status", "")
		if err := app.Save(bookLinkEdits); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		// Downgrade: drop all new collections
		names := []string{
			"book_link_edits", "book_link_votes", "book_links",
			"genre_ratings", "notifications", "book_follows",
			"author_follows", "book_stats", "activities",
			"thread_comments", "threads", "user_books",
		}
		for _, name := range names {
			coll, err := app.FindCollectionByNameOrId(name)
			if err == nil {
				if err := app.Delete(coll); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
