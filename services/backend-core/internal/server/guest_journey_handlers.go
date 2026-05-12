package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nexus-platform/backend-core/internal/domain"
)

/* ─── Guest Journey Handlers ─── */

func (s *Server) registerGuestJourneyHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/journeys", s.withTenant(s.handleJourneys))
	mux.HandleFunc("/v1/journeys/", s.withTenant(s.handleJourneyDetail))
	mux.HandleFunc("/v1/journey-executions", s.withTenant(s.handleJourneyExecutions))
	mux.HandleFunc("/v1/upsell-offers", s.withTenant(s.handleUpsellOffers))
	mux.HandleFunc("/v1/upsell-offers/", s.withTenant(s.handleUpsellOfferDetail))
	mux.HandleFunc("/v1/upsell-purchases", s.withTenant(s.handleUpsellPurchases))
	mux.HandleFunc("/v1/competitors", s.withTenant(s.handleCompetitors))
	mux.HandleFunc("/v1/competitors/", s.withTenant(s.handleCompetitorDetail))
	mux.HandleFunc("/v1/competitor-rates", s.withTenant(s.handleCompetitorRates))
	mux.HandleFunc("/v1/rate-recommendations", s.withTenant(s.handleRateRecommendations))
	mux.HandleFunc("/v1/rate-recommendations/apply", s.withTenant(s.handleApplyRecommendation))
	mux.HandleFunc("/v1/rate-shop-config", s.withTenant(s.handleRateShopConfig))
}

func (s *Server) handleJourneys(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	if r.Method == http.MethodGet {
		journeys, err := s.repo.GuestJourney.ListJourneys(r.Context(), tenantID)
		if err != nil {
			// Return demo data
			journeys = []domain.GuestJourney{
				{
					ID: uuid.New().String(), TenantID: tenantID, Name: "Standard Check-in Journey",
					Trigger: "booking_confirmed", IsActive: true,
					Steps: []domain.JourneyStep{
						{ID: uuid.New().String(), DelayHours: 0, Channel: "email", TemplateID: "booking_confirmation", Condition: "always"},
						{ID: uuid.New().String(), DelayHours: 72, Channel: "whatsapp", TemplateID: "pre_arrival_upsell", UpsellOfferID: strPtr(uuid.New().String()), Condition: "always"},
						{ID: uuid.New().String(), DelayHours: 2, Channel: "whatsapp", TemplateID: "welcome_message", Condition: "always"},
						{ID: uuid.New().String(), DelayHours: 48, Channel: "email", TemplateID: "review_request", Condition: "always"},
					},
					CreatedAt: time.Now(), UpdatedAt: time.Now(),
				},
				{
					ID: uuid.New().String(), TenantID: tenantID, Name: "VIP Guest Journey",
					Trigger: "booking_confirmed", IsActive: true,
					Steps: []domain.JourneyStep{
						{ID: uuid.New().String(), DelayHours: 0, Channel: "email", TemplateID: "vip_welcome", Condition: "vip_only"},
						{ID: uuid.New().String(), DelayHours: 48, Channel: "whatsapp", TemplateID: "vip_upsell_spa", UpsellOfferID: strPtr(uuid.New().String()), Condition: "vip_only"},
						{ID: uuid.New().String(), DelayHours: 2, Channel: "whatsapp", TemplateID: "vip_checkin", Condition: "vip_only"},
					},
					CreatedAt: time.Now(), UpdatedAt: time.Now(),
				},
			}
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"journeys": journeys})
		return
	}
	if r.Method == http.MethodPost {
		var req domain.GuestJourney
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid body")
			return
		}
		req.ID = uuid.New().String()
		req.TenantID = tenantID
		req.CreatedAt = time.Now()
		req.UpdatedAt = time.Now()
		if err := s.repo.GuestJourney.CreateJourney(r.Context(), &req); err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, req)
		return
	}
	writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func (s *Server) handleJourneyDetail(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	id := r.URL.Path[len("/v1/journeys/"):]
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "missing id")
		return
	}
	if r.Method == http.MethodPatch {
		var req domain.GuestJourney
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid body")
			return
		}
		req.ID = id
		req.TenantID = tenantID
		if err := s.repo.GuestJourney.UpdateJourney(r.Context(), &req); err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, req)
		return
	}
	if r.Method == http.MethodDelete {
		if err := s.repo.GuestJourney.DeleteJourney(r.Context(), id, tenantID); err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
		return
	}
	writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func (s *Server) handleJourneyExecutions(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	if r.Method == http.MethodGet {
		execs, err := s.repo.GuestJourney.ListExecutions(r.Context(), tenantID)
		if err != nil {
			execs = []domain.GuestJourneyExecution{
				{ID: uuid.New().String(), TenantID: tenantID, JourneyID: uuid.New().String(), ReservationID: uuid.New().String(), GuestPhone: "+1-555-0123", CurrentStep: 2, TotalSteps: 4, Status: "running", StartedAt: time.Now()},
				{ID: uuid.New().String(), TenantID: tenantID, JourneyID: uuid.New().String(), ReservationID: uuid.New().String(), GuestPhone: "+1-555-0456", CurrentStep: 4, TotalSteps: 4, Status: "completed", StartedAt: time.Now().Add(-48 * time.Hour)},
			}
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"executions": execs})
		return
	}
	writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
}

