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

func newListTenantLogicWithMock(t *testing.T, ctx context.Context) (*ListTenantLogic, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	conn := sqlx.NewSqlConnFromDB(db)
	return NewListTenantLogic(ctx, &svc.ServiceContext{
		TenantsModel: model.NewTenantsModel(conn),
	}), mock
}

func TestListTenantLogic_ListTenant(t *testing.T) {
	req := &types.ListTenantReq{PageIndex: 1, PageSize: 20}

	t.Run("unauthorized without role", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newListTenantLogicWithMock(t, context.Background())
		_, err := l.ListTenant(req)
		ast.ErrorIs(err, errUnauthorized)
	})

	t.Run("user is forbidden", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newListTenantLogicWithMock(t, operatorCtx(session.RoleTypeUser, 7))
		_, err := l.ListTenant(req)
		ast.ErrorIs(err, errForbidden)
	})

	t.Run("tenant admin is forbidden", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newListTenantLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		_, err := l.ListTenant(req)
		ast.ErrorIs(err, errForbidden)
	})

	t.Run("super admin lists tenants", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newListTenantLogicWithMock(t, operatorCtx(session.RoleTypeSuperAdmin, 1))
		now := time.Now()
		mock.ExpectQuery("select count\\(\\*\\) from `tenants` where `deleted_at` is null").
			WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(1))
		mock.ExpectQuery("select .+ from `tenants` where `deleted_at` is null order by `updated_at` desc limit \\? offset \\?").
			WithArgs(int64(20), int64(0)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "tenant_name", "email"}).
				AddRow(uint64(7), now, now, nil, "acme", "a@acme.com"))

		resp, err := l.ListTenant(req)
		ast.NoError(err)
		ast.Equal(int64(1), resp.Total)
		ast.Len(resp.List, 1)
		ast.Equal(uint64(7), resp.List[0].Id)
		ast.NoError(mock.ExpectationsWereMet())
	})
}
