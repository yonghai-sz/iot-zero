// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package login

import (
	"context"

	"iot-zero/services/platform-api/internal/authz"
	"iot-zero/services/platform-api/internal/session"
	"iot-zero/services/platform-api/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LogoutLogic) Logout() error {
	username, ok := session.UsernameFromContext(l.ctx)
	if !ok {
		return authz.ErrUnauthorized
	}
	return l.svcCtx.SessionStore.Delete(l.ctx, username)
}
