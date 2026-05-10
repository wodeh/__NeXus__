package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// ==================== IPTV CHANNEL HANDLERS ====================

func (h *Handler) listIPTVChannels(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	propertyID := r.URL.Query().Get("property_id")
	if propertyID == "" {
		respondError(w, http.StatusBadRequest, "property_id required")
		return
	}
	channels, err := h.store.IPTV.ListChannels(r.Context(), tenantID, propertyID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, channels)
}

func (h *Handler) createIPTVChannel(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var ch domain.IPTVChannel
	if err := json.NewDecoder(r.Body).Decode(&ch); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	ch.TenantID = tenantID
	if err := h.store.IPTV.CreateChannel(r.Context(), &ch); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, ch)
}

func (h *Handler) getIPTVChannel(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "channelId")
	ch, err := h.store.IPTV.GetChannel(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, ch)
}

func (h *Handler) updateIPTVChannel(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "channelId")
	var ch domain.IPTVChannel
	if err := json.NewDecoder(r.Body).Decode(&ch); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	ch.ID = id
	ch.TenantID = tenantID
	if err := h.store.IPTV.UpdateChannel(r.Context(), &ch); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, ch)
}

func (h *Handler) deleteIPTVChannel(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "channelId")
	if err := h.store.IPTV.DeleteChannel(r.Context(), id, tenantID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ==================== IPTV CONTENT HANDLERS ====================

func (h *Handler) listIPTVContent(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	propertyID := r.URL.Query().Get("property_id")
	if propertyID == "" {
		respondError(w, http.StatusBadRequest, "property_id required")
		return
	}
	items, err := h.store.IPTV.ListContent(r.Context(), tenantID, propertyID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *Handler) createIPTVContent(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var c domain.IPTVContent
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	c.TenantID = tenantID
	if err := h.store.IPTV.CreateContent(r.Context(), &c); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, c)
}

func (h *Handler) getIPTVContent(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "contentId")
	c, err := h.store.IPTV.GetContent(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, c)
}

func (h *Handler) updateIPTVContent(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "contentId")
	var c domain.IPTVContent
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	c.ID = id
	c.TenantID = tenantID
	if err := h.store.IPTV.UpdateContent(r.Context(), &c); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, c)
}

func (h *Handler) deleteIPTVContent(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "contentId")
	if err := h.store.IPTV.DeleteContent(r.Context(), id, tenantID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ==================== IPTV ROOM BINDING HANDLERS ====================

func (h *Handler) listIPTVRoomBindings(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	propertyID := r.URL.Query().Get("property_id")
	if propertyID == "" {
		respondError(w, http.StatusBadRequest, "property_id required")
		return
	}
	bindings, err := h.store.IPTV.ListRoomBindings(r.Context(), tenantID, propertyID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, bindings)
}

func (h *Handler) createIPTVRoomBinding(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var b domain.IPTVRoomBinding
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	b.TenantID = tenantID
	if err := h.store.IPTV.CreateRoomBinding(r.Context(), &b); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, b)
}

func (h *Handler) getIPTVRoomBinding(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "bindingId")
	b, err := h.store.IPTV.GetRoomBinding(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, b)
}

func (h *Handler) updateIPTVRoomBinding(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "bindingId")
	var b domain.IPTVRoomBinding
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	b.ID = id
	b.TenantID = tenantID
	if err := h.store.IPTV.UpdateRoomBinding(r.Context(), &b); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, b)
}

func (h *Handler) deleteIPTVRoomBinding(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "bindingId")
	if err := h.store.IPTV.DeleteRoomBinding(r.Context(), id, tenantID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ==================== IPTV WELCOME SCREEN ====================

func (h *Handler) getIPTVWelcomeScreen(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	roomID := chi.URLParam(r, "roomId")

	binding, err := h.store.IPTV.GetRoomBindingByRoom(r.Context(), tenantID, roomID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if binding == nil {
		respondError(w, http.StatusNotFound, "no IPTV binding for this room")
		return
	}

	// In production, fetch guest/reservation data. For demo, return mock welcome screen.
	screen := domain.IPTVWelcomeScreen{
		RoomNumber:     roomID,
		WelcomeMessage: binding.WelcomeMessage,
		HotelName:      "Grand Plaza Hotel",
		Language:       binding.LanguageOverride,
		LocalTime:      time.Now().Format("15:04"),
		Weather: &domain.IPTVWeatherInfo{
			Temp:      24,
			Condition: "Sunny",
			Icon:      "sun",
			High:      28,
			Low:       19,
		},
	}
	respondJSON(w, http.StatusOK, screen)
}

// ==================== IPTV ANALYTICS ====================

func (h *Handler) getIPTVAnalytics(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	propertyID := r.URL.Query().Get("property_id")
	if propertyID == "" {
		respondError(w, http.StatusBadRequest, "property_id required")
		return
	}
	from := time.Now().AddDate(0, 0, -7)
	to := time.Now()
	analytics, err := h.store.IPTV.GetAnalytics(r.Context(), tenantID, propertyID, from, to)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, analytics)
}
