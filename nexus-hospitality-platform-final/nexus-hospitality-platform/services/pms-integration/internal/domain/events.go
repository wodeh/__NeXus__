package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// DomainEvent is the canonical envelope for all PMS domain events.
type DomainEvent struct {
	ID            string          `json:"id"`
	Type          string          `json:"type"`
	AggregateID   string          `json:"aggregate_id"`
	AggregateType string          `json:"aggregate_type"`
	TenantID      string          `json:"tenant_id"`
	CorrelationID string          `json:"correlation_id"`
	OccurredAt    time.Time       `json:"occurred_at"`
	Payload       json.RawMessage `json:"payload"`
}

// NewDomainEvent constructs a domain event envelope.
func NewDomainEvent(eventType, aggregateID, aggregateType, tenantID, correlationID string, payload interface{}) (DomainEvent, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return DomainEvent{}, err
	}
	return DomainEvent{
		ID:            uuid.Must(uuid.NewRandom()).String(),
		Type:          eventType,
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		TenantID:      tenantID,
		CorrelationID: correlationID,
		OccurredAt:    time.Now().UTC(),
		Payload:       raw,
	}, nil
}

// ReservationCreatedEvent is emitted when a new reservation is created.
type ReservationCreatedEvent struct {
	ReservationID string    `json:"reservation_id"`
	GuestID       string    `json:"guest_id"`
	RoomID        string    `json:"room_id"`
	CheckIn       time.Time `json:"check_in"`
	CheckOut      time.Time `json:"check_out"`
}

// ReservationConfirmedEvent is emitted when a reservation is confirmed.
type ReservationConfirmedEvent struct {
	ReservationID string `json:"reservation_id"`
}

// GuestCheckedInEvent is emitted at check-in.
type GuestCheckedInEvent struct {
	ReservationID string `json:"reservation_id"`
	GuestID       string `json:"guest_id"`
	RoomID        string `json:"room_id"`
}

// FolioClosedEvent is emitted when a folio is closed.
type FolioClosedEvent struct {
	FolioID   string  `json:"folio_id"`
	Balance   float64 `json:"balance"`
	Currency  string  `json:"currency"`
}
