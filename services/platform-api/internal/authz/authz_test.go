package authz

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"iot-zero/services/platform-api/internal/session"
)

func operatorCtx(roleType string, tenantId uint64) context.Context {
	ctx := context.WithValue(context.Background(), "roleType", roleType)
	return context.WithValue(ctx, "tenantId", float64(tenantId))
}

func TestAuthorizeSuperAdmin(t *testing.T) {
	assert.ErrorIs(t, AuthorizeSuperAdmin(context.Background()), ErrUnauthorized)
	assert.ErrorIs(t, AuthorizeSuperAdmin(operatorCtx(session.RoleTypeUser, 7)), ErrForbidden)
	assert.ErrorIs(t, AuthorizeSuperAdmin(operatorCtx(session.RoleTypeTenantAdmin, 7)), ErrForbidden)
	assert.NoError(t, AuthorizeSuperAdmin(operatorCtx(session.RoleTypeSuperAdmin, 1)))
}

func TestAuthorizeAdminForTenant(t *testing.T) {
	assert.ErrorIs(t, AuthorizeAdminForTenant(context.Background(), 7), ErrUnauthorized)
	assert.ErrorIs(t, AuthorizeAdminForTenant(operatorCtx(session.RoleTypeUser, 7), 7), ErrForbidden)
	assert.ErrorIs(t, AuthorizeAdminForTenant(operatorCtx(session.RoleTypeTenantAdmin, 3), 7), ErrForbidden)
	assert.NoError(t, AuthorizeAdminForTenant(operatorCtx(session.RoleTypeTenantAdmin, 7), 7))
	assert.NoError(t, AuthorizeAdminForTenant(operatorCtx(session.RoleTypeSuperAdmin, 1), 7))
}
