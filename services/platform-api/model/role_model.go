package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ RoleModel = (*customRoleModel)(nil)

type (
	// RoleModel is an interface to be customized, add more methods here,
	// and implement the added methods in customRoleModel.
	RoleModel interface {
		roleModel
		FindOneByRoleType(ctx context.Context, roleType string) (*Role, error)
		FindOneByRoleTypeAndTenantId(ctx context.Context, roleType string, tenantId uint64) (*Role, error)
		FindPage(ctx context.Context, tenantId uint64, roleName string, pageIndex, pageSize int64) (int64, []*Role, error)
		withSession(session sqlx.Session) RoleModel
	}

	customRoleModel struct {
		*defaultRoleModel
	}
)

// NewRoleModel returns a model for the database table.
func NewRoleModel(conn sqlx.SqlConn) RoleModel {
	return &customRoleModel{
		defaultRoleModel: newRoleModel(conn),
	}
}

func (m *customRoleModel) withSession(session sqlx.Session) RoleModel {
	return NewRoleModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customRoleModel) Insert(ctx context.Context, data *Role) (sql.Result, error) {
	now := time.Now()
	data.CreatedAt = sql.NullTime{Time: now, Valid: true}
	data.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	data.DeletedAt = sql.NullTime{}
	if data.RoleType == "" {
		data.RoleType = "User"
	}
	if data.Enable == "" {
		data.Enable = "Enable"
	}
	query := fmt.Sprintf("insert into %s (`created_at`, `updated_at`, `role_type`, `role_name`, `enable`, `tenant_id`) values (?, ?, ?, ?, ?, ?)", m.table)
	return m.conn.ExecCtx(ctx, query, data.CreatedAt, data.UpdatedAt, data.RoleType, data.RoleName, data.Enable, data.TenantId)
}

func (m *customRoleModel) FindOne(ctx context.Context, id uint64) (*Role, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ? and `deleted_at` is null limit 1", roleRows, m.table)
	var resp Role
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customRoleModel) Update(ctx context.Context, data *Role) error {
	data.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}
	query := fmt.Sprintf("update %s set `updated_at` = ?, `role_type` = ?, `role_name` = ?, `enable` = ?, `tenant_id` = ? where `id` = ? and `deleted_at` is null", m.table)
	_, err := m.conn.ExecCtx(ctx, query, data.UpdatedAt, data.RoleType, data.RoleName, data.Enable, data.TenantId, data.Id)
	return err
}

func (m *customRoleModel) Delete(ctx context.Context, id uint64) error {
	now := time.Now()
	query := fmt.Sprintf("update %s set `deleted_at` = ?, `updated_at` = ? where `id` = ? and `deleted_at` is null", m.table)
	result, err := m.conn.ExecCtx(ctx, query, now, now, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (m *customRoleModel) FindOneByRoleType(ctx context.Context, roleType string) (*Role, error) {
	query := fmt.Sprintf("select %s from %s where `role_type` = ? and `deleted_at` is null limit 1", roleRows, m.table)
	var resp Role
	err := m.conn.QueryRowCtx(ctx, &resp, query, roleType)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customRoleModel) FindOneByRoleTypeAndTenantId(ctx context.Context, roleType string, tenantId uint64) (*Role, error) {
	query := fmt.Sprintf("select %s from %s where `role_type` = ? and `tenant_id` = ? and `deleted_at` is null limit 1", roleRows, m.table)
	var resp Role
	err := m.conn.QueryRowCtx(ctx, &resp, query, roleType, tenantId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customRoleModel) FindPage(ctx context.Context, tenantId uint64, roleName string, pageIndex, pageSize int64) (int64, []*Role, error) {
	var (
		conds []string
		args  []any
	)
	conds = append(conds, "`deleted_at` is null")
	if tenantId > 0 {
		conds = append(conds, "`tenant_id` = ?")
		args = append(args, tenantId)
	}
	if roleName != "" {
		conds = append(conds, "`role_name` like ?")
		args = append(args, "%"+roleName+"%")
	}
	where := " where " + strings.Join(conds, " and ")

	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s%s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return 0, nil, err
	}

	offset := (pageIndex - 1) * pageSize
	listQuery := fmt.Sprintf("select %s from %s%s order by `updated_at` desc limit ? offset ?", roleRows, m.table, where)
	listArgs := append(append([]any{}, args...), pageSize, offset)

	var list []*Role
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, listArgs...); err != nil {
		return 0, nil, err
	}
	return total, list, nil
}
