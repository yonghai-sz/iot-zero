package user

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

func newListUserLogicWithMock(t *testing.T, ctx context.Context) (*ListUserLogic, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	conn := sqlx.NewSqlConnFromDB(db)
	return NewListUserLogic(ctx, &svc.ServiceContext{
		UserModel: model.NewUserModel(conn),
	}), mock
}

func TestListUserLogic_ListUser(t *testing.T) {
	req := &types.ListUserReq{PageIndex: 1, PageSize: 20, TenantId: 7}

	t.Run("unauthorized without role", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newListUserLogicWithMock(t, context.Background())
		_, err := l.ListUser(req)
		ast.ErrorIs(err, errUnauthorized)
	})

	t.Run("user is forbidden", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newListUserLogicWithMock(t, operatorCtx(session.RoleTypeUser, 7))
		_, err := l.ListUser(req)
		ast.ErrorIs(err, errForbidden)
	})

	t.Run("tenant admin cannot list another tenant", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newListUserLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 3))
		_, err := l.ListUser(req)
		ast.ErrorIs(err, errForbidden)
	})

	t.Run("tenant admin cannot omit tenant id", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newListUserLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		_, err := l.ListUser(&types.ListUserReq{PageIndex: 1, PageSize: 20})
		ast.ErrorIs(err, errForbidden)
	})

	t.Run("tenant admin lists own tenant", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newListUserLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		now := time.Now()
		mock.ExpectQuery("select count\\(\\*\\) from `user` where `deleted_at` is null and `tenant_id` = \\?").
			WithArgs(uint64(7)).
			WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(1))
		mock.ExpectQuery("select .+ from `user` where `deleted_at` is null and `tenant_id` = \\? order by `updated_at` desc limit \\? offset \\?").
			WithArgs(uint64(7), int64(20), int64(0)).
			WillReturnRows(userRow(now, 8, "alice", 7))

		resp, err := l.ListUser(req)
		ast.NoError(err)
		ast.Equal(int64(1), resp.Total)
		ast.Len(resp.List, 1)
		ast.Equal("alice", resp.List[0].Username)
		ast.NoError(mock.ExpectationsWereMet())
	})

	t.Run("super admin can list any tenant", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newListUserLogicWithMock(t, operatorCtx(session.RoleTypeSuperAdmin, 1))
		now := time.Now()
		mock.ExpectQuery("select count\\(\\*\\) from `user` where `deleted_at` is null and `tenant_id` = \\?").
			WithArgs(uint64(7)).
			WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(1))
		mock.ExpectQuery("select .+ from `user` where `deleted_at` is null and `tenant_id` = \\? order by `updated_at` desc limit \\? offset \\?").
			WithArgs(uint64(7), int64(20), int64(0)).
			WillReturnRows(userRow(now, 8, "alice", 7))

		resp, err := l.ListUser(req)
		ast.NoError(err)
		ast.Equal(int64(1), resp.Total)
		ast.Len(resp.List, 1)
		ast.NoError(mock.ExpectationsWereMet())
	})
}
