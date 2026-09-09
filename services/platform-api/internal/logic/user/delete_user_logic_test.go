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

func newDeleteUserLogicWithMock(t *testing.T, ctx context.Context) (*DeleteUserLogic, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	conn := sqlx.NewSqlConnFromDB(db)
	return NewDeleteUserLogic(ctx, &svc.ServiceContext{
		UserModel: model.NewUserModel(conn),
	}), mock
}

func TestDeleteUserLogic_DeleteUser(t *testing.T) {
	req := &types.UserUsernamePathReq{Username: "alice"}

	t.Run("unauthorized without role", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newDeleteUserLogicWithMock(t, context.Background())
		err := l.DeleteUser(req)
		ast.ErrorIs(err, errUnauthorized)
	})

	t.Run("user is forbidden", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newDeleteUserLogicWithMock(t, operatorUserCtx(session.RoleTypeUser, 7, "alice"))
		err := l.DeleteUser(req)
		ast.ErrorIs(err, errForbidden)
	})

	t.Run("missing user", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newDeleteUserLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		mock.ExpectQuery("select .+ from `user` where `username` = \\? and `deleted_at` is null").
			WithArgs("alice").
			WillReturnError(sqlx.ErrNotFound)

		err := l.DeleteUser(req)
		ast.ErrorIs(err, errUserNotFound)
		ast.NoError(mock.ExpectationsWereMet())
	})

	t.Run("tenant admin cannot delete another tenant", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newDeleteUserLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		now := time.Now()
		mock.ExpectQuery("select .+ from `user` where `username` = \\? and `deleted_at` is null").
			WithArgs("alice").
			WillReturnRows(userRow(now, 8, "alice", 3))

		err := l.DeleteUser(req)
		ast.ErrorIs(err, errUserNotFound)
		ast.NoError(mock.ExpectationsWereMet())
	})

	t.Run("tenant admin deletes own tenant user", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newDeleteUserLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		now := time.Now()
		mock.ExpectQuery("select .+ from `user` where `username` = \\? and `deleted_at` is null").
			WithArgs("alice").
			WillReturnRows(userRow(now, 8, "alice", 7))
		mock.ExpectExec("update `user` set `deleted_at`").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), uint64(8)).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := l.DeleteUser(req)
		ast.NoError(err)
		ast.NoError(mock.ExpectationsWereMet())
	})

	t.Run("super admin can delete any tenant user", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newDeleteUserLogicWithMock(t, operatorCtx(session.RoleTypeSuperAdmin, 1))
		now := time.Now()
		mock.ExpectQuery("select .+ from `user` where `username` = \\? and `deleted_at` is null").
			WithArgs("alice").
			WillReturnRows(userRow(now, 8, "alice", 7))
		mock.ExpectExec("update `user` set `deleted_at`").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), uint64(8)).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := l.DeleteUser(req)
		ast.NoError(err)
		ast.NoError(mock.ExpectationsWereMet())
	})
}
