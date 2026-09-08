package role

import (
	"database/sql"
	"time"

	"iot-zero/services/platform-api/internal/authz"
	"iot-zero/services/platform-api/internal/types"
	"iot-zero/services/platform-api/model"
)

var (
	errUnauthorized = authz.ErrUnauthorized
	errForbidden    = authz.ErrForbidden
	errRoleNotFound = authz.NotFound("role")
)

func toRoleInfo(r *model.Role) types.RoleInfo {
	return types.RoleInfo{
		Id:        r.Id,
		RoleType:  r.RoleType,
		RoleName:  r.RoleName,
		Enable:    enableToBool(r.Enable),
		TenantId:  r.TenantId,
		CreatedAt: formatNullTime(r.CreatedAt),
		UpdatedAt: formatNullTime(r.UpdatedAt),
	}
}

func formatNullTime(v sql.NullTime) string {
	if !v.Valid {
		return ""
	}
	return v.Time.UTC().Format(time.RFC3339Nano)
}

func boolToEnable(enable bool) string {
	if enable {
		return "Enable"
	}
	return "Disable"
}

func enableToBool(enable string) bool {
	return enable != "Disable"
}

func toMenuInfo(m *model.Menu) types.MenuInfo {
	return types.MenuInfo{
		Id:       m.Id,
		MenuKey:  m.MenuKey,
		Title:    m.Title,
		ParentId: m.ParentId,
		HasChild: m.HasChild,
	}
}
