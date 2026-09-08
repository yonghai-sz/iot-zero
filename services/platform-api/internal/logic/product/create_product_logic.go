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

type CreateProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateProductLogic {
	return &CreateProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateProductLogic) CreateProduct(req *types.CreateProductReq) (resp *types.CreateProductResp, err error) {
	if err = authz.AuthorizeSuperAdmin(l.ctx); err != nil {
		return nil, err
	}

	code := strings.TrimSpace(req.ProductCode)
	name := strings.TrimSpace(req.ProductName)
	if code == "" {
		return nil, errors.New("productCode is required")
	}
	if name == "" {
		return nil, errors.New("productName is required")
	}

	_, err = l.svcCtx.ProductModel.FindOneByProductCode(l.ctx, code)
	if err == nil {
		return nil, errors.New("productCode already exists")
	}
	if !errors.Is(err, model.ErrNotFound) {
		return nil, err
	}

	_, err = l.svcCtx.ProductModel.FindOneByProductName(l.ctx, name)
	if err == nil {
		return nil, errors.New("productName already exists")
	}
	if !errors.Is(err, model.ErrNotFound) {
		return nil, err
	}

	result, err := l.svcCtx.ProductModel.Insert(l.ctx, &model.Product{
		ProductCode: code,
		ProductName: name,
	})
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &types.CreateProductResp{Id: uint64(id)}, nil
}
