package domain

import "time"

// CommunicationChannel for sending messages.
type CommunicationChannel string

const (
	CommChannelEmail    CommunicationChannel = "email"
	CommChannelSMS      CommunicationChannel = "sms"
	CommChannelWhatsApp CommunicationChannel = "whatsapp"
)

// CommunicationTemplate is a reusable message template.
type CommunicationTemplate struct {
	ID          string               `json:"id"`
	TenantID    string               `json:"tenant_id"`
	Name        string               `json:"name"`
	Subject     string               `json:"subject,omitempty"`
	Body        string               `json:"body"`
	Channel     CommunicationChannel `json:"channel"`
	Category    string               `json:"category"` // booking, pre_arrival, checkin, stay, checkout, review, general
	Variables   []string             `json:"variables,omitempty"`
	IsActive    bool                 `json:"is_active"`
	IsDefault   bool                 `json:"is_default"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

// CommunicationSequence is an ordered set of scheduled messages.
type CommunicationSequence struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Trigger     string    `json:"trigger"` // booking_created, checkin, checkout, custom_date
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CommunicationSequenceStep is a single step in a sequence.
type CommunicationSequenceStep struct {
	ID           string    `json:"id"`
	SequenceID   string    `json:"sequence_id"`
	TemplateID   string    `json:"template_id"`
	StepOrder    int       `json:"step_order"`
	DelayHours   int       `json:"delay_hours"` // from trigger
	Channel      CommunicationChannel `json:"channel"`
	ConditionsJSON string  `json:"conditions_json,omitempty"` // e.g. only_if_vip
	CreatedAt    time.Time `json:"created_at"`
}

// ScheduledCommunication is a message queued for delivery.
type ScheduledCommunication struct {
	ID             string               `json:"id"`
	TenantID       string               `json:"tenant_id"`
	ReservationID  string               `json:"reservation_id,omitempty"`
	GuestID        string               `json:"guest_id,omitempty"`
	GuestPhone     string               `json:"guest_phone,omitempty"`
	GuestEmail     string               `json:"guest_email,omitempty"`
	TemplateID     string               `json:"template_id,omitempty"`
	SequenceID     string               `json:"sequence_id,omitempty"`
	SequenceStepID string               `json:"sequence_step_id,omitempty"`
	Channel        CommunicationChannel `json:"channel"`
	Subject        string               `json:"subject,omitempty"`
	BodyRendered   string               `json:"body_rendered"`
	Status         string               `json:"status"` // scheduled, sending, sent, delivered, failed, bounced
	ScheduledAt    time.Time            `json:"scheduled_at"`
	SentAt         *time.Time           `json:"sent_at,omitempty"`
	DeliveredAt    *time.Time           `json:"delivered_at,omitempty"`
	FailedAt       *time.Time           `json:"failed_at,omitempty"`
	ErrorMessage   string               `json:"error_message,omitempty"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
}

// CommunicationLog is the audit trail of sent messages.
type CommunicationLog struct {
	ID          string               `json:"id"`
	TenantID    string               `json:"tenant_id"`
	GuestID     string               `json:"guest_id,omitempty"`
	ReservationID string             `json:"reservation_id,omitempty"`
	Channel     CommunicationChannel `json:"channel"`
	Direction   string               `json:"direction"` // outbound, inbound
	Subject     string               `json:"subject,omitempty"`
	Body        string               `json:"body"`
	Status      string               `json:"status"`
	ExternalID  string               `json:"external_id,omitempty"` // provider message ID
	SentAt      *time.Time           `json:"sent_at,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
}
