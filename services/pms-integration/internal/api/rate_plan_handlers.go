package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// ==================== RATE PLANS ====================

func (h *Handler) listRatePlans(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 500 { limit = 50 }
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 { page = 1 }
	offset := (page - 1) * limit

	plans, total, err := h.store.RatePlans.ListByTenant(ctx, tenant, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"data": plans, "total": total, "page": page, "limit": limit})
}

func (h *Handler) getRatePlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	planID := chi.URLParam(r, "planId")
	p, err := h.store.RatePlans.GetByID(ctx, tenant, planID)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, p)
}

func (h *Handler) createRatePlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	var req struct {
		PropertyID    string                `json:"property_id"`
		Name          string                `json:"name"`
		Code          string                `json:"code"`
		Description   string                `json:"description"`
		SeasonalRates []domain.SeasonalRate `json:"seasonal_rates"`
		Restrictions  domain.RateRestriction `json:"restrictions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	p := &domain.RatePlanSeasonal{
		TenantID:      tenant,
		PropertyID:    req.PropertyID,
		Name:          req.Name,
		Code:          req.Code,
		Description:   req.Description,
		SeasonalRates: req.SeasonalRates,
		Restrictions:  req.Restrictions,
	}
	if err := p.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.store.RatePlans.Create(ctx, p); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, p)
}

func (h *Handler) updateRatePlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	planID := chi.URLParam(r, "planId")
	var req struct {
		Name          string                `json:"name"`
		Description   string                `json:"description"`
		SeasonalRates []domain.SeasonalRate `json:"seasonal_rates"`
		IsActive      bool                  `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	updates := &domain.RatePlanSeasonal{Name: req.Name, Description: req.Description, SeasonalRates: req.SeasonalRates, IsActive: req.IsActive}
	if err := h.store.RatePlans.Update(ctx, tenant, planID, updates); err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *Handler) deleteRatePlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	planID := chi.URLParam(r, "planId")
	if err := h.store.RatePlans.Delete(ctx, tenant, planID); err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
