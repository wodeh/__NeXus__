package domain

import (
	"time"

	"github.com/google/uuid"
)

// SmartLock represents an OrbitaTech smart lock installed on a room door.
type SmartLock struct {
	ID                  uuid.UUID `json:"id"`
	TenantID            uuid.UUID `json:"tenant_id"`
	RoomID              uuid.UUID `json:"room_id"`
	RoomNumber          string    `json:"room_number"`
	SerialNumber        string    `json:"serial_number"`
	Model               string    `json:"model"`
	Manufacturer        string    `json:"manufacturer"`
	Status              string    `json:"status"` // online, offline, low_battery, warning
	BatteryLevel        int       `json:"battery_level"`
	LastCommunicationAt *time.Time `json:"last_communication_at,omitempty"`
	LastUnlockAt        *time.Time `json:"last_unlock_at,omitempty"`
	LastLockAt          *time.Time `json:"last_lock_at,omitempty"`
	FirmwareVersion     string    `json:"firmware_version"`
	RemoteUnlockEnabled bool      `json:"remote_unlock_enabled"`
	AutoLockEnabled     bool      `json:"auto_lock_enabled"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// LockEvent represents an audit event from a smart lock.
type LockEvent struct {
	ID           uuid.UUID `json:"id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	LockID       uuid.UUID `json:"lock_id"`
	RoomNumber   string    `json:"room_number"`
	EventType    string    `json:"event_type"`    // unlock, lock, code_created, code_deleted, alert
	EventSource  string    `json:"event_source"`  // guest_keycard, staff_master, remote, auto_lock, access_code
	Details      string    `json:"details"`
	OccurredAt   time.Time `json:"occurred_at"`
}

// AccessCode represents a temporary access code for a smart lock.
type AccessCode struct {
	ID         uuid.UUID  `json:"id"`
	TenantID   uuid.UUID  `json:"tenant_id"`
	LockID     uuid.UUID  `json:"lock_id"`
	Code       string     `json:"code"`
	Label      string     `json:"label"`
	IsActive   bool       `json:"is_active"`
	ValidFrom  time.Time  `json:"valid_from"`
	ValidUntil *time.Time `json:"valid_until,omitempty"`
	MaxUses    *int       `json:"max_uses,omitempty"`
	UseCount   int        `json:"use_count"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// LockStatusUpdateRequest updates a lock's status.
type LockStatusUpdateRequest struct {
	Status              string `json:"status"`
	BatteryLevel        int    `json:"battery_level,omitempty"`
	RemoteUnlockEnabled *bool  `json:"remote_unlock_enabled,omitempty"`
	AutoLockEnabled     *bool  `json:"auto_lock_enabled,omitempty"`
}

// CreateAccessCodeRequest creates a new access code.
type CreateAccessCodeRequest struct {
	Code       string     `json:"code"`
	Label      string     `json:"label"`
	ValidFrom  time.Time  `json:"valid_from"`
	ValidUntil *time.Time `json:"valid_until,omitempty"`
	MaxUses    *int       `json:"max_uses,omitempty"`
}

// RemoteUnlockRequest triggers a remote unlock.
type RemoteUnlockRequest struct {
	Reason string `json:"reason,omitempty"`
}
