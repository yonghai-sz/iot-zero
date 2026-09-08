package session

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTenantIdFromContext(t *testing.T) {
	t.Run("missing", func(t *testing.T) {
		id, ok := TenantIdFromContext(context.Background())
		assert.False(t, ok)
		assert.Zero(t, id)
	})

	t.Run("float64 from jwt", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "tenantId", float64(7))
		id, ok := TenantIdFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, uint64(7), id)
	})

	t.Run("uint64", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "tenantId", uint64(3))
		id, ok := TenantIdFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, uint64(3), id)
	})

	t.Run("json number", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "tenantId", json.Number("9"))
		id, ok := TenantIdFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, uint64(9), id)
	})

	t.Run("zero is present", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "tenantId", float64(0))
		id, ok := TenantIdFromContext(ctx)
		assert.True(t, ok)
		assert.Zero(t, id)
	})
}
