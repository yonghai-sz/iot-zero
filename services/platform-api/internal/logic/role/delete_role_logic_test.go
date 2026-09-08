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

func newDeleteRoleLogicWithMock(t *testing.T, ctx context.Context) (*DeleteRoleLogic, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	conn := sqlx.NewSqlConnFromDB(db)
	return NewDeleteRoleLogic(ctx, &svc.ServiceContext{
		RoleModel: model.NewRoleModel(conn),
	}), mock
}

func TestDeleteRoleLogic_DeleteRole(t *testing.T) {
	req := &types.RoleIdPathReq{Id: 4}

	t.Run("unauthorized without role", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newDeleteRoleLogicWithMock(t, context.Background())
		err := l.DeleteRole(req)
		ast.ErrorIs(err, errUnauthorized)
	})

	t.Run("user is forbidden", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newDeleteRoleLogicWithMock(t, operatorCtx(session.RoleTypeUser, 7))
		err := l.DeleteRole(req)
		ast.ErrorIs(err, errForbidden)
	})

	t.Run("missing role", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newDeleteRoleLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		mock.ExpectQuery("select .+ from `role` where `id` = \\? and `deleted_at` is null").
			WithArgs(uint64(4)).
			WillReturnError(sqlx.ErrNotFound)

		err := l.DeleteRole(req)
		ast.ErrorIs(err, errRoleNotFound)
		ast.NoError(mock.ExpectationsWereMet())
	})

	t.Run("tenant admin cannot delete another tenant", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newDeleteRoleLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		now := time.Now()
		mock.ExpectQuery("select .+ from `role` where `id` = \\? and `deleted_at` is null").
			WithArgs(uint64(4)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "role_type", "role_name", "enable", "tenant_id"}).
				AddRow(uint64(4), now, now, nil, session.RoleTypeUser, "operator", "Enable", uint64(3)))

		err := l.DeleteRole(req)
		ast.ErrorIs(err, errRoleNotFound)
		ast.NoError(mock.ExpectationsWereMet())
	})

	t.Run("tenant admin deletes own tenant role", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newDeleteRoleLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		now := time.Now()
		mock.ExpectQuery("select .+ from `role` where `id` = \\? and `deleted_at` is null").
			WithArgs(uint64(4)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "role_type", "role_name", "enable", "tenant_id"}).
				AddRow(uint64(4), now, now, nil, session.RoleTypeUser, "operator", "Enable", uint64(7)))
		mock.ExpectExec("update `role` set `deleted_at`").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), uint64(4)).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := l.DeleteRole(req)
		ast.NoError(err)
		ast.NoError(mock.ExpectationsWereMet())
	})

	t.Run("super admin can delete any tenant role", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newDeleteRoleLogicWithMock(t, operatorCtx(session.RoleTypeSuperAdmin, 1))
		now := time.Now()
		mock.ExpectQuery("select .+ from `role` where `id` = \\? and `deleted_at` is null").
			WithArgs(uint64(4)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "role_type", "role_name", "enable", "tenant_id"}).
				AddRow(uint64(4), now, now, nil, session.RoleTypeUser, "operator", "Enable", uint64(7)))
		mock.ExpectExec("update `role` set `deleted_at`").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), uint64(4)).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := l.DeleteRole(req)
		ast.NoError(err)
		ast.NoError(mock.ExpectationsWereMet())
	})
}
