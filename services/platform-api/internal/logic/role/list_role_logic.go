// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package role

import (
	"context"

	"iot-zero/services/platform-api/internal/authz"
	"iot-zero/services/platform-api/internal/svc"
	"iot-zero/services/platform-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRoleLogic {
	return &ListRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListRoleLogic) ListRole(req *types.ListRoleReq) (resp *types.ListRoleResp, err error) {
	if err = authz.AuthorizeAdminForTenant(l.ctx, req.TenantId); err != nil {
		return nil, err
	}

	pageIndex := req.PageIndex
	if pageIndex <= 0 {
		pageIndex = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	total, list, err := l.svcCtx.RoleModel.FindPage(l.ctx, req.TenantId, req.RoleName, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]types.RoleInfo, 0, len(list))
	for _, item := range list {
		items = append(items, toRoleInfo(item))
	}

	return &types.ListRoleResp{
		Total: total,
		List:  items,
	}, nil
}
