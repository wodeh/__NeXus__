package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// ==================== ROOM BLOCKS ====================

func (h *Handler) listRoomBlocks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 500 { limit = 50 }
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 { page = 1 }
	offset := (page - 1) * limit

	blocks, total, err := h.store.RoomBlocks.ListByTenant(ctx, tenant, status, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"data": blocks, "total": total, "page": page, "limit": limit})
}

func (h *Handler) getRoomBlock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	blockID := chi.URLParam(r, "blockId")
	b, err := h.store.RoomBlocks.GetByID(ctx, tenant, blockID)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, b)
}

func (h *Handler) createRoomBlock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	var req struct {
		PropertyID string   `json:"property_id"`
		RoomIDs    []string `json:"room_ids"`
		StartDate  string   `json:"start_date"`
		EndDate    string   `json:"end_date"`
		Reason     string   `json:"reason"`
		Type       string   `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	start, _ := time.Parse("2006-01-02", req.StartDate)
	end, _ := time.Parse("2006-01-02", req.EndDate)
	b := &domain.RoomBlock{
		TenantID:   tenant,
		PropertyID: req.PropertyID,
		RoomIDs:    req.RoomIDs,
		StartDate:  start,
		EndDate:    end,
		Reason:     req.Reason,
		Type:       domain.RoomBlockType(req.Type),
		CreatedBy:  "api",
	}
	if err := b.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.store.RoomBlocks.Create(ctx, b); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, b)
}

func (h *Handler) updateRoomBlockStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	blockID := chi.URLParam(r, "blockId")
	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.store.RoomBlocks.UpdateStatus(ctx, tenant, blockID, domain.RoomBlockStatus(req.Status)); err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *Handler) deleteRoomBlock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	blockID := chi.URLParam(r, "blockId")
	if err := h.store.RoomBlocks.Delete(ctx, tenant, blockID); err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
