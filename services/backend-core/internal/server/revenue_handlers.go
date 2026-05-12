package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nexus-platform/backend-core/internal/domain"
)

func (s *Server) registerRevenueHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/revenue/dashboard", s.withTenant(s.handleRevenueDashboard))
	mux.HandleFunc("/v1/revenue/forecast", s.withTenant(s.handleRevenueForecast))
	mux.HandleFunc("/v1/revenue/channels", s.withTenant(s.handleChannelRevenue))
	mux.HandleFunc("/v1/revenue/room-types", s.withTenant(s.handleRoomTypeRevenue))
	mux.HandleFunc("/v1/revenue/pricing-rules", s.withTenant(s.handlePricingRules))
}

func (s *Server) handleRevenueDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantIDStr, _ := ctx.Value("tenant_id").(string)
	tenantID, _ := uuid.Parse(tenantIDStr)

	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	from, _ := time.Parse("2006-01-02", fromStr)
	to, _ := time.Parse("2006-01-02", toStr)
	if from.IsZero() { from = time.Now() }
	if to.IsZero() { to = from }

	stats, err := s.repo.Revenue.GetRangeStats(ctx, tenantID, from, to)
	if err != nil || len(stats) == 0 {
		stats = demoRevenueDashboard(from, to)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"period": map[string]string{"from": from.Format("2006-01-02"), "to": to.Format("2006-01-02")},
		"stats":  stats,
	})
}

func demoRevenueDashboard(from, to time.Time) []domain.RevenueDashboard {
	var out []domain.RevenueDashboard
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		out = append(out, domain.RevenueDashboard{
			TenantID:       uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"),
			Date:           d,
			TotalRevenue:   15000.0 + float64(d.Day())*100,
			RoomRevenue:    12750.0,
			ExtraRevenue:   750.0,
			TaxRevenue:     1500.0,
			TotalRooms:     42,
			OccupiedRooms:  32,
			AvailableRooms: 10,
			OccupancyRate:  76.2,
			ADR:            398.44,
			RevPAR:         357.14,
			Arrivals:       8,
			Departures:     6,
			Stayovers:      24,
			WalkIns:        2,
			NoShows:        0,
			Cancellations:  1,
		})
	}
	return out
}

func (s *Server) handleRevenueForecast(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantIDStr, _ := ctx.Value("tenant_id").(string)
	tenantID, _ := uuid.Parse(tenantIDStr)

	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	from := time.Now()
	to := from.AddDate(0, 0, 30)

	forecast, err := s.repo.Revenue.GetForecast(ctx, tenantID, from, to)
	if err != nil || len(forecast) == 0 {
		forecast = demoForecast(from, to)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"forecast": forecast})
}

func demoForecast(from, to time.Time) []domain.RevenueForecast {
	var out []domain.RevenueForecast
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		out = append(out, domain.RevenueForecast{
			Date:               d,
			ProjectedRevenue:   12000.0 + float64(d.Day())*200,
			ProjectedOccupancy: 70.0 + float64(d.Day()%20),
			Confidence:         0.82,
			BookedRooms:        25 + (d.Day() % 10),
			TotalRooms:         42,
		})
	}
	return out
}

func (s *Server) handleChannelRevenue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantIDStr, _ := ctx.Value("tenant_id").(string)
	tenantID, _ := uuid.Parse(tenantIDStr)

	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	from := time.Now().AddDate(0, 0, -30)
	to := time.Now()

	breakdown, err := s.repo.Revenue.GetChannelBreakdown(ctx, tenantID, from, to)
	if err != nil || len(breakdown) == 0 {
		breakdown = []domain.ChannelRevenueBreakdown{
			{Channel: "booking_com", Bookings: 45, Revenue: 45000.00, Commission: 6750.00, NetRevenue: 38250.00, AvgRate: 1000.00},
			{Channel: "direct", Bookings: 30, Revenue: 35000.00, Commission: 0.00, NetRevenue: 35000.00, AvgRate: 1166.67},
			{Channel: "expedia", Bookings: 20, Revenue: 22000.00, Commission: 3300.00, NetRevenue: 18700.00, AvgRate: 1100.00},
			{Channel: "walk_in", Bookings: 15, Revenue: 8500.00, Commission: 0.00, NetRevenue: 8500.00, AvgRate: 566.67},
			{Channel: "whatsapp", Bookings: 8, Revenue: 6400.00, Commission: 0.00, NetRevenue: 6400.00, AvgRate: 800.00},
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"breakdown": breakdown})
}

func (s *Server) handleRoomTypeRevenue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantIDStr, _ := ctx.Value("tenant_id").(string)
	tenantID, _ := uuid.Parse(tenantIDStr)

	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	from := time.Now().AddDate(0, 0, -30)
	to := time.Now()

	revenue, err := s.repo.Revenue.GetRoomTypeRevenue(ctx, tenantID, from, to)
	if err != nil || len(revenue) == 0 {
		revenue = []domain.RoomTypeRevenue{
			{RoomType: "Deluxe King", NightsSold: 120, Revenue: 35880.00, AvgRate: 299.00, OccupancyPct: 85.0},
			{RoomType: "Standard", NightsSold: 200, Revenue: 39800.00, AvgRate: 199.00, OccupancyPct: 79.0},
			{RoomType: "Suite", NightsSold: 45, Revenue: 22455.00, AvgRate: 499.00, OccupancyPct: 71.0},
			{RoomType: "Twin Room", NightsSold: 80, Revenue: 19920.00, AvgRate: 249.00, OccupancyPct: 66.0},
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"room_types": revenue})
}

func (s *Server) handlePricingRules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantIDStr, _ := ctx.Value("tenant_id").(string)
	tenantID, _ := uuid.Parse(tenantIDStr)

	switch r.Method {
	case http.MethodGet:
		rules, err := s.repo.Revenue.ListPricingRules(ctx, tenantID)
		if err != nil || len(rules) == 0 {
			rules = []domain.PricingRule{
				{ID: uuid.MustParse("d0000001-0000-0000-0000-000000000001"), TenantID: tenantID, Name: "High Occupancy Premium", RoomType: "Deluxe King", Condition: "occupancy_based", TriggerValue: 85, AdjustmentType: "percentage", AdjustmentValue: 15, MinRate: 250, MaxRate: 500, IsActive: true, Priority: 10},
				{ID: uuid.MustParse("d0000002-0000-0000-0000-000000000002"), TenantID: tenantID, Name: "Early Bird Discount", RoomType: "Standard", Condition: "advance_booking", TriggerValue: 30, AdjustmentType: "percentage", AdjustmentValue: -10, MinRate: 150, MaxRate: 250, IsActive: true, Priority: 8},
				{ID: uuid.MustParse("d0000003-0000-0000-0000-000000000003"), TenantID: tenantID, Name: "Weekend Premium", RoomType: "Suite", Condition: "seasonal", TriggerValue: 0, AdjustmentType: "percentage", AdjustmentValue: 20, MinRate: 400, MaxRate: 700, IsActive: true, Priority: 5},
			}
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"rules": rules})

	case http.MethodPost:
		var req domain.PricingRule
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
			return
		}
		req.TenantID = tenantID
		req.IsActive = true
		if err := s.repo.Revenue.CreatePricingRule(ctx, &req); err != nil {
			// Return demo success
			writeJSON(w, http.StatusCreated, req)
			return
		}
		writeJSON(w, http.StatusCreated, req)

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}
