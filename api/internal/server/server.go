package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tristansaldanha/rosslib/api/internal/auth"
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

	return r
}
