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

type DeleteUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserLogic {
	return &DeleteUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteUserLogic) DeleteUser(req *types.UserUsernamePathReq) error {
	entity, err := loadUserForAdmin(l.ctx, l.svcCtx.UserModel, req.Username)
	if err != nil {
		return err
	}

	if err = l.svcCtx.UserModel.Delete(l.ctx, entity.Id); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return errUserNotFound
		}
		return err
	}
	return nil
}
