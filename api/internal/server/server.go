package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tristansaldanha/rosslib/api/internal/auth"
	"github.com/tristansaldanha/rosslib/api/internal/books"
	"github.com/tristansaldanha/rosslib/api/internal/collections"
	"github.com/tristansaldanha/rosslib/api/internal/imports"
	"github.com/tristansaldanha/rosslib/api/internal/middleware"
	"github.com/tristansaldanha/rosslib/api/internal/storage"
	"github.com/tristansaldanha/rosslib/api/internal/tags"
	"github.com/tristansaldanha/rosslib/api/internal/users"
)

func NewRouter(pool *pgxpool.Pool, jwtSecret string, store *storage.Client) http.Handler {
	r := gin.Default()

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

	booksHandler := books.NewHandler(pool)
	r.GET("/books/search", booksHandler.SearchBooks)
	r.GET("/books/lookup", booksHandler.LookupBook)
	r.GET("/books/:workId", booksHandler.GetBook)
	r.GET("/books/:workId/reviews", middleware.OptionalAuth(secret), booksHandler.GetBookReviews)

	usersHandler := users.NewHandler(pool, store)
	r.GET("/users", usersHandler.SearchUsers)
	r.GET("/users/:username", middleware.OptionalAuth(secret), usersHandler.GetProfile)
	r.GET("/users/:username/reviews", usersHandler.GetUserReviews)

	collectionsHandler := collections.NewHandler(pool)
	r.GET("/users/:username/shelves", collectionsHandler.GetUserShelves)
	r.GET("/users/:username/shelves/:slug", collectionsHandler.GetShelfBySlug)
	r.GET("/users/:username/tags/*path", collectionsHandler.GetTagBooks)

	authed := r.Group("/")
	authed.Use(middleware.Auth(secret))
	authed.PATCH("/users/me", usersHandler.UpdateMe)
	authed.POST("/me/avatar", usersHandler.UploadAvatar)
	authed.POST("/users/:username/follow", usersHandler.Follow)
	authed.DELETE("/users/:username/follow", usersHandler.Unfollow)
	importsHandler := imports.NewHandler(pool)
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
	r.GET("/users/:username/labels/:keySlug/*valuePath", tagsHandler.GetLabelBooks)
	authed.GET("/me/tag-keys", tagsHandler.ListTagKeys)
	authed.POST("/me/tag-keys", tagsHandler.CreateTagKey)
	authed.DELETE("/me/tag-keys/:keyId", tagsHandler.DeleteTagKey)
	authed.POST("/me/tag-keys/:keyId/values", tagsHandler.CreateTagValue)
	authed.DELETE("/me/tag-keys/:keyId/values/:valueId", tagsHandler.DeleteTagValue)
	authed.GET("/me/books/:olId/status", booksHandler.GetMyBookStatus)
	authed.GET("/me/books/:olId/tags", tagsHandler.GetBookTags)
	authed.PUT("/me/books/:olId/tags/:keyId", tagsHandler.SetBookTag)
	authed.DELETE("/me/books/:olId/tags/:keyId", tagsHandler.UnsetBookTag)
	authed.DELETE("/me/books/:olId/tags/:keyId/values/:valueId", tagsHandler.UnsetBookTagValue)

	return r
}
