// Package server implements the tenant-aware HTTP API server for backend-core.
package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/nexus-platform/backend-core/internal/config"
	"github.com/nexus-platform/backend-core/internal/events"
	"github.com/nexus-platform/backend-core/internal/health"
	"github.com/nexus-platform/backend-core/internal/metrics"
	"github.com/nexus-platform/backend-core/internal/repository"
)

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
	s.registerChannelHandlers(mux)
	s.registerLockHandlers(mux)
	s.registerWhatsAppHandlers(mux)
	s.registerReviewHandlers(mux)
	s.registerAuditHandlers(mux)
	s.registerIPTVHandlers(mux)
	s.registerCommHandlers(mux)

	s.httpServer = &http.Server{
		Addr:         ":" + s.cfg.HTTPPort,
		Handler:      s.withMetrics(mux),
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
			tenantID = "unknown"
		}
		s.metrics.IncrementCounter("http_requests_total", tenantID)
		s.metrics.HistogramObserve("http_request_duration_seconds", duration, tenantID)
	})
}

func (s *Server) withTenant(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			http.Error(w, `{"error":"missing X-Tenant-ID header"}`, http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(r.Context(), "tenant_id", tenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (s *Server) handleTenantAPI(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenant_id").(string)
	if !ok || tenantID == "" {
		http.Error(w, `{"error":"missing or invalid tenant context"}`, http.StatusInternalServerError)
		return
	}

	dbStatus := "not_configured"
	if s.repo != nil {
		_, err := s.repo.GetTenantConfig(r.Context(), tenantID)
		if err != nil {
			dbStatus = "unavailable"
		} else {
			dbStatus = "connected"
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tenant_id": tenantID,
		"message":   "backend-core API",
		"db_status": dbStatus,
		"timestamp": time.Now().UTC(),
	})
}

func (s *Server) handlePublishEvent(w http.ResponseWriter, r *http.Request) {
	if s.producer == nil {
		http.Error(w, `{"error":"event producer not configured"}`, http.StatusServiceUnavailable)
		return
	}

	tenantID, ok := r.Context().Value("tenant_id").(string)
	if !ok || tenantID == "" {
		http.Error(w, `{"error":"missing or invalid tenant context"}`, http.StatusInternalServerError)
		return
	}

	var req struct {
		EventType string                 `json:"event_type"`
		Payload   map[string]interface{} `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	event, err := events.NewEvent(req.EventType, "1.0", tenantID, "", "backend-core", req.Payload)
	if err != nil {
		http.Error(w, `{"error":"failed to create event"}`, http.StatusInternalServerError)
		return
	}

	if err := s.producer.Publish(r.Context(), event); err != nil {
		http.Error(w, `{"error":"failed to publish event"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{"status": "published"})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
