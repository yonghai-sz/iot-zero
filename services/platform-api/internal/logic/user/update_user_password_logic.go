// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"
	"errors"
	"strings"

	"iot-zero/pkg/utils"
	"iot-zero/services/platform-api/internal/svc"
	"iot-zero/services/platform-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserPasswordLogic {
	return &UpdateUserPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserPasswordLogic) UpdateUserPassword(req *types.UpdateUserPasswordReq) error {
	password := strings.TrimSpace(req.Password)
	if password == "" {
		return errors.New("password is required")
	}

	entity, err := loadUserForSelfOrAdmin(l.ctx, l.svcCtx.UserModel, req.Username)
	if err != nil {
		return err
	}

	salt := utils.GenerateCryptoRandString(6)
	hashed := utils.HashPassword(password, salt)

	entity.Password = hashed
	entity.Salt = salt

	return l.svcCtx.UserModel.Update(l.ctx, entity)
}
