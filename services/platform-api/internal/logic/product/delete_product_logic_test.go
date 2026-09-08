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

func newDeleteProductLogicWithMock(t *testing.T, ctx context.Context) (*DeleteProductLogic, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	conn := sqlx.NewSqlConnFromDB(db)
	return NewDeleteProductLogic(ctx, &svc.ServiceContext{
		ProductModel: model.NewProductModel(conn),
	}), mock
}

func TestDeleteProductLogic_DeleteProduct(t *testing.T) {
	req := &types.ProductIdPathReq{Id: 5}

	t.Run("unauthorized without role", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newDeleteProductLogicWithMock(t, context.Background())
		err := l.DeleteProduct(req)
		ast.ErrorIs(err, authz.ErrUnauthorized)
	})

	t.Run("user is forbidden", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newDeleteProductLogicWithMock(t, operatorCtx(session.RoleTypeUser, 7))
		err := l.DeleteProduct(req)
		ast.ErrorIs(err, authz.ErrForbidden)
	})

	t.Run("tenant admin is forbidden", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newDeleteProductLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		err := l.DeleteProduct(req)
		ast.ErrorIs(err, authz.ErrForbidden)
	})

	t.Run("super admin deletes product", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newDeleteProductLogicWithMock(t, operatorCtx(session.RoleTypeSuperAdmin, 1))
		mock.ExpectExec("update `product` set `deleted_at`").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), uint64(5)).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := l.DeleteProduct(req)
		ast.NoError(err)
		ast.NoError(mock.ExpectationsWereMet())
	})
}
