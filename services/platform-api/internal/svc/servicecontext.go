// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"iot-zero/services/platform-api/internal/config"
	"iot-zero/services/platform-api/internal/middleware"
	"iot-zero/services/platform-api/internal/session"
	"iot-zero/services/platform-api/model"
	"iot-zero/services/rpc-transform/transformer"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config        config.Config
	Transformer   transformer.Transformer
	TenantsModel  model.TenantsModel
	ProductModel  model.ProductModel
	RoleModel     model.RoleModel
	UserModel     model.UserModel
	MenuModel     model.MenuModel
	RoleMenuModel model.RoleMenuModel
	SessionStore  *session.Store
	Session       rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	store := session.NewStore(redis.MustNewRedis(c.Redis))
	return &ServiceContext{
		Config:        c,
		Transformer:   transformer.NewTransformer(zrpc.MustNewClient(c.Transform)),
		TenantsModel:  model.NewTenantsModel(conn),
		ProductModel:  model.NewProductModel(conn),
		RoleModel:     model.NewRoleModel(conn),
		UserModel:     model.NewUserModel(conn),
		MenuModel:     model.NewMenuModel(conn),
		RoleMenuModel: model.NewRoleMenuModel(conn),
		SessionStore:  store,
		Session:       middleware.NewSessionMiddleware(store).Handle,
	}
}