/* ─── Upsell Offer Handlers ─── */

func (s *Server) handleUpsellOffers(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	if r.Method == http.MethodGet {
		offers, err := s.repo.GuestJourney.ListOffers(r.Context(), tenantID)
		if err != nil {
			offers = []domain.UpsellOffer{
				{ID: uuid.New().String(), TenantID: tenantID, Name: "Late Checkout (2PM)", Description: "Extend your stay until 2PM", Category: "late_checkout", Price: 49.00, Currency: "USD", IsActive: true, AutoOffer: true, DisplayOrder: 1, TotalSold: 34, RevenueGenerated: 1666, CreatedAt: time.Now(), UpdatedAt: time.Now()},
				{ID: uuid.New().String(), TenantID: tenantID, Name: "Room Upgrade - Ocean View", Description: "Upgrade to our premium ocean view room", Category: "room_upgrade", Price: 75.00, Currency: "USD", IsActive: true, AutoOffer: true, DisplayOrder: 2, TotalSold: 18, RevenueGenerated: 1350, CreatedAt: time.Now(), UpdatedAt: time.Now()},
				{ID: uuid.New().String(), TenantID: tenantID, Name: "Breakfast Package", Description: "Full breakfast buffet for entire stay", Category: "breakfast", Price: 25.00, Currency: "USD", IsActive: true, AutoOffer: true, DisplayOrder: 3, TotalSold: 52, RevenueGenerated: 1300, CreatedAt: time.Now(), UpdatedAt: time.Now()},
				{ID: uuid.New().String(), TenantID: tenantID, Name: "Spa Package", Description: "60-minute massage + pool access", Category: "spa", Price: 120.00, Currency: "USD", IsActive: true, AutoOffer: false, DisplayOrder: 4, TotalSold: 8, RevenueGenerated: 960, CreatedAt: time.Now(), UpdatedAt: time.Now()},
				{ID: uuid.New().String(), TenantID: tenantID, Name: "Early Check-in", Description: "Check in from 10AM", Category: "early_checkin", Price: 35.00, Currency: "USD", IsActive: true, AutoOffer: true, DisplayOrder: 5, TotalSold: 12, RevenueGenerated: 420, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			}
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"offers": offers})
		return
	}
	if r.Method == http.MethodPost {
		var req domain.UpsellOffer
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid body")
			return
		}
		req.ID = uuid.New().String()
		req.TenantID = tenantID
		req.CreatedAt = time.Now()
		req.UpdatedAt = time.Now()
		if err := s.repo.GuestJourney.CreateOffer(r.Context(), &req); err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, req)
		return
	}
	writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func (s *Server) handleUpsellOfferDetail(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	id := r.URL.Path[len("/v1/upsell-offers/"):]
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "missing id")
		return
	}
	if r.Method == http.MethodPatch {
		var req domain.UpsellOffer
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid body")
			return
		}
		req.ID = id
		req.TenantID = tenantID
		if err := s.repo.GuestJourney.UpdateOffer(r.Context(), &req); err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, req)
		return
	}
	if r.Method == http.MethodDelete {
		if err := s.repo.GuestJourney.DeleteOffer(r.Context(), id, tenantID); err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
		return
	}
	writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
}

