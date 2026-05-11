package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// ============== BOOKING ENGINE HANDLERS ==============

func (h *Handler) listPromoCodes(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	codes, err := h.store.BookingEngine.ListPromoCodes(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, codes)
}

func (h *Handler) createPromoCode(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var p domain.PromoCode
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	res, err := h.store.BookingEngine.CreatePromoCode(r.Context(), tenantID, &p)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, res)
}

func (h *Handler) deletePromoCode(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "codeId")
	if err := h.store.BookingEngine.DeletePromoCode(r.Context(), tenantID, id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *Handler) validatePromoCode(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	code := chi.URLParam(r, "code")
	p, err := h.store.BookingEngine.GetPromoCodeByCode(r.Context(), tenantID, code)
	if err != nil {
		respondError(w, http.StatusNotFound, "invalid or expired promo code")
		return
	}
	now := time.Now()
	if now.After(p.ValidUntil) || (p.MaxUses > 0 && p.UsesCount >= p.MaxUses) {
		respondError(w, http.StatusBadRequest, "promo code expired or max uses reached")
		return
	}
	respondJSON(w, http.StatusOK, p)
}

// ============== WIDGET CONFIG ==============

func (h *Handler) getWidgetConfig(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	cfg, err := h.store.BookingEngine.GetWidgetConfig(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, cfg)
}

func (h *Handler) saveWidgetConfig(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var cfg domain.DirectBookingWidgetConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.store.BookingEngine.SaveWidgetConfig(r.Context(), tenantID, &cfg); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

// ============== UPSELLS ==============

func (h *Handler) listUpsells(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	upsells, err := h.store.BookingEngine.ListUpsells(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, upsells)
}

func (h *Handler) createUpsell(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var u domain.BookingUpsell
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	res, err := h.store.BookingEngine.CreateUpsell(r.Context(), tenantID, &u)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, res)
}

// ============== DIRECT BOOKING SESSIONS ==============

func (h *Handler) createBookingSession(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var s domain.DirectBookingSession
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	s.Status = "pending"
	s.ExpiresAt = time.Now().Add(1 * time.Hour)
	res, err := h.store.BookingEngine.CreateBookingSession(r.Context(), tenantID, &s)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, res)
}

func (h *Handler) getBookingSession(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "sessionId")
	s, err := h.store.BookingEngine.GetBookingSession(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, s)
}

func (h *Handler) confirmBookingSession(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "sessionId")
	if err := h.store.BookingEngine.UpdateBookingSessionStatus(r.Context(), tenantID, id, "confirmed"); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "confirmed"})
}

// ============== PUBLIC WIDGET ENDPOINTS (No Auth) ==============

func (h *Handler) publicGetWidgetConfig(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	cfg, err := h.store.BookingEngine.GetWidgetConfig(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	// Strip sensitive fields
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"enabled":              cfg.Enabled,
		"theme_color":          cfg.ThemeColor,
		"logo_url":             cfg.LogoURL,
		"title":                cfg.Title,
		"subtitle":             cfg.Subtitle,
		"show_promo_code":      cfg.ShowPromoCode,
		"show_extras":          cfg.ShowExtras,
		"upsell_enabled":       cfg.UpsellEnabled,
		"require_deposit":      cfg.RequireDeposit,
		"deposit_pct":          cfg.DepositPct,
		"success_redirect_url": cfg.SuccessRedirectURL,
	})
}

func (h *Handler) publicCreateBooking(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	var s domain.DirectBookingSession
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}

	// Validate promo code if provided
	if s.PromoCodeID != "" {
		// In real implementation, look up promo and apply discount
	}

	s.Status = "pending"
	s.ExpiresAt = time.Now().Add(1 * time.Hour)
	res, err := h.store.BookingEngine.CreateBookingSession(r.Context(), tenantID, &s)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, res)
}

func (h *Handler) publicCheckAvailability(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "tenantId")
	checkIn := r.URL.Query().Get("check_in")
	checkOut := r.URL.Query().Get("check_out")
	adults, _ := strconv.Atoi(r.URL.Query().Get("adults"))
	if adults == 0 {
		adults = 2
	}

	// In real implementation, query availability from rooms
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"check_in":  checkIn,
		"check_out": checkOut,
		"adults":    adults,
		"available": true,
		"room_types": []map[string]interface{}{
			{"id": "rt1", "name": "Standard", "price_per_night": 120, "available_rooms": 5},
			{"id": "rt2", "name": "Deluxe", "price_per_night": 180, "available_rooms": 3},
			{"id": "rt3", "name": "Suite", "price_per_night": 280, "available_rooms": 1},
		},
	})
}
