// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package role

import (
	"context"
	"errors"
	"strings"

	"iot-zero/services/platform-api/internal/svc"
	"iot-zero/services/platform-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRoleLogic {
	return &UpdateRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateRoleLogic) UpdateRole(req *types.UpdateRoleReq) (resp *types.RoleInfo, err error) {
	entity, err := loadRoleForAdmin(l.ctx, l.svcCtx.RoleModel, req.Id)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(req.RoleName)
	if name == "" && req.Enable == nil {
		return nil, errors.New("roleName or enable is required")
	}

	if name != "" {
		entity.RoleName = name
	}
	if req.Enable != nil {
		entity.Enable = boolToEnable(*req.Enable)
	}

	if err = l.svcCtx.RoleModel.Update(l.ctx, entity); err != nil {
		return nil, err
	}

	updated, err := l.svcCtx.RoleModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return nil, err
	}
	info := toRoleInfo(updated)
	return &info, nil
}
