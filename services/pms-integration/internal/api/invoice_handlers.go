package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// ==================== INVOICES ====================

func (h *Handler) listInvoices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 500 { limit = 50 }
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 { page = 1 }
	offset := (page - 1) * limit

	invoices, total, err := h.store.Invoices.ListByTenant(ctx, tenant, status, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"data": invoices, "total": total, "page": page, "limit": limit})
}

func (h *Handler) getInvoice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	invID := chi.URLParam(r, "invoiceId")
	inv, err := h.store.Invoices.GetByID(ctx, tenant, invID)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, inv)
}

func (h *Handler) createInvoice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	var req struct {
		FolioID     string                  `json:"folio_id"`
		GuestID     string                  `json:"guest_id"`
		Currency    string                  `json:"currency"`
		DueDate     string                  `json:"due_date"`
		LineItems   []domain.InvoiceLineItem `json:"line_items"`
		Notes       string                  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	due, _ := time.Parse("2006-01-02", req.DueDate)
	inv := &domain.Invoice{
		TenantID:   tenant,
		FolioID:    req.FolioID,
		GuestID:    req.GuestID,
		Currency:   req.Currency,
		DueDate:    due,
		LineItems:  req.LineItems,
		Notes:      req.Notes,
	}
	if err := inv.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.store.Invoices.Create(ctx, inv); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, inv)
}

func (h *Handler) updateInvoice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	invID := chi.URLParam(r, "invoiceId")
	var req struct {
		Status    string                  `json:"status"`
		LineItems []domain.InvoiceLineItem `json:"line_items"`
		Notes     string                  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	updates := &domain.Invoice{Status: req.Status, LineItems: req.LineItems, Notes: req.Notes}
	if err := h.store.Invoices.Update(ctx, tenant, invID, updates); err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *Handler) deleteInvoice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	invID := chi.URLParam(r, "invoiceId")
	if err := h.store.Invoices.Delete(ctx, tenant, invID); err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
