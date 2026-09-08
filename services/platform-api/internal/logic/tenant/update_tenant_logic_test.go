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

func newUpdateTenantLogicWithMock(t *testing.T, ctx context.Context) (*UpdateTenantLogic, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	conn := sqlx.NewSqlConnFromDB(db)
	return NewUpdateTenantLogic(ctx, &svc.ServiceContext{
		TenantsModel: model.NewTenantsModel(conn),
	}), mock
}

func TestUpdateTenantLogic_UpdateTenant(t *testing.T) {
	req := &types.UpdateTenantReq{Id: 7, TenantName: "renamed"}

	t.Run("unauthorized without role", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newUpdateTenantLogicWithMock(t, context.Background())
		_, err := l.UpdateTenant(req)
		ast.ErrorIs(err, errUnauthorized)
	})

	t.Run("user is forbidden", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newUpdateTenantLogicWithMock(t, operatorCtx(session.RoleTypeUser, 7))
		_, err := l.UpdateTenant(req)
		ast.ErrorIs(err, errForbidden)
	})

	t.Run("tenant admin is forbidden", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newUpdateTenantLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		_, err := l.UpdateTenant(req)
		ast.ErrorIs(err, errForbidden)
	})

	t.Run("super admin updates tenant", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newUpdateTenantLogicWithMock(t, operatorCtx(session.RoleTypeSuperAdmin, 1))
		now := time.Now()
		tenantRows := []string{"id", "created_at", "updated_at", "deleted_at", "tenant_name", "email"}
		mock.ExpectQuery("select .+ from `tenants` where `id` = \\? and `deleted_at` is null").
			WithArgs(uint64(7)).
			WillReturnRows(sqlmock.NewRows(tenantRows).
				AddRow(uint64(7), now, now, nil, "acme", "a@acme.com"))
		mock.ExpectQuery("select .+ from `tenants` where `tenant_name` = \\? and `deleted_at` is null").
			WithArgs("renamed").
			WillReturnError(sqlx.ErrNotFound)
		mock.ExpectExec("update `tenants` set `updated_at`").
			WithArgs(sqlmock.AnyArg(), "renamed", "a@acme.com", uint64(7)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery("select .+ from `tenants` where `id` = \\? and `deleted_at` is null").
			WithArgs(uint64(7)).
			WillReturnRows(sqlmock.NewRows(tenantRows).
				AddRow(uint64(7), now, now, nil, "renamed", "a@acme.com"))

		resp, err := l.UpdateTenant(req)
		ast.NoError(err)
		ast.Equal("renamed", resp.TenantName)
		ast.NoError(mock.ExpectationsWereMet())
	})
}
