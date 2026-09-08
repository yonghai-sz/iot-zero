// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package role

import (
	"context"

	"iot-zero/services/platform-api/internal/svc"
	"iot-zero/services/platform-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRoleLogic {
	return &GetRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRoleLogic) GetRole(req *types.RoleIdPathReq) (resp *types.RoleInfo, err error) {
	entity, err := loadRoleForAdmin(l.ctx, l.svcCtx.RoleModel, req.Id)
	if err != nil {
		return nil, err
	}
	info := toRoleInfo(entity)
	return &info, nil
}
