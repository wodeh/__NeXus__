// Package api provides channel manager REST handlers.
package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/nexus-platform/pms-integration/internal/domain"
)

func (h *Handler) listChannelIntegrations(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	integrations, err := h.store.ChannelIntegrations.ListByTenant(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, integrations)
}

func (h *Handler) createChannelIntegration(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var req struct {
		PropertyID    string  `json:"property_id"`
		ChannelType   string  `json:"channel_type"`
		Name          string  `json:"name"`
		APIKey        string  `json:"api_key"`
		APISecret     string  `json:"api_secret"`
		HotelID       string  `json:"hotel_id"`
		CommissionPct float64 `json:"commission_pct"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ct, err := domain.ValidateChannelType(req.ChannelType)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	ci := domain.NewChannelIntegration(tenantID, req.PropertyID, ct, req.Name)
	ci.APIKey = req.APIKey
	ci.APISecret = req.APISecret
	ci.HotelID = req.HotelID
	if req.CommissionPct > 0 {
		ci.CommissionPct = req.CommissionPct
	}

	if err := h.store.ChannelIntegrations.Create(r.Context(), ci); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, ci)
}

func (h *Handler) getChannelIntegration(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "channelId")

	ci, err := h.store.ChannelIntegrations.GetByID(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "channel integration not found")
		return
	}

	// Don't return sensitive credentials in GET
	ci.APIKey = ""
	ci.APISecret = ""
	respondJSON(w, http.StatusOK, ci)
}

func (h *Handler) updateChannelIntegration(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "channelId")

	ci, err := h.store.ChannelIntegrations.GetByID(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "channel integration not found")
		return
	}

	var req struct {
		Name          string  `json:"name"`
		APIKey        string  `json:"api_key"`
		APISecret     string  `json:"api_secret"`
		HotelID       string  `json:"hotel_id"`
		CommissionPct float64 `json:"commission_pct"`
		IsActive      *bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name != "" { ci.Name = req.Name }
	if req.APIKey != "" { ci.APIKey = req.APIKey }
	if req.APISecret != "" { ci.APISecret = req.APISecret }
	if req.HotelID != "" { ci.HotelID = req.HotelID }
	if req.CommissionPct > 0 { ci.CommissionPct = req.CommissionPct }
	if req.IsActive != nil { ci.IsActive = *req.IsActive }
	ci.UpdatedAt = time.Now().UTC()

	if err := h.store.ChannelIntegrations.Update(r.Context(), ci); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, ci)
}

func (h *Handler) deleteChannelIntegration(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "channelId")

	if err := h.store.ChannelIntegrations.Delete(r.Context(), tenantID, id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *Handler) syncChannelIntegration(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "channelId")

	ci, err := h.store.ChannelIntegrations.GetByID(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "channel integration not found")
		return
	}

	// In production, trigger async sync job
	ci.Status = domain.ChannelStatusSyncing
	ci.LastSyncAt = func() *time.Time { t := time.Now().UTC(); return &t }()
	_ = h.store.ChannelIntegrations.Update(r.Context(), ci)

	respondJSON(w, http.StatusOK, map[string]string{
		"status": "syncing",
		"channel_id": id,
		"channel_type": string(ci.ChannelType),
	})
}

func (h *Handler) receiveBookingComReservation(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")

	var payload struct {
		ReservationID string `json:"reservation_id"`
		GuestName     string `json:"guest_name"`
		GuestEmail    string `json:"guest_email"`
		GuestPhone    string `json:"guest_phone"`
		RoomTypeID    string `json:"room_type_id"`
		CheckIn       string `json:"check_in"`
		CheckOut      string `json:"check_out"`
		Adults        int    `json:"adults"`
		Children      int    `json:"children"`
		TotalAmount   float64 `json:"total_amount"`
		Currency      string `json:"currency"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	// Find channel integration
	integrations, err := h.store.ChannelIntegrations.ListByTenant(r.Context(), tenantID)
	if err != nil || len(integrations) == 0 {
		respondError(w, http.StatusNotFound, "no channel integration found")
		return
	}

	ci := integrations[0] // simplified

	cr := domain.NewChannelReservation(tenantID, string(ci.ID), payload.ReservationID)
	cr.GuestName = payload.GuestName
	cr.GuestEmail = payload.GuestEmail
	cr.GuestPhone = payload.GuestPhone
	cr.RoomTypeID = payload.RoomTypeID
	cr.CheckInDate, _ = time.Parse("2006-01-02", payload.CheckIn)
	cr.CheckOutDate, _ = time.Parse("2006-01-02", payload.CheckOut)
	cr.Adults = payload.Adults
	cr.Children = payload.Children
	cr.TotalAmount = payload.TotalAmount
	cr.CurrencyCode = payload.Currency
	cr.CommissionAmount = ci.CalculateCommission(payload.TotalAmount)

	raw, _ := json.Marshal(payload)
	cr.RawData = string(raw)

	if err := h.store.ChannelReservations.Create(r.Context(), cr); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, cr)
}

func (h *Handler) listPendingChannelReservations(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	reservations, err := h.store.ChannelReservations.ListPending(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, reservations)
}

func (h *Handler) acceptChannelReservation(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "channelReservationId")

	cr, err := h.store.ChannelReservations.GetByOTAConf(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "reservation not found")
		return
	}

	// Create local reservation
	guest, _ := domain.NewGuest(tenantID, cr.GuestName, cr.GuestEmail, cr.GuestPhone, "", "", "", "", "", "", "")
	_ = h.store.Guests.Create(r.Context(), guest)

	// Find available room for room type
	var roomID string
	row := h.store.pool.QueryRow(r.Context(), `
		SELECT id FROM rooms WHERE tenant_id = $1 AND room_type_id = $2 AND status = 'available' LIMIT 1
	`, tenantID, cr.RoomTypeID)
	_ = row.Scan(&roomID)

	if roomID == "" {
		respondError(w, http.StatusConflict, "no available rooms for this room type")
		return
	}

	res, _ := domain.NewReservation(tenantID, string(guest.ID), roomID, cr.CheckInDate, cr.CheckOutDate, cr.Adults, cr.Children)
	// Get channel integration for source
	ci, _ := h.store.ChannelIntegrations.GetByID(r.Context(), tenantID, cr.ChannelIntegrationID)
	if ci != nil {
		res.Source = string(ci.ChannelType)
	} else {
		res.Source = "ota"
	}
	_ = h.store.Reservations.Create(r.Context(), res)

	// Update channel reservation
	resID := string(res.ID)
	_ = h.store.ChannelReservations.UpdateStatus(r.Context(), tenantID, string(cr.ID), "accepted", &resID)

	respondJSON(w, http.StatusOK, map[string]string{
		"status": "accepted",
		"reservation_id": resID,
	})
}
