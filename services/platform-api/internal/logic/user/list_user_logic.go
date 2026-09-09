// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"

	"iot-zero/services/platform-api/internal/authz"
	"iot-zero/services/platform-api/internal/svc"
	"iot-zero/services/platform-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserLogic {
	return &ListUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListUserLogic) ListUser(req *types.ListUserReq) (resp *types.ListUserResp, err error) {
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

	total, list, err := l.svcCtx.UserModel.FindPage(l.ctx, req.TenantId, req.Username, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]types.UserInfo, 0, len(list))
	for _, item := range list {
		items = append(items, toUserInfo(item))
	}

	return &types.ListUserResp{
		Total: total,
		List:  items,
	}, nil
}
