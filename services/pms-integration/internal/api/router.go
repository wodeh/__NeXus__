// Package api provides HTTP REST handlers for the PMS integration service.
package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nexus-platform/pms-integration/internal/auth"
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

	// Auth (no license check for registration/login)
	jwtSvc := auth.NewService(auth.DefaultConfig())
	r.Route("/tenants/{tenantId}/auth", func(r chi.Router) {
		r.Post("/register", h.register)
		r.Post("/login", h.login)
		r.With(apiMiddleware.Auth(jwtSvc)).Get("/me", h.me)
	})

	// Users (requires auth)
	r.Route("/tenants/{tenantId}/users", func(r chi.Router) {
		r.Use(apiMiddleware.Auth(jwtSvc))
		r.Get("/", h.listUsers)
	})

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
		r.Get("/", h.listReservations)
		r.Post("/", h.createReservation)
		r.Get("/{reservationId}", h.getReservation)
		r.Post("/{reservationId}/checkin", h.checkIn)
		r.Post("/{reservationId}/checkout", h.checkOut)
		r.Patch("/{reservationId}/cancel", h.cancelReservation)
		r.Patch("/{reservationId}/room", h.moveReservation)
	})

	// Tapechart
	r.Get("/tenants/{tenantId}/tapechart", h.getTapechart)

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

	// Room Blocks
	r.Route("/tenants/{tenantId}/room-blocks", func(r chi.Router) {
		r.Use(apiMiddleware.License(h.licenseSvc, domain.CapRoomBlocks))
		r.Get("/", h.listRoomBlocks)
		r.Post("/", h.createRoomBlock)
		r.Get("/{blockId}", h.getRoomBlock)
		r.Patch("/{blockId}/status", h.updateRoomBlockStatus)
		r.Delete("/{blockId}", h.deleteRoomBlock)
	})

	// Group Reservations
	r.Route("/tenants/{tenantId}/groups", func(r chi.Router) {
		r.Use(apiMiddleware.License(h.licenseSvc, domain.CapGroupReservations))
		r.Get("/", h.listGroups)
		r.Post("/", h.createGroup)
		r.Get("/{groupId}", h.getGroup)
		r.Patch("/{groupId}", h.updateGroup)
		r.Delete("/{groupId}", h.deleteGroup)
	})

	// Guest CRM / Profiles
	r.Route("/tenants/{tenantId}/guest-profiles", func(r chi.Router) {
		r.Use(apiMiddleware.License(h.licenseSvc, domain.CapAdvancedCRM))
		r.Get("/", h.listGuestProfiles)
		r.Post("/", h.createGuestProfile)
		r.Get("/{profileId}", h.getGuestProfile)
		r.Patch("/{profileId}", h.updateGuestProfile)
		r.Post("/{profileId}/communications", h.addCommunication)
	})

	// Agents / Partners
	r.Route("/tenants/{tenantId}/agents", func(r chi.Router) {
		r.Use(apiMiddleware.License(h.licenseSvc, domain.CapAgentManagement))
		r.Get("/", h.listAgents)
		r.Post("/", h.createAgent)
		r.Get("/{agentId}", h.getAgent)
		r.Patch("/{agentId}", h.updateAgent)
		r.Delete("/{agentId}", h.deleteAgent)
	})

	// Rate Plans
	r.Route("/tenants/{tenantId}/rate-plans", func(r chi.Router) {
		r.Use(apiMiddleware.License(h.licenseSvc, domain.CapDynamicPricing))
		r.Get("/", h.listRatePlans)
		r.Post("/", h.createRatePlan)
		r.Get("/{planId}", h.getRatePlan)
		r.Patch("/{planId}", h.updateRatePlan)
		r.Delete("/{planId}", h.deleteRatePlan)
	})

	// Check-In / Check-Out
	r.Route("/tenants/{tenantId}/checkins", func(r chi.Router) {
		r.Use(apiMiddleware.License(h.licenseSvc, domain.CapReservations))
		r.Post("/", h.createCheckIn)
		r.Get("/{checkinId}", h.getCheckIn)
		r.Patch("/{checkinId}", h.updateCheckIn)
	})
	r.Route("/tenants/{tenantId}/checkouts", func(r chi.Router) {
		r.Use(apiMiddleware.License(h.licenseSvc, domain.CapReservations))
		r.Post("/", h.createCheckOut)
		r.Get("/{checkoutId}", h.getCheckOut)
		r.Patch("/{checkoutId}", h.updateCheckOut)
	})

	// Invoices
	r.Route("/tenants/{tenantId}/invoices", func(r chi.Router) {
		r.Use(apiMiddleware.License(h.licenseSvc, domain.CapReservations))
		r.Get("/", h.listInvoices)
		r.Post("/", h.createInvoice)
		r.Get("/{invoiceId}", h.getInvoice)
		r.Patch("/{invoiceId}", h.updateInvoice)
		r.Delete("/{invoiceId}", h.deleteInvoice)
	})

	// IPTV Module
	r.Route("/tenants/{tenantId}/iptv", func(r chi.Router) {
		r.Use(apiMiddleware.License(h.licenseSvc, domain.CapIPTVBasic))
		r.Get("/channels", h.listIPTVChannels)
		r.Post("/channels", h.createIPTVChannel)
		r.Get("/channels/{channelId}", h.getIPTVChannel)
		r.Patch("/channels/{channelId}", h.updateIPTVChannel)
		r.Delete("/channels/{channelId}", h.deleteIPTVChannel)
		r.Get("/content", h.listIPTVContent)
		r.Post("/content", h.createIPTVContent)
		r.Get("/content/{contentId}", h.getIPTVContent)
		r.Patch("/content/{contentId}", h.updateIPTVContent)
		r.Delete("/content/{contentId}", h.deleteIPTVContent)
		r.Get("/room-bindings", h.listIPTVRoomBindings)
		r.Post("/room-bindings", h.createIPTVRoomBinding)
		r.Get("/room-bindings/{bindingId}", h.getIPTVRoomBinding)
		r.Patch("/room-bindings/{bindingId}", h.updateIPTVRoomBinding)
		r.Delete("/room-bindings/{bindingId}", h.deleteIPTVRoomBinding)
		r.Get("/rooms/{roomId}/welcome", h.getIPTVWelcomeScreen)
		r.Get("/analytics", h.getIPTVAnalytics)
	})

	// Smart Lock Module
	r.Route("/tenants/{tenantId}/locks", func(r chi.Router) {
		r.Use(apiMiddleware.License(h.licenseSvc, domain.CapSmartLocks))
		r.Get("/overview", h.getLockOverview)
		r.Get("/", h.listSmartLocks)
		r.Post("/", h.createSmartLock)
		r.Get("/{lockId}", h.getSmartLock)
		r.Patch("/{lockId}", h.updateSmartLock)
		r.Delete("/{lockId}", h.deleteSmartLock)
		r.Post("/{lockId}/unlock", h.remoteUnlock)
		r.Post("/{lockId}/lock", h.remoteLock)
		r.Get("/{lockId}/access-codes", h.listAccessCodes)
		r.Post("/{lockId}/access-codes", h.createAccessCode)
		r.Post("/{lockId}/access-codes/{codeId}/revoke", h.revokeAccessCode)
		r.Get("/{lockId}/events", h.listLockEvents)
		r.Get("/by-room/{roomId}", h.getLockByRoom)
	})

	// Channel Manager
	r.Route("/tenants/{tenantId}/channels", func(r chi.Router) {
		r.Use(apiMiddleware.License(h.licenseSvc, domain.CapChannelManager))
		r.Get("/", h.listChannelConnections)
		r.Post("/", h.createChannelConnection)
		r.Get("/health", h.getChannelHealth)
		r.Get("/sync-logs", h.listSyncLogs)
		r.Get("/reservations", h.listChannelReservations)
		r.Post("/reservations", h.createChannelReservation)
		r.Route("/{connId}", func(r chi.Router) {
			r.Patch("/", h.updateChannelConnection)
			r.Delete("/", h.deleteChannelConnection)
			r.Post("/sync", h.syncChannelConnection)
		})
	})

	// WhatsApp Bot
	r.Route("/tenants/{tenantId}/whatsapp", func(r chi.Router) {
		r.Use(apiMiddleware.License(h.licenseSvc, domain.CapWhatsAppBot))
		r.Get("/config", h.getWhatsAppConfig)
		r.Post("/config", h.saveWhatsAppConfig)
		r.Get("/conversations", h.listWhatsAppConversations)
		r.Get("/conversations/{phone}", h.getWhatsAppConversation)
		r.Get("/conversations/{convId}/messages", h.listWhatsAppMessages)
	})
	r.Post("/webhooks/whatsapp/{tenantId}", h.whatsAppWebhook)

	// Booking Engine
	r.Route("/tenants/{tenantId}/booking", func(r chi.Router) {
		r.Use(apiMiddleware.License(h.licenseSvc, domain.CapDirectBooking))
		r.Get("/promo-codes", h.listPromoCodes)
		r.Post("/promo-codes", h.createPromoCode)
		r.Get("/promo-codes/{code}/validate", h.validatePromoCode)
		r.Delete("/promo-codes/{codeId}", h.deletePromoCode)
		r.Get("/widget-config", h.getWidgetConfig)
		r.Post("/widget-config", h.saveWidgetConfig)
		r.Get("/upsells", h.listUpsells)
		r.Post("/upsells", h.createUpsell)
		r.Post("/sessions", h.createBookingSession)
		r.Get("/sessions/{sessionId}", h.getBookingSession)
		r.Post("/sessions/{sessionId}/confirm", h.confirmBookingSession)
	})

	// Public booking widget (no auth)
	r.Route("/public/{tenantId}/booking", func(r chi.Router) {
		r.Get("/widget-config", h.publicGetWidgetConfig)
		r.Get("/availability", h.publicCheckAvailability)
		r.Post("/", h.publicCreateBooking)
	})

	// Reviews
	r.Route("/tenants/{tenantId}/reviews", func(r chi.Router) {
		r.Use(apiMiddleware.License(h.licenseSvc, domain.CapGuestReviews))
		r.Get("/", h.listReviews)
		r.Post("/", h.createReview)
		r.Get("/snapshot", h.getRatingSnapshot)
		r.Get("/requests", h.listReviewRequests)
		r.Post("/requests", h.createReviewRequest)
		r.Post("/{reviewId}/respond", h.respondToReview)
	})
	r.Post("/public/{tenantId}/reviews", h.publicSubmitReview)

	// Communications
	r.Route("/tenants/{tenantId}/communications", func(r chi.Router) {
		r.Use(apiMiddleware.License(h.licenseSvc, domain.CapCommunications))
		r.Get("/templates", h.listTemplates)
		r.Post("/templates", h.createTemplate)
		r.Patch("/templates/{templateId}", h.updateTemplate)
		r.Delete("/templates/{templateId}", h.deleteTemplate)
		r.Get("/sequences", h.listSequences)
		r.Post("/sequences", h.createSequence)
		r.Get("/sequences/{seqId}/steps", h.getSequenceSteps)
		r.Post("/sequences/{seqId}/steps", h.addSequenceStep)
		r.Get("/scheduled", h.listScheduledCommunications)
		r.Post("/scheduled", h.scheduleCommunication)
		r.Post("/scheduled/{commId}/cancel", h.cancelScheduledCommunication)
		r.Get("/logs", h.listCommunicationLogs)
		r.Post("/send", h.sendImmediate)
	})

	// Housekeeping Tasks + Staff + Manager Dashboard
	r.Route("/tenants/{tenantId}/housekeeping", func(r chi.Router) {
		r.Use(apiMiddleware.License(h.licenseSvc, domain.CapHousekeeping))
		r.Get("/dashboard", h.getHousekeepingDashboard)
		r.Get("/tasks", h.listHousekeepingTasks)
		r.Post("/tasks", h.createHousekeepingTask)
		r.Get("/tasks/{taskId}", h.getHousekeepingTask)
		r.Patch("/tasks/{taskId}", h.updateHousekeepingTask)
		r.Delete("/tasks/{taskId}", h.deleteHousekeepingTask)
		r.Post("/tasks/{taskId}/assign", h.assignHousekeepingTask)
		r.Post("/tasks/{taskId}/start", h.startHousekeepingTask)
		r.Post("/tasks/{taskId}/complete", h.completeHousekeepingTask)
		r.Post("/tasks/bulk-assign", h.bulkAssignTasks)
		r.Get("/staff", h.listHousekeepingStaff)
		r.Post("/staff", h.createHousekeepingStaff)
		r.Patch("/staff/{staffId}", h.updateHousekeepingStaff)
		r.Delete("/staff/{staffId}", h.deleteHousekeepingStaff)
		r.Get("/staff/{staffId}/performance", h.getStaffPerformance)
		r.Get("/checklists", h.getChecklistTemplate)
		r.Post("/checklists", h.saveChecklistTemplate)
		r.Get("/cleaner-cards", h.listCleanerCards)
		r.Post("/cleaner-cards", h.createCleanerCard)
		r.Post("/cleaner-cards/{cardId}/revoke", h.revokeCleanerCard)
		r.Get("/supplies", h.listSupplies)
		r.Post("/supplies/update", h.updateSupplyStock)
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
			"guests":           "GET/POST /tenants/{tenantId}/guests",
			"reservations":     "GET/POST /tenants/{tenantId}/reservations",
			"properties":       "GET/POST /tenants/{tenantId}/properties",
			"room_types":       "GET/POST /tenants/{tenantId}/properties/{propertyId}/room-types",
			"rooms":            "GET/POST /tenants/{tenantId}/properties/{propertyId}/rooms",
			"audit_logs":       "GET/POST /tenants/{tenantId}/audit",
			"gdpr":             "GET/POST/DELETE /tenants/{tenantId}/gdpr",
			"revenue":          "GET/POST /tenants/{tenantId}/revenue",
			"iptv_channels":    "GET/POST /tenants/{tenantId}/iptv/channels",
			"iptv_content":     "GET/POST /tenants/{tenantId}/iptv/content",
			"iptv_bindings":    "GET/POST /tenants/{tenantId}/iptv/room-bindings",
			"iptv_welcome":     "GET /tenants/{tenantId}/iptv/rooms/{roomId}/welcome",
			"smart_locks":      "GET/POST /tenants/{tenantId}/locks",
			"lock_unlock":      "POST /tenants/{tenantId}/locks/{lockId}/unlock",
			"lock_codes":       "GET/POST /tenants/{tenantId}/locks/{lockId}/access-codes",
			"lock_events":      "GET /tenants/{tenantId}/locks/{lockId}/events",
			"metrics":          "GET /metrics (port 9090)",
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
