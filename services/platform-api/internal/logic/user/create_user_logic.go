// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"
	"errors"
	"strings"
	"unicode"

	"iot-zero/pkg/utils"
	"iot-zero/services/platform-api/internal/authz"
	"iot-zero/services/platform-api/internal/svc"
	"iot-zero/services/platform-api/internal/types"
	"iot-zero/services/platform-api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserLogic {
	return &CreateUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateUserLogic) CreateUser(req *types.CreateUserReq) (resp *types.CreateUserResp, err error) {
	if err = authz.AuthorizeAdminForTenant(l.ctx, req.TenantId); err != nil {
		return nil, err
	}
	if req.TenantId == 0 {
		return nil, errors.New("tenantId is required")
	}
	if req.RoleId == 0 {
		return nil, errors.New("roleId is required")
	}
	username := strings.TrimSpace(req.Username)
	if err = validateUsername(username); err != nil {
		return nil, err
	}
	password := strings.TrimSpace(req.Password)
	if password == "" {
		return nil, errors.New("password is required")
	}

	_, err = l.svcCtx.RoleModel.FindOne(l.ctx, req.RoleId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}

	_, err = l.svcCtx.UserModel.FindOneByUsername(l.ctx, username)
	if err == nil {
		return nil, errors.New("username already exists")
	}
	if !errors.Is(err, model.ErrNotFound) {
		return nil, err
	}

	salt := utils.GenerateCryptoRandString(6)
	hashed := utils.HashPassword(password, salt)

	result, err := l.svcCtx.UserModel.Insert(l.ctx, &model.User{
		Username: username,
		Password: hashed,
		Salt:     salt,
		Enable:   boolToEnable(req.Enable),
		RoleId:   req.RoleId,
		TenantId: req.TenantId,
	})
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &types.CreateUserResp{Id: uint64(id)}, nil
}

func validateUsername(username string) error {
	if len(username) < 3 || len(username) > 63 {
		return errors.New("username length must be between 3 and 63")
	}
	for _, r := range username {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return errors.New("username must be alphanumeric")
		}
	}
	return nil
}
