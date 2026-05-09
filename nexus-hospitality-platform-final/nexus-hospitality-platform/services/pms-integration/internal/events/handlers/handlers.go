// Package handlers provides Kafka event handlers for PMS integration.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	bcEvents "github.com/nexus-platform/backend-core/internal/events"
	"github.com/nexus-platform/pms-integration/internal/domain"
	"github.com/nexus-platform/pms-integration/internal/repository"
)

// Registry maps event types to their handlers.
type Registry struct {
	handlers map[string]bcEvents.Handler
}

// NewRegistry creates a handler registry with all PMS event handlers.
func NewRegistry(store *repository.Store) *Registry {
	r := &Registry{handlers: make(map[string]bcEvents.Handler)}
	r.handlers["ReservationCreated"] = &ReservationCreatedHandler{store: store}
	r.handlers["ReservationUpdated"] = &ReservationUpdatedHandler{store: store}
	r.handlers["GuestCreated"] = &GuestCreatedHandler{store: store}
	return r
}

// Get returns the handler for a given event type.
func (r *Registry) Get(eventType string) (bcEvents.Handler, bool) {
	h, ok := r.handlers[eventType]
	return h, ok
}

// ReservationCreatedHandler processes ReservationCreated events.
type ReservationCreatedHandler struct {
	store *repository.Store
}

// Handle implements bcEvents.Handler.
func (h *ReservationCreatedHandler) Handle(ctx context.Context, event bcEvents.Event) error {
	slog.Info("handling ReservationCreated", slog.String("event_id", event.ID), slog.String("tenant", event.TenantID))

	var payload domain.ReservationCreatedEvent
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	// Idempotency: check if reservation already exists
	_, err := h.store.Reservations.GetByID(ctx, event.TenantID, payload.ReservationID)
	if err == nil {
		slog.Info("reservation already exists, skipping", slog.String("reservation_id", payload.ReservationID))
		return nil
	}

	res := domain.NewReservation(event.TenantID, payload.RoomID, payload.GuestID, payload.RoomID, payload.CheckIn, payload.CheckOut)
	if err := h.store.Reservations.Create(ctx, res); err != nil {
		return fmt.Errorf("create reservation: %w", err)
	}

	return nil
}

// ReservationUpdatedHandler processes ReservationUpdated events.
type ReservationUpdatedHandler struct {
	store *repository.Store
}

// Handle implements bcEvents.Handler.
func (h *ReservationUpdatedHandler) Handle(ctx context.Context, event bcEvents.Event) error {
	slog.Info("handling ReservationUpdated", slog.String("event_id", event.ID), slog.String("tenant", event.TenantID))

	var payload struct {
		ReservationID string                 `json:"reservation_id"`
		Status        string                 `json:"status"`
		Updates       map[string]interface{} `json:"updates"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	res, err := h.store.Reservations.GetByID(ctx, event.TenantID, payload.ReservationID)
	if err != nil {
		return fmt.Errorf("get reservation: %w", err)
	}

	// Apply state transition if valid
	switch payload.Status {
	case "confirmed":
		if err := res.Confirm(); err != nil {
			return fmt.Errorf("confirm reservation: %w", err)
		}
	case "checked_in":
		if err := res.CheckIn(); err != nil {
			return fmt.Errorf("check in reservation: %w", err)
		}
	case "checked_out":
		if err := res.CheckOut(); err != nil {
			return fmt.Errorf("check out reservation: %w", err)
		}
	case "cancelled":
		if err := res.Cancel(); err != nil {
			return fmt.Errorf("cancel reservation: %w", err)
		}
	}

	if err := h.store.Reservations.Update(ctx, res); err != nil {
		return fmt.Errorf("update reservation: %w", err)
	}

	return nil
}

// GuestCreatedHandler processes GuestCreated events.
type GuestCreatedHandler struct {
	store *repository.Store
}

// Handle implements bcEvents.Handler.
func (h *GuestCreatedHandler) Handle(ctx context.Context, event bcEvents.Event) error {
	slog.Info("handling GuestCreated", slog.String("event_id", event.ID), slog.String("tenant", event.TenantID))

	var payload struct {
		GuestID   string `json:"guest_id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		PropertyID string `json:"property_id"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	// Idempotency: check if guest already exists
	_, err := h.store.Guests.GetByID(ctx, event.TenantID, payload.GuestID)
	if err == nil {
		slog.Info("guest already exists, skipping", slog.String("guest_id", payload.GuestID))
		return nil
	}

	guest := domain.NewGuest(event.TenantID, payload.PropertyID, payload.FirstName, payload.LastName, payload.Email)
	if err := h.store.Guests.Create(ctx, guest); err != nil {
		return fmt.Errorf("create guest: %w", err)
	}

	return nil
}
