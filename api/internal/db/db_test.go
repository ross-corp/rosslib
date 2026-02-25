package db_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tristansaldanha/rosslib/api/internal/testutil"
)

func TestConnect(t *testing.T) {
	pool := testutil.GetDB(t)
	assert.NotNil(t, pool)

	var result int
	err := pool.QueryRow(context.Background(), "SELECT 1").Scan(&result)
	assert.NoError(t, err)
	assert.Equal(t, 1, result)
}
