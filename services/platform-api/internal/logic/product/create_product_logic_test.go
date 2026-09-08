package product

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"iot-zero/services/platform-api/internal/authz"
	"iot-zero/services/platform-api/internal/session"
	"iot-zero/services/platform-api/internal/svc"
	"iot-zero/services/platform-api/internal/types"
	"iot-zero/services/platform-api/model"
)

func newCreateProductLogicWithMock(t *testing.T, ctx context.Context) (*CreateProductLogic, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	conn := sqlx.NewSqlConnFromDB(db)
	return NewCreateProductLogic(ctx, &svc.ServiceContext{
		ProductModel: model.NewProductModel(conn),
	}), mock
}

func operatorCtx(roleType string, tenantId uint64) context.Context {
	ctx := context.WithValue(context.Background(), "roleType", roleType)
	return context.WithValue(ctx, "tenantId", float64(tenantId))
}

func TestCreateProductLogic_CreateProduct(t *testing.T) {
	req := &types.CreateProductReq{ProductCode: "P001", ProductName: "Sensor"}

	t.Run("unauthorized without role", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newCreateProductLogicWithMock(t, context.Background())
		_, err := l.CreateProduct(req)
		ast.ErrorIs(err, authz.ErrUnauthorized)
	})

	t.Run("user is forbidden", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newCreateProductLogicWithMock(t, operatorCtx(session.RoleTypeUser, 7))
		_, err := l.CreateProduct(req)
		ast.ErrorIs(err, authz.ErrForbidden)
	})

	t.Run("tenant admin is forbidden", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newCreateProductLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		_, err := l.CreateProduct(req)
		ast.ErrorIs(err, authz.ErrForbidden)
	})

	t.Run("super admin creates product", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newCreateProductLogicWithMock(t, operatorCtx(session.RoleTypeSuperAdmin, 1))
		mock.ExpectQuery("select .+ from `product` where `product_code` = \\? and `deleted_at` is null").
			WithArgs("P001").
			WillReturnError(sqlx.ErrNotFound)
		mock.ExpectQuery("select .+ from `product` where `product_name` = \\? and `deleted_at` is null").
			WithArgs("Sensor").
			WillReturnError(sqlx.ErrNotFound)
		mock.ExpectExec("insert into `product`").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "P001", "Sensor").
			WillReturnResult(sqlmock.NewResult(5, 1))

		resp, err := l.CreateProduct(req)
		ast.NoError(err)
		ast.Equal(uint64(5), resp.Id)
		ast.NoError(mock.ExpectationsWereMet())
	})
}
