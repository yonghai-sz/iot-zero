// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package tenant

import (
	"context"
	"errors"
	"strings"

	"iot-zero/services/platform-api/internal/authz"
	"iot-zero/services/platform-api/internal/session"
	"iot-zero/services/platform-api/internal/svc"
	"iot-zero/services/platform-api/internal/types"
	"iot-zero/services/platform-api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTenantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTenantLogic {
	return &CreateTenantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTenantLogic) CreateTenant(req *types.CreateTenantReq) (resp *types.CreateTenantResp, err error) {
	if err = authz.AuthorizeSuperAdmin(l.ctx); err != nil {
		return nil, err
	}

	name := strings.TrimSpace(req.TenantName)
	if name == "" {
		return nil, errors.New("tenantName is required")
	}

	_, err = l.svcCtx.TenantsModel.FindOneByTenantName(l.ctx, name)
	if err == nil {
		return nil, errors.New("tenantName already exists")
	}
	if !errors.Is(err, model.ErrNotFound) {
		return nil, err
	}

	result, err := l.svcCtx.TenantsModel.Insert(l.ctx, &model.Tenants{
		TenantName: name,
		Email:      emailToNullString(strings.TrimSpace(req.Email)),
	})
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	_, err = l.svcCtx.RoleModel.Insert(l.ctx, &model.Role{
		RoleType: session.RoleTypeTenantAdmin,
		RoleName: "tenant administrator",
		Enable:   "Enable",
		TenantId: uint64(id),
	})
	if err != nil {
		return nil, err
	}

	return &types.CreateTenantResp{Id: uint64(id)}, nil
}
