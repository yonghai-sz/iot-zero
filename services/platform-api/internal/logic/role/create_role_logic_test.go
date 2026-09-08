package role

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

func newCreateRoleLogicWithMock(t *testing.T, ctx context.Context) (*CreateRoleLogic, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	conn := sqlx.NewSqlConnFromDB(db)
	return NewCreateRoleLogic(ctx, &svc.ServiceContext{
		RoleModel: model.NewRoleModel(conn),
	}), mock
}

func operatorCtx(roleType string, tenantId uint64) context.Context {
	ctx := context.WithValue(context.Background(), "roleType", roleType)
	return context.WithValue(ctx, "tenantId", float64(tenantId))
}

func TestCreateRoleLogic_CreateRole(t *testing.T) {
	req := &types.CreateRoleReq{TenantId: 7, RoleName: "operator", Enable: true}

	t.Run("unauthorized without role", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newCreateRoleLogicWithMock(t, context.Background())
		_, err := l.CreateRole(req)
		ast.ErrorIs(err, errUnauthorized)
	})

	t.Run("user is forbidden", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newCreateRoleLogicWithMock(t, operatorCtx(session.RoleTypeUser, 7))
		_, err := l.CreateRole(req)
		ast.ErrorIs(err, errForbidden)
	})

	t.Run("tenant admin cannot create for another tenant", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newCreateRoleLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 3))
		_, err := l.CreateRole(req)
		ast.ErrorIs(err, errForbidden)
	})

	t.Run("tenant admin creates for own tenant", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newCreateRoleLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		mock.ExpectExec("insert into `role`").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), session.RoleTypeUser, "operator", "Enable", uint64(7)).
			WillReturnResult(sqlmock.NewResult(11, 1))

		resp, err := l.CreateRole(req)
		ast.NoError(err)
		ast.Equal(uint64(11), resp.Id)
		ast.NoError(mock.ExpectationsWereMet())
	})

	t.Run("super admin can create for any tenant", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newCreateRoleLogicWithMock(t, operatorCtx(session.RoleTypeSuperAdmin, 1))
		mock.ExpectExec("insert into `role`").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), session.RoleTypeUser, "operator", "Enable", uint64(7)).
			WillReturnResult(sqlmock.NewResult(12, 1))

		resp, err := l.CreateRole(req)
		ast.NoError(err)
		ast.Equal(uint64(12), resp.Id)
		ast.NoError(mock.ExpectationsWereMet())
	})
}
