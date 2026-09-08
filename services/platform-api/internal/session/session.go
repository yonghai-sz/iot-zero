package session

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	keyPrefix           = "login:"
	RoleTypeSuperAdmin  = "SuperAdmin"
	RoleTypeTenantAdmin = "TenantAdmin"
	RoleTypeUser        = "User"
)

type Store struct {
	rds *redis.Redis
}

func NewStore(rds *redis.Redis) *Store {
	return &Store{rds: rds}
}

func key(username string) string {
	return keyPrefix + username
}

func (s *Store) Save(ctx context.Context, username, token string, expireSeconds int) error {
	return s.rds.SetexCtx(ctx, key(username), token, expireSeconds)
}

func (s *Store) Match(ctx context.Context, username, token string) (bool, error) {
	stored, err := s.rds.GetCtx(ctx, key(username))
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return stored != "" && stored == token, nil
}

func (s *Store) Delete(ctx context.Context, username string) error {
	_, err := s.rds.DelCtx(ctx, key(username))
	return err
}

func UsernameFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value("username")
	s, ok := v.(string)
	return s, ok && s != ""
}

func RoleTypeFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value("roleType")
	s, ok := v.(string)
	return s, ok && s != ""
}

func TenantIdFromContext(ctx context.Context) (uint64, bool) {
	return uint64FromContext(ctx, "tenantId")
}

func uint64FromContext(ctx context.Context, key string) (uint64, bool) {
	switch v := ctx.Value(key).(type) {
	case uint64:
		return v, true
	case float64:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case json.Number:
		n, err := v.Int64()
		if err != nil || n < 0 {
			return 0, false
		}
		return uint64(n), true
	default:
		return 0, false
	}
}

func BearerToken(r *http.Request) string {
	value := r.Header.Get("Authorization")
	if len(value) > 7 && strings.EqualFold(value[:7], "Bearer ") {
		return value[7:]
	}
	return value
}
