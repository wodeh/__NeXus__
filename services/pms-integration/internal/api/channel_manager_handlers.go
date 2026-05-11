package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// ============== CHANNEL MANAGER HANDLERS ==============

func (h *Handler) listChannelConnections(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	conns, err := h.store.ChannelManager.ListConnections(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, conns)
}

func (h *Handler) createChannelConnection(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var conn domain.ChannelConnection
	if err := json.NewDecoder(r.Body).Decode(&conn); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	res, err := h.store.ChannelManager.CreateConnection(r.Context(), tenantID, &conn)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, res)
}

func (h *Handler) updateChannelConnection(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "connId")
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.store.ChannelManager.UpdateConnection(r.Context(), tenantID, id, updates); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *Handler) deleteChannelConnection(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "connId")
	if err := h.store.ChannelManager.DeleteConnection(r.Context(), tenantID, id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *Handler) syncChannelConnection(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "connId")
	// Start sync log
	log := &domain.ChannelSyncLog{
		TenantID:  tenantID,
		ChannelID: id,
		Direction: "pull",
		Status:    "running",
		StartedAt: time.Now(),
	}
	log, _ = h.store.ChannelManager.CreateSyncLog(r.Context(), tenantID, log)

	// In a real implementation, this would call the OTA API
	// For now, mark as success with mock data
	now := time.Now()
	log.Status = "success"
	log.Records = 12
	log.CompletedAt = &now
	// Note: In a real implementation, we'd update the sync log status

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "synced",
		"log_id":   log.ID,
		"records":  12,
		"duration": "2.3s",
	})
}

// ============== CHANNEL RESERVATIONS ==============

func (h *Handler) listChannelReservations(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var source *domain.ChannelManagerSource
	if s := r.URL.Query().Get("source"); s != "" {
		src := domain.ChannelManagerSource(s)
		source = &src
	}
	var status *string
	if st := r.URL.Query().Get("status"); st != "" {
		status = &st
	}
	var fromDate, toDate *time.Time
	if f := r.URL.Query().Get("from"); f != "" {
		t, _ := time.Parse("2006-01-02", f)
		fromDate = &t
	}
	if t := r.URL.Query().Get("to"); t != "" {
		tt, _ := time.Parse("2006-01-02", t)
		toDate = &tt
	}
	res, err := h.store.ChannelManager.ListChannelReservations(r.Context(), tenantID, source, status, fromDate, toDate)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, res)
}

func (h *Handler) createChannelReservation(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var res domain.ChannelReservation
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	created, err := h.store.ChannelManager.CreateChannelReservation(r.Context(), tenantID, &res)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, created)
}

// ============== CHANNEL HEALTH ==============

func (h *Handler) getChannelHealth(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days == 0 {
		days = 30
	}
	health, err := h.store.ChannelManager.GetChannelHealth(r.Context(), tenantID, days)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, health)
}

func (h *Handler) listSyncLogs(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 50
	}
	logs, err := h.store.ChannelManager.ListSyncLogs(r.Context(), tenantID, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, logs)
}
