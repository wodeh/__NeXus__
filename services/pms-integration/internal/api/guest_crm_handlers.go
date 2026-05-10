package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// ==================== GUEST CRM ====================

func (h *Handler) listGuestProfiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 500 { limit = 50 }
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 { page = 1 }
	offset := (page - 1) * limit

	profiles, total, err := h.store.GuestProfiles.ListByTenant(ctx, tenant, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"data": profiles, "total": total, "page": page, "limit": limit})
}

func (h *Handler) getGuestProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	profileID := chi.URLParam(r, "profileId")
	p, err := h.store.GuestProfiles.GetByID(ctx, tenant, profileID)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, p)
}

func (h *Handler) createGuestProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	var req struct {
		FirstName string                 `json:"first_name"`
		LastName  string                 `json:"last_name"`
		Email     string                 `json:"email"`
		Phone     string                 `json:"phone"`
		Notes     string                 `json:"notes"`
		Preferences domain.GuestPreferences `json:"preferences"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	p := &domain.GuestProfile{
		TenantID:    tenant,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Email:       req.Email,
		Phone:       req.Phone,
		Notes:       req.Notes,
		Preferences: req.Preferences,
	}
	if err := h.store.GuestProfiles.Create(ctx, p); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, p)
}

func (h *Handler) updateGuestProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	profileID := chi.URLParam(r, "profileId")
	var req struct {
		VIPStatus   string `json:"vip_status"`
		LoyaltyTier string `json:"loyalty_tier"`
		Notes       string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	updates := &domain.GuestProfile{VIPStatus: req.VIPStatus, LoyaltyTier: req.LoyaltyTier, Notes: req.Notes}
	if err := h.store.GuestProfiles.Update(ctx, tenant, profileID, updates); err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *Handler) addCommunication(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	profileID := chi.URLParam(r, "profileId")
	var req struct {
		Channel   string `json:"channel"`
		Direction string `json:"direction"`
		Subject   string `json:"subject"`
		Body      string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	entry := domain.CommunicationEntry{
		Channel:   req.Channel,
		Direction: req.Direction,
		Subject:   req.Subject,
		Body:      req.Body,
		SentAt:    time.Now(),
	}
	if err := h.store.GuestProfiles.AddCommunication(ctx, tenant, profileID, entry); err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, entry)
}
