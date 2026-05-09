// Package api provides notification REST handlers.
package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/nexus-platform/pms-integration/internal/domain"
	"github.com/nexus-platform/pms-integration/internal/middleware"
)

func (h *Handler) listNotificationLogs(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 { limit = 50 }
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	logs, err := h.store.NotificationLogs.ListByTenant(r.Context(), tenantID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, logs)
}

func (h *Handler) sendNotification(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var req struct {
		Channel   string `json:"channel"`
		Recipient string `json:"recipient"`
		Subject   string `json:"subject"`
		Body      string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var ch domain.NotificationChannel
	switch req.Channel {
	case "email":
		ch = domain.NotificationEmail
	case "sms":
		ch = domain.NotificationSMS
	default:
		respondError(w, http.StatusBadRequest, "invalid channel")
		return
	}

	log := domain.NewNotificationLog(tenantID, "", ch, req.Recipient, req.Subject, req.Body)
	if err := h.store.NotificationLogs.Create(r.Context(), log); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// In production, enqueue to Redis for async delivery via SendGrid/Twilio
	respondJSON(w, http.StatusAccepted, map[string]interface{}{
		"id":     log.ID,
		"status": "queued",
	})
}

func (h *Handler) listGuestMessages(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	guestID := chi.URLParam(r, "guestId")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 { limit = 50 }
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	msgs, err := h.store.GuestMessages.ListByGuest(r.Context(), tenantID, guestID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, msgs)
}

func (h *Handler) createGuestMessage(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	guestID := chi.URLParam(r, "guestId")

	var req struct {
		Content       string `json:"content"`
		Channel       string `json:"channel"`
		ReservationID string `json:"reservation_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var ch domain.NotificationChannel = domain.NotificationInApp
	if req.Channel == "sms" {
		ch = domain.NotificationSMS
	} else if req.Channel == "email" {
		ch = domain.NotificationEmail
	}

	msg := domain.NewGuestMessage(tenantID, guestID, domain.GuestMessageOutbound, ch, req.Content)
	if req.ReservationID != "" {
		msg.ReservationID = &req.ReservationID
	}
	claims := middleware.ClaimsFromContext(r.Context())
	if claims != nil {
		msg.SentBy = &claims.UserID
	}

	if err := h.store.GuestMessages.Create(r.Context(), msg); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, msg)
}

func (h *Handler) markMessageAsRead(w http.ResponseWriter, r *http.Request) {
	msgID := chi.URLParam(r, "messageId")
	if err := h.store.GuestMessages.MarkAsRead(r.Context(), msgID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "read"})
}

func (h *Handler) listUnreadMessages(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	msgs, err := h.store.GuestMessages.ListUnread(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, msgs)
}
