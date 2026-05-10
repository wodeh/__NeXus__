// Package api provides HTTP REST handlers for the PMS integration service.
package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nexus-platform/pms-integration/internal/domain"
	"github.com/nexus-platform/pms-integration/internal/license"
	apiMiddleware "github.com/nexus-platform/pms-integration/internal/api/middleware"
	"github.com/nexus-platform/pms-integration/internal/repository"
)

// Handler holds all HTTP handlers.
type Handler struct {
	store      *repository.Store
	licenseSvc *license.Service
}

// NewHandler creates a new HTTP handler with the given store and license service.
func NewHandler(store *repository.Store, licenseSvc *license.Service) *Handler {
	return &Handler{store: store, licenseSvc: licenseSvc}
}

// Router builds the chi router with all routes.
func (h *Handler) Router() chi.Router {
	r := chi.NewRouter()
	r.Use(recoverer)
	r.Use(requestLogger)

	// Root API info
	r.Get("/", h.apiInfo)
	r.Get("/health", h.health)
	r.Get("/ready", h.ready)
	r.Get("/live", h.live)

	// Demo / test endpoints (no DB required)
	r.Get("/demo/info", h.demoInfo)

	// Tenant configuration (no license check needed for reading config)
	th := NewTenantHandler(h.licenseSvc)
	r.Route("/tenants/{tenantId}", func(r chi.Router) {
		r.Get("/config", th.getTenantConfig)
	})
	r.Get("/property-types", th.listPropertyTypes)
	r.Get("/license-tiers", th.listLicenseTiers)
	r.Get("/capabilities", th.listCapabilities)

	// Guests
	r.Route("/tenants/{tenantId}/guests", func(r chi.Router) {
		r.Use(apiMiddleware.License(h.licenseSvc, domain.CapGuests))
		r.Get("/", h.listGuests)
		r.Post("/", h.createGuest)
		r.Get("/{guestId}", h.getGuest)
		r.Patch("/{guestId}", h.updateGuest)
	})

	// Reservations
	r.Route("/tenants/{tenantId}/reservations", func(r chi.Router) {
		r.Use(apiMiddleware.License(h.licenseSvc, domain.CapReservations))
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
		r.Use(apiMiddleware.License(h.licenseSvc, domain.CapProperties))
		r.Get("/", h.listProperties)
		r.Post("/", h.createProperty)
		r.Get("/{propertyId}", h.getProperty)
		r.Put("/{propertyId}", h.updateProperty)
	})

	// Room Types
	r.Route("/tenants/{tenantId}/properties/{propertyId}/room-types", func(r chi.Router) {
		r.Use(apiMiddleware.License(h.licenseSvc, domain.CapRooms))
		r.Get("/", h.listRoomTypes)
		r.Post("/", h.createRoomType)
		r.Get("/{roomTypeId}", h.getRoomType)
	})

	// Rooms
	r.Route("/tenants/{tenantId}/properties/{propertyId}/rooms", func(r chi.Router) {
		r.Use(apiMiddleware.License(h.licenseSvc, domain.CapRooms))
		r.Get("/", h.listRooms)
		r.Post("/", h.createRoom)
		r.Get("/{roomId}", h.getRoom)
		r.Put("/{roomId}", h.updateRoom)
		r.Patch("/{roomId}/status", h.updateRoomStatus)
		r.Patch("/{roomId}/housekeeping", h.updateHousekeeping)
	})

	// Audit Logs
	r.Route("/tenants/{tenantId}/audit", func(r chi.Router) {
		r.Use(apiMiddleware.License(h.licenseSvc, domain.CapAuditLogs))
		r.Get("/", h.listAuditLogs)
		r.Post("/", h.createAuditLog)
		r.Get("/resource/{resourceType}/{resourceId}", h.getResourceAuditTrail)
	})

	// GDPR
	r.Route("/tenants/{tenantId}/gdpr", func(r chi.Router) {
		r.Use(apiMiddleware.License(h.licenseSvc, domain.CapAuditLogs))
		r.Post("/requests", h.createGDPRRequest)
		r.Get("/guests/{guestId}/export", h.getGuestDataExport)
		r.Delete("/guests/{guestId}", h.deleteGuestData)
	})

	// Revenue Management
	r.Route("/tenants/{tenantId}/revenue", func(r chi.Router) {
		r.Use(apiMiddleware.License(h.licenseSvc, domain.CapDynamicPricing, domain.CapRevenueForecast))
		r.Get("/forecasts", h.listRevenueForecasts)
		r.Post("/pricing-rules", h.createDynamicPricingRule)
		r.Get("/pricing-rules", h.listDynamicPricingRules)
		r.Get("/recommendations", h.getPriceRecommendations)
		r.Post("/recommendations/{recommendationId}/apply", h.applyPriceRecommendation)
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

func (h *Handler) apiInfo(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"service":    "pms-integration",
		"version":    "1.0.0",
		"health":     "/health",
		"ready":      "/ready",
		"live":       "/live",
		"demo":       "/demo/info",
		"endpoints": map[string]string{
			"guests":        "GET/POST /tenants/{tenantId}/guests",
			"reservations":  "GET/POST /tenants/{tenantId}/reservations",
			"properties":    "GET/POST /tenants/{tenantId}/properties",
			"room_types":    "GET/POST /tenants/{tenantId}/properties/{propertyId}/room-types",
			"rooms":         "GET/POST /tenants/{tenantId}/properties/{propertyId}/rooms",
			"audit_logs":    "GET/POST /tenants/{tenantId}/audit",
			"gdpr":          "GET/POST/DELETE /tenants/{tenantId}/gdpr",
			"revenue":       "GET/POST /tenants/{tenantId}/revenue",
			"metrics":       "GET /metrics (port 9090)",
		},
	})
}

func (h *Handler) demoInfo(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Demo endpoints for live testing",
		"examples": []map[string]string{
			{"method": "GET",  "url": "http://localhost:8080/",                       "desc": "API info"},
			{"method": "GET",  "url": "http://localhost:8080/health",                  "desc": "Health check"},
			{"method": "GET",  "url": "http://localhost:8080/ready",                   "desc": "Readiness check"},
			{"method": "GET",  "url": "http://localhost:8080/demo/info",               "desc": "This endpoint"},
			{"method": "GET",  "url": "http://localhost:8080/tenants/demo/guests",      "desc": "List guests (needs DB)"},
			{"method": "GET",  "url": "http://localhost:8080/tenants/demo/properties", "desc": "List properties (needs DB)"},
			{"method": "GET",  "url": "http://localhost:9090/metrics",                 "desc": "Prometheus metrics"},
		},
		"note": "Replace 'demo' with a real tenant_id. These endpoints require PostgreSQL. Set DATABASE_URL env var.",
	})
}
