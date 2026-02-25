package testutil

import (
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tristansaldanha/rosslib/api/internal/db"
)

// GetDB connects to the database specified by DATABASE_URL environment variable.
// If DATABASE_URL is not set, it skips the test.
func GetDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("skipping integration test: DATABASE_URL not set")
	}

	pool, err := db.Connect(dsn)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.Migrate(pool); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	return pool
}
