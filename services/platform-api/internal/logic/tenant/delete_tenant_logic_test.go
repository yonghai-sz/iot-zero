package tenant

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"iot-zero/services/platform-api/internal/session"
	"iot-zero/services/platform-api/internal/svc"
	"iot-zero/services/platform-api/internal/types"
	"iot-zero/services/platform-api/model"
)

func newDeleteTenantLogicWithMock(t *testing.T, ctx context.Context) (*DeleteTenantLogic, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	conn := sqlx.NewSqlConnFromDB(db)
	return NewDeleteTenantLogic(ctx, &svc.ServiceContext{
		TenantsModel: model.NewTenantsModel(conn),
		RoleModel:    model.NewRoleModel(conn),
	}), mock
}

func TestDeleteTenantLogic_DeleteTenant(t *testing.T) {
	req := &types.TenantIdPathReq{Id: 7}

	t.Run("unauthorized without role", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newDeleteTenantLogicWithMock(t, context.Background())
		err := l.DeleteTenant(req)
		ast.ErrorIs(err, errUnauthorized)
	})

	t.Run("user is forbidden", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newDeleteTenantLogicWithMock(t, operatorCtx(session.RoleTypeUser, 7))
		err := l.DeleteTenant(req)
		ast.ErrorIs(err, errForbidden)
	})

	t.Run("tenant admin is forbidden", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newDeleteTenantLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		err := l.DeleteTenant(req)
		ast.ErrorIs(err, errForbidden)
	})

	t.Run("super admin deletes tenant", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newDeleteTenantLogicWithMock(t, operatorCtx(session.RoleTypeSuperAdmin, 1))
		now := time.Now()
		mock.ExpectExec("update `tenants` set `deleted_at`").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), uint64(7)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery("select .+ from `role` where `role_type` = \\? and `tenant_id` = \\? and `deleted_at` is null").
			WithArgs(session.RoleTypeTenantAdmin, uint64(7)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "role_type", "role_name", "enable", "tenant_id"}).
				AddRow(uint64(21), now, now, nil, session.RoleTypeTenantAdmin, "tenant administrator", "Enable", uint64(7)))
		mock.ExpectExec("update `role` set `deleted_at`").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), uint64(21)).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := l.DeleteTenant(req)
		ast.NoError(err)
		ast.NoError(mock.ExpectationsWereMet())
	})
}
