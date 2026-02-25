package users_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/tristansaldanha/rosslib/api/internal/testutil"
	"github.com/tristansaldanha/rosslib/api/internal/users"
)

func TestSearchUsers(t *testing.T) {
	pool := testutil.GetDB(t)
	h := users.NewHandler(pool, nil)

	_, err := pool.Exec(context.Background(), `
		INSERT INTO users (id, username, display_name, email, password_hash)
		VALUES ('00000000-0000-0000-0000-000000000001', 'testuser1', 'Test User 1', 'test1@example.com', 'hash')
		ON CONFLICT (id) DO NOTHING
	`)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/users?q=testuser1", nil)

	h.SearchUsers(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var results []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &results); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	found := false
	for _, r := range results {
		if r["username"] == "testuser1" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected testuser1 in results, got %v", results)
	}
}
