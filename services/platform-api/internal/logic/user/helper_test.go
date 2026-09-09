package user

import (
	"context"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func operatorCtx(roleType string, tenantId uint64) context.Context {
	return operatorUserCtx(roleType, tenantId, "operator")
}

func operatorUserCtx(roleType string, tenantId uint64, username string) context.Context {
	ctx := context.WithValue(context.Background(), "roleType", roleType)
	ctx = context.WithValue(ctx, "tenantId", float64(tenantId))
	return context.WithValue(ctx, "username", username)
}

func userColumns() []string {
	return []string{"id", "created_at", "updated_at", "deleted_at", "username", "password", "salt", "enable", "role_id", "tenant_id"}
}

func userRow(now time.Time, id uint64, username string, tenantId uint64) *sqlmock.Rows {
	return sqlmock.NewRows(userColumns()).
		AddRow(id, now, now, nil, username, "hash", "salt", "Enable", uint64(2), tenantId)
}

func roleRow(now time.Time, id uint64, tenantId uint64) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "role_type", "role_name", "enable", "tenant_id"}).
		AddRow(id, now, now, nil, "User", "operator", "Enable", tenantId)
}
