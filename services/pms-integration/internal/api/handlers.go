package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// ==================== GUESTS ====================

func (h *Handler) listGuests(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 500 { limit = 50 }
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 { page = 1 }
	offset := (page - 1) * limit

	guests, err := h.store.Guests.ListByTenant(ctx, tenant, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"guests": guests,
		"page":   page,
		"limit":  limit,
	})
}

func (h *Handler) createGuest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)

	var req struct {
		FirstName    string            `json:"first_name"`
		LastName     string            `json:"last_name"`
		Email        string            `json:"email"`
		Phone        string            `json:"phone"`
		PropertyID   string            `json:"property_id"`
		Preferences  map[string]string `json:"preferences"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.FirstName == "" || req.LastName == "" || req.Email == "" {
		respondError(w, http.StatusBadRequest, "first_name, last_name, email required")
		return
	}

	g := domain.NewGuest(tenant, req.PropertyID, req.FirstName, req.LastName, req.Email)
	g.Phone = req.Phone
	g.Preferences = req.Preferences

	if err := h.store.Guests.Create(ctx, g); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, g)
}

func (h *Handler) getGuest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	guestID := chi.URLParam(r, "guestId")

	g, err := h.store.Guests.GetByID(ctx, tenant, guestID)
	if err != nil {
		respondError(w, http.StatusNotFound, "guest not found")
		return
	}
	respondJSON(w, http.StatusOK, g)
}

func (h *Handler) updateGuest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	guestID := chi.URLParam(r, "guestId")

	g, err := h.store.Guests.GetByID(ctx, tenant, guestID)
	if err != nil {
		respondError(w, http.StatusNotFound, "guest not found")
		return
	}

	var req struct {
		FirstName   string            `json:"first_name"`
		LastName    string            `json:"last_name"`
		Email       string            `json:"email"`
		Phone       string            `json:"phone"`
		Preferences map[string]string `json:"preferences"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.FirstName != "" { g.FirstName = req.FirstName }
	if req.LastName != "" { g.LastName = req.LastName }
	if req.Email != "" { g.Email = req.Email }
	if req.Phone != "" { g.Phone = req.Phone }
	if req.Preferences != nil { g.Preferences = req.Preferences }

	if err := h.store.Guests.Update(ctx, g); err != nil {
		respondError(w, http.StatusConflict, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, g)
}

// ==================== RESERVATIONS ====================

func (h *Handler) createReservation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)

	var req struct {
		GuestID         string `json:"guest_id"`
		PropertyID      string `json:"property_id"`
		RoomTypeID      string `json:"room_type_id"`
		RoomID          string `json:"room_id"`
		CheckInDate     string `json:"check_in_date"`
		CheckOutDate    string `json:"check_out_date"`
		Adults          int    `json:"adults"`
		Children        int    `json:"children"`
		SpecialRequests string `json:"special_requests"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.GuestID == "" || req.CheckInDate == "" || req.CheckOutDate == "" {
		respondError(w, http.StatusBadRequest, "guest_id, check_in_date, check_out_date required")
		return
	}

	checkIn, err := parseDate(req.CheckInDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid check_in_date")
		return
	}
	checkOut, err := parseDate(req.CheckOutDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid check_out_date")
		return
	}
	if !checkOut.After(checkIn) {
		respondError(w, http.StatusBadRequest, "check_out must be after check_in")
		return
	}

	res := domain.NewReservation(tenant, req.PropertyID, req.GuestID, req.RoomID, checkIn, checkOut)
	if req.SpecialRequests != "" {
		res.SpecialRequests = []string{req.SpecialRequests}
	}

	if err := res.Confirm(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.store.Reservations.Create(ctx, res); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, res)
}

func (h *Handler) getReservation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	resID := chi.URLParam(r, "reservationId")

	res, err := h.store.Reservations.GetByID(ctx, tenant, resID)
	if err != nil {
		respondError(w, http.StatusNotFound, "reservation not found")
		return
	}
	respondJSON(w, http.StatusOK, res)
}

func (h *Handler) checkIn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	resID := chi.URLParam(r, "reservationId")

	res, err := h.store.Reservations.GetByID(ctx, tenant, resID)
	if err != nil {
		respondError(w, http.StatusNotFound, "reservation not found")
		return
	}

	var req struct {
		RoomID string `json:"room_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.RoomID != "" {
		res.RoomID = req.RoomID
		room, err := h.store.Rooms.GetByID(ctx, tenant, req.RoomID)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid room_id")
			return
		}
		if err := room.Assign(); err != nil {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		if err := h.store.Rooms.Update(ctx, room); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	if err := res.CheckIn(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.store.Reservations.Update(ctx, res); err != nil {
		respondError(w, http.StatusConflict, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"reservation_id": res.ID,
		"status":         res.Status,
		"room_id":        res.RoomID,
	})
}

func (h *Handler) checkOut(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	resID := chi.URLParam(r, "reservationId")

	res, err := h.store.Reservations.GetByID(ctx, tenant, resID)
	if err != nil {
		respondError(w, http.StatusNotFound, "reservation not found")
		return
	}

	if err := res.CheckOut(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.store.Reservations.Update(ctx, res); err != nil {
		respondError(w, http.StatusConflict, err.Error())
		return
	}

	if res.RoomID != "" {
		room, err := h.store.Rooms.GetByID(ctx, tenant, res.RoomID)
		if err == nil {
			room.Vacate()
			h.store.Rooms.Update(ctx, room)
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"reservation_id": res.ID,
		"status":         res.Status,
	})
}

func (h *Handler) cancelReservation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	resID := chi.URLParam(r, "reservationId")

	res, err := h.store.Reservations.GetByID(ctx, tenant, resID)
	if err != nil {
		respondError(w, http.StatusNotFound, "reservation not found")
		return
	}

	if err := res.Cancel(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.store.Reservations.Update(ctx, res); err != nil {
		respondError(w, http.StatusConflict, err.Error())
		return
	}

	if res.RoomID != "" {
		room, err := h.store.Rooms.GetByID(ctx, tenant, res.RoomID)
		if err == nil {
			room.Status = domain.RoomStatusAvailable
			h.store.Rooms.Update(ctx, room)
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"reservation_id": res.ID,
		"status":         res.Status,
	})
}

// ==================== AVAILABILITY ====================

func (h *Handler) checkAvailability(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)

	propertyID := r.URL.Query().Get("propertyId")
	checkInStr := r.URL.Query().Get("checkIn")
	checkOutStr := r.URL.Query().Get("checkOut")

	if propertyID == "" || checkInStr == "" || checkOutStr == "" {
		respondError(w, http.StatusBadRequest, "propertyId, checkIn, checkOut required")
		return
	}

	checkIn, err := parseDate(checkInStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid checkIn")
		return
	}
	checkOut, err := parseDate(checkOutStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid checkOut")
		return
	}

	rooms, err := h.store.Rooms.ListAvailable(ctx, tenant, propertyID, checkIn, checkOut, 100, 0)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Group by room type for the response
	grouped := make(map[string]*struct {
		TypeID        string `json:"room_type_id"`
		TypeName      string `json:"room_type_name"`
		Available     int    `json:"available_rooms"`
		Total         int    `json:"total_rooms"`
	})

	for _, room := range rooms {
		if _, ok := grouped[room.RoomTypeID]; !ok {
			grouped[room.RoomTypeID] = &struct {
				TypeID    string `json:"room_type_id"`
				TypeName  string `json:"room_type_name"`
				Available int    `json:"available_rooms"`
				Total     int    `json:"total_rooms"`
			}{
				TypeID: room.RoomTypeID,
			}
		}
		grouped[room.RoomTypeID].Available++
	}

	var results []interface{}
	for _, v := range grouped {
		results = append(results, v)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"property_id":    propertyID,
		"check_in_date":  checkInStr,
		"check_out_date": checkOutStr,
		"results":        results,
	})
}

func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}
