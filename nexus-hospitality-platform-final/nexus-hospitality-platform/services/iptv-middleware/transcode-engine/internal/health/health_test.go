package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	s := NewServer("8081")
	if s == nil {
		t.Fatal("expected non-nil server")
	}
	if !s.ready.Load() {
		t.Error("expected ready to be true by default")
	}
	if !s.live.Load() {
		t.Error("expected live to be true by default")
	}
}

func TestHandleHealth(t *testing.T) {
	s := NewServer("0")
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	s.handleHealth(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleReady(t *testing.T) {
	s := NewServer("0")
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()
	s.handleReady(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleReady_NotReady(t *testing.T) {
	s := NewServer("0")
	s.SetReady(false)
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()
	s.handleReady(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}

func TestHandleLive_NotAlive(t *testing.T) {
	s := NewServer("0")
	s.SetLive(false)
	req := httptest.NewRequest(http.MethodGet, "/live", nil)
	w := httptest.NewRecorder()
	s.handleLive(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}

func TestStartAndStop(t *testing.T) {
	s := NewServer("0")
	if err := s.Start(); err != nil {
		t.Fatalf("Start error = %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	s.Stop()
}
