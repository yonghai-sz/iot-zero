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

func newCreateUserLogicWithMock(t *testing.T, ctx context.Context) (*CreateUserLogic, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	conn := sqlx.NewSqlConnFromDB(db)
	return NewCreateUserLogic(ctx, &svc.ServiceContext{
		RoleModel: model.NewRoleModel(conn),
		UserModel: model.NewUserModel(conn),
	}), mock
}

func TestCreateUserLogic_CreateUser(t *testing.T) {
	req := &types.CreateUserReq{TenantId: 7, Username: "alice", Password: "secret", RoleId: 2, Enable: true}

	t.Run("unauthorized without role", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newCreateUserLogicWithMock(t, context.Background())
		_, err := l.CreateUser(req)
		ast.ErrorIs(err, errUnauthorized)
	})

	t.Run("user is forbidden", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newCreateUserLogicWithMock(t, operatorCtx(session.RoleTypeUser, 7))
		_, err := l.CreateUser(req)
		ast.ErrorIs(err, errForbidden)
	})

	t.Run("tenant admin cannot create for another tenant", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newCreateUserLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 3))
		_, err := l.CreateUser(req)
		ast.ErrorIs(err, errForbidden)
	})

	t.Run("tenant admin creates for own tenant", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newCreateUserLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		now := time.Now()
		mock.ExpectQuery("select .+ from `role` where `id` = \\? and `deleted_at` is null").
			WithArgs(uint64(2)).
			WillReturnRows(roleRow(now, 2, 7))
		mock.ExpectQuery("select .+ from `user` where `username` = \\? and `deleted_at` is null").
			WithArgs("alice").
			WillReturnError(sqlx.ErrNotFound)
		mock.ExpectExec("insert into `user`").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "alice", sqlmock.AnyArg(), sqlmock.AnyArg(), "Enable", uint64(2), uint64(7)).
			WillReturnResult(sqlmock.NewResult(8, 1))

		resp, err := l.CreateUser(req)
		ast.NoError(err)
		ast.Equal(uint64(8), resp.Id)
		ast.NoError(mock.ExpectationsWereMet())
	})

	t.Run("super admin can create for any tenant", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newCreateUserLogicWithMock(t, operatorCtx(session.RoleTypeSuperAdmin, 1))
		now := time.Now()
		mock.ExpectQuery("select .+ from `role` where `id` = \\? and `deleted_at` is null").
			WithArgs(uint64(2)).
			WillReturnRows(roleRow(now, 2, 7))
		mock.ExpectQuery("select .+ from `user` where `username` = \\? and `deleted_at` is null").
			WithArgs("alice").
			WillReturnError(sqlx.ErrNotFound)
		mock.ExpectExec("insert into `user`").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "alice", sqlmock.AnyArg(), sqlmock.AnyArg(), "Enable", uint64(2), uint64(7)).
			WillReturnResult(sqlmock.NewResult(9, 1))

		resp, err := l.CreateUser(req)
		ast.NoError(err)
		ast.Equal(uint64(9), resp.Id)
		ast.NoError(mock.ExpectationsWereMet())
	})
}
