package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// ==================== PROPERTIES ====================

func (h *Handler) listProperties(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 { limit = 50 }
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 { page = 1 }
	offset := (page - 1) * limit

	props, err := h.store.Properties.ListByTenant(ctx, tenant, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"properties": props,
		"page":       page,
		"limit":      limit,
	})
}

func (h *Handler) createProperty(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)

	var req struct {
		Name         string            `json:"name"`
		Slug         string            `json:"slug"`
		Address      map[string]string `json:"address"`
		Timezone     string            `json:"timezone"`
		CurrencyCode string            `json:"currency_code"`
		ContactEmail string            `json:"contact_email"`
		ContactPhone string            `json:"contact_phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Slug == "" {
		respondError(w, http.StatusBadRequest, "name, slug required")
		return
	}
	if req.Timezone == "" { req.Timezone = "UTC" }
	if req.CurrencyCode == "" { req.CurrencyCode = "USD" }

	p := domain.NewProperty(tenant, req.Name, req.Slug, req.Timezone, req.CurrencyCode)
	p.Address = req.Address
	p.ContactEmail = req.ContactEmail
	p.ContactPhone = req.ContactPhone

	if err := h.store.Properties.Create(ctx, p); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, p)
}

func (h *Handler) getProperty(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	propID := chi.URLParam(r, "propertyId")

	p, err := h.store.Properties.GetByID(ctx, tenant, propID)
	if err != nil {
		respondError(w, http.StatusNotFound, "property not found")
		return
	}
	respondJSON(w, http.StatusOK, p)
}

func (h *Handler) updateProperty(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	propID := chi.URLParam(r, "propertyId")

	p, err := h.store.Properties.GetByID(ctx, tenant, propID)
	if err != nil {
		respondError(w, http.StatusNotFound, "property not found")
		return
	}

	var req struct {
		Name         string            `json:"name"`
		Slug         string            `json:"slug"`
		Address      map[string]string `json:"address"`
		Timezone     string            `json:"timezone"`
		CurrencyCode string            `json:"currency_code"`
		ContactEmail string            `json:"contact_email"`
		ContactPhone string            `json:"contact_phone"`
		IsActive     *bool             `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name != "" { p.Name = req.Name }
	if req.Slug != "" { p.Slug = req.Slug }
	if req.Address != nil { p.Address = req.Address }
	if req.Timezone != "" { p.Timezone = req.Timezone }
	if req.CurrencyCode != "" { p.CurrencyCode = req.CurrencyCode }
	if req.ContactEmail != "" { p.ContactEmail = req.ContactEmail }
	if req.ContactPhone != "" { p.ContactPhone = req.ContactPhone }
	if req.IsActive != nil { p.IsActive = *req.IsActive }

	if err := h.store.Properties.Update(ctx, p); err != nil {
		respondError(w, http.StatusConflict, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, p)
}

// ==================== ROOM TYPES ====================

func (h *Handler) listRoomTypes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	propertyID := chi.URLParam(r, "propertyId")

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 { limit = 50 }
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 { page = 1 }
	offset := (page - 1) * limit

	rts, err := h.store.RoomTypes.ListByProperty(ctx, tenant, propertyID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"room_types": rts,
		"page":       page,
		"limit":      limit,
	})
}

