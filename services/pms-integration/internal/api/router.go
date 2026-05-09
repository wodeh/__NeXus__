// Package api provides HTTP REST handlers for the PMS integration service.
package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nexus-platform/pms-integration/internal/domain"
	"github.com/nexus-platform/pms-integration/internal/repository"
)

// Handler holds all HTTP handlers.
type Handler struct {
	store *repository.Store
}

// NewHandler creates a new HTTP handler with the given store.
func NewHandler(store *repository.Store) *Handler {
	return &Handler{store: store}
}

// Router builds the chi router with all routes.
func (h *Handler) Router() chi.Router {
	r := chi.NewRouter()
	r.Use(recoverer)
	r.Use(requestLogger)

	r.Get("/health", h.health)
	r.Get("/ready", h.ready)
	r.Get("/live", h.live)

	// Guests
	r.Route("/tenants/{tenantId}/guests", func(r chi.Router) {
		r.Get("/", h.listGuests)
		r.Post("/", h.createGuest)
		r.Get("/{guestId}", h.getGuest)
		r.Patch("/{guestId}", h.updateGuest)
	})

	// Reservations
	r.Route("/tenants/{tenantId}/reservations", func(r chi.Router) {
		r.Post("/", h.createReservation)
		r.Get("/{reservationId}", h.getReservation)
		r.Post("/{reservationId}/checkin", h.checkIn)
		r.Post("/{reservationId}/checkout", h.checkOut)
		r.Patch("/{reservationId}/cancel", h.cancelReservation)
	})

	// Availability
	r.Get("/tenants/{tenantId}/availability", h.checkAvailability)

	// Properties
	r.Route("/tenants/{tenantId}/properties", func(r chi.Router) {
		r.Get("/", h.listProperties)
		r.Post("/", h.createProperty)
		r.Get("/{propertyId}", h.getProperty)
		r.Put("/{propertyId}", h.updateProperty)
	})

	// Room Types
	r.Route("/tenants/{tenantId}/properties/{propertyId}/room-types", func(r chi.Router) {
		r.Get("/", h.listRoomTypes)
		r.Post("/", h.createRoomType)
		r.Get("/{roomTypeId}", h.getRoomType)
	})

	// Rooms
	r.Route("/tenants/{tenantId}/properties/{propertyId}/rooms", func(r chi.Router) {
		r.Get("/", h.listRooms)
		r.Post("/", h.createRoom)
		r.Get("/{roomId}", h.getRoom)
		r.Put("/{roomId}", h.updateRoom)
		r.Patch("/{roomId}/status", h.updateRoomStatus)
		r.Patch("/{roomId}/housekeeping", h.updateHousekeeping)
	})

	return r
}

// ==================== MIDDLEWARE ====================

func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				respondError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		// In production, use structured logging. Here we keep it minimal.
		_ = start
	})
}

func tenantID(r *http.Request) string {
	return chi.URLParam(r, "tenantId")
}

// ==================== RESPONSE HELPERS ====================

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// ==================== HEALTH ====================

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "healthy", "service": "pms-integration"})
}

func (h *Handler) ready(w http.ResponseWriter, r *http.Request) {
	status := "ready"
	if h.store != nil {
		if err := h.store.Ping(r.Context()); err != nil {
			status = "not_ready"
		}
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": status})
}

func (h *Handler) live(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "alive"})
}
