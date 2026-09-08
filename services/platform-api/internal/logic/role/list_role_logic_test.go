package role

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

func newListRoleLogicWithMock(t *testing.T, ctx context.Context) (*ListRoleLogic, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	conn := sqlx.NewSqlConnFromDB(db)
	return NewListRoleLogic(ctx, &svc.ServiceContext{
		RoleModel: model.NewRoleModel(conn),
	}), mock
}

func TestListRoleLogic_ListRole(t *testing.T) {
	req := &types.ListRoleReq{PageIndex: 1, PageSize: 20, TenantId: 7}

	t.Run("unauthorized without role", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newListRoleLogicWithMock(t, context.Background())
		_, err := l.ListRole(req)
		ast.ErrorIs(err, errUnauthorized)
	})

	t.Run("user is forbidden", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newListRoleLogicWithMock(t, operatorCtx(session.RoleTypeUser, 7))
		_, err := l.ListRole(req)
		ast.ErrorIs(err, errForbidden)
	})

	t.Run("tenant admin cannot list another tenant", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newListRoleLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 3))
		_, err := l.ListRole(req)
		ast.ErrorIs(err, errForbidden)
	})

	t.Run("tenant admin cannot omit tenant id", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newListRoleLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		_, err := l.ListRole(&types.ListRoleReq{PageIndex: 1, PageSize: 20})
		ast.ErrorIs(err, errForbidden)
	})

	t.Run("tenant admin lists own tenant", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newListRoleLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		now := time.Now()
		mock.ExpectQuery("select count\\(\\*\\) from `role` where `deleted_at` is null and `tenant_id` = \\?").
			WithArgs(uint64(7)).
			WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(1))
		mock.ExpectQuery("select .+ from `role` where `deleted_at` is null and `tenant_id` = \\? order by `updated_at` desc limit \\? offset \\?").
			WithArgs(uint64(7), int64(20), int64(0)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "role_type", "role_name", "enable", "tenant_id"}).
				AddRow(uint64(4), now, now, nil, session.RoleTypeUser, "operator", "Enable", uint64(7)))

		resp, err := l.ListRole(req)
		ast.NoError(err)
		ast.Equal(int64(1), resp.Total)
		ast.Len(resp.List, 1)
		ast.Equal(uint64(4), resp.List[0].Id)
		ast.NoError(mock.ExpectationsWereMet())
	})

	t.Run("super admin can list any tenant", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newListRoleLogicWithMock(t, operatorCtx(session.RoleTypeSuperAdmin, 1))
		now := time.Now()
		mock.ExpectQuery("select count\\(\\*\\) from `role` where `deleted_at` is null and `tenant_id` = \\?").
			WithArgs(uint64(7)).
			WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(1))
		mock.ExpectQuery("select .+ from `role` where `deleted_at` is null and `tenant_id` = \\? order by `updated_at` desc limit \\? offset \\?").
			WithArgs(uint64(7), int64(20), int64(0)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "role_type", "role_name", "enable", "tenant_id"}).
				AddRow(uint64(4), now, now, nil, session.RoleTypeUser, "operator", "Enable", uint64(7)))

		resp, err := l.ListRole(req)
		ast.NoError(err)
		ast.Equal(int64(1), resp.Total)
		ast.Len(resp.List, 1)
		ast.NoError(mock.ExpectationsWereMet())
	})
}
