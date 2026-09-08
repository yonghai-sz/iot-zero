// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package tenant

import (
	"context"

	"iot-zero/services/platform-api/internal/authz"
	"iot-zero/services/platform-api/internal/svc"
	"iot-zero/services/platform-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTenantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTenantLogic {
	return &ListTenantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTenantLogic) ListTenant(req *types.ListTenantReq) (resp *types.ListTenantResp, err error) {
	if err = authz.AuthorizeSuperAdmin(l.ctx); err != nil {
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

	total, list, err := l.svcCtx.TenantsModel.FindPage(l.ctx, req.TenantName, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]types.TenantInfo, 0, len(list))
	for _, item := range list {
		items = append(items, toTenantInfo(item))
	}

	return &types.ListTenantResp{
		Total: total,
		List:  items,
	}, nil
}
