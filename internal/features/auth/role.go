package auth

import "cnmt/internal/infra/db"

type Role string

const RoleAdmin Role = "admin"

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin:
		return true
	default:
		return false
	}
}

func roleFromDB(role db.UserRole) Role {
	return Role(role)
}

func roleToDB(role Role) db.UserRole {
	return db.UserRole(role)
}

func hasRole(userRole Role, allowed ...Role) bool {
	for _, role := range allowed {
		if userRole == role {
			return true
		}
	}
	return false
}
