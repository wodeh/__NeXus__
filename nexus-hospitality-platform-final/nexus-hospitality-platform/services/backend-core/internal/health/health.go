// Package health provides Kubernetes-compatible health, readiness, and liveness probes.
package health

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"
)

// Server exposes health endpoints.
type Server struct {
	port      string
	server    *http.Server
	ready     atomic.Bool
	live      atomic.Bool
	startTime time.Time
}

// NewServer creates a health probe server.
func NewServer(port string) *Server {
	s := &Server{port: port, startTime: time.Now()}
	s.ready.Store(true)
	s.live.Store(true)
	return s
}

// Start launches the health HTTP server.
func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/ready", s.handleReady)
	mux.HandleFunc("/live", s.handleLive)
	s.server = &http.Server{Addr: ":" + s.port, Handler: mux}
	go func() {
		slog.Info("health server starting", slog.String("port", s.port))
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("health server failed", slog.String("error", err.Error()))
		}
	}()
	return nil
}

// Stop gracefully shuts down the health server.
func (s *Server) Stop() {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.server.Shutdown(ctx)
	}
}

// SetReady controls readiness state.
func (s *Server) SetReady(ready bool) { s.ready.Store(ready) }

// SetLive controls liveness state.
func (s *Server) SetLive(live bool) { s.live.Store(live) }

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"status": "healthy", "uptime": time.Since(s.startTime).String()})
}
func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	if s.ready.Load() {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
		return
	}
	writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not ready"})
}
func (s *Server) handleLive(w http.ResponseWriter, r *http.Request) {
	if s.live.Load() {
		writeJSON(w, http.StatusOK, map[string]string{"status": "alive"})
		return
	}
	writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not alive"})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
