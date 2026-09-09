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

func newUpdateUserPasswordLogicWithMock(t *testing.T, ctx context.Context) (*UpdateUserPasswordLogic, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	conn := sqlx.NewSqlConnFromDB(db)
	return NewUpdateUserPasswordLogic(ctx, &svc.ServiceContext{
		UserModel: model.NewUserModel(conn),
	}), mock
}

func TestUpdateUserPasswordLogic_UpdateUserPassword(t *testing.T) {
	req := &types.UpdateUserPasswordReq{Username: "alice", Password: "newpass"}

	t.Run("unauthorized without role", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newUpdateUserPasswordLogicWithMock(t, context.Background())
		err := l.UpdateUserPassword(req)
		ast.ErrorIs(err, errUnauthorized)
	})

	t.Run("user cannot update others", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newUpdateUserPasswordLogicWithMock(t, operatorUserCtx(session.RoleTypeUser, 7, "bob"))
		err := l.UpdateUserPassword(req)
		ast.ErrorIs(err, errForbidden)
	})

	t.Run("user can update self", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newUpdateUserPasswordLogicWithMock(t, operatorUserCtx(session.RoleTypeUser, 7, "alice"))
		now := time.Now()
		mock.ExpectQuery("select .+ from `user` where `username` = \\? and `deleted_at` is null").
			WithArgs("alice").
			WillReturnRows(userRow(now, 8, "alice", 7))
		mock.ExpectExec("update `user` set `updated_at`").
			WithArgs(sqlmock.AnyArg(), "alice", sqlmock.AnyArg(), sqlmock.AnyArg(), "Enable", uint64(2), uint64(7), uint64(8)).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := l.UpdateUserPassword(req)
		ast.NoError(err)
		ast.NoError(mock.ExpectationsWereMet())
	})

	t.Run("tenant admin cannot update another tenant", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newUpdateUserPasswordLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		now := time.Now()
		mock.ExpectQuery("select .+ from `user` where `username` = \\? and `deleted_at` is null").
			WithArgs("alice").
			WillReturnRows(userRow(now, 8, "alice", 3))

		err := l.UpdateUserPassword(req)
		ast.ErrorIs(err, errUserNotFound)
		ast.NoError(mock.ExpectationsWereMet())
	})

	t.Run("super admin can update any tenant user", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newUpdateUserPasswordLogicWithMock(t, operatorCtx(session.RoleTypeSuperAdmin, 1))
		now := time.Now()
		mock.ExpectQuery("select .+ from `user` where `username` = \\? and `deleted_at` is null").
			WithArgs("alice").
			WillReturnRows(userRow(now, 8, "alice", 7))
		mock.ExpectExec("update `user` set `updated_at`").
			WithArgs(sqlmock.AnyArg(), "alice", sqlmock.AnyArg(), sqlmock.AnyArg(), "Enable", uint64(2), uint64(7), uint64(8)).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := l.UpdateUserPassword(req)
		ast.NoError(err)
		ast.NoError(mock.ExpectationsWereMet())
	})
}
