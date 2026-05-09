// Package domain defines financial, housekeeping, and maintenance aggregates.
package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ==================== FOLIO ====================

type FolioID string

type FolioStatus string

const (
	FolioStatusOpen       FolioStatus = "open"
	FolioStatusClosed     FolioStatus = "closed"
	FolioStatusTransferred FolioStatus = "transferred"
)

type Folio struct {
	ID                  FolioID
	TenantID            string
	ReservationID       string
	GuestID             string
	Status              FolioStatus
	Balance             float64
	TotalCharges        float64
	TotalPayments       float64
	TotalAdjustments    float64
	CurrencyCode        string
	IsMaster            bool
	MasterFolioID       *string
	Notes               string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Version             int
}

func NewFolio(tenantID, reservationID, guestID, currency string) *Folio {
	now := time.Now().UTC()
	return &Folio{
		ID:            FolioID(uuid.Must(uuid.NewRandom()).String()),
		TenantID:      tenantID,
		ReservationID: reservationID,
		GuestID:       guestID,
		Status:        FolioStatusOpen,
		CurrencyCode:  currency,
		Balance:       0,
		TotalCharges:  0,
		TotalPayments: 0,
		CreatedAt:     now,
		UpdatedAt:     now,
		Version:       1,
	}
}

func (f *Folio) AddCharge(amount float64) {
	f.TotalCharges += amount
	f.Balance += amount
	f.bumpVersion()
}

func (f *Folio) AddPayment(amount float64) {
	f.TotalPayments += amount
	f.Balance -= amount
	f.bumpVersion()
}

func (f *Folio) AddAdjustment(amount float64) {
	f.TotalAdjustments += amount
	f.Balance -= amount
	f.bumpVersion()
}

func (f *Folio) Close() error {
	if f.Balance != 0 {
		return fmt.Errorf("cannot close folio %s: balance is %.2f", f.ID, f.Balance)
	}
	f.Status = FolioStatusClosed
	f.bumpVersion()
	return nil
}

func (f *Folio) bumpVersion() {
	f.UpdatedAt = time.Now().UTC()
	f.Version++
}

// ==================== CHARGE ====================

type ChargeID string

