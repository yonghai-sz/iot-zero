package role

import (
	"context"
	"errors"

	"iot-zero/services/platform-api/internal/authz"
	"iot-zero/services/platform-api/model"
)

func loadRoleForAdmin(ctx context.Context, roleModel model.RoleModel, id uint64) (*model.Role, error) {
	if err := authz.AuthorizeAdmin(ctx); err != nil {
		return nil, err
	}

	entity, err := roleModel.FindOne(ctx, id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errRoleNotFound
		}
		return nil, err
	}

	if err = authz.AuthorizeAdminForTenant(ctx, entity.TenantId); err != nil {
		if errors.Is(err, authz.ErrForbidden) {
			return nil, errRoleNotFound
		}
		return nil, err
	}
	return entity, nil
}
