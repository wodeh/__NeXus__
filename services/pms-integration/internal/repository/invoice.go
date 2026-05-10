package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nexus-platform/pms-integration/internal/domain"
)

// InvoiceRepository manages invoices.
type InvoiceRepository struct {
	mu       sync.RWMutex
	invoices map[string]*domain.Invoice
	seq      int64
}

// NewInvoiceRepository creates a new invoice repository.
func NewInvoiceRepository() *InvoiceRepository {
	return &InvoiceRepository{
		invoices: make(map[string]*domain.Invoice),
		seq:      5000,
	}
}

// ListByTenant returns all invoices for a tenant.
func (r *InvoiceRepository) ListByTenant(ctx context.Context, tenant, status string, limit, offset int) ([]*domain.Invoice, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*domain.Invoice
	for _, inv := range r.invoices {
		if inv.TenantID != tenant {
			continue
		}
		if status != "" && inv.Status != status {
			continue
		}
		results = append(results, inv)
	}
	total := len(results)
	if offset >= total {
		return []*domain.Invoice{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return results[offset:end], total, nil
}

// GetByID returns a specific invoice.
func (r *InvoiceRepository) GetByID(ctx context.Context, tenant, id string) (*domain.Invoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	inv, ok := r.invoices[id]
	if !ok || inv.TenantID != tenant {
		return nil, fmt.Errorf("invoice not found")
	}
	return inv, nil
}

// Create stores a new invoice.
func (r *InvoiceRepository) Create(ctx context.Context, inv *domain.Invoice) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.seq++
	inv.ID = fmt.Sprintf("inv-%d", r.seq)
	inv.InvoiceNumber = fmt.Sprintf("INV-%06d", r.seq)
	inv.CreatedAt = time.Now()
	inv.UpdatedAt = inv.CreatedAt
	inv.Status = domain.InvoiceStatusDraft
	inv.CalculateTotals()
	r.invoices[inv.ID] = inv
	return nil
}

// Update modifies an existing invoice.
func (r *InvoiceRepository) Update(ctx context.Context, tenant, id string, updates *domain.Invoice) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	inv, ok := r.invoices[id]
	if !ok || inv.TenantID != tenant {
		return fmt.Errorf("invoice not found")
	}
	if updates.Status != "" {
		inv.Status = updates.Status
		if updates.Status == domain.InvoiceStatusPaid {
			now := time.Now()
			inv.PaidDate = &now
		}
	}
	if len(updates.LineItems) > 0 {
		inv.LineItems = updates.LineItems
		inv.CalculateTotals()
	}
	if updates.Notes != "" {
		inv.Notes = updates.Notes
	}
	inv.UpdatedAt = time.Now()
	return nil
}

// Delete removes an invoice.
func (r *InvoiceRepository) Delete(ctx context.Context, tenant, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	inv, ok := r.invoices[id]
	if !ok || inv.TenantID != tenant {
		return fmt.Errorf("invoice not found")
	}
	delete(r.invoices, id)
	return nil
}

// SeedInvoices pre-populates demo data.
func (r *InvoiceRepository) SeedInvoices() {
	now := time.Now()
	r.Create(context.Background(), &domain.Invoice{
		TenantID: "demo",
		FolioID:  "f-001",
		GuestID:  "g-001",
		Currency: "USD",
		IssueDate: now,
		DueDate:  now.AddDate(0, 0, 7),
		LineItems: []domain.InvoiceLineItem{
			{Description: "Room 101 - 4 nights", Quantity: 4, UnitPrice: 129.00, TaxRate: 14.0},
			{Description: "Breakfast buffet", Quantity: 4, UnitPrice: 25.00, TaxRate: 14.0},
		},
	})
	r.Create(context.Background(), &domain.Invoice{
		TenantID: "demo",
		FolioID:  "f-002",
		GuestID:  "g-002",
		Currency: "USD",
		IssueDate: now,
		DueDate:  now.AddDate(0, 0, 7),
		LineItems: []domain.InvoiceLineItem{
			{Description: "Room 201 - 3 nights", Quantity: 3, UnitPrice: 189.00, TaxRate: 14.0},
			{Description: "Spa package", Quantity: 1, UnitPrice: 150.00, TaxRate: 14.0},
		},
	})
}
