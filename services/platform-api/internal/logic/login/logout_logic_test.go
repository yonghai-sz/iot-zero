package login

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/stores/redis"

	"iot-zero/services/platform-api/internal/authz"
	"iot-zero/services/platform-api/internal/session"
	"iot-zero/services/platform-api/internal/svc"
)

func TestLogoutLogic_Logout(t *testing.T) {
	store := session.NewStore(redis.New(miniredis.RunT(t).Addr()))
	ctx := context.WithValue(context.Background(), "username", "alice")
	assert.NoError(t, store.Save(ctx, "alice", "token-1", 3600))

	l := NewLogoutLogic(ctx, &svc.ServiceContext{SessionStore: store})
	assert.NoError(t, l.Logout())

	matched, err := store.Match(ctx, "alice", "token-1")
	assert.NoError(t, err)
	assert.False(t, matched)
}

func TestLogoutLogic_LogoutUnauthorized(t *testing.T) {
	l := NewLogoutLogic(context.Background(), &svc.ServiceContext{})
	assert.ErrorIs(t, l.Logout(), authz.ErrUnauthorized)
}
