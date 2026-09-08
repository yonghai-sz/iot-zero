package authz

import (
	"context"
	"errors"

	"iot-zero/services/platform-api/internal/session"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)

type NotFoundError struct {
	Resource string
}

func (e *NotFoundError) Error() string {
	if e.Resource == "" {
		return "not found"
	}
	return e.Resource + " not found"
}

func NotFound(resource string) error {
	return &NotFoundError{Resource: resource}
}

func AuthorizeSuperAdmin(ctx context.Context) error {
	roleType, ok := session.RoleTypeFromContext(ctx)
	if !ok {
		return ErrUnauthorized
	}
	if roleType != session.RoleTypeSuperAdmin {
		return ErrForbidden
	}
	return nil
}

func AuthorizeAdmin(ctx context.Context) error {
	roleType, ok := session.RoleTypeFromContext(ctx)
	if !ok {
		return ErrUnauthorized
	}
	switch roleType {
	case session.RoleTypeSuperAdmin, session.RoleTypeTenantAdmin:
		return nil
	default:
		return ErrForbidden
	}
}

func AuthorizeAdminForTenant(ctx context.Context, tenantId uint64) error {
	if err := AuthorizeAdmin(ctx); err != nil {
		return err
	}
	roleType, _ := session.RoleTypeFromContext(ctx)
	if roleType != session.RoleTypeTenantAdmin {
		return nil
	}
	operatorTenantId, ok := session.TenantIdFromContext(ctx)
	if !ok {
		return ErrUnauthorized
	}
	if tenantId != operatorTenantId {
		return ErrForbidden
	}
	return nil
}
