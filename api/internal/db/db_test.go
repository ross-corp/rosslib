package db_test

import (
	"context"
	"testing"

	"github.com/tristansaldanha/rosslib/api/internal/testutil"
)

func TestConnect(t *testing.T) {
	pool := testutil.GetDB(t)

	var result int
	err := pool.QueryRow(context.Background(), "SELECT 1").Scan(&result)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if result != 1 {
		t.Errorf("expected 1, got %d", result)
	}
}
