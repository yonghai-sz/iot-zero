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

func newUpdateRoleLogicWithMock(t *testing.T, ctx context.Context) (*UpdateRoleLogic, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	conn := sqlx.NewSqlConnFromDB(db)
	return NewUpdateRoleLogic(ctx, &svc.ServiceContext{
		RoleModel: model.NewRoleModel(conn),
	}), mock
}

func TestUpdateRoleLogic_UpdateRole(t *testing.T) {
	req := &types.UpdateRoleReq{Id: 4, RoleName: "renamed"}

	t.Run("unauthorized without role", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newUpdateRoleLogicWithMock(t, context.Background())
		_, err := l.UpdateRole(req)
		ast.ErrorIs(err, errUnauthorized)
	})

	t.Run("user is forbidden", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newUpdateRoleLogicWithMock(t, operatorCtx(session.RoleTypeUser, 7))
		_, err := l.UpdateRole(req)
		ast.ErrorIs(err, errForbidden)
	})

	t.Run("missing role", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newUpdateRoleLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		mock.ExpectQuery("select .+ from `role` where `id` = \\? and `deleted_at` is null").
			WithArgs(uint64(4)).
			WillReturnError(sqlx.ErrNotFound)

		_, err := l.UpdateRole(req)
		ast.ErrorIs(err, errRoleNotFound)
		ast.NoError(mock.ExpectationsWereMet())
	})

	t.Run("tenant admin cannot update another tenant", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newUpdateRoleLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		now := time.Now()
		mock.ExpectQuery("select .+ from `role` where `id` = \\? and `deleted_at` is null").
			WithArgs(uint64(4)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "role_type", "role_name", "enable", "tenant_id"}).
				AddRow(uint64(4), now, now, nil, session.RoleTypeUser, "operator", "Enable", uint64(3)))

		_, err := l.UpdateRole(req)
		ast.ErrorIs(err, errRoleNotFound)
		ast.NoError(mock.ExpectationsWereMet())
	})

	t.Run("tenant admin updates own tenant role", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newUpdateRoleLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		now := time.Now()
		roleRows := []string{"id", "created_at", "updated_at", "deleted_at", "role_type", "role_name", "enable", "tenant_id"}
		mock.ExpectQuery("select .+ from `role` where `id` = \\? and `deleted_at` is null").
			WithArgs(uint64(4)).
			WillReturnRows(sqlmock.NewRows(roleRows).
				AddRow(uint64(4), now, now, nil, session.RoleTypeUser, "operator", "Enable", uint64(7)))
		mock.ExpectExec("update `role` set `updated_at`").
			WithArgs(sqlmock.AnyArg(), session.RoleTypeUser, "renamed", "Enable", uint64(7), uint64(4)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery("select .+ from `role` where `id` = \\? and `deleted_at` is null").
			WithArgs(uint64(4)).
			WillReturnRows(sqlmock.NewRows(roleRows).
				AddRow(uint64(4), now, now, nil, session.RoleTypeUser, "renamed", "Enable", uint64(7)))

		resp, err := l.UpdateRole(req)
		ast.NoError(err)
		ast.Equal("renamed", resp.RoleName)
		ast.NoError(mock.ExpectationsWereMet())
	})

	t.Run("super admin can update any tenant role", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newUpdateRoleLogicWithMock(t, operatorCtx(session.RoleTypeSuperAdmin, 1))
		now := time.Now()
		roleRows := []string{"id", "created_at", "updated_at", "deleted_at", "role_type", "role_name", "enable", "tenant_id"}
		mock.ExpectQuery("select .+ from `role` where `id` = \\? and `deleted_at` is null").
			WithArgs(uint64(4)).
			WillReturnRows(sqlmock.NewRows(roleRows).
				AddRow(uint64(4), now, now, nil, session.RoleTypeUser, "operator", "Enable", uint64(7)))
		mock.ExpectExec("update `role` set `updated_at`").
			WithArgs(sqlmock.AnyArg(), session.RoleTypeUser, "renamed", "Enable", uint64(7), uint64(4)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery("select .+ from `role` where `id` = \\? and `deleted_at` is null").
			WithArgs(uint64(4)).
			WillReturnRows(sqlmock.NewRows(roleRows).
				AddRow(uint64(4), now, now, nil, session.RoleTypeUser, "renamed", "Enable", uint64(7)))

		resp, err := l.UpdateRole(req)
		ast.NoError(err)
		ast.Equal(uint64(7), resp.TenantId)
		ast.NoError(mock.ExpectationsWereMet())
	})
}
