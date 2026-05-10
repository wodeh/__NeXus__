package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nexus-platform/pms-integration/internal/domain"
)

// CheckInRepository manages check-in/out workflows.
type CheckInRepository struct {
	mu       sync.RWMutex
	checkins map[string]*domain.CheckInWorkflow
	checkouts map[string]*domain.CheckOutWorkflow
	seq      int64
}

// NewCheckInRepository creates a new check-in repository.
func NewCheckInRepository() *CheckInRepository {
	return &CheckInRepository{
		checkins:  make(map[string]*domain.CheckInWorkflow),
		checkouts: make(map[string]*domain.CheckOutWorkflow),
		seq:       4000,
	}
}

// CheckIn CRUD
func (r *CheckInRepository) CreateCheckIn(ctx context.Context, ci *domain.CheckInWorkflow) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	ci.ID = fmt.Sprintf("ci-%d", r.seq)
	ci.CreatedAt = time.Now()
	ci.UpdatedAt = ci.CreatedAt
	ci.Status = domain.CheckInPending
	r.checkins[ci.ID] = ci
	return nil
}

func (r *CheckInRepository) GetCheckIn(ctx context.Context, tenant, id string) (*domain.CheckInWorkflow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ci, ok := r.checkins[id]
	if !ok || ci.TenantID != tenant {
		return nil, fmt.Errorf("check-in not found")
	}
	return ci, nil
}

func (r *CheckInRepository) UpdateCheckIn(ctx context.Context, tenant, id string, updates map[string]interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	ci, ok := r.checkins[id]
	if !ok || ci.TenantID != tenant {
		return fmt.Errorf("check-in not found")
	}
	if v, ok := updates["id_verified"]; ok {
		ci.IDVerified = v.(bool)
	}
	if v, ok := updates["payment_collected"]; ok {
		ci.PaymentCollected = v.(bool)
	}
	if v, ok := updates["key_issued"]; ok {
		ci.KeyIssued = v.(bool)
	}
	if v, ok := updates["room_inspected"]; ok {
		ci.RoomInspected = v.(bool)
	}
	if v, ok := updates["welcome_sent"]; ok {
		ci.WelcomeSent = v.(bool)
	}
	if ci.IsComplete() {
		ci.Status = domain.CheckInComplete
		now := time.Now()
		ci.CompletedAt = &now
	} else if ci.Status == domain.CheckInPending && (ci.IDVerified || ci.PaymentCollected) {
		ci.Status = domain.CheckInProgress
		now := time.Now()
		ci.StartedAt = &now
	}
	ci.UpdatedAt = time.Now()
	return nil
}

// CheckOut CRUD
func (r *CheckInRepository) CreateCheckOut(ctx context.Context, co *domain.CheckOutWorkflow) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	co.ID = fmt.Sprintf("co-%d", r.seq)
	co.CreatedAt = time.Now()
	co.UpdatedAt = co.CreatedAt
	co.Status = domain.CheckInPending
	r.checkouts[co.ID] = co
	return nil
}

func (r *CheckInRepository) GetCheckOut(ctx context.Context, tenant, id string) (*domain.CheckOutWorkflow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	co, ok := r.checkouts[id]
	if !ok || co.TenantID != tenant {
		return nil, fmt.Errorf("check-out not found")
	}
	return co, nil
}

func (r *CheckInRepository) UpdateCheckOut(ctx context.Context, tenant, id string, updates map[string]interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	co, ok := r.checkouts[id]
	if !ok || co.TenantID != tenant {
		return fmt.Errorf("check-out not found")
	}
	if v, ok := updates["balance_settled"]; ok {
		co.BalanceSettled = v.(bool)
	}
	if v, ok := updates["key_returned"]; ok {
		co.KeyReturned = v.(bool)
	}
	if v, ok := updates["room_inspected"]; ok {
		co.RoomInspected = v.(bool)
	}
	if v, ok := updates["feedback_sent"]; ok {
		co.FeedbackSent = v.(bool)
	}
	if co.BalanceSettled && co.KeyReturned && co.RoomInspected {
		co.Status = domain.CheckInComplete
		now := time.Now()
		co.CompletedAt = &now
	}
	co.UpdatedAt = time.Now()
	return nil
}
