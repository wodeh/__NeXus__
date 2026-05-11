package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// ============== WHATSAPP BOT HANDLERS ==============

func (h *Handler) getWhatsAppConfig(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	cfg, err := h.store.WhatsApp.GetConfig(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, cfg)
}

func (h *Handler) saveWhatsAppConfig(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var cfg domain.WhatsAppBotConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.store.WhatsApp.SaveConfig(r.Context(), tenantID, &cfg); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (h *Handler) listWhatsAppConversations(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	convs, err := h.store.WhatsApp.ListConversations(r.Context(), tenantID, 50)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, convs)
}

func (h *Handler) getWhatsAppConversation(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	phone := chi.URLParam(r, "phone")
	conv, err := h.store.WhatsApp.GetConversationByPhone(r.Context(), tenantID, phone)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, conv)
}

func (h *Handler) listWhatsAppMessages(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	convID := chi.URLParam(r, "convId")
	msgs, err := h.store.WhatsApp.ListMessages(r.Context(), tenantID, convID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, msgs)
}

// Webhook handler for incoming WhatsApp messages
func (h *Handler) whatsAppWebhook(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")

	var payload struct {
		Entry []struct {
			Changes []struct {
				Value struct {
					Messages []struct {
						From string `json:"from"`
						Body string `json:"body"`
						Type string `json:"type"`
					} `json:"messages"`
				} `json:"value"`
			} `json:"changes"`
		} `json:"entry"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid webhook payload")
		return
	}

	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			for _, msg := range change.Value.Messages {
				// Process incoming message
				conv, err := h.store.WhatsApp.GetConversationByPhone(r.Context(), tenantID, msg.From)
				if err != nil {
					// Create new conversation
					conv = &domain.WhatsAppConversation{
						TenantID:     tenantID,
						GuestPhone:   msg.From,
						CurrentState: "greeting",
					}
					conv, _ = h.store.WhatsApp.CreateConversation(r.Context(), tenantID, conv)
				}

				// Save inbound message
				h.store.WhatsApp.AddMessage(r.Context(), tenantID, &domain.WhatsAppMessage{
					ConversationID: conv.ID,
					Direction:      "inbound",
					Body:           msg.Body,
					MessageType:    "text",
					Status:         "delivered",
				})

				// Simple auto-reply logic (in production, use NLP)
				reply := "Thank you for your message! Our team will assist you shortly. To make a reservation, reply with 'BOOK'."
				if msg.Body == "BOOK" || msg.Body == "book" {
					reply = "Great! Please provide your check-in date (YYYY-MM-DD):"
					conv.CurrentState = "ask_dates"
					h.store.WhatsApp.UpdateConversation(r.Context(), tenantID, conv.ID, "ask_dates", "")
				}

				// Save outbound auto-reply
				h.store.WhatsApp.AddMessage(r.Context(), tenantID, &domain.WhatsAppMessage{
					ConversationID: conv.ID,
					Direction:      "outbound",
					Body:           reply,
					MessageType:    "text",
					Status:         "sent",
				})
			}
		}
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "processed"})
}
