package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// ============== COMMUNICATION TEMPLATES ==============

func (h *Handler) listTemplates(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var category *string
	if c := r.URL.Query().Get("category"); c != "" {
		category = &c
	}
	tmpls, err := h.store.Communications.ListTemplates(r.Context(), tenantID, category)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, tmpls)
}

func (h *Handler) createTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var t domain.CommunicationTemplate
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	res, err := h.store.Communications.CreateTemplate(r.Context(), tenantID, &t)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, res)
}

func (h *Handler) updateTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "templateId")
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.store.Communications.UpdateTemplate(r.Context(), tenantID, id, updates); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *Handler) deleteTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "templateId")
	if err := h.store.Communications.DeleteTemplate(r.Context(), tenantID, id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ============== COMMUNICATION SEQUENCES ==============

func (h *Handler) listSequences(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	seqs, err := h.store.Communications.ListSequences(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, seqs)
}

func (h *Handler) createSequence(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var s domain.CommunicationSequence
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	res, err := h.store.Communications.CreateSequence(r.Context(), tenantID, &s)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, res)
}

func (h *Handler) addSequenceStep(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	seqID := chi.URLParam(r, "seqId")
	var step domain.CommunicationSequenceStep
	if err := json.NewDecoder(r.Body).Decode(&step); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	step.SequenceID = seqID
	res, err := h.store.Communications.AddSequenceStep(r.Context(), tenantID, &step)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, res)
}

func (h *Handler) getSequenceSteps(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	seqID := chi.URLParam(r, "seqId")
	steps, err := h.store.Communications.GetSequenceSteps(r.Context(), tenantID, seqID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, steps)
}

// ============== SCHEDULED COMMUNICATIONS ==============

func (h *Handler) listScheduledCommunications(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var status *string
	if s := r.URL.Query().Get("status"); s != "" {
		status = &s
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 50
	}
	comms, err := h.store.Communications.ListScheduledCommunications(r.Context(), tenantID, status, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, comms)
}

func (h *Handler) scheduleCommunication(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var c domain.ScheduledCommunication
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	c.Status = "scheduled"
	if c.ScheduledAt.IsZero() {
		c.ScheduledAt = time.Now().Add(1 * time.Hour)
	}
	res, err := h.store.Communications.ScheduleCommunication(r.Context(), tenantID, &c)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, res)
}

func (h *Handler) cancelScheduledCommunication(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "commId")
	if err := h.store.Communications.UpdateScheduledStatus(r.Context(), tenantID, id, "cancelled", nil); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

// ============== COMMUNICATION LOGS ==============

func (h *Handler) listCommunicationLogs(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 50
	}
	logs, err := h.store.Communications.ListLogs(r.Context(), tenantID, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, logs)
}

// ============== SEND IMMEDIATE ==============

func (h *Handler) sendImmediate(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var body struct {
		GuestID       string               `json:"guest_id,omitempty"`
		GuestPhone    string               `json:"guest_phone,omitempty"`
		GuestEmail    string               `json:"guest_email,omitempty"`
		ReservationID string               `json:"reservation_id,omitempty"`
		Channel       domain.CommunicationChannel `json:"channel"`
		Subject       string               `json:"subject,omitempty"`
		Body          string               `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}

	// Log the communication
	log := &domain.CommunicationLog{
		TenantID:      tenantID,
		GuestID:       body.GuestID,
		ReservationID: body.ReservationID,
		Channel:       body.Channel,
		Direction:     "outbound",
		Subject:       body.Subject,
		Body:          body.Body,
		Status:        "sent",
		SentAt:        timePtr(time.Now()),
	}
	_, err := h.store.Communications.LogCommunication(r.Context(), tenantID, log)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "sent"})
}

func timePtr(t time.Time) *time.Time {
	return &t
}
