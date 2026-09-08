// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package tenant

import (
	"context"
	"errors"
	"strings"

	"iot-zero/services/platform-api/internal/authz"
	"iot-zero/services/platform-api/internal/svc"
	"iot-zero/services/platform-api/internal/types"
	"iot-zero/services/platform-api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTenantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTenantLogic {
	return &UpdateTenantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTenantLogic) UpdateTenant(req *types.UpdateTenantReq) (resp *types.TenantInfo, err error) {
	if err = authz.AuthorizeSuperAdmin(l.ctx); err != nil {
		return nil, err
	}

	entity, err := l.svcCtx.TenantsModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errTenantNotFound
		}
		return nil, err
	}

	name := strings.TrimSpace(req.TenantName)
	email := strings.TrimSpace(req.Email)
	if name == "" && email == "" {
		return nil, errors.New("tenantName or email is required")
	}

	if name != "" && name != entity.TenantName {
		existing, findErr := l.svcCtx.TenantsModel.FindOneByTenantName(l.ctx, name)
		if findErr == nil && existing.Id != entity.Id {
			return nil, errors.New("tenantName already exists")
		}
		if findErr != nil && !errors.Is(findErr, model.ErrNotFound) {
			return nil, findErr
		}
		entity.TenantName = name
	}
	if email != "" {
		entity.Email = emailToNullString(email)
	}

	if err = l.svcCtx.TenantsModel.Update(l.ctx, entity); err != nil {
		return nil, err
	}

	updated, err := l.svcCtx.TenantsModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return nil, err
	}
	info := toTenantInfo(updated)
	return &info, nil
}
