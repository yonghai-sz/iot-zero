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

func newUpdateUserLogicWithMock(t *testing.T, ctx context.Context) (*UpdateUserLogic, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	conn := sqlx.NewSqlConnFromDB(db)
	return NewUpdateUserLogic(ctx, &svc.ServiceContext{
		RoleModel: model.NewRoleModel(conn),
		UserModel: model.NewUserModel(conn),
	}), mock
}

func TestUpdateUserLogic_UpdateUser(t *testing.T) {
	enable := false
	req := &types.UpdateUserReq{Username: "alice", Enable: &enable}

	t.Run("unauthorized without role", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newUpdateUserLogicWithMock(t, context.Background())
		_, err := l.UpdateUser(req)
		ast.ErrorIs(err, errUnauthorized)
	})

	t.Run("user cannot update others", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newUpdateUserLogicWithMock(t, operatorUserCtx(session.RoleTypeUser, 7, "bob"))
		_, err := l.UpdateUser(req)
		ast.ErrorIs(err, errForbidden)
	})

	t.Run("user can update self", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newUpdateUserLogicWithMock(t, operatorUserCtx(session.RoleTypeUser, 7, "alice"))
		now := time.Now()
		mock.ExpectQuery("select .+ from `user` where `username` = \\? and `deleted_at` is null").
			WithArgs("alice").
			WillReturnRows(userRow(now, 8, "alice", 7))
		mock.ExpectExec("update `user` set `updated_at`").
			WithArgs(sqlmock.AnyArg(), "alice", "hash", "salt", "Disable", uint64(2), uint64(7), uint64(8)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery("select .+ from `user` where `id` = \\? and `deleted_at` is null").
			WithArgs(uint64(8)).
			WillReturnRows(sqlmock.NewRows(userColumns()).
				AddRow(uint64(8), now, now, nil, "alice", "hash", "salt", "Disable", uint64(2), uint64(7)))

		resp, err := l.UpdateUser(req)
		ast.NoError(err)
		ast.Equal("alice", resp.Username)
		ast.False(resp.Enable)
		ast.NoError(mock.ExpectationsWereMet())
	})

	t.Run("tenant admin cannot update another tenant", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newUpdateUserLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		now := time.Now()
		mock.ExpectQuery("select .+ from `user` where `username` = \\? and `deleted_at` is null").
			WithArgs("alice").
			WillReturnRows(userRow(now, 8, "alice", 3))

		_, err := l.UpdateUser(req)
		ast.ErrorIs(err, errUserNotFound)
		ast.NoError(mock.ExpectationsWereMet())
	})

	t.Run("tenant admin updates own tenant user", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newUpdateUserLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		now := time.Now()
		mock.ExpectQuery("select .+ from `user` where `username` = \\? and `deleted_at` is null").
			WithArgs("alice").
			WillReturnRows(userRow(now, 8, "alice", 7))
		mock.ExpectExec("update `user` set `updated_at`").
			WithArgs(sqlmock.AnyArg(), "alice", "hash", "salt", "Disable", uint64(2), uint64(7), uint64(8)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery("select .+ from `user` where `id` = \\? and `deleted_at` is null").
			WithArgs(uint64(8)).
			WillReturnRows(sqlmock.NewRows(userColumns()).
				AddRow(uint64(8), now, now, nil, "alice", "hash", "salt", "Disable", uint64(2), uint64(7)))

		resp, err := l.UpdateUser(req)
		ast.NoError(err)
		ast.False(resp.Enable)
		ast.NoError(mock.ExpectationsWereMet())
	})

	t.Run("super admin can update any tenant user", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newUpdateUserLogicWithMock(t, operatorCtx(session.RoleTypeSuperAdmin, 1))
		now := time.Now()
		mock.ExpectQuery("select .+ from `user` where `username` = \\? and `deleted_at` is null").
			WithArgs("alice").
			WillReturnRows(userRow(now, 8, "alice", 7))
		mock.ExpectExec("update `user` set `updated_at`").
			WithArgs(sqlmock.AnyArg(), "alice", "hash", "salt", "Disable", uint64(2), uint64(7), uint64(8)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery("select .+ from `user` where `id` = \\? and `deleted_at` is null").
			WithArgs(uint64(8)).
			WillReturnRows(sqlmock.NewRows(userColumns()).
				AddRow(uint64(8), now, now, nil, "alice", "hash", "salt", "Disable", uint64(2), uint64(7)))

		resp, err := l.UpdateUser(req)
		ast.NoError(err)
		ast.Equal(uint64(7), resp.TenantId)
		ast.NoError(mock.ExpectationsWereMet())
	})
}
