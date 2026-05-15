package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nexus-platform/backend-core/internal/db"
	"github.com/nexus-platform/backend-core/internal/domain"
)

// GuestJourneyRepository provides guest journey data access.
type GuestJourneyRepository struct {
	pool *db.Pool
}

// NewGuestJourneyRepository creates a guest journey repository.
func NewGuestJourneyRepository(pool *db.Pool) *GuestJourneyRepository {
	return &GuestJourneyRepository{pool: pool}
}

// CreateEvent adds a journey event.
func (r *GuestJourneyRepository) CreateEvent(ctx context.Context, tenantID, reservationID uuid.UUID, eventType, createdBy string, eventData map[string]interface{}) (*domain.GuestJourneyEvent, error) {
	query := `
		INSERT INTO guest_journey_events (tenant_id, reservation_id, event_type, event_data, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, tenant_id, reservation_id, event_type, event_data, occurred_at, created_by, created_at
	`
	var event domain.GuestJourneyEvent
	var eventDataJSON []byte
	if eventData != nil {
		eventDataJSON, _ = json.Marshal(eventData)
	}
	row := r.pool.QueryRow(ctx, query, tenantID, reservationID, eventType, eventDataJSON, createdBy)
	if err := row.Scan(&event.ID, &event.TenantID, &event.ReservationID, &event.EventType, &event.EventData, &event.OccurredAt, &event.CreatedBy, &event.CreatedAt); err != nil {
		return nil, fmt.Errorf("create journey event: %w", err)
	}
	return &event, nil
}

// GetTimeline returns all events for a reservation.
func (r *GuestJourneyRepository) GetTimeline(ctx context.Context, tenantID, reservationID uuid.UUID) (*domain.GuestJourneyTimeline, error) {
	// Get guest name
	var guestName string
	_ = r.pool.QueryRow(ctx, `SELECT guest_name FROM reservations WHERE id = $1 AND tenant_id = $2`, reservationID, tenantID).Scan(&guestName)

	query := `
		SELECT id, tenant_id, reservation_id, event_type, event_data, occurred_at, created_by, created_at
		FROM guest_journey_events
		WHERE tenant_id = $1 AND reservation_id = $2
		ORDER BY occurred_at ASC
	`
	rows, err := r.pool.Query(ctx, query, tenantID, reservationID)
	if err != nil {
		return nil, fmt.Errorf("get timeline: %w", err)
	}
	defer rows.Close()

	var events []domain.GuestJourneyEvent
	for rows.Next() {
		var event domain.GuestJourneyEvent
		var eventDataJSON []byte
		if err := rows.Scan(&event.ID, &event.TenantID, &event.ReservationID, &event.EventType, &eventDataJSON, &event.OccurredAt, &event.CreatedBy, &event.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan journey event: %w", err)
		}
		if len(eventDataJSON) > 0 {
			json.Unmarshal(eventDataJSON, &event.EventData)
		}
		events = append(events, event)
	}

	return &domain.GuestJourneyTimeline{
		ReservationID: reservationID,
		GuestName:     guestName,
		Events:        events,
	}, nil
}

// ListUpcomingCheckIns returns reservations checking in today/tomorrow without pre-arrival email.
func (r *GuestJourneyRepository) ListUpcomingCheckIns(ctx context.Context, tenantID uuid.UUID, date string) ([]uuid.UUID, error) {
	query := `
		SELECT r.id FROM reservations r
		WHERE r.tenant_id = $1 AND r.status = 'confirmed'
		AND r.check_in = $2
		AND NOT EXISTS (
			SELECT 1 FROM guest_journey_events e
			WHERE e.reservation_id = r.id AND e.event_type = 'pre_arrival_email_sent'
		)
	`
	rows, err := r.pool.Query(ctx, query, tenantID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	return ids, nil
}
