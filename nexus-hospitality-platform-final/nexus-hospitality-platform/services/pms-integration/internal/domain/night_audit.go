// Package domain provides night audit aggregates.
package domain

import (
	"time"
)

// NightAuditRun represents a single night audit execution.
type NightAuditRun struct {
	ID                string
	TenantID          string
	PropertyID        *string
	RunDate           time.Time
	Status            string // pending, running, completed, failed
	StartedAt         *time.Time
	CompletedAt       *time.Time
	ArrivalsProcessed int
	DeparturesProcessed int
	NoShowsProcessed  int
	ChargesPosted     int
	InvoicesGenerated int
	RevenueTotal      float64
	Errors            []string
	CreatedAt         time.Time
}

// NightAuditTask is a sub-task within a night audit run.
type NightAuditTask struct {
	ID             string
	NightAuditID   string
	TaskType       string // post_room_charges, check_no_shows, generate_invoices, close_folios, sync_channels
	Status         string // pending, running, completed, failed
	StartedAt      *time.Time
	CompletedAt    *time.Time
	RecordsAffected int
	ErrorMessage   string
}