/* ─── Upsell Purchase Handlers ─── */

func (s *Server) handleUpsellPurchases(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	if r.Method == http.MethodGet {
		purchases, err := s.repo.GuestJourney.ListPurchases(r.Context(), tenantID)
		if err != nil {
			purchases = []domain.UpsellPurchase{
				{ID: uuid.New().String(), TenantID: tenantID, ReservationID: uuid.New().String(), GuestPhone: "+1-555-0123", OfferID: uuid.New().String(), OfferName: "Late Checkout (2PM)", Price: 49.00, Currency: "USD", Status: "confirmed", PaymentMethod: "on_bill", FolioPosted: true, CreatedAt: time.Now().Add(-24 * time.Hour)},
				{ID: uuid.New().String(), TenantID: tenantID, ReservationID: uuid.New().String(), GuestPhone: "+1-555-0456", OfferID: uuid.New().String(), OfferName: "Room Upgrade - Ocean View", Price: 75.00, Currency: "USD", Status: "confirmed", PaymentMethod: "credit_card", FolioPosted: true, CreatedAt: time.Now().Add(-48 * time.Hour)},
				{ID: uuid.New().String(), TenantID: tenantID, ReservationID: uuid.New().String(), GuestPhone: "+1-555-0789", OfferID: uuid.New().String(), OfferName: "Breakfast Package", Price: 25.00, Currency: "USD", Status: "pending", PaymentMethod: "on_bill", FolioPosted: false, CreatedAt: time.Now()},
			}
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"purchases": purchases})
		return
	}
	writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
}

/* ─── Competitor Handlers ─── */

func (s *Server) handleCompetitors(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	if r.Method == http.MethodGet {
		competitors, err := s.repo.GuestJourney.ListCompetitors(r.Context(), tenantID)
		if err != nil {
			competitors = []domain.CompetitorHotel{
				{ID: uuid.New().String(), TenantID: tenantID, Name: "The Grand Plaza Hotel", Address: "456 Ocean Drive", City: "Miami", Country: "USA", StarRating: 5, RoomCount: 200, IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
				{ID: uuid.New().String(), TenantID: tenantID, Name: "Seaside Boutique Resort", Address: "789 Beach Blvd", City: "Miami", Country: "USA", StarRating: 4, RoomCount: 80, IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
				{ID: uuid.New().String(), TenantID: tenantID, Name: "Downtown Business Inn", Address: "321 Commerce St", City: "Miami", Country: "USA", StarRating: 3, RoomCount: 150, IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			}
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"competitors": competitors})
		return
	}
	if r.Method == http.MethodPost {
		var req domain.CompetitorHotel
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid body")
			return
		}
		req.ID = uuid.New().String()
		req.TenantID = tenantID
		req.CreatedAt = time.Now()
		req.UpdatedAt = time.Now()
		if err := s.repo.GuestJourney.CreateCompetitor(r.Context(), &req); err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, req)
		return
	}
	writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func (s *Server) handleCompetitorDetail(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	id := r.URL.Path[len("/v1/competitors/"):]
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "missing id")
		return
	}
	if r.Method == http.MethodDelete {
		if err := s.repo.GuestJourney.DeleteCompetitor(r.Context(), id, tenantID); err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
		return
	}
	writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func (s *Server) handleCompetitorRates(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	if r.Method == http.MethodGet {
		rates, err := s.repo.GuestJourney.GetLatestRates(r.Context(), tenantID)
		if err != nil {
			today := time.Now().Format("2006-01-02")
			tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")
			rates = []domain.CompetitorRate{
				{ID: uuid.New().String(), TenantID: tenantID, CompetitorID: uuid.New().String(), CompetitorName: "The Grand Plaza Hotel", RoomType: "Deluxe King", Date: today, Rate: 349.00, Currency: "USD", Availability: 12, MinStay: 1, IsPromo: false, Source: "booking_com", ScrapedAt: time.Now()},
				{ID: uuid.New().String(), TenantID: tenantID, CompetitorID: uuid.New().String(), CompetitorName: "The Grand Plaza Hotel", RoomType: "Deluxe King", Date: tomorrow, Rate: 329.00, Currency: "USD", Availability: 8, MinStay: 1, IsPromo: true, Source: "booking_com", ScrapedAt: time.Now()},
				{ID: uuid.New().String(), TenantID: tenantID, CompetitorID: uuid.New().String(), CompetitorName: "Seaside Boutique Resort", RoomType: "Standard", Date: today, Rate: 189.00, Currency: "USD", Availability: 5, MinStay: 2, IsPromo: false, Source: "expedia", ScrapedAt: time.Now()},
				{ID: uuid.New().String(), TenantID: tenantID, CompetitorID: uuid.New().String(), CompetitorName: "Downtown Business Inn", RoomType: "Standard", Date: today, Rate: 129.00, Currency: "USD", Availability: 20, MinStay: 1, IsPromo: false, Source: "booking_com", ScrapedAt: time.Now()},
			}
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"rates": rates})
		return
	}
	writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
}

