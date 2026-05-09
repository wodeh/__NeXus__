package auth

import (
	"context"
	"fmt"
)

// Permission is a granular capability string.
type Permission string

// Role is a named collection of permissions.
type Role string

// Built-in role definitions.
var rolePermissions = map[Role][]Permission{
	"admin":      {"*"},
	"executive":  {"read:analytics", "read:tenant", "read:property"},
	"manager":    {"read:tenant", "write:tenant", "read:property", "write:property", "read:reservation", "write:reservation"},
	"front_desk": {"read:reservation", "write:reservation", "read:guest"},
	"housekeeping":{"read:room", "write:room"},
	"guest":      {"read:own_reservation", "read:own_profile"},
	"msp_engineer":{"read:system", "write:system"},
}

// RBACEngine evaluates role-based access control.
type RBACEngine struct {
	permissions map[Role][]Permission
}

// NewRBACEngine creates an RBAC engine with default roles.
func NewRBACEngine() *RBACEngine {
	return &RBACEngine{permissions: rolePermissions}
}

// HasPermission returns true if the role grants the requested permission.
func (e *RBACEngine) HasPermission(role Role, perm Permission) bool {
	perms, ok := e.permissions[role]
	if !ok {
		return false
	}
	for _, p := range perms {
		if p == "*" || p == perm {
			return true
		}
	}
	return false
}

// Enforce returns an error if the role does not have the permission.
func (e *RBACEngine) Enforce(ctx context.Context, role Role, perm Permission) error {
	if !e.HasPermission(role, perm) {
		return fmt.Errorf("role %s lacks permission %s", role, perm)
	}
	return nil
}
