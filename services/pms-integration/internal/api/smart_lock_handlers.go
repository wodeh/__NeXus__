package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// ==================== SMART LOCK HANDLERS ====================

func (h *Handler) listSmartLocks(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	propertyID := r.URL.Query().Get("property_id")
	if propertyID == "" {
		respondError(w, http.StatusBadRequest, "property_id required")
		return
	}
	locks, err := h.store.SmartLocks.ListLocks(r.Context(), tenantID, propertyID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, locks)
}

func (h *Handler) createSmartLock(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var lock domain.SmartLock
	if err := json.NewDecoder(r.Body).Decode(&lock); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	lock.TenantID = tenantID
	lock.Status = "online"
	if err := h.store.SmartLocks.CreateLock(r.Context(), &lock); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, lock)
}

func (h *Handler) getSmartLock(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "lockId")
	lock, err := h.store.SmartLocks.GetLock(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, lock)
}

func (h *Handler) updateSmartLock(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "lockId")
	var lock domain.SmartLock
	if err := json.NewDecoder(r.Body).Decode(&lock); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	lock.ID = id
	lock.TenantID = tenantID
	if err := h.store.SmartLocks.UpdateLock(r.Context(), &lock); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, lock)
}

func (h *Handler) deleteSmartLock(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "lockId")
	if err := h.store.SmartLocks.DeleteLock(r.Context(), id, tenantID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ==================== REMOTE LOCK/UNLOCK ====================

func (h *Handler) remoteUnlock(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "lockId")

	lock, err := h.store.SmartLocks.GetLock(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	if !lock.RemoteUnlockEnabled {
		respondError(w, http.StatusForbidden, "remote unlock disabled for this lock")
		return
	}

	// In production, send MQTT/HTTP command to OrbitaTech gateway.
	// For demo, simulate success and log event.
	now := time.Now()
	lock.LastUnlockAt = &now
	lock.LastCommunicationAt = &now
	_ = h.store.SmartLocks.UpdateLock(r.Context(), lock)

	event := domain.LockEvent{
		TenantID:    tenantID,
		PropertyID:  lock.PropertyID,
		SmartLockID: id,
		EventType:   "unlock",
		EventSource: "remote",
		Details:     map[string]interface{}{"method": "api", "user_agent": r.UserAgent()},
		OccurredAt:  now,
	}
	_ = h.store.SmartLocks.CreateEvent(r.Context(), &event)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":       "unlocked",
		"lock_id":      id,
		"timestamp":    now,
		"battery_level": lock.BatteryLevel,
	})
}

func (h *Handler) remoteLock(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "lockId")

	lock, err := h.store.SmartLocks.GetLock(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	now := time.Now()
	lock.LastLockAt = &now
	lock.LastCommunicationAt = &now
	_ = h.store.SmartLocks.UpdateLock(r.Context(), lock)

	event := domain.LockEvent{
		TenantID:    tenantID,
		PropertyID:  lock.PropertyID,
		SmartLockID: id,
		EventType:   "lock",
		EventSource: "remote",
		Details:     map[string]interface{}{"method": "api", "user_agent": r.UserAgent()},
		OccurredAt:  now,
	}
	_ = h.store.SmartLocks.CreateEvent(r.Context(), &event)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":       "locked",
		"lock_id":      id,
		"timestamp":    now,
		"battery_level": lock.BatteryLevel,
	})
}

// ==================== ACCESS CODE HANDLERS ====================

func (h *Handler) listAccessCodes(w http.ResponseWriter, r *http.Request) {
	lockID := chi.URLParam(r, "lockId")
	codes, err := h.store.SmartLocks.ListAccessCodes(r.Context(), lockID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, codes)
}

func (h *Handler) createAccessCode(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	lockID := chi.URLParam(r, "lockId")
	var code domain.LockAccessCode
	if err := json.NewDecoder(r.Body).Decode(&code); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	code.TenantID = tenantID
	code.SmartLockID = lockID
	code.IsActive = true
	if code.Code == "" {
		// Generate 4-digit code
		code.Code = fmt.Sprintf("%04d", time.Now().Unix()%10000)
	}
	if err := h.store.SmartLocks.CreateAccessCode(r.Context(), &code); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, code)
}

func (h *Handler) revokeAccessCode(w http.ResponseWriter, r *http.Request) {
	codeID := chi.URLParam(r, "codeId")
	var body struct {
		Reason string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	if err := h.store.SmartLocks.RevokeAccessCode(r.Context(), codeID, body.Reason); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}

// ==================== LOCK EVENT HANDLERS ====================

func (h *Handler) listLockEvents(w http.ResponseWriter, r *http.Request) {
	lockID := chi.URLParam(r, "lockId")
	limit := 50
	events, err := h.store.SmartLocks.ListEvents(r.Context(), lockID, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, events)
}

// ==================== LOCK OVERVIEW ====================

func (h *Handler) getLockOverview(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	propertyID := r.URL.Query().Get("property_id")
	if propertyID == "" {
		respondError(w, http.StatusBadRequest, "property_id required")
		return
	}
	overview, err := h.store.SmartLocks.GetOverview(r.Context(), tenantID, propertyID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, overview)
}

// ==================== LOCK BY ROOM ====================

func (h *Handler) getLockByRoom(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	roomID := chi.URLParam(r, "roomId")
	lock, err := h.store.SmartLocks.GetLockByRoom(r.Context(), tenantID, roomID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if lock == nil {
		respondError(w, http.StatusNotFound, "no lock assigned to this room")
		return
	}
	respondJSON(w, http.StatusOK, lock)
}
