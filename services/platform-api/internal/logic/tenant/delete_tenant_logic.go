// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package tenant

import (
	"context"
	"errors"

	"iot-zero/services/platform-api/internal/authz"
	"iot-zero/services/platform-api/internal/session"
	"iot-zero/services/platform-api/internal/svc"
	"iot-zero/services/platform-api/internal/types"
	"iot-zero/services/platform-api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTenantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTenantLogic {
	return &DeleteTenantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteTenantLogic) DeleteTenant(req *types.TenantIdPathReq) error {
	if err := authz.AuthorizeSuperAdmin(l.ctx); err != nil {
		return err
	}

	err := l.svcCtx.TenantsModel.Delete(l.ctx, req.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return errTenantNotFound
		}
		return err
	}

	role, err := l.svcCtx.RoleModel.FindOneByRoleTypeAndTenantId(l.ctx, session.RoleTypeTenantAdmin, req.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil
		}
		return err
	}
	if err = l.svcCtx.RoleModel.Delete(l.ctx, role.Id); err != nil && !errors.Is(err, model.ErrNotFound) {
		return err
	}
	return nil
}
