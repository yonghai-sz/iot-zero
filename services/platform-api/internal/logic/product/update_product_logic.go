// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package product

import (
	"context"
	"errors"
	"strings"

	"iot-zero/services/platform-api/internal/authz"
	"iot-zero/services/platform-api/internal/svc"
	"iot-zero/services/platform-api/internal/types"
	"iot-zero/services/platform-api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductLogic {
	return &UpdateProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateProductLogic) UpdateProduct(req *types.UpdateProductReq) (resp *types.ProductInfo, err error) {
	if err = authz.AuthorizeSuperAdmin(l.ctx); err != nil {
		return nil, err
	}

	entity, err := l.svcCtx.ProductModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, authz.NotFound("product")
		}
		return nil, err
	}

	code := strings.TrimSpace(req.ProductCode)
	name := strings.TrimSpace(req.ProductName)
	if code == "" && name == "" {
		return nil, errors.New("productCode or productName is required")
	}

	if code != "" && code != entity.ProductCode {
		existing, findErr := l.svcCtx.ProductModel.FindOneByProductCode(l.ctx, code)
		if findErr == nil && existing.Id != entity.Id {
			return nil, errors.New("productCode already exists")
		}
		if findErr != nil && !errors.Is(findErr, model.ErrNotFound) {
			return nil, findErr
		}
		entity.ProductCode = code
	}

	if name != "" && name != entity.ProductName {
		existing, findErr := l.svcCtx.ProductModel.FindOneByProductName(l.ctx, name)
		if findErr == nil && existing.Id != entity.Id {
			return nil, errors.New("productName already exists")
		}
		if findErr != nil && !errors.Is(findErr, model.ErrNotFound) {
			return nil, findErr
		}
		entity.ProductName = name
	}

	if err = l.svcCtx.ProductModel.Update(l.ctx, entity); err != nil {
		return nil, err
	}

	updated, err := l.svcCtx.ProductModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return nil, err
	}
	info := toProductInfo(updated)
	return &info, nil
}
