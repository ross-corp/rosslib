package main

import (
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	"github.com/tristansaldanha/rosslib/api/handlers"
	_ "github.com/tristansaldanha/rosslib/api/migrations"
)

func main() {
	app := pocketbase.New()

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: true,
	})

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// ── Auth (public) ────────────────────────────────────────
		se.Router.POST("/auth/login", handlers.Login(app))
		se.Router.POST("/auth/register", handlers.Register(app))
		se.Router.POST("/auth/google", handlers.GoogleAuth(app))

		// ── Books (public) ───────────────────────────────────────
		se.Router.GET("/books/search", handlers.SearchBooks(app))
		se.Router.GET("/books/popular", handlers.GetPopularBooks(app))
		se.Router.GET("/books/lookup", handlers.LookupBook(app))
		se.Router.GET("/books/{workId}", handlers.GetBookDetail(app))
		se.Router.GET("/books/{workId}/editions", handlers.GetBookEditions(app))
		se.Router.GET("/books/{workId}/stats", handlers.GetBookStats(app))
		se.Router.GET("/books/{workId}/genre-ratings", handlers.GetBookGenreRatings(app))

		// ── Books (optional auth) ────────────────────────────────
		se.Router.GET("/books/{workId}/reviews", handlers.GetBookReviews(app)).BindFunc(handlers.OptionalAuthFunc(app))
		se.Router.GET("/books/{workId}/links", handlers.GetBookLinks(app)).BindFunc(handlers.OptionalAuthFunc(app))
		se.Router.GET("/books/{workId}/threads", handlers.GetBookThreads(app))

		// ── Authors (public) ─────────────────────────────────────
		se.Router.GET("/authors/search", handlers.SearchAuthors(app))
		se.Router.GET("/authors/{authorKey}", handlers.GetAuthorDetail(app))

		// ── Users (public / optional auth) ───────────────────────
		se.Router.GET("/users", handlers.SearchUsers(app))
		se.Router.GET("/users/{username}", handlers.GetProfile(app)).BindFunc(handlers.OptionalAuthFunc(app))
		se.Router.GET("/users/{username}/reviews", handlers.GetUserReviews(app)).BindFunc(handlers.OptionalAuthFunc(app))
		se.Router.GET("/users/{username}/books", handlers.GetUserBooks(app)).BindFunc(handlers.OptionalAuthFunc(app))
		se.Router.GET("/users/{username}/shelves", handlers.GetUserShelves(app)).BindFunc(handlers.OptionalAuthFunc(app))
		se.Router.GET("/users/{username}/shelves/{slug}", handlers.GetShelfDetail(app)).BindFunc(handlers.OptionalAuthFunc(app))
		se.Router.GET("/users/{username}/tag-keys", handlers.GetUserTagKeys(app)).BindFunc(handlers.OptionalAuthFunc(app))
		se.Router.GET("/users/{username}/tags/{path...}", handlers.GetUserTagBooks(app)).BindFunc(handlers.OptionalAuthFunc(app))
		se.Router.GET("/users/{username}/labels/{keySlug}/{valuePath...}", handlers.GetUserLabelBooks(app)).BindFunc(handlers.OptionalAuthFunc(app))
		se.Router.GET("/users/{username}/activity", handlers.GetUserActivity(app)).BindFunc(handlers.OptionalAuthFunc(app))

		// ── Threads (public GET) ─────────────────────────────────
		se.Router.GET("/threads/{threadId}", handlers.GetThread(app))

		// ── Authenticated routes ─────────────────────────────────
		authed := se.Router.Group("").Bind(apis.RequireAuth())

		// Account
		authed.GET("/me/account", handlers.GetAccount(app))
		authed.PUT("/me/password", handlers.ChangePassword(app))
		authed.DELETE("/me/account/data", handlers.DeleteAllData(app))

		// Profile
		authed.PATCH("/users/me", handlers.UpdateProfile(app))
		authed.POST("/me/avatar", handlers.UploadAvatar(app))

		// Feed
		authed.GET("/me/feed", handlers.GetFeed(app))

		// User books
		authed.POST("/me/books", handlers.AddBook(app))
		authed.PATCH("/me/books/{olId}", handlers.UpdateBook(app))
		authed.DELETE("/me/books/{olId}", handlers.DeleteBook(app))
		authed.GET("/me/books/{olId}/status", handlers.GetBookStatus(app))
		authed.PUT("/me/books/{olId}/status", handlers.SetBookStatus(app))
		authed.GET("/me/books/status-map", handlers.GetStatusMap(app))

		// Tags
		authed.GET("/me/tag-keys", handlers.GetTagKeys(app))
		authed.POST("/me/tag-keys", handlers.CreateTagKey(app))
		authed.DELETE("/me/tag-keys/{keyId}", handlers.DeleteTagKey(app))
		authed.POST("/me/tag-keys/{keyId}/values", handlers.CreateTagValue(app))
		authed.DELETE("/me/tag-keys/{keyId}/values/{valueId}", handlers.DeleteTagValue(app))
		authed.GET("/me/books/{olId}/tags", handlers.GetBookTags(app))
		authed.PUT("/me/books/{olId}/tags/{keyId}", handlers.SetBookTag(app))
		authed.DELETE("/me/books/{olId}/tags/{keyId}", handlers.UnsetBookTag(app))
		authed.DELETE("/me/books/{olId}/tags/{keyId}/values/{valueId}", handlers.UnsetBookTagValue(app))

		// Shelves
		authed.GET("/me/shelves", handlers.GetMyShelves(app))
		authed.POST("/me/shelves", handlers.CreateShelf(app))
		authed.PATCH("/me/shelves/{id}", handlers.UpdateShelf(app))
		authed.DELETE("/me/shelves/{id}", handlers.DeleteShelf(app))
		authed.POST("/shelves/{shelfId}/books", handlers.AddBookToShelf(app))
		authed.PATCH("/shelves/{shelfId}/books/{olId}", handlers.UpdateShelfBook(app))
		authed.DELETE("/shelves/{shelfId}/books/{olId}", handlers.RemoveBookFromShelf(app))

		// Export
		authed.GET("/me/export/csv", handlers.ExportCSV(app))

		// Follow
		authed.POST("/users/{username}/follow", handlers.FollowUser(app))
		authed.DELETE("/users/{username}/follow", handlers.UnfollowUser(app))
		authed.GET("/me/follow-requests", handlers.GetFollowRequests(app))
		authed.POST("/me/follow-requests/{userId}/accept", handlers.AcceptFollowRequest(app))
		authed.DELETE("/me/follow-requests/{userId}/reject", handlers.RejectFollowRequest(app))

		// Threads (auth required for mutations)
		authed.POST("/books/{workId}/threads", handlers.CreateThread(app))
		authed.DELETE("/threads/{threadId}", handlers.DeleteThread(app))
		authed.POST("/threads/{threadId}/comments", handlers.AddComment(app))
		authed.DELETE("/threads/{threadId}/comments/{commentId}", handlers.DeleteComment(app))

		// Book scan
		authed.POST("/books/scan", handlers.ScanBook(app))

		// Notifications
		authed.GET("/me/notifications", handlers.GetNotifications(app))
		authed.GET("/me/notifications/unread-count", handlers.GetUnreadCount(app))
		authed.POST("/me/notifications/{notifId}/read", handlers.MarkNotificationRead(app))
		authed.POST("/me/notifications/read-all", handlers.MarkAllRead(app))

		// Notification preferences
		authed.GET("/me/notification-preferences", handlers.GetNotificationPreferences(app))
		authed.PUT("/me/notification-preferences", handlers.UpdateNotificationPreferences(app))

		// Imports
		authed.POST("/me/import/goodreads/preview", handlers.PreviewGoodreadsImport(app))
		authed.POST("/me/import/goodreads/commit", handlers.CommitGoodreadsImport(app))
		authed.GET("/me/imports/pending", handlers.GetPendingImports(app))
		authed.PATCH("/me/imports/pending/{id}", handlers.ResolvePendingImport(app))
		authed.DELETE("/me/imports/pending/{id}", handlers.DeletePendingImport(app))

		// Genre ratings
		authed.GET("/me/books/{olId}/genre-ratings", handlers.GetMyGenreRatings(app))
		authed.PUT("/me/books/{olId}/genre-ratings", handlers.SetGenreRatings(app))

		// Book links
		authed.POST("/books/{workId}/links", handlers.CreateBookLink(app))
		authed.DELETE("/links/{linkId}", handlers.DeleteBookLink(app))
		authed.POST("/links/{linkId}/vote", handlers.VoteLink(app))
		authed.DELETE("/links/{linkId}/vote", handlers.UnvoteLink(app))
		authed.POST("/links/{linkId}/edits", handlers.ProposeLinkEdit(app))

		// Author/book follows
		authed.POST("/authors/{authorKey}/follow", handlers.FollowAuthor(app))
		authed.DELETE("/authors/{authorKey}/follow", handlers.UnfollowAuthor(app))
		authed.GET("/me/followed-authors", handlers.GetFollowedAuthors(app))
		authed.POST("/books/{workId}/follow", handlers.FollowBook(app))
		authed.DELETE("/books/{workId}/follow", handlers.UnfollowBook(app))
		authed.GET("/me/followed-books", handlers.GetFollowedBooks(app))

		// Feedback
		authed.POST("/feedback", handlers.CreateFeedback(app))
		authed.GET("/me/feedback", handlers.GetMyFeedback(app))
		authed.DELETE("/me/feedback/{feedbackId}", handlers.DeleteMyFeedback(app))

		// ── Admin routes ─────────────────────────────────────────
		admin := se.Router.Group("/admin").Bind(apis.RequireAuth()).BindFunc(handlers.RequireModerator(app))
		admin.GET("/feedback", handlers.GetFeedback(app))
		admin.PATCH("/feedback/{feedbackId}", handlers.UpdateFeedbackStatus(app))
		admin.DELETE("/feedback/{feedbackId}", handlers.DeleteFeedback(app))
		admin.POST("/ghosts/seed", handlers.SeedGhosts(app))
		admin.POST("/ghosts/simulate", handlers.SimulateGhosts(app))
		admin.GET("/ghosts/status", handlers.GetGhostStatus(app))
		admin.GET("/link-edits", handlers.GetPendingLinkEdits(app))
		admin.PUT("/link-edits/{editId}", handlers.ReviewLinkEdit(app))

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
