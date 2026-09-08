package tenant

import (
	"database/sql"
	"time"

	"iot-zero/services/platform-api/internal/authz"
	"iot-zero/services/platform-api/internal/types"
	"iot-zero/services/platform-api/model"
)

var (
	errUnauthorized   = authz.ErrUnauthorized
	errForbidden      = authz.ErrForbidden
	errTenantNotFound = authz.NotFound("tenant")
)

func toTenantInfo(t *model.Tenants) types.TenantInfo {
	return types.TenantInfo{
		Id:         t.Id,
		TenantName: t.TenantName,
		Email:      nullString(t.Email),
		CreatedAt:  formatNullTime(t.CreatedAt),
		UpdatedAt:  formatNullTime(t.UpdatedAt),
	}
}

func nullString(v sql.NullString) string {
	if v.Valid {
		return v.String
	}
	return ""
}

func formatNullTime(v sql.NullTime) string {
	if !v.Valid {
		return ""
	}
	return v.Time.UTC().Format(time.RFC3339Nano)
}

func emailToNullString(email string) sql.NullString {
	if email == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: email, Valid: true}
}
