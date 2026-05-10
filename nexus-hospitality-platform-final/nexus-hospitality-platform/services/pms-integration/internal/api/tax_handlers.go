package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/nexus-platform/pms-integration/internal/domain"
	"github.com/nexus-platform/pms-integration/internal/middleware"
)

func (h *Handler) listTaxRates(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	rates, err := h.store.TaxRates.ListActive(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, rates)
}

func (h *Handler) createTaxRate(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var tr domain.TaxRate
	if err := json.NewDecoder(r.Body).Decode(&tr); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	tr.TenantID = tenantID
	if err := h.store.TaxRates.Create(r.Context(), &tr); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, tr)
}

func (h *Handler) getTaxRate(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "taxRateId")
	tr, err := h.store.TaxRates.GetByID(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, tr)
}

func (h *Handler) updateTaxRate(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "taxRateId")
	var tr domain.TaxRate
	if err := json.NewDecoder(r.Body).Decode(&tr); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	tr.ID = id
	tr.TenantID = tenantID
	if err := h.store.TaxRates.Update(r.Context(), tenantID, &tr); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, tr)
}

func (h *Handler) deleteTaxRate(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "taxRateId")
	if err := h.store.TaxRates.Delete(r.Context(), tenantID, id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *Handler) calculateTax(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var req struct {
		Amount      float64 `json:"amount"`
		CountryCode string  `json:"country_code"`
		StateCode   string  `json:"state_code"`
		Category    string  `json:"category"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	rates, err := h.store.TaxRates.ListByJurisdiction(r.Context(), tenantID, req.CountryCode, req.StateCode)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var calc domain.TaxCalculation
	calc.BaseAmount = req.Amount
	for _, rate := range rates {
		if rate.AppliesTo != "all" && rate.AppliesTo != req.Category {
			continue
		}
		taxAmount := req.Amount * rate.Rate
		calc.TaxBreakdown = append(calc.TaxBreakdown, domain.TaxLine{
			TaxRateID: rate.ID,
			Name:      rate.Name,
			Rate:      rate.Rate,
			TaxAmount: taxAmount,
		})
		calc.TotalTax += taxAmount
		if rate.IsCompound {
			req.Amount += taxAmount
		}
	}
	calc.TotalWithTax = calc.BaseAmount + calc.TotalTax
	respondJSON(w, http.StatusOK, calc)
}

func (h *Handler) listTaxExemptions(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	guestID := r.URL.Query().Get("guest_id")
	if guestID != "" {
		exemptions, err := h.store.TaxExemptions.ListByGuest(r.Context(), tenantID, guestID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, exemptions)
		return
	}
	respondError(w, http.StatusBadRequest, "guest_id required")
}

func (h *Handler) createTaxExemption(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var te domain.TaxExemption
	if err := json.NewDecoder(r.Body).Decode(&te); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	te.TenantID = tenantID
	if err := h.store.TaxExemptions.Create(r.Context(), &te); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, te)
}
