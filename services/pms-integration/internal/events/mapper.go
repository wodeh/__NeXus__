// Package events provides domain-to-transport event mapping for PMS integration.
package events

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nexus-platform/pms-integration/internal/domain"
)

// TransportEvent represents an event envelope for external systems.
type TransportEvent struct {
	ID            string          `json:"id"`
	Type          string          `json:"type"`
	Version       string          `json:"version"`
	TenantID      string          `json:"tenant_id"`
	CorrelationID string          `json:"correlation_id"`
	Timestamp     time.Time       `json:"timestamp"`
	Source        string          `json:"source"`
	Payload       json.RawMessage `json:"payload"`
}

// NewTransportEvent creates a new transport event.
func NewTransportEvent(eventType, version, tenantID, correlationID, source string, payload interface{}) (TransportEvent, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return TransportEvent{}, fmt.Errorf("marshal payload: %w", err)
	}
	return TransportEvent{
		ID:            generateEventID(),
		Type:          eventType,
		Version:       version,
		TenantID:      tenantID,
		CorrelationID: correlationID,
		Timestamp:     time.Now().UTC(),
		Source:        source,
		Payload:       data,
	}, nil
}

// DomainEventMapper converts PMS domain events to transport events.
type DomainEventMapper struct {
	source string
}

// NewDomainEventMapper creates a new event mapper.
func NewDomainEventMapper(source string) *DomainEventMapper {
	return &DomainEventMapper{source: source}
}

// Map converts a PMS DomainEvent to a transport Event envelope.
func (m *DomainEventMapper) Map(event DomainEvent) (TransportEvent, error) {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return TransportEvent{}, fmt.Errorf("marshal payload: %w", err)
	}

	return TransportEvent{
		ID:            generateEventID(),
		Type:          event.Type,
		Version:       "1.0",
		TenantID:      event.TenantID,
		CorrelationID: event.CorrelationID,
		Timestamp:     event.OccurredAt,
		Source:        m.source,
		Payload:       payload,
	}, nil
}

// MapReservationCreated converts a ReservationCreatedEvent to a transport Event.
func (m *DomainEventMapper) MapReservationCreated(e domain.ReservationCreatedEvent, tenantID, correlationID string) (TransportEvent, error) {
	return NewTransportEvent(
		"ReservationCreated",
		"1.0",
		tenantID,
		correlationID,
		m.source,
		e,
	)
}

// MapReservationConfirmed converts a ReservationConfirmedEvent to a transport Event.
func (m *DomainEventMapper) MapReservationConfirmed(e domain.ReservationConfirmedEvent, tenantID, correlationID string) (TransportEvent, error) {
	return NewTransportEvent(
		"ReservationConfirmed",
		"1.0",
		tenantID,
		correlationID,
		m.source,
		e,
	)
}

// MapGuestCheckedIn converts a GuestCheckedInEvent to a transport Event.
func (m *DomainEventMapper) MapGuestCheckedIn(e domain.GuestCheckedInEvent, tenantID, correlationID string) (TransportEvent, error) {
	return NewTransportEvent(
		"GuestCheckedIn",
		"1.0",
		tenantID,
		correlationID,
		m.source,
		e,
	)
}

// MapFolioClosed converts a FolioClosedEvent to a transport Event.
func (m *DomainEventMapper) MapFolioClosed(e domain.FolioClosedEvent, tenantID, correlationID string) (TransportEvent, error) {
	return NewTransportEvent(
		"FolioClosed",
		"1.0",
		tenantID,
		correlationID,
		m.source,
		e,
	)
}

func generateEventID() string {
	return "evt_" + time.Now().UTC().Format("20060102T150405.000000000") + randomSuffix()
}

func randomSuffix() string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
