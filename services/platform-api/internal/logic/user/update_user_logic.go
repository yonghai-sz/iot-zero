// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"
	"errors"

	"iot-zero/services/platform-api/internal/svc"
	"iot-zero/services/platform-api/internal/types"
	"iot-zero/services/platform-api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserLogic) UpdateUser(req *types.UpdateUserReq) (resp *types.UserInfo, err error) {
	entity, err := loadUserForSelfOrAdmin(l.ctx, l.svcCtx.UserModel, req.Username)
	if err != nil {
		return nil, err
	}

	if req.RoleId == 0 && req.Enable == nil {
		return nil, errors.New("roleId or enable is required")
	}

	if req.RoleId > 0 {
		if _, err = l.svcCtx.RoleModel.FindOne(l.ctx, req.RoleId); err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return nil, errors.New("role not found")
			}
			return nil, err
		}
		entity.RoleId = req.RoleId
	}
	if req.Enable != nil {
		entity.Enable = boolToEnable(*req.Enable)
	}

	if err = l.svcCtx.UserModel.Update(l.ctx, entity); err != nil {
		return nil, err
	}

	updated, err := l.svcCtx.UserModel.FindOne(l.ctx, entity.Id)
	if err != nil {
		return nil, err
	}
	info := toUserInfo(updated)
	return &info, nil
}
