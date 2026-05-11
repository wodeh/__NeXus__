package domain

import (
	"time"

	"github.com/google/uuid"
)

// HousekeepingStaff represents a cleaning/maintenance worker.
type HousekeepingStaff struct {
	ID              uuid.UUID `json:"id"`
	TenantID        uuid.UUID `json:"tenant_id"`
	Name            string    `json:"name"`
	Role            string    `json:"role"` // cleaner, inspector, supervisor
	ActiveShift     string    `json:"active_shift"` // day, evening, night
	MaxRoomsPerDay  int       `json:"max_rooms_per_day"`
	Phone           *string   `json:"phone,omitempty"`
	Email           *string   `json:"email,omitempty"`
	Active          bool      `json:"active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
	Version         int       `json:"version"`
}

// HousekeepingTask represents a cleaning or maintenance task.
type HousekeepingTask struct {
	ID           uuid.UUID  `json:"id"`
	TenantID     uuid.UUID  `json:"tenant_id"`
	RoomNumber   string     `json:"room_number"`
	TaskType     string     `json:"task_type"` // full_clean, refill, maintenance, inspection
	Status       string     `json:"status"`    // pending, in_progress, completed, skipped, blocked
	AssignedTo   *uuid.UUID `json:"assigned_to,omitempty"`
	Priority     string     `json:"priority"` // low, normal, high, urgent
	Notes        *string    `json:"notes,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
	Version      int        `json:"version"`
}

// HousekeepingTaskCreateRequest is the payload for creating a task.
type HousekeepingTaskCreateRequest struct {
	RoomNumber string    `json:"room_number"`
	TaskType   string    `json:"task_type"`
	Priority   string    `json:"priority"`
	Notes      string    `json:"notes,omitempty"`
	AssignedTo *uuid.UUID `json:"assigned_to,omitempty"`
}

// HousekeepingTaskUpdateRequest updates a task.
type HousekeepingTaskUpdateRequest struct {
	Status     *string    `json:"status,omitempty"`
	AssignedTo *uuid.UUID `json:"assigned_to,omitempty"`
	Notes      *string    `json:"notes,omitempty"`
}
