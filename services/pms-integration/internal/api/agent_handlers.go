package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// ==================== AGENTS ====================

func (h *Handler) listAgents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	agentType := r.URL.Query().Get("type")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 500 { limit = 50 }
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 { page = 1 }
	offset := (page - 1) * limit

	agents, total, err := h.store.Agents.ListByTenant(ctx, tenant, agentType, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"data": agents, "total": total, "page": page, "limit": limit})
}

func (h *Handler) getAgent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	agentID := chi.URLParam(r, "agentId")
	a, err := h.store.Agents.GetByID(ctx, tenant, agentID)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, a)
}

func (h *Handler) createAgent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	var req struct {
		Name          string  `json:"name"`
		Type          string  `json:"type"`
		CommissionPct float64 `json:"commission_pct"`
		ContactName   string  `json:"contact_name"`
		ContactEmail  string  `json:"contact_email"`
		ContactPhone  string  `json:"contact_phone"`
		ContractRef   string  `json:"contract_ref"`
		SourceCode    string  `json:"source_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	a := &domain.Agent{
		TenantID:      tenant,
		Name:          req.Name,
		Type:          req.Type,
		CommissionPct: req.CommissionPct,
		ContactName:   req.ContactName,
		ContactEmail:  req.ContactEmail,
		ContactPhone:  req.ContactPhone,
		ContractRef:   req.ContractRef,
		SourceCode:    req.SourceCode,
		IsActive:      true,
	}
	if err := a.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.store.Agents.Create(ctx, a); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, a)
}

func (h *Handler) updateAgent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	agentID := chi.URLParam(r, "agentId")
	var req struct {
		Name          string  `json:"name"`
		CommissionPct float64 `json:"commission_pct"`
		ContactName   string  `json:"contact_name"`
		IsActive      bool    `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	updates := &domain.Agent{Name: req.Name, CommissionPct: req.CommissionPct, ContactName: req.ContactName, IsActive: req.IsActive}
	if err := h.store.Agents.Update(ctx, tenant, agentID, updates); err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *Handler) deleteAgent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	agentID := chi.URLParam(r, "agentId")
	if err := h.store.Agents.Delete(ctx, tenant, agentID); err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
