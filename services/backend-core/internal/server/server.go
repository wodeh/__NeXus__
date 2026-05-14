// Package server implements the tenant-aware HTTP API server for backend-core.
package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexus-platform/backend-core/internal/auth"
	"github.com/nexus-platform/backend-core/internal/config"
	"github.com/nexus-platform/backend-core/internal/events"
	"github.com/nexus-platform/backend-core/internal/health"
	"github.com/nexus-platform/backend-core/internal/metrics"
	"github.com/nexus-platform/backend-core/internal/repository"
)

// jwtAudience is the expected JWT audience claim.
const jwtAudience = "nexus-api"

// isPublicRoute returns true for routes that do not require JWT authentication.
func isPublicRoute(path string) bool {
	// Exact match routes
	if path == "/health" || path == "/ready" || path == "/live" {
		return true
	}
	// Auth endpoints (allow registration too if added later)
	if path == "/v1/auth/login" || strings.HasPrefix(path, "/v1/auth/login") {
		return true
	}
	return false
}

// jwtSecretFromEnv returns the HMAC secret for JWT validation.
// Falls back to a development-only key (NOT safe for production).
func jwtSecretFromEnv() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret != "" {
		return []byte(secret)
	}
	// Fallback development key — NOT for production
	return []byte("nexus-dev-jwt-secret-do-not-use-in-production-2024")
}

// Server wraps the HTTP server with observability and tenant isolation.
type Server struct {
	cfg        *config.Config
	repo       *repository.Store
	producer   *events.ProducerManager
	httpServer *http.Server
	healthSrv  *health.Server
	metrics    *metrics.Collector
}

// New creates a new backend-core server.
func New(cfg *config.Config, repo *repository.Store, producer *events.ProducerManager, metricsCollector *metrics.Collector) *Server {
	return &Server{
		cfg:       cfg,
		repo:      repo,
		producer:  producer,
		metrics:   metricsCollector,
		healthSrv: health.NewServer("8081"),
	}
}

// Start initializes and runs the HTTP server.
func (s *Server) Start(ctx context.Context) error {
	if err := s.healthSrv.Start(); err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "healthy", "service": "backend-core"})
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	mux.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "alive"})
	})
	mux.HandleFunc("/v1/tenant/", s.withTenant(s.handleTenantAPI))
	mux.HandleFunc("/v1/events/publish", s.withTenant(s.handlePublishEvent))
	s.registerReservationHandlers(mux)
	s.registerReservationDetailHandlers(mux)
	s.registerRoomHandlers(mux)
	s.registerPropertyHandlers(mux)
	s.registerAdminHandlers(mux)
	s.registerChannelManagerHandlers(mux)
	s.registerRevenueHandlers(mux)
	s.registerAgentHandlers(mux)
	s.registerLockHandlers(mux)
	s.registerWhatsAppHandlers(mux)
	s.registerReviewHandlers(mux)
	s.registerAuditHandlers(mux)
	s.registerIPTVHandlers(mux)
	s.registerCommHandlers(mux)
	s.registerGuestJourneyHandlers(mux)
	s.registerAuthHandlers(mux)
	s.registerVillaHandlers(mux)
	s.registerHousekeepingHandlers(mux)

	// Build middleware chain: CORS → JWT → Metrics
	// CORS is outermost, then JWT auth, then metrics tracking
	jwtValidator := auth.NewHMACValidator(jwtSecretFromEnv(), "nexus-backend", jwtAudience)
	jwtMiddleware := auth.JWTMiddleware(jwtValidator)

	// CORS wraps everything
	corsHandler := withCORS(mux.ServeHTTP)

	// Apply JWT middleware to the CORS-wrapped handler, but skip for public routes
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always let CORS handle OPTIONS preflight before any auth check
		if r.Method == http.MethodOptions {
			corsHandler(w, r)
			return
		}
		if isPublicRoute(r.URL.Path) {
			corsHandler(w, r)
			return
		}
		// Protected non-OPTIONS route: validate JWT first, then continue through CORS
		jwtMiddleware(http.HandlerFunc(corsHandler)).ServeHTTP(w, r)
	})

	wrapped := s.withMetrics(handler)

	s.httpServer = &http.Server{
		Addr:         ":" + s.cfg.HTTPPort,
		Handler:      wrapped,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	slog.Info("backend-core starting", slog.String("port", s.cfg.HTTPPort))
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.healthSrv.SetReady(false)
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			return err
		}
	}
	s.healthSrv.Stop()
	if s.repo != nil {
		s.repo.Close()
	}
	if s.producer != nil {
		_ = s.producer.Close()
	}
	return s.metrics.Shutdown(ctx)
}

func (s *Server) withMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start).Seconds()
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			// Try to get tenant_id from JWT context
			if tid, ok := auth.TenantIDFromContext(r.Context()); ok {
				tenantID = tid
			}
		}
		if tenantID == "" {
			tenantID = "unknown"
		}
		s.metrics.IncrementCounter("http_requests_total", tenantID)
		s.metrics.HistogramObserve("http_request_duration_seconds", duration, tenantID)
	})
}

