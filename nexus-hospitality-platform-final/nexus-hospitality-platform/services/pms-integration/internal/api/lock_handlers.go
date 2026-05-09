// Package api provides HTTP REST handlers for smart lock and digital key operations.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// ==================== SMART LOCK ====================

type createSmartLockRequest struct {
	PropertyID      string `json:"property_id"`
	RoomID          string `json:"room_id"`
	DeviceID        string `json:"device_id"`
	DeviceModel     string `json:"device_model"`
	FirmwareVersion string `json:"firmware_version"`
}

type smartLockResponse struct {
	ID              string    `json:"id"`
	PropertyID      string    `json:"property_id"`
	RoomID          string    `json:"room_id"`
	DeviceID        string    `json:"device_id"`
	DeviceModel     string    `json:"device_model"`
	FirmwareVersion string    `json:"firmware_version"`
	BatteryLevel    int       `json:"battery_level"`
	IsOnline        bool      `json:"is_online"`
	LastSeenAt      time.Time `json:"last_seen_at"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
}

func toSmartLockResponse(l *domain.SmartLock) smartLockResponse {
	return smartLockResponse{
		ID:              string(l.ID),
		PropertyID:      l.PropertyID,
		RoomID:          l.RoomID,
		DeviceID:        l.DeviceID,
		DeviceModel:     l.DeviceModel,
		FirmwareVersion: l.FirmwareVersion,
		BatteryLevel:    l.BatteryLevel,
		IsOnline:        l.IsOnline,
		LastSeenAt:      l.LastSeenAt,
		IsActive:        l.IsActive,
		CreatedAt:       l.CreatedAt,
	}
}

func (h *Handler) createSmartLock(w http.ResponseWriter, r *http.Request) {
	var req createSmartLockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := tenantID(r)
	lock := domain.NewSmartLock(tenantID, req.PropertyID, req.RoomID, req.DeviceID)
	lock.DeviceModel = req.DeviceModel
	lock.FirmwareVersion = req.FirmwareVersion

	if err := h.store.SmartLocks.Create(r.Context(), lock); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, toSmartLockResponse(lock))
}

func (h *Handler) getSmartLock(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "lockId")

	lock, err := h.store.SmartLocks.GetByID(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "smart lock not found")
		return
	}

	respondJSON(w, http.StatusOK, toSmartLockResponse(lock))
}

func (h *Handler) getSmartLockByRoom(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	roomID := chi.URLParam(r, "roomId")

	lock, err := h.store.SmartLocks.GetByRoom(r.Context(), tenantID, roomID)
	if err != nil {
		respondError(w, http.StatusNotFound, "no smart lock for room")
		return
	}

	respondJSON(w, http.StatusOK, toSmartLockResponse(lock))
}

func (h *Handler) listSmartLocks(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	propertyID := r.URL.Query().Get("property_id")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	locks, err := h.store.SmartLocks.ListByProperty(r.Context(), tenantID, propertyID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]smartLockResponse, len(locks))
	for i, l := range locks {
		resp[i] = toSmartLockResponse(l)
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"smart_locks": resp,
		"total":       len(resp),
	})
}

func (h *Handler) updateSmartLockStatus(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "lockId")

	var req struct {
		IsOnline     bool `json:"is_online"`
		BatteryLevel int  `json:"battery_level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.store.SmartLocks.UpdateStatus(r.Context(), tenantID, id, req.IsOnline, req.BatteryLevel); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// ==================== DIGITAL KEY ====================

type createDigitalKeyRequest struct {
	SmartLockID   string `json:"smart_lock_id"`
	ReservationID string `json:"reservation_id"`
	GuestID       string `json:"guest_id"`
	GuestEmail    string `json:"guest_email"`
	GuestPhone    string `json:"guest_phone"`
	ValidFrom     string `json:"valid_from"`
	ValidUntil    string `json:"valid_until"`
	MaxUses       int    `json:"max_uses"`
}

type digitalKeyResponse struct {
	ID            string    `json:"id"`
	SmartLockID   string    `json:"smart_lock_id"`
	ReservationID string    `json:"reservation_id"`
	GuestID       string    `json:"guest_id"`
	GuestEmail    string    `json:"guest_email"`
	Status        string    `json:"status"`
	KeyCode       string    `json:"key_code"`
	PinCode       string    `json:"pin_code,omitempty"`
	ValidFrom     time.Time `json:"valid_from"`
	ValidUntil    time.Time `json:"valid_until"`
	MaxUses       int       `json:"max_uses"`
	UsesRemaining int       `json:"uses_remaining"`
	IssuedAt      time.Time `json:"issued_at"`
	CreatedAt     time.Time `json:"created_at"`
}

func toDigitalKeyResponse(k *domain.DigitalKey) digitalKeyResponse {
	return digitalKeyResponse{
		ID:            string(k.ID),
		SmartLockID:   k.SmartLockID,
		ReservationID: k.ReservationID,
		GuestID:       k.GuestID,
		GuestEmail:    k.GuestEmail,
		Status:        string(k.Status),
		KeyCode:       k.KeyCode,
		PinCode:       k.PinCode,
		ValidFrom:     k.ValidFrom,
		ValidUntil:    k.ValidUntil,
		MaxUses:       k.MaxUses,
		UsesRemaining: k.UsesRemaining,
		IssuedAt:      k.IssuedAt,
		CreatedAt:     k.CreatedAt,
	}
}

func (h *Handler) createDigitalKey(w http.ResponseWriter, r *http.Request) {
	var req createDigitalKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := tenantID(r)
	validFrom, _ := time.Parse(time.RFC3339, req.ValidFrom)
	validUntil, _ := time.Parse(time.RFC3339, req.ValidUntil)

	key := domain.NewDigitalKey(tenantID, req.SmartLockID, req.ReservationID, req.GuestID, validFrom, validUntil)
	key.GuestEmail = req.GuestEmail
	key.GuestPhone = req.GuestPhone
	if req.MaxUses > 0 {
		key.MaxUses = req.MaxUses
		key.UsesRemaining = req.MaxUses
	}

	// Generate key code
	key.KeyCode = generateKeyCode()
	key.PinCode = generatePIN()

	if err := h.store.DigitalKeys.Create(r.Context(), key); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Activate key
	key.Activate()
	_ = h.store.DigitalKeys.Update(r.Context(), key)

	// TODO: Sync with Orbita API

	respondJSON(w, http.StatusCreated, toDigitalKeyResponse(key))
}

func (h *Handler) getDigitalKey(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "keyId")

	key, err := h.store.DigitalKeys.GetByID(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "digital key not found")
		return
	}

	respondJSON(w, http.StatusOK, toDigitalKeyResponse(key))
}

