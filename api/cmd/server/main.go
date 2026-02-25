package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tristansaldanha/rosslib/api/internal/config"
	"github.com/tristansaldanha/rosslib/api/internal/db"
	"github.com/tristansaldanha/rosslib/api/internal/notifications"
	"github.com/tristansaldanha/rosslib/api/internal/olhttp"
	"github.com/tristansaldanha/rosslib/api/internal/search"
	"github.com/tristansaldanha/rosslib/api/internal/server"
	"github.com/tristansaldanha/rosslib/api/internal/storage"
)

func main() {
	cfg := config.Load()

	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := db.Migrate(pool); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	store, err := storage.NewMinIOClient(
		cfg.MinIOEndpoint,
		cfg.MinIOAccessKey,
		cfg.MinIOSecretKey,
		cfg.MinIOBucket,
		cfg.MinIOPublicURL,
		false,
	)
	if err != nil {
		log.Fatalf("failed to create storage client: %v", err)
	}
	if err := store.EnsureBucket(context.Background()); err != nil {
		log.Printf("warning: could not configure storage bucket: %v", err)
	}

	searchClient, err := search.NewClient(cfg.MeiliURL, cfg.MeiliMasterKey)
	if err != nil {
		log.Printf("warning: meilisearch unavailable: %v", err)
	}
	if searchClient != nil {
		go func() {
			if err := searchClient.SyncBooks(context.Background(), pool); err != nil {
				log.Printf("warning: meilisearch sync failed: %v", err)
			}
		}()
	}

	router := server.NewRouter(pool, cfg.JWTSecret, store, searchClient)

	// Start background poller for new publications by followed authors.
	pollerCtx, pollerCancel := context.WithCancel(context.Background())
	defer pollerCancel()
	notifications.StartPoller(pollerCtx, pool, olhttp.DefaultClient())

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("server listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	pollerCancel()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("server exited")
}
