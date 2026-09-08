// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package role

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

type CreateRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRoleLogic {
	return &CreateRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateRoleLogic) CreateRole(req *types.CreateRoleReq) (resp *types.CreateRoleResp, err error) {
	if err = authz.AuthorizeAdminForTenant(l.ctx, req.TenantId); err != nil {
		return nil, err
	}
	if req.TenantId == 0 {
		return nil, errors.New("tenantId is required")
	}
	name := strings.TrimSpace(req.RoleName)
	if name == "" {
		return nil, errors.New("roleName is required")
	}

	result, err := l.svcCtx.RoleModel.Insert(l.ctx, &model.Role{
		RoleType: session.RoleTypeUser,
		RoleName: name,
		Enable:   boolToEnable(req.Enable),
		TenantId: req.TenantId,
	})
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &types.CreateRoleResp{Id: uint64(id)}, nil
}
