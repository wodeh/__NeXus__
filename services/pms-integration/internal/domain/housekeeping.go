package domain

import (
	"time"
)

// HousekeepingTask represents a cleaning or maintenance task for a room
type HousekeepingTask struct {
	ID            string     `json:"id"`
	TenantID      string     `json:"tenant_id"`
	RoomID        string     `json:"room_id"`
	RoomNumber    string     `json:"room_number"`
	Floor         string     `json:"floor"`
	TaskType      string     `json:"task_type"` // clean, refill, maintenance, inspection
	Priority      int        `json:"priority"`  // 1=low, 2=normal, 3=high, 4=urgent, 5=vip
	Status        string     `json:"status"`    // pending, in_progress, paused, completed, cancelled
	AssignedTo    *string    `json:"assigned_to,omitempty"`
	AssignedBy    *string    `json:"assigned_by,omitempty"`
	Shift         string     `json:"shift"`     // morning, afternoon, evening, night
	EstimatedMin  int        `json:"estimated_minutes"`
	ActualMin     *int       `json:"actual_minutes,omitempty"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	ChecklistDone []string   `json:"checklist_done"`
	SuppliesUsed  []SupplyUsage `json:"supplies_used,omitempty"`
	Notes         *string    `json:"notes,omitempty"`
	DamageReport  *DamageReport `json:"damage_report,omitempty"`
	CardMode      *string    `json:"card_mode,omitempty"` // clean_full, clean_refill
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// SupplyUsage tracks what was consumed/restocked during cleaning
type SupplyUsage struct {
	Item     string `json:"item"`     // soap, shampoo, towel, linen, coffee, tea, sugar, minibar_item
	Quantity int    `json:"quantity"`
	Action   string `json:"action"`   // consumed, restocked
}

// DamageReport for logging room damage with photos
type DamageReport struct {
	Description string   `json:"description"`
	Severity    string   `json:"severity"` // minor, moderate, major
	Photos      []string `json:"photos,omitempty"`
	ReportedAt  time.Time `json:"reported_at"`
}

// HousekeepingStaff represents a housekeeping employee
type HousekeepingStaff struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id"`
	UserID         *string   `json:"user_id,omitempty"`
	Name           string    `json:"name"`
	Role           string    `json:"role"`         // cleaner, inspector, manager
	Active         bool      `json:"active"`
	Shift          string    `json:"shift"`        // morning, afternoon, evening, night
	Floors         []string  `json:"floors,omitempty"`
	MaxRoomsPerDay int       `json:"max_rooms_per_day"`
	CurrentLoad    int       `json:"current_load"`
	Rating         float64   `json:"rating"`       // avg quality score 1-5
	CompletedToday int       `json:"completed_today"`
	CreatedAt    time.Time `json:"created_at"`
}

// StaffPerformance holds daily/weekly stats for a staff member
type StaffPerformance struct {
	StaffID          string    `json:"staff_id"`
	StaffName        string    `json:"staff_name"`
	Date             string    `json:"date"`
	TasksCompleted   int       `json:"tasks_completed"`
	AvgMinutesPerRoom float64  `json:"avg_minutes_per_room"`
	QualityScore     float64   `json:"quality_score"`
	Complaints       int       `json:"complaints"`
	RoomsCleaned     int       `json:"rooms_cleaned"`
	RoomsInspected   int       `json:"rooms_inspected"`
}

// CleaningChecklistTemplate defines what needs to be checked per room type
type CleaningChecklistTemplate struct {
	ID         string   `json:"id"`
	TenantID   string   `json:"tenant_id"`
	RoomType   string   `json:"room_type"`  // standard, deluxe, suite
	TaskType   string   `json:"task_type"`  // clean, refill
	Items      []ChecklistItem `json:"items"`
	CreatedAt  time.Time `json:"created_at"`
}

// ChecklistItem is a single checklist step
type ChecklistItem struct {
	ID       string `json:"id"`
	Label    string `json:"label"`     // "Make bed", "Replace towels", "Restock minibar"
	Category string `json:"category"`  // bed, bath, amenities, minibar, general
	Required bool   `json:"required"`
	Order    int    `json:"order"`
}

// CleanerCardMode for smart lock integration
type CleanerCardMode string

const (
	CleanerCardCleanFull  CleanerCardMode = "clean_full"   // Full room clean
	CleanerCardCleanRefill CleanerCardMode = "clean_refill" // Refill only: coffee, tea, sugar, minibar
	CleanerCardMaintenance CleanerCardMode = "maintenance"  // Maintenance access
)

// LockAccessCodeWithMode extends the basic access code with cleaner mode
type LockAccessCodeWithMode struct {
	AccessCodeID string          `json:"access_code_id"`
	Code         string          `json:"code"`
	CleanerMode  CleanerCardMode `json:"cleaner_mode"`
	AssignedTo   string          `json:"assigned_to"` // staff member name or ID
	ValidFrom    time.Time       `json:"valid_from"`
	ValidUntil   time.Time       `json:"valid_until"`
	Floor        string          `json:"floor"`
	RoomIDs      []string        `json:"room_ids,omitempty"`
}

// HousekeepingDashboardStats for the manager dashboard
type HousekeepingDashboardStats struct {
	TotalRooms       int                     `json:"total_rooms"`
	DirtyRooms       int                     `json:"dirty_rooms"`
	InProgressRooms  int                     `json:"in_progress_rooms"`
	ReadyRooms       int                     `json:"ready_rooms"`
	InspectedRooms   int                     `json:"inspected_rooms"`
	BlockedRooms     int                     `json:"blocked_rooms"`
	MaintenanceRooms int                     `json:"maintenance_rooms"`
	UnassignedTasks  int                     `json:"unassigned_tasks"`
	StaffOnDuty      int                     `json:"staff_on_duty"`
	AvgCleanTime     float64                 `json:"avg_clean_time_minutes"`
	PredictedReadyBy string                  `json:"predicted_ready_by"`
	StaffLoad        []StaffLoadItem          `json:"staff_load"`
	FloorProgress    []FloorProgressItem      `json:"floor_progress"`
}

// StaffLoadItem shows how much work each staff member has
type StaffLoadItem struct {
	StaffID      string `json:"staff_id"`
	StaffName    string `json:"staff_name"`
	Assigned     int    `json:"assigned"`
	InProgress   int    `json:"in_progress"`
	Completed    int    `json:"completed"`
	EstimatedMin int    `json:"estimated_minutes"`
}

// FloorProgressItem shows cleaning progress per floor
type FloorProgressItem struct {
	Floor      string `json:"floor"`
	Total      int    `json:"total"`
	Dirty      int    `json:"dirty"`
	InProgress int    `json:"in_progress"`
	Ready      int    `json:"ready"`
	PercentDone int   `json:"percent_done"`
}

// SuppliesInventoryItem tracks consumable stock levels.
type SuppliesInventoryItem struct {
	ID            string     `json:"id"`
	TenantID      string     `json:"tenant_id"`
	ItemName      string     `json:"item_name"`
	Category      string     `json:"category"`
	Unit          string     `json:"unit"`
	CurrentStock  int        `json:"current_stock"`
	ReorderLevel  int        `json:"reorder_level"`
	CostPerUnit   float64    `json:"cost_per_unit"`
	Supplier      string     `json:"supplier,omitempty"`
	LastRestocked *time.Time `json:"last_restocked,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}
