// Package events provides domain-to-transport event mapping for PMS integration.
package events

import (
	"encoding/json"
	"fmt"
	"time"

	bcEvents "github.com/nexus-platform/backend-core/internal/events"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// DomainEventMapper converts PMS domain events to backend-core transport events.
type DomainEventMapper struct {
	source string
}

// NewDomainEventMapper creates a new event mapper.
func NewDomainEventMapper(source string) *DomainEventMapper {
	return &DomainEventMapper{source: source}
}

// Map converts a PMS DomainEvent to a backend-core Event envelope.
func (m *DomainEventMapper) Map(event DomainEvent) (bcEvents.Event, error) {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return bcEvents.Event{}, fmt.Errorf("marshal payload: %w", err)
	}

	return bcEvents.Event{
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
func (m *DomainEventMapper) MapReservationCreated(e domain.ReservationCreatedEvent, tenantID, correlationID string) (bcEvents.Event, error) {
	return bcEvents.NewEvent(
		"ReservationCreated",
		"1.0",
		tenantID,
		correlationID,
		m.source,
		e,
	)
}

// MapReservationConfirmed converts a ReservationConfirmedEvent to a transport Event.
func (m *DomainEventMapper) MapReservationConfirmed(e domain.ReservationConfirmedEvent, tenantID, correlationID string) (bcEvents.Event, error) {
	return bcEvents.NewEvent(
		"ReservationConfirmed",
		"1.0",
		tenantID,
		correlationID,
		m.source,
		e,
	)
}

// MapGuestCheckedIn converts a GuestCheckedInEvent to a transport Event.
func (m *DomainEventMapper) MapGuestCheckedIn(e domain.GuestCheckedInEvent, tenantID, correlationID string) (bcEvents.Event, error) {
	return bcEvents.NewEvent(
		"GuestCheckedIn",
		"1.0",
		tenantID,
		correlationID,
		m.source,
		e,
	)
}

// MapFolioClosed converts a FolioClosedEvent to a transport Event.
func (m *DomainEventMapper) MapFolioClosed(e domain.FolioClosedEvent, tenantID, correlationID string) (bcEvents.Event, error) {
	return bcEvents.NewEvent(
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
