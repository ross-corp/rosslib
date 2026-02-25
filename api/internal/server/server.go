package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tristansaldanha/rosslib/api/internal/activity"
	"github.com/tristansaldanha/rosslib/api/internal/auth"
	"github.com/tristansaldanha/rosslib/api/internal/books"
	"github.com/tristansaldanha/rosslib/api/internal/collections"
	"github.com/tristansaldanha/rosslib/api/internal/docs"
	"github.com/tristansaldanha/rosslib/api/internal/ghosts"
	"github.com/tristansaldanha/rosslib/api/internal/imports"
	"github.com/tristansaldanha/rosslib/api/internal/links"
	"github.com/tristansaldanha/rosslib/api/internal/middleware"
	"github.com/tristansaldanha/rosslib/api/internal/olhttp"
	"github.com/tristansaldanha/rosslib/api/internal/search"
	"github.com/tristansaldanha/rosslib/api/internal/storage"
	"github.com/tristansaldanha/rosslib/api/internal/tags"
	"github.com/tristansaldanha/rosslib/api/internal/threads"
	"github.com/tristansaldanha/rosslib/api/internal/userbooks"
	"github.com/tristansaldanha/rosslib/api/internal/users"
)

func NewRouter(pool *pgxpool.Pool, jwtSecret string, store *storage.Client, searchClient *search.Client) http.Handler {
	r := gin.Default()

	docs.Register(r)

	r.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		dbStatus := "ok"
		if err := pool.Ping(ctx); err != nil {
			dbStatus = "error"
		}

		status := http.StatusOK
		if dbStatus != "ok" {
			status = http.StatusServiceUnavailable
		}

		c.JSON(status, gin.H{
			"status": "ok",
			"db":     dbStatus,
		})
	})

	authHandler := auth.NewHandler(pool, jwtSecret)
	r.POST("/auth/register", authHandler.Register)
	r.POST("/auth/login", authHandler.Login)

	secret := []byte(jwtSecret)

	olClient := olhttp.DefaultClient()

	booksHandler := books.NewHandler(pool, searchClient, olClient)
	r.GET("/books/search", booksHandler.SearchBooks)
	r.GET("/books/lookup", booksHandler.LookupBook)
	r.GET("/authors/search", booksHandler.SearchAuthors)
	r.GET("/authors/:authorKey", booksHandler.GetAuthor)
	r.GET("/genres", booksHandler.GetGenres)
	r.GET("/genres/:slug/books", booksHandler.GetGenreBooks)
	r.GET("/books/:workId", booksHandler.GetBook)
	r.GET("/books/:workId/editions", booksHandler.GetEditions)
	r.GET("/books/:workId/reviews", middleware.OptionalAuth(secret), booksHandler.GetBookReviews)

	usersHandler := users.NewHandler(pool, store)
	r.GET("/users", usersHandler.SearchUsers)
	r.GET("/users/:username", middleware.OptionalAuth(secret), usersHandler.GetProfile)
	r.GET("/users/:username/reviews", middleware.OptionalAuth(secret), usersHandler.GetUserReviews)

	collectionsHandler := collections.NewHandler(pool)
	r.GET("/users/:username/shelves", middleware.OptionalAuth(secret), collectionsHandler.GetUserShelves)
	r.GET("/users/:username/shelves/:slug", middleware.OptionalAuth(secret), collectionsHandler.GetShelfBySlug)
	r.GET("/users/:username/tags/*path", middleware.OptionalAuth(secret), collectionsHandler.GetTagBooks)

	userbooksHandler := userbooks.NewHandler(pool, searchClient)
	r.GET("/users/:username/books", middleware.OptionalAuth(secret), userbooksHandler.GetUserBooks)

	activityHandler := activity.NewHandler(pool)
	r.GET("/users/:username/activity", middleware.OptionalAuth(secret), activityHandler.GetUserActivity)

	threadsHandler := threads.NewHandler(pool)
	r.GET("/books/:workId/threads", threadsHandler.ListThreads)
	r.GET("/threads/:threadId", threadsHandler.GetThread)

	linksHandler := links.NewHandler(pool)
	r.GET("/books/:workId/links", middleware.OptionalAuth(secret), linksHandler.ListLinks)

	authed := r.Group("/")
	authed.Use(middleware.Auth(secret))
	authed.GET("/me/feed", activityHandler.GetFeed)
	authed.PATCH("/users/me", usersHandler.UpdateMe)
	authed.POST("/me/avatar", usersHandler.UploadAvatar)
	authed.POST("/users/:username/follow", usersHandler.Follow)
	authed.DELETE("/users/:username/follow", usersHandler.Unfollow)
	authed.GET("/me/follow-requests", usersHandler.GetFollowRequests)
	authed.POST("/me/follow-requests/:userId/accept", usersHandler.AcceptFollowRequest)
	authed.DELETE("/me/follow-requests/:userId/reject", usersHandler.RejectFollowRequest)
	importsHandler := imports.NewHandler(pool, searchClient, olClient)
	authed.POST("/me/import/goodreads/preview", importsHandler.Preview)
	authed.POST("/me/import/goodreads/commit", importsHandler.Commit)

	authed.GET("/me/export/csv", collectionsHandler.ExportCSV)
	authed.GET("/me/shelves", collectionsHandler.GetMyShelves)
	authed.POST("/me/shelves", collectionsHandler.CreateShelf)
	authed.PATCH("/me/shelves/:id", collectionsHandler.UpdateShelf)
	authed.DELETE("/me/shelves/:id", collectionsHandler.DeleteShelf)
	authed.POST("/shelves/:shelfId/books", collectionsHandler.AddBookToShelf)
	authed.PATCH("/shelves/:shelfId/books/:olId", collectionsHandler.UpdateBookInShelf)
	authed.DELETE("/shelves/:shelfId/books/:olId", collectionsHandler.RemoveBookFromShelf)

	tagsHandler := tags.NewHandler(pool)
	r.GET("/users/:username/tag-keys", middleware.OptionalAuth(secret), tagsHandler.GetUserTagKeys)
	r.GET("/users/:username/labels/:keySlug/*valuePath", middleware.OptionalAuth(secret), tagsHandler.GetLabelBooks)
	authed.GET("/me/tag-keys", tagsHandler.ListTagKeys)
	authed.POST("/me/tag-keys", tagsHandler.CreateTagKey)
	authed.DELETE("/me/tag-keys/:keyId", tagsHandler.DeleteTagKey)
	authed.POST("/me/tag-keys/:keyId/values", tagsHandler.CreateTagValue)
	authed.DELETE("/me/tag-keys/:keyId/values/:valueId", tagsHandler.DeleteTagValue)
	authed.POST("/me/books", userbooksHandler.AddBook)
	authed.GET("/me/books/status-map", userbooksHandler.GetStatusMap)
	authed.GET("/me/books/:olId/status", userbooksHandler.GetMyBookStatus)
	authed.PATCH("/me/books/:olId", userbooksHandler.UpdateBook)
	authed.DELETE("/me/books/:olId", userbooksHandler.RemoveBook)
	authed.GET("/me/books/:olId/tags", tagsHandler.GetBookTags)
	authed.PUT("/me/books/:olId/tags/:keyId", tagsHandler.SetBookTag)
	authed.DELETE("/me/books/:olId/tags/:keyId", tagsHandler.UnsetBookTag)
	authed.DELETE("/me/books/:olId/tags/:keyId/values/:valueId", tagsHandler.UnsetBookTagValue)

	authed.POST("/books/:workId/threads", threadsHandler.CreateThread)
	authed.DELETE("/threads/:threadId", threadsHandler.DeleteThread)
	authed.POST("/threads/:threadId/comments", threadsHandler.CreateComment)
	authed.DELETE("/threads/:threadId/comments/:commentId", threadsHandler.DeleteComment)

	authed.POST("/books/:workId/links", linksHandler.CreateLink)
	authed.DELETE("/links/:linkId", linksHandler.DeleteLink)
	authed.POST("/links/:linkId/vote", linksHandler.Vote)
	authed.DELETE("/links/:linkId/vote", linksHandler.Unvote)

	admin := authed.Group("/admin")
	admin.Use(middleware.RequireModerator())
	admin.GET("/users", usersHandler.ListAllUsers)
	admin.PUT("/users/:userId/moderator", usersHandler.SetModerator)
	ghostsHandler := ghosts.NewHandler(pool)
	admin.POST("/ghosts/seed", ghostsHandler.Seed)
	admin.POST("/ghosts/simulate", ghostsHandler.Simulate)
	admin.GET("/ghosts/status", ghostsHandler.Status)

	return r
}
