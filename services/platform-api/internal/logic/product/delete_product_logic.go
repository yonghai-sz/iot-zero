// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package product

import (
	"context"
	"errors"

	"iot-zero/services/platform-api/internal/authz"
	"iot-zero/services/platform-api/internal/svc"
	"iot-zero/services/platform-api/internal/types"
	"iot-zero/services/platform-api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteProductLogic {
	return &DeleteProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteProductLogic) DeleteProduct(req *types.ProductIdPathReq) error {
	if err := authz.AuthorizeSuperAdmin(l.ctx); err != nil {
		return err
	}

	err := l.svcCtx.ProductModel.Delete(l.ctx, req.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return authz.NotFound("product")
		}
		return err
	}
	return nil
}