func (h *Handler) getDigitalKeyByReservation(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	reservationID := chi.URLParam(r, "reservationId")

	key, err := h.store.DigitalKeys.GetByReservation(r.Context(), tenantID, reservationID)
	if err != nil {
		respondError(w, http.StatusNotFound, "no active key for reservation")
		return
	}

	respondJSON(w, http.StatusOK, toDigitalKeyResponse(key))
}

func (h *Handler) listDigitalKeys(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	lockID := chi.URLParam(r, "lockId")
	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	keys, err := h.store.DigitalKeys.ListBySmartLock(r.Context(), tenantID, lockID, status, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]digitalKeyResponse, len(keys))
	for i, k := range keys {
		resp[i] = toDigitalKeyResponse(k)
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"digital_keys": resp,
		"total":        len(resp),
	})
}

func (h *Handler) revokeDigitalKey(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "keyId")

	var req struct {
		Reason string `json:"reason"`
		By     string `json:"by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.store.DigitalKeys.Revoke(r.Context(), tenantID, id, req.By, req.Reason); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// TODO: Revoke on Orbita API

	respondJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}

// ==================== LOCK EVENTS (WEBHOOK) ====================

type lockEventRequest struct {
	DeviceID     string                 `json:"device_id"`
	EventType    string                 `json:"event_type"`
	KeyCode      string                 `json:"key_code,omitempty"`
	PinCode      string                 `json:"pin_code,omitempty"`
	UserID       string                 `json:"user_id,omitempty"`
	Method       string                 `json:"method,omitempty"`
	Success      bool                   `json:"success"`
	BatteryLevel int                    `json:"battery_level,omitempty"`
	Timestamp    string                 `json:"timestamp"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

func (h *Handler) receiveLockEvent(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")

	var req lockEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Find lock by device ID
	lock, err := h.store.SmartLocks.GetByDeviceID(r.Context(), tenantID, req.DeviceID)
	if err != nil {
		respondError(w, http.StatusNotFound, "device not found")
		return
	}

	// Update lock status
	if req.EventType == "online" || req.EventType == "offline" {
		_ = h.store.SmartLocks.UpdateStatus(r.Context(), tenantID, string(lock.ID), req.EventType == "online", req.BatteryLevel)
	}

	// Create event record
	event := domain.NewLockEvent(tenantID, string(lock.ID), domain.LockEventType(req.EventType))
	event.KeyCode = req.KeyCode
	event.PinCode = req.PinCode
	event.UserID = req.UserID
	event.Method = req.Method
	event.Success = req.Success
	event.BatteryLevel = req.BatteryLevel
	event.Metadata = req.Metadata
	if req.Timestamp != "" {
		t, _ := time.Parse(time.RFC3339, req.Timestamp)
		event.OccurredAt = t
	}

	// TODO: Store event in audit_log or dedicated events table

	// If key used successfully, record usage
	if req.EventType == "key_used" && req.KeyCode != "" {
		// Find key by code and record use
		// This would need a GetByKeyCode method
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "received"})
}

// ==================== HELPERS ====================

func generateKeyCode() string {
	// 8-char alphanumeric code
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

func generatePIN() string {
	// 6-digit PIN
	return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
}
