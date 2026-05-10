package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// ==================== GROUP RESERVATIONS ====================

func (h *Handler) listGroups(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 500 { limit = 50 }
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 { page = 1 }
	offset := (page - 1) * limit

	groups, total, err := h.store.GroupReservations.ListByTenant(ctx, tenant, status, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"data": groups, "total": total, "page": page, "limit": limit})
}

func (h *Handler) getGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	groupID := chi.URLParam(r, "groupId")
	g, err := h.store.GroupReservations.GetByID(ctx, tenant, groupID)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, g)
}

func (h *Handler) createGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	var req struct {
		PropertyID      string `json:"property_id"`
		GroupName       string `json:"group_name"`
		ContactName     string `json:"contact_name"`
		ContactEmail    string `json:"contact_email"`
		ContactPhone    string `json:"contact_phone"`
		NumRooms        int    `json:"num_rooms"`
		NumGuests       int    `json:"num_guests"`
		CheckIn         string `json:"check_in"`
		CheckOut        string `json:"check_out"`
		RateCode        string `json:"rate_code"`
		SpecialRequests string `json:"special_requests"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	in, _ := time.Parse("2006-01-02", req.CheckIn)
	out, _ := time.Parse("2006-01-02", req.CheckOut)
	g := &domain.GroupReservation{
		TenantID:        tenant,
		PropertyID:      req.PropertyID,
		GroupName:       req.GroupName,
		ContactName:     req.ContactName,
		ContactEmail:    req.ContactEmail,
		ContactPhone:    req.ContactPhone,
		NumRooms:        req.NumRooms,
		NumGuests:       req.NumGuests,
		CheckIn:         in,
		CheckOut:        out,
		RateCode:        req.RateCode,
		SpecialRequests: req.SpecialRequests,
	}
	if err := g.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.store.GroupReservations.Create(ctx, g); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, g)
}

func (h *Handler) updateGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	groupID := chi.URLParam(r, "groupId")
	var req struct {
		GroupName string `json:"group_name"`
		Status    string `json:"status"`
		NumRooms  int    `json:"num_rooms"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	updates := &domain.GroupReservation{GroupName: req.GroupName, Status: req.Status, NumRooms: req.NumRooms}
	if err := h.store.GroupReservations.Update(ctx, tenant, groupID, updates); err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *Handler) deleteGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	groupID := chi.URLParam(r, "groupId")
	if err := h.store.GroupReservations.Delete(ctx, tenant, groupID); err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
