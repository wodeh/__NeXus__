package domain

import (
	"errors"
	"time"
)

// InvoiceLineItem represents a single charge.
type InvoiceLineItem struct {
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TaxRate     float64 `json:"tax_rate"`
	Total       float64 `json:"total"`
}

// Invoice tracks billing for a folio.
type Invoice struct {
	ID          string
	TenantID    string
	FolioID     string
	GuestID     string
	InvoiceNumber string
	Status      string // draft, sent, paid, overdue, cancelled
	LineItems   []InvoiceLineItem
	Subtotal    float64
	TaxTotal    float64
	GrandTotal  float64
	Currency    string
	IssueDate   time.Time
	DueDate     time.Time
	PaidDate    *time.Time
	Notes       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Validate checks an invoice.
func (inv *Invoice) Validate() error {
	if inv.TenantID == "" {
		return errors.New("tenant_id is required")
	}
	if inv.FolioID == "" {
		return errors.New("folio_id is required")
	}
	return nil
}

// CalculateTotals computes subtotal, tax, and grand total.
func (inv *Invoice) CalculateTotals() {
	inv.Subtotal = 0
	inv.TaxTotal = 0
	for _, li := range inv.LineItems {
		li.Total = float64(li.Quantity) * li.UnitPrice
		inv.Subtotal += li.Total
		inv.TaxTotal += li.Total * (li.TaxRate / 100)
	}
	inv.GrandTotal = inv.Subtotal + inv.TaxTotal
}

const (
	InvoiceStatusDraft    = "draft"
	InvoiceStatusSent     = "sent"
	InvoiceStatusPaid     = "paid"
	InvoiceStatusOverdue  = "overdue"
	InvoiceStatusCancelled = "cancelled"
)
