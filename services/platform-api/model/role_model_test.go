package model

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func newMockRoleModel(t *testing.T) (RoleModel, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewRoleModel(sqlx.NewSqlConnFromDB(db)), mock
}

func TestRoleModel_InsertAndFindOne(t *testing.T) {
	ast := assert.New(t)
	m, mock := newMockRoleModel(t)
	ctx := context.Background()

	mock.ExpectExec("insert into `role`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "User", "admin", "Enable", uint64(1)).
		WillReturnResult(sqlmock.NewResult(9, 1))

	result, err := m.Insert(ctx, &Role{
		RoleType: "",
		RoleName: "admin",
		Enable:   "Enable",
		TenantId: 1,
	})
	ast.NoError(err)
	id, err := result.LastInsertId()
	ast.NoError(err)
	ast.Equal(int64(9), id)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "role_type", "role_name", "enable", "tenant_id"}).
		AddRow(uint64(9), now, now, nil, "9", "admin", "Enable", uint64(1))
	mock.ExpectQuery("select .+ from `role` where `id` = \\? and `deleted_at` is null").
		WithArgs(uint64(9)).
		WillReturnRows(rows)

	got, err := m.FindOne(ctx, 9)
	ast.NoError(err)
	ast.Equal("9", got.RoleType)
	ast.Equal("admin", got.RoleName)
}

func TestRoleModel_FindOneByRoleType(t *testing.T) {
	ast := assert.New(t)
	m, mock := newMockRoleModel(t)
	ctx := context.Background()
	now := time.Now()

	mock.ExpectQuery("select .+ from `role` where `role_type` = \\? and `deleted_at` is null").
		WithArgs("admin").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "role_type", "role_name", "enable", "tenant_id"}).
			AddRow(uint64(2), now, now, nil, "admin", "Admin", "Enable", uint64(1)))

	got, err := m.FindOneByRoleType(ctx, "admin")
	ast.NoError(err)
	ast.Equal(uint64(2), got.Id)
	ast.Equal("admin", got.RoleType)

	mock.ExpectQuery("select .+ from `role` where `role_type` = \\? and `deleted_at` is null").
		WithArgs("missing").
		WillReturnError(sqlx.ErrNotFound)
	_, err = m.FindOneByRoleType(ctx, "missing")
	ast.ErrorIs(err, ErrNotFound)
}

func TestRoleModel_FindOneByRoleTypeAndTenantId(t *testing.T) {
	ast := assert.New(t)
	m, mock := newMockRoleModel(t)
	ctx := context.Background()
	now := time.Now()

	mock.ExpectQuery("select .+ from `role` where `role_type` = \\? and `tenant_id` = \\? and `deleted_at` is null").
		WithArgs("TenantAdmin", uint64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "role_type", "role_name", "enable", "tenant_id"}).
			AddRow(uint64(3), now, now, nil, "TenantAdmin", "tenant administrator", "Enable", uint64(7)))

	got, err := m.FindOneByRoleTypeAndTenantId(ctx, "TenantAdmin", 7)
	ast.NoError(err)
	ast.Equal(uint64(3), got.Id)
	ast.Equal(uint64(7), got.TenantId)

	mock.ExpectQuery("select .+ from `role` where `role_type` = \\? and `tenant_id` = \\? and `deleted_at` is null").
		WithArgs("TenantAdmin", uint64(99)).
		WillReturnError(sqlx.ErrNotFound)
	_, err = m.FindOneByRoleTypeAndTenantId(ctx, "TenantAdmin", 99)
	ast.ErrorIs(err, ErrNotFound)
}

func TestRoleModel_FindPage(t *testing.T) {
	ast := assert.New(t)
	m, mock := newMockRoleModel(t)
	ctx := context.Background()

	mock.ExpectQuery("select count\\(\\*\\) from `role` where `deleted_at` is null and `tenant_id` = \\? and `role_name` like \\?").
		WithArgs(uint64(1), "%admin%").
		WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(1))

	now := time.Now()
	mock.ExpectQuery("select .+ from `role` where `deleted_at` is null and `tenant_id` = \\? and `role_name` like \\? order by `updated_at` desc limit \\? offset \\?").
		WithArgs(uint64(1), "%admin%", int64(20), int64(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "role_type", "role_name", "enable", "tenant_id"}).
			AddRow(uint64(1), now, now, nil, "1", "admin", "Enable", uint64(1)))

	total, list, err := m.FindPage(ctx, 1, "admin", 1, 20)
	ast.NoError(err)
	ast.Equal(int64(1), total)
	ast.Len(list, 1)
	ast.Equal("admin", list[0].RoleName)
}

func TestRoleModel_SoftDelete(t *testing.T) {
	ast := assert.New(t)
	m, mock := newMockRoleModel(t)
	ctx := context.Background()

	mock.ExpectExec(fmt.Sprintf("update %s set `deleted_at`", "`role`")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), uint64(3)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	ast.NoError(m.Delete(ctx, 3))

	mock.ExpectExec(fmt.Sprintf("update %s set `deleted_at`", "`role`")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), uint64(99)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	ast.ErrorIs(m.Delete(ctx, 99), ErrNotFound)
}
