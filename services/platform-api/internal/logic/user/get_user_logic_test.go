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

func newGetUserLogicWithMock(t *testing.T, ctx context.Context) (*GetUserLogic, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	conn := sqlx.NewSqlConnFromDB(db)
	return NewGetUserLogic(ctx, &svc.ServiceContext{
		UserModel: model.NewUserModel(conn),
	}), mock
}

func TestGetUserLogic_GetUser(t *testing.T) {
	req := &types.UserUsernamePathReq{Username: "alice"}

	t.Run("unauthorized without role", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newGetUserLogicWithMock(t, context.Background())
		_, err := l.GetUser(req)
		ast.ErrorIs(err, errUnauthorized)
	})

	t.Run("user cannot get others", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newGetUserLogicWithMock(t, operatorUserCtx(session.RoleTypeUser, 7, "bob"))
		_, err := l.GetUser(req)
		ast.ErrorIs(err, errForbidden)
	})

	t.Run("user can get self", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newGetUserLogicWithMock(t, operatorUserCtx(session.RoleTypeUser, 7, "alice"))
		now := time.Now()
		mock.ExpectQuery("select .+ from `user` where `username` = \\? and `deleted_at` is null").
			WithArgs("alice").
			WillReturnRows(userRow(now, 8, "alice", 7))

		resp, err := l.GetUser(req)
		ast.NoError(err)
		ast.Equal("alice", resp.Username)
		ast.NoError(mock.ExpectationsWereMet())
	})

	t.Run("missing user", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newGetUserLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		mock.ExpectQuery("select .+ from `user` where `username` = \\? and `deleted_at` is null").
			WithArgs("alice").
			WillReturnError(sqlx.ErrNotFound)

		_, err := l.GetUser(req)
		ast.ErrorIs(err, errUserNotFound)
		ast.NoError(mock.ExpectationsWereMet())
	})

	t.Run("tenant admin cannot get another tenant", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newGetUserLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		now := time.Now()
		mock.ExpectQuery("select .+ from `user` where `username` = \\? and `deleted_at` is null").
			WithArgs("alice").
			WillReturnRows(userRow(now, 8, "alice", 3))

		_, err := l.GetUser(req)
		ast.ErrorIs(err, errUserNotFound)
		ast.NoError(mock.ExpectationsWereMet())
	})

	t.Run("tenant admin gets own tenant user", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newGetUserLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		now := time.Now()
		mock.ExpectQuery("select .+ from `user` where `username` = \\? and `deleted_at` is null").
			WithArgs("alice").
			WillReturnRows(userRow(now, 8, "alice", 7))

		resp, err := l.GetUser(req)
		ast.NoError(err)
		ast.Equal(uint64(8), resp.Id)
		ast.Equal(uint64(7), resp.TenantId)
		ast.NoError(mock.ExpectationsWereMet())
	})

	t.Run("super admin can get any tenant user", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newGetUserLogicWithMock(t, operatorCtx(session.RoleTypeSuperAdmin, 1))
		now := time.Now()
		mock.ExpectQuery("select .+ from `user` where `username` = \\? and `deleted_at` is null").
			WithArgs("alice").
			WillReturnRows(userRow(now, 8, "alice", 7))

		resp, err := l.GetUser(req)
		ast.NoError(err)
		ast.Equal("alice", resp.Username)
		ast.NoError(mock.ExpectationsWereMet())
	})
}
