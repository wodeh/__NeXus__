package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserWithRole extends User with tenant-specific role info.
type UserWithRole struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	Role       string    `json:"role"`
	IsActive   bool      `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// RolePermission defines what actions a role can perform.
type RolePermission struct {
	Role       string   `json:"role"`
	Actions    []string `json:"actions"`
	Resources  []string `json:"resources"`
}

// SystemConfig holds tenant-wide settings.
type SystemConfig struct {
	TenantID           uuid.UUID `json:"tenant_id"`
	DefaultCheckInTime string    `json:"default_check_in_time"`
	DefaultCheckOutTime string   `json:"default_check_out_time"`
	AutoConfirm        bool      `json:"auto_confirm"`
	RequireDeposit     bool      `json:"require_deposit"`
	DepositPercent     int       `json:"deposit_percent"`
	AllowWalkIn        bool      `json:"allow_walk_in"`
	OverbookingEnabled bool      `json:"overbooking_enabled"`
	UpdatedAt          time.Time `json:"updated_at"`
}
