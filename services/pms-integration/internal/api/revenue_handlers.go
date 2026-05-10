package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/nexus-platform/pms-integration/internal/domain"
)

func (h *Handler) listRevenueForecasts(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	propertyID := r.URL.Query().Get("property_id")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	if startDate == "" {
		startDate = time.Now().Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().AddDate(0, 0, 30).Format("2006-01-02")
	}
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)

	forecasts, err := h.store.RevenueForecasts.ListByProperty(r.Context(), tenantID, propertyID, start, end)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, forecasts)
}

func (h *Handler) createDynamicPricingRule(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var rule domain.DynamicPricingRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	rule.TenantID = tenantID
	if err := h.store.DynamicPricingRules.Create(r.Context(), &rule); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, rule)
}

func (h *Handler) listDynamicPricingRules(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	rules, err := h.store.DynamicPricingRules.List(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, rules)
}

func (h *Handler) getPriceRecommendations(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 { limit = 20 }

	recs, err := h.store.PriceRecommendations.ListPending(r.Context(), tenantID, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, recs)
}

func (h *Handler) applyPriceRecommendation(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	recID := chi.URLParam(r, "recommendationId")
	if err := h.store.PriceRecommendations.MarkApplied(r.Context(), tenantID, recID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "applied"})
}