func (h *Handler) createRoomType(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	propertyID := chi.URLParam(r, "propertyId")

	var req struct {
		Code          string   `json:"code"`
		Name          string   `json:"name"`
		Description   string   `json:"description"`
		BaseOccupancy int      `json:"base_occupancy"`
		MaxOccupancy  int      `json:"max_occupancy"`
		Amenities     []string `json:"amenities"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Code == "" || req.Name == "" {
		respondError(w, http.StatusBadRequest, "code, name required")
		return
	}
	if req.BaseOccupancy <= 0 { req.BaseOccupancy = 2 }
	if req.MaxOccupancy <= 0 { req.MaxOccupancy = 4 }

	rt := domain.NewRoomType(tenant, propertyID, req.Code, req.Name, req.BaseOccupancy, req.MaxOccupancy)
	rt.Description = req.Description
	rt.Amenities = req.Amenities

	if err := h.store.RoomTypes.Create(ctx, rt); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, rt)
}

func (h *Handler) getRoomType(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	rtID := chi.URLParam(r, "roomTypeId")

	rt, err := h.store.RoomTypes.GetByID(ctx, tenant, rtID)
	if err != nil {
		respondError(w, http.StatusNotFound, "room type not found")
		return
	}
	respondJSON(w, http.StatusOK, rt)
}

// ==================== ROOMS ====================

func (h *Handler) listRooms(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	propertyID := chi.URLParam(r, "propertyId")

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 { limit = 50 }
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 { page = 1 }
	offset := (page - 1) * limit

	rooms, err := h.store.Rooms.ListByProperty(ctx, tenant, propertyID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"rooms": rooms,
		"page":  page,
		"limit": limit,
	})
}

func (h *Handler) createRoom(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	propertyID := chi.URLParam(r, "propertyId")

	var req struct {
		RoomTypeID        string   `json:"room_type_id"`
		RoomNumber        string   `json:"room_number"`
		Floor             string   `json:"floor"`
		IsSmoking         bool     `json:"is_smoking"`
		HasAC             bool     `json:"has_ac"`
		Attributes        []string `json:"attributes"`
		SmartLockDeviceID string   `json:"smart_lock_device_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.RoomTypeID == "" || req.RoomNumber == "" {
		respondError(w, http.StatusBadRequest, "room_type_id, room_number required")
		return
	}

	room := domain.NewRoom(tenant, propertyID, req.RoomTypeID, req.RoomNumber)
	room.Floor = req.Floor
	room.IsSmoking = req.IsSmoking
	room.HasAC = req.HasAC
	room.Attributes = req.Attributes
	room.SmartLockDeviceID = req.SmartLockDeviceID

	if err := h.store.Rooms.Create(ctx, room); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, room)
}

func (h *Handler) getRoom(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	roomID := chi.URLParam(r, "roomId")

	room, err := h.store.Rooms.GetByID(ctx, tenant, roomID)
	if err != nil {
		respondError(w, http.StatusNotFound, "room not found")
		return
	}
	respondJSON(w, http.StatusOK, room)
}

func (h *Handler) updateRoom(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	roomID := chi.URLParam(r, "roomId")

	room, err := h.store.Rooms.GetByID(ctx, tenant, roomID)
	if err != nil {
		respondError(w, http.StatusNotFound, "room not found")
		return
	}

	var req struct {
		RoomTypeID        string   `json:"room_type_id"`
		RoomNumber        string   `json:"room_number"`
		Floor             string   `json:"floor"`
		IsSmoking         *bool    `json:"is_smoking"`
		HasAC             *bool    `json:"has_ac"`
		Attributes        []string `json:"attributes"`
		SmartLockDeviceID string   `json:"smart_lock_device_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RoomTypeID != "" { room.RoomTypeID = req.RoomTypeID }
	if req.RoomNumber != "" { room.RoomNumber = req.RoomNumber }
	if req.Floor != "" { room.Floor = req.Floor }
	if req.IsSmoking != nil { room.IsSmoking = *req.IsSmoking }
	if req.HasAC != nil { room.HasAC = *req.HasAC }
	if req.Attributes != nil { room.Attributes = req.Attributes }
	if req.SmartLockDeviceID != "" { room.SmartLockDeviceID = req.SmartLockDeviceID }

	if err := h.store.Rooms.Update(ctx, room); err != nil {
		respondError(w, http.StatusConflict, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, room)
}

func (h *Handler) updateRoomStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	roomID := chi.URLParam(r, "roomId")

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Status == "" {
		respondError(w, http.StatusBadRequest, "status required")
		return
	}

	if err := h.store.Rooms.UpdateStatus(ctx, tenant, roomID, domain.RoomStatus(req.Status)); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"room_id": roomID, "status": req.Status})
}

func (h *Handler) updateHousekeeping(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	roomID := chi.URLParam(r, "roomId")

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Status == "" {
		respondError(w, http.StatusBadRequest, "status required")
		return
	}

	if err := h.store.Rooms.UpdateHousekeeping(ctx, tenant, roomID, domain.HousekeepingStatus(req.Status)); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"room_id": roomID, "housekeeping_status": req.Status})
}