func (s *Server) withTenant(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// First: try to get tenant_id from JWT claims (set by JWT middleware)
		tenantID, hasJWTTenant := auth.TenantIDFromContext(ctx)

		// Second: if no JWT tenant, fall back to X-Tenant-ID header for lookup
		if !hasJWTTenant || tenantID == "" {
			headerValue := r.Header.Get("X-Tenant-ID")
			if headerValue == "" {
				writeJSONError(w, http.StatusBadRequest, "missing X-Tenant-ID header")
				return
			}

			// Try parsing as UUID directly
			tenantID = headerValue
			if _, err := uuid.Parse(headerValue); err != nil {
				// Not a UUID — look up by external_id
				if s.repo == nil {
					writeJSONError(w, http.StatusServiceUnavailable, "database not configured")
					return
				}
				tenant, err := s.repo.Tenants.GetByExternalID(r.Context(), headerValue)
				if err != nil {
					writeJSONError(w, http.StatusNotFound, "tenant not found")
					return
				}
				tenantID = tenant.ID.String()
			}
		}

		// Bind tenant to DB session for RLS
		if s.repo != nil {
			if err := s.repo.SetTenant(r.Context(), tenantID); err != nil {
				slog.Warn("failed to set db tenant", slog.String("error", err.Error()))
			}
		}

		// Ensure tenant_id is in context for downstream handlers.
		// We always set the raw string key for compatibility with existing
		// handlers that use ctx.Value("tenant_id").(string).
		ctx = context.WithValue(ctx, "tenant_id", tenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (s *Server) handleTenantAPI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := auth.TenantIDFromContext(ctx)
	if !ok || tenantID == "" {
		tenantID, _ = ctx.Value("tenant_id").(string)
	}
	if tenantID == "" {
		writeJSONError(w, http.StatusInternalServerError, "missing or invalid tenant context")
		return
	}

	// Extract external_id from path: /v1/tenant/{external_id}
	externalID := r.URL.Path[len("/v1/tenant/"):]

	var response map[string]interface{}

	if s.repo != nil {
		tenant, err := s.repo.Tenants.GetByExternalID(r.Context(), externalID)
		if err == nil && tenant != nil {
			response = map[string]interface{}{
				"id":               tenant.ExternalID,
				"name":             tenant.Name,
				"property_type":    "boutique",
				"license_tier":     tenant.Tier,
				"license_status":   "active",
				"license_expires_at": "",
				"max_rooms":        5000,
				"max_users":        200,
				"capabilities": []string{
					"core:reservations", "core:guests", "core:properties", "core:rooms",
					"core:housekeeping", "core:settings", "core:audit_logs",
					"operations:floor_dashboard", "operations:room_blocks",
					"operations:group_reservations", "operations:maintenance", "operations:front_desk",
					"revenue:dynamic_pricing", "revenue:ota_integration",
					"revenue:revenue_forecasting", "revenue:agent_management",
					"enterprise:multi_property", "enterprise:advanced_crm",
					"enterprise:api_access", "enterprise:white_label", "enterprise:custom_reports",
					"villa:dashboard", "villa:properties", "villa:reservations", "villa:revenue", "villa:cleaner_tracking",
				},
				"settings": map[string]interface{}{
					"timezone":      "UTC",
					"currency_code": "USD",
					"date_format":   "YYYY-MM-DD",
					"language":      "en",
				},
				"created_at": tenant.CreatedAt.Format(time.RFC3339),
				"updated_at": tenant.UpdatedAt.Format(time.RFC3339),
			}
		}
	}

	if response == nil {
		// Fallback response
		response = map[string]interface{}{
			"id":               externalID,
			"name":             "Demo Hotel",
			"property_type":    "boutique",
			"license_tier":     "enterprise",
			"license_status":   "active",
			"license_expires_at": "",
			"max_rooms":        5000,
			"max_users":        200,
			"capabilities": []string{
				"core:reservations", "core:guests", "core:properties", "core:rooms",
				"core:housekeeping", "core:settings", "core:audit_logs",
				"operations:floor_dashboard", "operations:room_blocks",
				"operations:group_reservations", "operations:maintenance", "operations:front_desk",
				"revenue:dynamic_pricing", "revenue:ota_integration",
				"revenue:revenue_forecasting", "revenue:agent_management",
				"enterprise:multi_property", "enterprise:advanced_crm",
				"enterprise:api_access", "enterprise:white_label", "enterprise:custom_reports",
				"villa:dashboard", "villa:properties", "villa:reservations", "villa:revenue", "villa:cleaner_tracking",
			},
			"settings": map[string]interface{}{
				"timezone":      "UTC",
				"currency_code": "USD",
				"date_format":   "YYYY-MM-DD",
				"language":      "en",
			},
			"created_at": time.Now().UTC().Format(time.RFC3339),
			"updated_at": time.Now().UTC().Format(time.RFC3339),
		}
	}

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) handlePublishEvent(w http.ResponseWriter, r *http.Request) {
	if s.producer == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "event producer not configured")
		return
	}

	ctx := r.Context()
	tenantID, ok := auth.TenantIDFromContext(ctx)
	if !ok || tenantID == "" {
		tenantID, _ = ctx.Value("tenant_id").(string)
	}
	if tenantID == "" {
		writeJSONError(w, http.StatusInternalServerError, "missing or invalid tenant context")
		return
	}

	var req struct {
		EventType string                 `json:"event_type"`
		Payload   map[string]interface{} `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	event, err := events.NewEvent(req.EventType, "1.0", tenantID, "", "backend-core", req.Payload)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to create event")
		return
	}

	if err := s.producer.Publish(r.Context(), event); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to publish event")
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{"status": "published"})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
