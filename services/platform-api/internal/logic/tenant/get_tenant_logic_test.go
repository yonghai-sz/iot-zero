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

func newGetTenantLogicWithMock(t *testing.T, ctx context.Context) (*GetTenantLogic, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	conn := sqlx.NewSqlConnFromDB(db)
	return NewGetTenantLogic(ctx, &svc.ServiceContext{
		TenantsModel: model.NewTenantsModel(conn),
	}), mock
}

func TestGetTenantLogic_GetTenant(t *testing.T) {
	req := &types.TenantIdPathReq{Id: 7}

	t.Run("unauthorized without role", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newGetTenantLogicWithMock(t, context.Background())
		_, err := l.GetTenant(req)
		ast.ErrorIs(err, errUnauthorized)
	})

	t.Run("user is forbidden", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newGetTenantLogicWithMock(t, operatorCtx(session.RoleTypeUser, 7))
		_, err := l.GetTenant(req)
		ast.ErrorIs(err, errForbidden)
	})

	t.Run("tenant admin cannot get another tenant", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newGetTenantLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 3))
		_, err := l.GetTenant(req)
		ast.ErrorIs(err, errForbidden)
	})

	t.Run("missing tenant", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newGetTenantLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		mock.ExpectQuery("select .+ from `tenants` where `id` = \\? and `deleted_at` is null").
			WithArgs(uint64(7)).
			WillReturnError(sqlx.ErrNotFound)

		_, err := l.GetTenant(req)
		ast.ErrorIs(err, errTenantNotFound)
		ast.NoError(mock.ExpectationsWereMet())
	})

	t.Run("tenant admin gets own tenant", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newGetTenantLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		now := time.Now()
		mock.ExpectQuery("select .+ from `tenants` where `id` = \\? and `deleted_at` is null").
			WithArgs(uint64(7)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "tenant_name", "email"}).
				AddRow(uint64(7), now, now, nil, "acme", "a@acme.com"))

		resp, err := l.GetTenant(req)
		ast.NoError(err)
		ast.Equal(uint64(7), resp.Id)
		ast.Equal("acme", resp.TenantName)
		ast.NoError(mock.ExpectationsWereMet())
	})

	t.Run("super admin can get any tenant", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newGetTenantLogicWithMock(t, operatorCtx(session.RoleTypeSuperAdmin, 1))
		now := time.Now()
		mock.ExpectQuery("select .+ from `tenants` where `id` = \\? and `deleted_at` is null").
			WithArgs(uint64(7)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "tenant_name", "email"}).
				AddRow(uint64(7), now, now, nil, "acme", "a@acme.com"))

		resp, err := l.GetTenant(req)
		ast.NoError(err)
		ast.Equal(uint64(7), resp.Id)
		ast.NoError(mock.ExpectationsWereMet())
	})
}
