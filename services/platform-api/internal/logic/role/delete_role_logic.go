// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package role

import (
	"context"
	"errors"

	"iot-zero/services/platform-api/internal/svc"
	"iot-zero/services/platform-api/internal/types"
	"iot-zero/services/platform-api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteRoleLogic {
	return &DeleteRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteRoleLogic) DeleteRole(req *types.RoleIdPathReq) error {
	if _, err := loadRoleForAdmin(l.ctx, l.svcCtx.RoleModel, req.Id); err != nil {
		return err
	}

	err := l.svcCtx.RoleModel.Delete(l.ctx, req.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return errRoleNotFound
		}
		return err
	}
	return nil
}
