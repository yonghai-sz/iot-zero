package tenant

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"iot-zero/services/platform-api/internal/session"
	"iot-zero/services/platform-api/internal/svc"
	"iot-zero/services/platform-api/internal/types"
	"iot-zero/services/platform-api/model"
)

func newCreateTenantLogicWithMock(t *testing.T, ctx context.Context) (*CreateTenantLogic, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	conn := sqlx.NewSqlConnFromDB(db)
	return NewCreateTenantLogic(ctx, &svc.ServiceContext{
		TenantsModel: model.NewTenantsModel(conn),
		RoleModel:    model.NewRoleModel(conn),
	}), mock
}

func operatorCtx(roleType string, tenantId uint64) context.Context {
	ctx := context.WithValue(context.Background(), "roleType", roleType)
	return context.WithValue(ctx, "tenantId", float64(tenantId))
}

func TestCreateTenantLogic_CreateTenant(t *testing.T) {
	req := &types.CreateTenantReq{TenantName: "acme", Email: "a@acme.com"}

	t.Run("unauthorized without role", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newCreateTenantLogicWithMock(t, context.Background())
		_, err := l.CreateTenant(req)
		ast.ErrorIs(err, errUnauthorized)
	})

	t.Run("user is forbidden", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newCreateTenantLogicWithMock(t, operatorCtx(session.RoleTypeUser, 7))
		_, err := l.CreateTenant(req)
		ast.ErrorIs(err, errForbidden)
	})

	t.Run("tenant admin is forbidden", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newCreateTenantLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		_, err := l.CreateTenant(req)
		ast.ErrorIs(err, errForbidden)
	})

	t.Run("super admin creates tenant", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newCreateTenantLogicWithMock(t, operatorCtx(session.RoleTypeSuperAdmin, 1))
		mock.ExpectQuery("select .+ from `tenants` where `tenant_name` = \\? and `deleted_at` is null").
			WithArgs("acme").
			WillReturnError(sqlx.ErrNotFound)
		mock.ExpectExec("insert into `tenants`").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "acme", "a@acme.com").
			WillReturnResult(sqlmock.NewResult(11, 1))
		mock.ExpectExec("insert into `role`").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), session.RoleTypeTenantAdmin, "tenant administrator", "Enable", uint64(11)).
			WillReturnResult(sqlmock.NewResult(21, 1))

		resp, err := l.CreateTenant(req)
		ast.NoError(err)
		ast.Equal(uint64(11), resp.Id)
		ast.NoError(mock.ExpectationsWereMet())
	})
}
