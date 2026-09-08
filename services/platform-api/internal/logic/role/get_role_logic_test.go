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

func newGetRoleLogicWithMock(t *testing.T, ctx context.Context) (*GetRoleLogic, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	conn := sqlx.NewSqlConnFromDB(db)
	return NewGetRoleLogic(ctx, &svc.ServiceContext{
		RoleModel: model.NewRoleModel(conn),
	}), mock
}

func TestGetRoleLogic_GetRole(t *testing.T) {
	req := &types.RoleIdPathReq{Id: 4}

	t.Run("unauthorized without role", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newGetRoleLogicWithMock(t, context.Background())
		_, err := l.GetRole(req)
		ast.ErrorIs(err, errUnauthorized)
	})

	t.Run("user is forbidden", func(t *testing.T) {
		ast := assert.New(t)
		l, _ := newGetRoleLogicWithMock(t, operatorCtx(session.RoleTypeUser, 7))
		_, err := l.GetRole(req)
		ast.ErrorIs(err, errForbidden)
	})

	t.Run("missing role", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newGetRoleLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		mock.ExpectQuery("select .+ from `role` where `id` = \\? and `deleted_at` is null").
			WithArgs(uint64(4)).
			WillReturnError(sqlx.ErrNotFound)

		_, err := l.GetRole(req)
		ast.ErrorIs(err, errRoleNotFound)
		ast.NoError(mock.ExpectationsWereMet())
	})

	t.Run("tenant admin cannot see another tenant", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newGetRoleLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		now := time.Now()
		mock.ExpectQuery("select .+ from `role` where `id` = \\? and `deleted_at` is null").
			WithArgs(uint64(4)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "role_type", "role_name", "enable", "tenant_id"}).
				AddRow(uint64(4), now, now, nil, session.RoleTypeUser, "operator", "Enable", uint64(3)))

		_, err := l.GetRole(req)
		ast.ErrorIs(err, errRoleNotFound)
		ast.NoError(mock.ExpectationsWereMet())
	})

	t.Run("tenant admin gets own tenant role", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newGetRoleLogicWithMock(t, operatorCtx(session.RoleTypeTenantAdmin, 7))
		now := time.Now()
		mock.ExpectQuery("select .+ from `role` where `id` = \\? and `deleted_at` is null").
			WithArgs(uint64(4)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "role_type", "role_name", "enable", "tenant_id"}).
				AddRow(uint64(4), now, now, nil, session.RoleTypeUser, "operator", "Enable", uint64(7)))

		resp, err := l.GetRole(req)
		ast.NoError(err)
		ast.Equal(uint64(4), resp.Id)
		ast.Equal(uint64(7), resp.TenantId)
		ast.NoError(mock.ExpectationsWereMet())
	})

	t.Run("super admin can get any tenant role", func(t *testing.T) {
		ast := assert.New(t)
		l, mock := newGetRoleLogicWithMock(t, operatorCtx(session.RoleTypeSuperAdmin, 1))
		now := time.Now()
		mock.ExpectQuery("select .+ from `role` where `id` = \\? and `deleted_at` is null").
			WithArgs(uint64(4)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "role_type", "role_name", "enable", "tenant_id"}).
				AddRow(uint64(4), now, now, nil, session.RoleTypeUser, "operator", "Enable", uint64(7)))

		resp, err := l.GetRole(req)
		ast.NoError(err)
		ast.Equal(uint64(4), resp.Id)
		ast.Equal(uint64(7), resp.TenantId)
		ast.NoError(mock.ExpectationsWereMet())
	})
}
