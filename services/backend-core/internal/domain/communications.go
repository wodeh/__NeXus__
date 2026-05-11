package domain

import (
	"time"

	"github.com/google/uuid"
)

// CommTemplate represents a communication template (email, SMS, WhatsApp).
type CommTemplate struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	Name       string    `json:"name"`
	Subject    *string   `json:"subject,omitempty"`
	Body       string    `json:"body"`
	Channel    string    `json:"channel"` // email, sms, whatsapp
	Category   string    `json:"category"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CommSequence represents an automated communication sequence.
type CommSequence struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Trigger     string    `json:"trigger"`
	IsActive    bool      `json:"is_active"`
	Steps       int       `json:"steps"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CommScheduled represents a scheduled communication.
type CommScheduled struct {
	ID           uuid.UUID  `json:"id"`
	TenantID     uuid.UUID  `json:"tenant_id"`
	GuestName    string     `json:"guest_name"`
	Channel      string     `json:"channel"`
	Subject      *string    `json:"subject,omitempty"`
	Body         *string    `json:"body,omitempty"`
	Status       string     `json:"status"` // scheduled, sent, delivered, failed
	ScheduledAt  time.Time  `json:"scheduled_at"`
	SentAt       *time.Time `json:"sent_at,omitempty"`
	Error        *string    `json:"error,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}
