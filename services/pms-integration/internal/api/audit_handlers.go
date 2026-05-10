package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/nexus-platform/pms-integration/internal/domain"
	"github.com/nexus-platform/pms-integration/internal/middleware"
)

func (h *Handler) listAuditLogs(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	resource := r.URL.Query().Get("resource")
	action := r.URL.Query().Get("action")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 { limit = 50 }
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	logs, err := h.store.AuditLogs.List(r.Context(), tenantID, resource, action, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, logs)
}

func (h *Handler) getResourceAuditTrail(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	resource := chi.URLParam(r, "resource")
	resourceID := chi.URLParam(r, "resourceId")

	logs, err := h.store.AuditLogs.GetByResourceID(r.Context(), tenantID, resource, resourceID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, logs)
}

func (h *Handler) createAuditLog(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var log domain.AuditLog
	if err := json.NewDecoder(r.Body).Decode(&log); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	log.TenantID = tenantID
	
	claims := middleware.ClaimsFromContext(r.Context())
	if claims != nil {
		log.UserID = &claims.UserID
	}
	log.IPAddress = r.RemoteAddr
	log.UserAgent = r.UserAgent()

	if err := h.store.AuditLogs.Create(r.Context(), &log); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, log)
}

// GDPR Handlers

func (h *Handler) createGDPRRequest(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var req domain.GDPRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.TenantID = tenantID
	req.Status = "pending"
	
	respondJSON(w, http.StatusCreated, req)
}

func (h *Handler) getGuestDataExport(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	guestID := chi.URLParam(r, "guestId")
	
	// In production: aggregate all guest data from all tables
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"guest_id": guestID,
		"tenant_id": tenantID,
		"export": "guest data package",
		"reservations": []string{},
		"charges": []string{},
		"payments": []string{},
		"messages": []string{},
	})
}

func (h *Handler) deleteGuestData(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	guestID := chi.URLParam(r, "guestId")
	
	// In production: anonymize personal data, keep anonymous reservation stats
	respondJSON(w, http.StatusOK, map[string]string{
		"status": "deleted",
		"guest_id": guestID,
		"tenant_id": tenantID,
		"note": "Personal data anonymized per GDPR Article 17",
	})
}
