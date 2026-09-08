package product

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"iot-zero/services/platform-api/internal/authz"
	"iot-zero/services/platform-api/internal/session"
	"iot-zero/services/platform-api/internal/svc"
	"iot-zero/services/platform-api/internal/types"
	"iot-zero/services/platform-api/model"
)

func newUpdateProductLogicWithMock(t *testing.T, ctx context.Context) (*UpdateProductLogic, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	conn := sqlx.NewSqlConnFromDB(db)
	return NewUpdateProductLogic(ctx, &svc.ServiceContext{
		ProductModel: model.NewProductModel(conn),
	}), mock
}

func TestUpdateProductLogic_UpdateProduct(t *testing.T) {
	req := &types.UpdateProductReq{Id: 5, ProductName: "Renamed"}

	t.Run("unauthorized without role", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newUpdateProductLogicWithMock(t, context.Background())
		_, err := l.UpdateProduct(req)
		ast.ErrorIs(err, authz.ErrUnauthorized)
	})

	t.Run("user is forbidden", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newUpdateProductLogicWithMock(t, operatorCtx(session.RoleTypeUser, 7))
		_, err := l.UpdateProduct(req)
		ast.ErrorIs(err, authz.ErrForbidden)
	})

	t.Run("tenant admin is forbidden", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newUpdateProductLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		_, err := l.UpdateProduct(req)
		ast.ErrorIs(err, authz.ErrForbidden)
	})

	t.Run("super admin updates product", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newUpdateProductLogicWithMock(t, operatorCtx(session.RoleTypeSuperAdmin, 1))
		now := time.Now()
		productRows := []string{"id", "created_at", "updated_at", "deleted_at", "product_code", "product_name"}
		mock.ExpectQuery("select .+ from `product` where `id` = \\? and `deleted_at` is null").
			WithArgs(uint64(5)).
			WillReturnRows(sqlmock.NewRows(productRows).
				AddRow(uint64(5), now, now, nil, "P001", "Sensor"))
		mock.ExpectQuery("select .+ from `product` where `product_name` = \\? and `deleted_at` is null").
			WithArgs("Renamed").
			WillReturnError(sqlx.ErrNotFound)
		mock.ExpectExec("update `product` set `updated_at`").
			WithArgs(sqlmock.AnyArg(), "P001", "Renamed", uint64(5)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery("select .+ from `product` where `id` = \\? and `deleted_at` is null").
			WithArgs(uint64(5)).
			WillReturnRows(sqlmock.NewRows(productRows).
				AddRow(uint64(5), now, now, nil, "P001", "Renamed"))

		resp, err := l.UpdateProduct(req)
		ast.NoError(err)
		ast.Equal("Renamed", resp.ProductName)
		ast.NoError(mock.ExpectationsWereMet())
	})
}
