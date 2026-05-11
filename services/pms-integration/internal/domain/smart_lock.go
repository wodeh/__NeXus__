package domain

import (
	"time"
)

// SmartLock represents a physical door lock device
type SmartLock struct {
	ID                   string     `json:"id" db:"id"`
	TenantID             string     `json:"tenant_id" db:"tenant_id"`
	PropertyID           string     `json:"property_id" db:"property_id"`
	RoomID               *string    `json:"room_id,omitempty" db:"room_id"`
	DeviceName           string     `json:"device_name" db:"device_name"`
	DeviceSerial         string     `json:"device_serial" db:"device_serial"`
	Manufacturer         string     `json:"manufacturer" db:"manufacturer"`
	Model                string     `json:"model" db:"model"`
	FirmwareVersion      string     `json:"firmware_version,omitempty" db:"firmware_version"`
	Status               string     `json:"status" db:"status"`
	BatteryLevel         int        `json:"battery_level" db:"battery_level"`
	ConnectionType       string     `json:"connection_type" db:"connection_type"`
	IPAddress            string     `json:"ip_address,omitempty" db:"ip_address"`
	MACAddress           string     `json:"mac_address,omitempty" db:"mac_address"`
	LastCommunicationAt  *time.Time `json:"last_communication_at,omitempty" db:"last_communication_at"`
	LastUnlockAt         *time.Time `json:"last_unlock_at,omitempty" db:"last_unlock_at"`
	LastLockAt           *time.Time `json:"last_lock_at,omitempty" db:"last_lock_at"`
	RemoteUnlockEnabled  bool       `json:"remote_unlock_enabled" db:"remote_unlock_enabled"`
	AutoLockEnabled      bool       `json:"auto_lock_enabled" db:"auto_lock_enabled"`
	AutoLockDelaySeconds int        `json:"auto_lock_delay_seconds" db:"auto_lock_delay_seconds"`
	Settings             JSONMap    `json:"settings,omitempty" db:"settings"`
	Metadata             JSONMap    `json:"metadata,omitempty" db:"metadata"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
}

// LockAccessCode represents a generated access code for a lock
type LockAccessCode struct {
	ID            string     `json:"id" db:"id"`
	TenantID      string     `json:"tenant_id" db:"tenant_id"`
	PropertyID    string     `json:"property_id" db:"property_id"`
	SmartLockID   string     `json:"smart_lock_id" db:"smart_lock_id"`
	ReservationID *string    `json:"reservation_id,omitempty" db:"reservation_id"`
	GuestID       *string    `json:"guest_id,omitempty" db:"guest_id"`
	Code          string     `json:"code" db:"code"`
	CodeType      string     `json:"code_type" db:"code_type"`
	Label         string     `json:"label,omitempty" db:"label"`
	ValidFrom     time.Time  `json:"valid_from" db:"valid_from"`
	ValidUntil    *time.Time `json:"valid_until,omitempty" db:"valid_until"`
	MaxUses       *int       `json:"max_uses,omitempty" db:"max_uses"`
	UsesCount     int        `json:"uses_count" db:"uses_count"`
	IsActive      bool       `json:"is_active" db:"is_active"`
	IsMaster      bool       `json:"is_master" db:"is_master"`
	CreatedBy     *string    `json:"created_by,omitempty" db:"created_by"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	RevokedAt     *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
	RevokedReason string     `json:"revoked_reason,omitempty" db:"revoked_reason"`
	Metadata      JSONMap    `json:"metadata,omitempty" db:"metadata"`
}

// LockEvent represents an activity log entry for a smart lock
type LockEvent struct {
	ID           string    `json:"id" db:"id"`
	TenantID     string    `json:"tenant_id" db:"tenant_id"`
	PropertyID   string    `json:"property_id" db:"property_id"`
	SmartLockID  string    `json:"smart_lock_id" db:"smart_lock_id"`
	AccessCodeID *string   `json:"access_code_id,omitempty" db:"access_code_id"`
	EventType    string    `json:"event_type" db:"event_type"`
	EventSource  string    `json:"event_source" db:"event_source"`
	UserID       *string   `json:"user_id,omitempty" db:"user_id"`
	GuestID      *string   `json:"guest_id,omitempty" db:"guest_id"`
	Details      JSONMap   `json:"details,omitempty" db:"details"`
	OccurredAt   time.Time `json:"occurred_at" db:"occurred_at"`
}

// LockCommand represents a command sent to a lock
type LockCommand struct {
	ID        string    `json:"id" db:"id"`
	LockID    string    `json:"lock_id" db:"lock_id"`
	Command   string    `json:"command" db:"command"`
	Status    string    `json:"status" db:"status"`
	Result    JSONMap   `json:"result,omitempty" db:"result"`
	SentAt    time.Time `json:"sent_at" db:"sent_at"`
	AckedAt   *time.Time `json:"acked_at,omitempty" db:"acked_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`
}

// LockStatusUpdate is used for real-time status pushes
type LockStatusUpdate struct {
	LockID       string     `json:"lock_id"`
	Status       string     `json:"status"`
	BatteryLevel int        `json:"battery_level"`
	IsLocked     bool       `json:"is_locked"`
	LastActivity *time.Time `json:"last_activity,omitempty"`
	Timestamp    time.Time  `json:"timestamp"`
}

// LockOverview aggregates lock data for dashboard
type LockOverview struct {
	TotalLocks        int `json:"total_locks"`
	OnlineLocks       int `json:"online_locks"`
	OfflineLocks      int `json:"offline_locks"`
	LowBatteryLocks   int `json:"low_battery_locks"`
	ActiveAccessCodes int `json:"active_access_codes"`
	TodayEvents       int `json:"today_events"`
	RecentUnlocks     int `json:"recent_unlocks"`
}
