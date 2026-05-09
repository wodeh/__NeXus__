// Package api provides HTTP REST handlers for rate management.
package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// ==================== RATE PLANS ====================

type createRatePlanRequest struct {
	PropertyID         string  `json:"property_id"`
	RoomTypeID         string  `json:"room_type_id"`
	Code               string  `json:"code"`
	Name               string  `json:"name"`
	Description        string  `json:"description"`
	BaseRate           float64 `json:"base_rate"`
	CurrencyCode       string  `json:"currency_code"`
	MinLOS             int     `json:"min_los"`
	MaxLOS             int     `json:"max_los"`
	AdvanceBookingDays int     `json:"advance_booking_days"`
}

type ratePlanResponse struct {
	ID                 string                 `json:"id"`
	PropertyID         string                 `json:"property_id"`
	RoomTypeID         string                 `json:"room_type_id"`
	Code               string                 `json:"code"`
	Name               string                 `json:"name"`
	Description        string                 `json:"description"`
	BaseRate           float64                `json:"base_rate"`
	CurrencyCode       string                 `json:"currency_code"`
	MinLOS             int                    `json:"min_los"`
	MaxLOS             int                    `json:"max_los"`
	AdvanceBookingDays int                    `json:"advance_booking_days"`
	IsActive           bool                   `json:"is_active"`
	Version            int                    `json:"version"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

func toRatePlanResponse(rp *domain.RatePlan) ratePlanResponse {
	return ratePlanResponse{
		ID:                 string(rp.ID),
		PropertyID:         rp.PropertyID,
		RoomTypeID:         rp.RoomTypeID,
		Code:               rp.Code,
		Name:               rp.Name,
		Description:        rp.Description,
		BaseRate:           rp.BaseRate,
		CurrencyCode:       rp.CurrencyCode,
		MinLOS:             rp.MinLOS,
		MaxLOS:             rp.MaxLOS,
		AdvanceBookingDays: rp.AdvanceBookingDays,
		IsActive:           rp.IsActive,
		Version:            rp.Version,
		CreatedAt:          rp.CreatedAt,
		UpdatedAt:          rp.UpdatedAt,
	}
}

func (h *Handler) createRatePlan(w http.ResponseWriter, r *http.Request) {
	var req createRatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := tenantID(r)
	rp := domain.NewRatePlan(tenantID, req.PropertyID, req.RoomTypeID, req.Code, req.Name, req.BaseRate, req.CurrencyCode)
	rp.Description = req.Description
	if req.MinLOS > 0 {
		rp.MinLOS = req.MinLOS
	}
	rp.MaxLOS = req.MaxLOS
	rp.AdvanceBookingDays = req.AdvanceBookingDays

	if err := h.store.RatePlans.Create(r.Context(), rp); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, toRatePlanResponse(rp))
}

func (h *Handler) getRatePlan(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "ratePlanId")

	rp, err := h.store.RatePlans.GetByID(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "rate plan not found")
		return
	}

	respondJSON(w, http.StatusOK, toRatePlanResponse(rp))
}

func (h *Handler) listRatePlans(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	propertyID := r.URL.Query().Get("property_id")
	roomTypeID := r.URL.Query().Get("room_type_id")

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	plans, err := h.store.RatePlans.List(r.Context(), tenantID, propertyID, roomTypeID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]ratePlanResponse, len(plans))
	for i, rp := range plans {
		resp[i] = toRatePlanResponse(rp)
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"rate_plans": resp,
		"total":      len(resp),
	})
}

func (h *Handler) updateRatePlan(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "ratePlanId")

	rp, err := h.store.RatePlans.GetByID(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "rate plan not found")
		return
	}

	var req createRatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name != "" {
		rp.Name = req.Name
	}
	if req.BaseRate > 0 {
		rp.BaseRate = req.BaseRate
	}
	if req.MinLOS > 0 {
		rp.MinLOS = req.MinLOS
	}
	rp.MaxLOS = req.MaxLOS
	rp.AdvanceBookingDays = req.AdvanceBookingDays
	rp.bumpVersion()

	if err := h.store.RatePlans.Update(r.Context(), rp); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, toRatePlanResponse(rp))
}

// ==================== DAILY RATES ====================

type createDailyRateRequest struct {
	RatePlanID       string  `json:"rate_plan_id"`
	RateDate         string  `json:"rate_date"`
	Rate             float64 `json:"rate"`
	Availability     int     `json:"availability"`
	MinStay          int     `json:"min_stay"`
	CloseToArrival   bool    `json:"close_to_arrival"`
	CloseToDeparture bool    `json:"close_to_departure"`
	StopSell         bool    `json:"stop_sell"`
}

type dailyRateResponse struct {
	ID               string    `json:"id"`
	RatePlanID       string    `json:"rate_plan_id"`
	RateDate         time.Time `json:"rate_date"`
	Rate             float64   `json:"rate"`
	Availability     int       `json:"availability"`
	MinStay          int       `json:"min_stay"`
	CloseToArrival   bool      `json:"close_to_arrival"`
	CloseToDeparture bool      `json:"close_to_departure"`
	StopSell         bool      `json:"stop_sell"`
	CreatedAt        time.Time `json:"created_at"`
}

func toDailyRateResponse(dr *domain.DailyRate) dailyRateResponse {
	return dailyRateResponse{
		ID:               string(dr.ID),
		RatePlanID:       dr.RatePlanID,
		RateDate:         dr.RateDate,
		Rate:             dr.Rate,
		Availability:     dr.Availability,
		MinStay:          dr.MinStay,
		CloseToArrival:   dr.CloseToArrival,
		CloseToDeparture: dr.CloseToDeparture,
		StopSell:         dr.StopSell,
		CreatedAt:        dr.CreatedAt,
	}
}

func (h *Handler) createDailyRate(w http.ResponseWriter, r *http.Request) {
	var req createDailyRateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	date, err := time.Parse("2006-01-02", req.RateDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "rate_date must be YYYY-MM-DD")
		return
	}

	tenantID := tenantID(r)
	dr := domain.NewDailyRate(tenantID, req.RatePlanID, date, req.Rate, req.Availability)
	dr.MinStay = req.MinStay
	if dr.MinStay <= 0 {
		dr.MinStay = 1
	}
	dr.CloseToArrival = req.CloseToArrival
	dr.CloseToDeparture = req.CloseToDeparture
	dr.StopSell = req.StopSell

	if err := h.store.DailyRates.Create(r.Context(), dr); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, toDailyRateResponse(dr))
}

func (h *Handler) getDailyRate(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "dailyRateId")

	dr, err := h.store.DailyRates.GetByID(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "daily rate not found")
		return
	}

	respondJSON(w, http.StatusOK, toDailyRateResponse(dr))
}

func (h *Handler) listDailyRates(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	ratePlanID := r.URL.Query().Get("rate_plan_id")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	var start, end time.Time
	var err error
	if startDate != "" {
		start, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			respondError(w, http.StatusBadRequest, "start_date must be YYYY-MM-DD")
			return
		}
	}
	if endDate != "" {
		end, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			respondError(w, http.StatusBadRequest, "end_date must be YYYY-MM-DD")
			return
		}
	}

	rates, err := h.store.DailyRates.ListByPlanAndDateRange(r.Context(), tenantID, ratePlanID, start, end)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]dailyRateResponse, len(rates))
	for i, dr := range rates {
		resp[i] = toDailyRateResponse(dr)
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"daily_rates": resp,
		"total":       len(resp),
	})
}

func (h *Handler) updateDailyRate(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "dailyRateId")

	dr, err := h.store.DailyRates.GetByID(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "daily rate not found")
		return
	}

	var req createDailyRateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Rate > 0 {
		dr.Rate = req.Rate
	}
	if req.Availability >= 0 {
		dr.Availability = req.Availability
	}
	dr.MinStay = req.MinStay
	if dr.MinStay <= 0 {
		dr.MinStay = 1
	}
	dr.CloseToArrival = req.CloseToArrival
	dr.CloseToDeparture = req.CloseToDeparture
	dr.StopSell = req.StopSell
	dr.UpdatedAt = time.Now().UTC()

	if err := h.store.DailyRates.Update(r.Context(), dr); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, toDailyRateResponse(dr))
}
