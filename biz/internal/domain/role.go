package domain

const (
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
	RoleUser       = "user"
)

func NormalizeRole(role string) string {
	switch role {
	case RoleSuperAdmin, RoleAdmin, RoleUser:
		return role
	default:
		return RoleUser
	}
}

func IsAdminRole(role string) bool {
	return role == RoleSuperAdmin || role == RoleAdmin
}

func IsSuperAdminRole(role string) bool {
	return role == RoleSuperAdmin
}
