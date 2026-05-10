package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// ==================== CHECK-IN / CHECK-OUT ====================

func (h *Handler) createCheckIn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	var req struct {
		ReservationID string `json:"reservation_id"`
		GuestID       string `json:"guest_id"`
		RoomID        string `json:"room_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	ci := &domain.CheckInWorkflow{
		TenantID:      tenant,
		ReservationID: req.ReservationID,
		GuestID:       req.GuestID,
		RoomID:        req.RoomID,
	}
	if err := ci.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.store.CheckIns.CreateCheckIn(ctx, ci); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, ci)
}

func (h *Handler) getCheckIn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	ciID := chi.URLParam(r, "checkinId")
	ci, err := h.store.CheckIns.GetCheckIn(ctx, tenant, ciID)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, ci)
}

func (h *Handler) updateCheckIn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	ciID := chi.URLParam(r, "checkinId")
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.store.CheckIns.UpdateCheckIn(ctx, tenant, ciID, updates); err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *Handler) createCheckOut(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	var req struct {
		ReservationID string `json:"reservation_id"`
		GuestID       string `json:"guest_id"`
		RoomID        string `json:"room_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	co := &domain.CheckOutWorkflow{
		TenantID:      tenant,
		ReservationID: req.ReservationID,
		GuestID:       req.GuestID,
		RoomID:        req.RoomID,
	}
	if err := h.store.CheckIns.CreateCheckOut(ctx, co); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, co)
}

func (h *Handler) getCheckOut(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	coID := chi.URLParam(r, "checkoutId")
	co, err := h.store.CheckIns.GetCheckOut(ctx, tenant, coID)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, co)
}

func (h *Handler) updateCheckOut(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	coID := chi.URLParam(r, "checkoutId")
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.store.CheckIns.UpdateCheckOut(ctx, tenant, coID, updates); err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}
