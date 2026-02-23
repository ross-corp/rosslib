package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tristansaldanha/rosslib/api/internal/auth"
	"github.com/tristansaldanha/rosslib/api/internal/books"
	"github.com/tristansaldanha/rosslib/api/internal/middleware"
	"github.com/tristansaldanha/rosslib/api/internal/users"
)

func NewRouter(pool *pgxpool.Pool, jwtSecret string) http.Handler {
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

	booksHandler := books.NewHandler()
	r.GET("/books/search", booksHandler.SearchBooks)

	usersHandler := users.NewHandler(pool)
	r.GET("/users", usersHandler.SearchUsers)
	r.GET("/users/:username", middleware.OptionalAuth(secret), usersHandler.GetProfile)

	authed := r.Group("/")
	authed.Use(middleware.Auth(secret))
	authed.PATCH("/users/me", usersHandler.UpdateMe)
	authed.POST("/users/:username/follow", usersHandler.Follow)
	authed.DELETE("/users/:username/follow", usersHandler.Unfollow)

	return r
}
