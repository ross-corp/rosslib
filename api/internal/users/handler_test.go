package users_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/tristansaldanha/rosslib/api/internal/testutil"
	"github.com/tristansaldanha/rosslib/api/internal/users"
)

func TestSearchUsers(t *testing.T) {
	pool := testutil.GetDB(t)
	h := users.NewHandler(pool, nil)

	// Setup data
	// Using ON CONFLICT to avoid errors if run multiple times on same DB
	_, err := pool.Exec(context.Background(), `
		INSERT INTO users (id, username, display_name, email, password_hash)
		VALUES ('00000000-0000-0000-0000-000000000001', 'testuser1', 'Test User 1', 'test1@example.com', 'hash')
		ON CONFLICT (id) DO NOTHING
	`)
	assert.NoError(t, err)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/users?q=testuser1", nil)

	h.SearchUsers(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var results []map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &results)
	assert.NoError(t, err)

	assert.NotEmpty(t, results)
	found := false
	for _, r := range results {
		if r["username"] == "testuser1" {
			found = true
			break
		}
	}
	assert.True(t, found)
}
