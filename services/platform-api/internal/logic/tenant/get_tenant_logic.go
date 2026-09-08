// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package tenant

import (
	"context"
	"errors"

	"iot-zero/services/platform-api/internal/authz"
	"iot-zero/services/platform-api/internal/svc"
	"iot-zero/services/platform-api/internal/types"
	"iot-zero/services/platform-api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantLogic {
	return &GetTenantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTenantLogic) GetTenant(req *types.TenantIdPathReq) (resp *types.TenantInfo, err error) {
	if err = authz.AuthorizeAdminForTenant(l.ctx, req.Id); err != nil {
		return nil, err
	}

	entity, err := l.svcCtx.TenantsModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errTenantNotFound
		}
		return nil, err
	}

	info := toTenantInfo(entity)
	return &info, nil
}
