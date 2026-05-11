package domain

import (
	"time"

	"github.com/google/uuid"
)

// AuditLog represents a system audit record.
type AuditLog struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	UserID     *uuid.UUID `json:"user_id,omitempty"`
	UserEmail  *string   `json:"user_email,omitempty"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`   // reservation, room, lock, user, setting
	ResourceID string    `json:"resource_id,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
	IPAddress  string    `json:"ip_address,omitempty"`
	UserAgent  string    `json:"user_agent,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// AuditStats holds audit metrics.
type AuditStats struct {
	TenantID        string `json:"tenant_id"`
	TotalEvents     int    `json:"total_events"`
	TodayEvents     int    `json:"today_events"`
	UserActions     int    `json:"user_actions"`
	SystemActions   int    `json:"system_actions"`
	FailedActions   int    `json:"failed_actions"`
}

// CreateAuditLogRequest creates an audit entry.
type CreateAuditLogRequest struct {
	UserID     string                 `json:"user_id,omitempty"`
	UserEmail  string                 `json:"user_email,omitempty"`
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource"`
	ResourceID string                 `json:"resource_id,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
	IPAddress  string                 `json:"ip_address,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
}
