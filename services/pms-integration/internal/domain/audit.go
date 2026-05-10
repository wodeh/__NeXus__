// Package domain provides audit trail aggregates.
package domain

import (
	"time"
)

// AuditLog captures every significant system action.
type AuditLog struct {
	ID          string
	TenantID    string
	UserID      *string
	GuestID     *string
	Action      string // create, update, delete, view, export, login, logout
	Resource    string // reservation, guest, folio, rate_plan, etc.
	ResourceID  string
	OldValues   map[string]interface{}
	NewValues   map[string]interface{}
	IPAddress   string
	UserAgent   string
	CreatedAt   time.Time
}

// DataRetentionPolicy defines GDPR/data retention rules.
type DataRetentionPolicy struct {
	ID              string
	TenantID        string
	GuestDataMonths int      // Delete guest personal data after X months
	ReservationMonths int    // Keep reservation summaries after X months
	FolioYears      int      // Keep folio records for X years
	AuditLogMonths  int      // Keep audit logs for X months
	AutoPurgeEnabled bool
	LastPurgeRun    *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// GDPRRequest tracks data subject requests.
type GDPRRequest struct {
	ID          string
	TenantID    string
	GuestID     string
	Type        string // access, deletion, portability, rectification
	Status      string // pending, processing, completed, rejected
	RequestData map[string]interface{}
	CompletedAt *time.Time
	CreatedAt   time.Time
}