type Charge struct {
	ID             ChargeID
	TenantID       string
	FolioID        string
	ReservationID  *string
	RoomID         *string
	ChargeType     string
	Category       string
	Description    string
	Quantity       float64
	UnitPrice      float64
	TotalAmount    float64
	TaxAmount      float64
	TaxRate        float64
	CurrencyCode   string
	ChargeDate     time.Time
	PostingDate    time.Time
	PostedBy       string
	IsRevenue      bool
	RevenueCenter  string
	Source         string
	IsVoided       bool
	VoidedAt       *time.Time
	VoidedBy       string
	VoidReason     string
	Metadata       map[string]interface{}
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewCharge(tenantID, folioID, chargeType, description string, quantity, unitPrice float64) *Charge {
	now := time.Now().UTC()
	total := quantity * unitPrice
	return &Charge{
		ID:          ChargeID(uuid.Must(uuid.NewRandom()).String()),
		TenantID:    tenantID,
		FolioID:     folioID,
		ChargeType:  chargeType,
		Description: description,
		Quantity:    quantity,
		UnitPrice:   unitPrice,
		TotalAmount: total,
		TaxAmount:   0,
		TaxRate:     0,
		CurrencyCode: "USD",
		ChargeDate:  now,
		PostingDate: now,
		IsRevenue:   true,
		Source:      "manual",
		IsVoided:    false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func (c *Charge) SetTax(taxRate float64) {
	c.TaxRate = taxRate
	c.TaxAmount = c.TotalAmount * taxRate
	c.TotalAmount += c.TaxAmount
}

func (c *Charge) Void(by, reason string) {
	now := time.Now().UTC()
	c.IsVoided = true
	c.VoidedAt = &now
	c.VoidedBy = by
	c.VoidReason = reason
	c.UpdatedAt = now
}

// ==================== PAYMENT ====================

type PaymentID string

type PaymentStatus string

const (
	PaymentStatusCompleted  PaymentStatus = "completed"
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusRefunded   PaymentStatus = "refunded"
	PaymentStatusPartialRefund PaymentStatus = "partial_refund"
)

type Payment struct {
	ID                PaymentID
	TenantID          string
	FolioID           string
	ReservationID     *string
	GuestID           *string
	PaymentMethod     string
	PaymentType       string
	Amount            float64
	CurrencyCode      string
	ExchangeRate      float64
	ReferenceNumber   string
	TransactionID     string
	Status            PaymentStatus
	Processor         string
	LastFourDigits    string
	CardBrand         string
	ExpiryMonth       int
	ExpiryYear        int
	AuthorizationCode string
	CapturedAt        *time.Time
	RefundedAt        *time.Time
	RefundAmount      float64
	RefundReason      string
	ProcessedBy       string
	Metadata          map[string]interface{}
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func NewPayment(tenantID, folioID string, amount float64, method string) *Payment {
	now := time.Now().UTC()
	return &Payment{
		ID:           PaymentID(uuid.Must(uuid.NewRandom()).String()),
		TenantID:     tenantID,
		FolioID:      folioID,
		Amount:       amount,
		CurrencyCode: "USD",
		ExchangeRate: 1.0,
		Status:       PaymentStatusCompleted,
		PaymentMethod: method,
		PaymentType:  "charge",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (p *Payment) Refund(amount float64, reason string) error {
	if amount > p.Amount {
		return fmt.Errorf("refund amount %.2f exceeds payment amount %.2f", amount, p.Amount)
	}
	now := time.Now().UTC()
	p.RefundedAt = &now
	p.RefundAmount = amount
	p.RefundReason = reason
	if amount == p.Amount {
		p.Status = PaymentStatusRefunded
	} else {
		p.Status = PaymentStatusPartialRefund
	}
	p.UpdatedAt = now
	return nil
}

// ==================== FOLIO ADJUSTMENT ====================

type FolioAdjustmentID string

type FolioAdjustment struct {
	ID            FolioAdjustmentID
	TenantID      string
	FolioID       string
	ChargeID      *string
	AdjustmentType string
	Amount        float64
	CurrencyCode  string
	Reason        string
	ApprovedBy    string
	ApprovedAt    *time.Time
	IsApproved    bool
	Metadata      map[string]interface{}
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewFolioAdjustment(tenantID, folioID, adjType, reason string, amount float64) *FolioAdjustment {
	now := time.Now().UTC()
	return &FolioAdjustment{
		ID:             FolioAdjustmentID(uuid.Must(uuid.NewRandom()).String()),
		TenantID:       tenantID,
		FolioID:        folioID,
		AdjustmentType: adjType,
		Reason:         reason,
		Amount:         amount,
		CurrencyCode:   "USD",
		IsApproved:     false,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (a *FolioAdjustment) Approve(by string) {
	now := time.Now().UTC()
	a.IsApproved = true
	a.ApprovedBy = by
	a.ApprovedAt = &now
	a.UpdatedAt = now
}

// ==================== HOUSEKEEPING TASK ====================

type HousekeepingTaskID string

type TaskPriority string

const (
	TaskPriorityLow      TaskPriority = "low"
	TaskPriorityNormal   TaskPriority = "normal"
	TaskPriorityHigh     TaskPriority = "high"
	TaskPriorityUrgent   TaskPriority = "urgent"
)

type TaskStatus string

const (
	TaskStatusPending     TaskStatus = "pending"
	TaskStatusAssigned    TaskStatus = "assigned"
	TaskStatusInProgress  TaskStatus = "in_progress"
	TaskStatusCompleted   TaskStatus = "completed"
	TaskStatusInspected   TaskStatus = "inspected"
	TaskStatusSkipped     TaskStatus = "skipped"
)

type HousekeepingTask struct {
	ID                     HousekeepingTaskID
	TenantID               string
	PropertyID             string
	RoomID                 string
	TaskType               string
	Priority               TaskPriority
	Status                 TaskStatus
	AssignedTo             string
	ScheduledAt            *time.Time
	StartedAt              *time.Time
	CompletedAt            *time.Time
	EstimatedDuration      int
	ActualDuration         int
	Notes                  string
	InspectionScore        int
	InspectionNotes        string
	InspectedBy            string
	CreatedBy              string
	CreatedAt              time.Time
	UpdatedAt              time.Time
	Version                int
}

func NewHousekeepingTask(tenantID, propertyID, roomID, taskType string) *HousekeepingTask {
	now := time.Now().UTC()
	return &HousekeepingTask{
		ID:         HousekeepingTaskID(uuid.Must(uuid.NewRandom()).String()),
		TenantID:   tenantID,
		PropertyID: propertyID,
		RoomID:     roomID,
		TaskType:   taskType,
		Priority:   TaskPriorityNormal,
		Status:     TaskStatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
		Version:    1,
	}
}

func (t *HousekeepingTask) Assign(to string) {
	t.AssignedTo = to
	t.Status = TaskStatusAssigned
	t.bumpVersion()
}

func (t *HousekeepingTask) Start() {
	now := time.Now().UTC()
	t.StartedAt = &now
	t.Status = TaskStatusInProgress
	t.bumpVersion()
}

func (t *HousekeepingTask) Complete() {
	now := time.Now().UTC()
	t.CompletedAt = &now
	if t.StartedAt != nil {
		t.ActualDuration = int(now.Sub(*t.StartedAt).Minutes())
	}
	t.Status = TaskStatusCompleted
	t.bumpVersion()
}

func (t *HousekeepingTask) Inspect(score int, by, notes string) {
	t.InspectionScore = score
	t.InspectedBy = by
	t.InspectionNotes = notes
	t.Status = TaskStatusInspected
	t.bumpVersion()
}

func (t *HousekeepingTask) bumpVersion() {
	t.UpdatedAt = time.Now().UTC()
	t.Version++
}

// ==================== WORK ORDER ====================

type WorkOrderID string

type WorkOrderPriority string

const (
	WorkOrderPriorityLow      WorkOrderPriority = "low"
	WorkOrderPriorityMedium   WorkOrderPriority = "medium"
	WorkOrderPriorityHigh     WorkOrderPriority = "high"
	WorkOrderPriorityCritical WorkOrderPriority = "critical"
)

type WorkOrderStatus string

const (
	WorkOrderStatusOpen        WorkOrderStatus = "open"
	WorkOrderStatusAssigned    WorkOrderStatus = "assigned"
	WorkOrderStatusInProgress  WorkOrderStatus = "in_progress"
	WorkOrderStatusOnHold      WorkOrderStatus = "on_hold"
	WorkOrderStatusCompleted   WorkOrderStatus = "completed"
	WorkOrderStatusCancelled   WorkOrderStatus = "cancelled"
)

type WorkOrder struct {
	ID              WorkOrderID
	TenantID        string
	PropertyID      string
	RoomID          *string
	AssetID         *string
	WorkOrderNumber string
	Category        string
	Priority        WorkOrderPriority
	Status          WorkOrderStatus
	Title           string
	Description     string
	ReportedBy      string
	AssignedTo      string
	EstimatedCost   float64
	ActualCost      float64
	ScheduledDate   *time.Time
	StartedAt       *time.Time
	CompletedAt     *time.Time
	ResolutionNotes string
	PartsUsed       []map[string]interface{}
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Version         int
}

func NewWorkOrder(tenantID, propertyID, woNumber, category, title string) *WorkOrder {
	now := time.Now().UTC()
	return &WorkOrder{
		ID:              WorkOrderID(uuid.Must(uuid.NewRandom()).String()),
		TenantID:        tenantID,
		PropertyID:      propertyID,
		WorkOrderNumber: woNumber,
		Category:        category,
		Priority:        WorkOrderPriorityMedium,
		Status:          WorkOrderStatusOpen,
		Title:           title,
		CreatedAt:       now,
		UpdatedAt:       now,
		Version:         1,
	}
}

func (w *WorkOrder) Assign(to string) {
	w.AssignedTo = to
	w.Status = WorkOrderStatusAssigned
	w.bumpVersion()
}

func (w *WorkOrder) Start() {
	now := time.Now().UTC()
	w.StartedAt = &now
	w.Status = WorkOrderStatusInProgress
	w.bumpVersion()
}

func (w *WorkOrder) Complete(resolution string, actualCost float64) {
	now := time.Now().UTC()
	w.CompletedAt = &now
	w.ResolutionNotes = resolution
	w.ActualCost = actualCost
	w.Status = WorkOrderStatusCompleted
	w.bumpVersion()
}

func (w *WorkOrder) bumpVersion() {
	w.UpdatedAt = time.Now().UTC()
	w.Version++
}

// ==================== AUDIT LOG ====================

type AuditLogID string

type AuditLog struct {
	ID            AuditLogID
	TenantID      string
	EntityType    string
	EntityID      string
	Action        string
	PerformedBy   string
	PerformedByID *string
	IPAddress     string
	UserAgent     string
	OldValues     map[string]interface{}
	NewValues     map[string]interface{}
	Metadata      map[string]interface{}
	CreatedAt     time.Time
}

func NewAuditLog(tenantID, entityType, entityID, action, performedBy string) *AuditLog {
	return &AuditLog{
		ID:          AuditLogID(uuid.Must(uuid.NewRandom()).String()),
		TenantID:    tenantID,
		EntityType:  entityType,
		EntityID:    entityID,
		Action:      action,
		PerformedBy: performedBy,
		CreatedAt:   time.Now().UTC(),
	}
}