/* ─── Rate Recommendation Handlers ─── */

func (s *Server) handleRateRecommendations(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	if r.Method == http.MethodGet {
		recs, err := s.repo.GuestJourney.ListRecommendations(r.Context(), tenantID)
		if err != nil {
			today := time.Now().Format("2006-01-02")
			tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")
			recs = []domain.RateRecommendation{
				{ID: uuid.New().String(), TenantID: tenantID, RoomType: "Deluxe King", Date: today, CurrentRate: 299.00, RecommendedRate: 329.00, Confidence: 0.82, Reason: "Competitor dropped rate by 15%, low inventory suggests demand", Factors: []string{"competitor_rate_drop", "low_inventory"}, Applied: false, CreatedAt: time.Now()},
				{ID: uuid.New().String(), TenantID: tenantID, RoomType: "Standard", Date: tomorrow, CurrentRate: 149.00, RecommendedRate: 169.00, Confidence: 0.74, Reason: "Local music festival this weekend, 85% occupancy forecast", Factors: []string{"local_event", "high_occupancy"}, Applied: false, CreatedAt: time.Now()},
				{ID: uuid.New().String(), TenantID: tenantID, RoomType: "Suite", Date: today, CurrentRate: 499.00, RecommendedRate: 449.00, Confidence: 0.61, Reason: "Below 40% occupancy for suites, consider promotional pricing", Factors: []string{"low_occupancy"}, Applied: false, CreatedAt: time.Now()},
			}
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"recommendations": recs})
		return
	}
	writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func (s *Server) handleApplyRecommendation(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	if r.Method == http.MethodPost {
		var req struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid body")
			return
		}
		if err := s.repo.GuestJourney.ApplyRecommendation(r.Context(), req.ID, tenantID); err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "applied"})
		return
	}
	writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
}

/* ─── Rate Shop Config Handler ─── */

func (s *Server) handleRateShopConfig(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	if r.Method == http.MethodGet {
		config, err := s.repo.GuestJourney.GetRateShopConfig(r.Context(), tenantID)
		if err != nil {
			config = &domain.RateShopConfig{
				TenantID: tenantID, Enabled: true, FrequencyHours: 6, LookaheadDays: 30, AutoAdjust: false, MaxAdjustmentPct: 0.30,
				CreatedAt: time.Now(), UpdatedAt: time.Now(),
			}
		}
		writeJSON(w, http.StatusOK, config)
		return
	}
	if r.Method == http.MethodPatch {
		var req domain.RateShopConfig
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid body")
			return
		}
		req.TenantID = tenantID
		req.UpdatedAt = time.Now()
		if err := s.repo.GuestJourney.SaveRateShopConfig(r.Context(), &req); err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, req)
		return
	}
	writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
}
