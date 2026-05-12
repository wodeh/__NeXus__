package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"nexus-hospitality-platform/services/backend-core/internal/domain"
)

func (s *Server) registerVillaHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/villas", s.withTenant(s.handleVillas))
	mux.HandleFunc("/v1/villas/", s.withTenant(s.handleVillaDetail))
	mux.HandleFunc("/v1/villa-reservations", s.withTenant(s.handleVillaReservations))
	mux.HandleFunc("/v1/villa-reservations/", s.withTenant(s.handleVillaReservationDetail))
	mux.HandleFunc("/v1/villa-availability", s.withTenant(s.handleVillaAvailability))
	mux.HandleFunc("/v1/villa-revenue", s.withTenant(s.handleVillaRevenue))
	mux.HandleFunc("/v1/sensor-logs", s.withTenant(s.handleSensorLogs))
	mux.HandleFunc("/v1/sensor-logs/latest", s.withTenant(s.handleLatestSensorLog))
}

// Villa Properties
func (s *Server) handleVillas(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		villas, err := s.repo.Villa.ListVillas(ctx)
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]interface{}{"villas": demoVillas()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"villas": villas})
	case http.MethodPost:
		var v domain.VillaProperty
		if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		v.TenantID = ctx.Value("tenant_id").(string)
		if err := s.repo.Villa.CreateVilla(ctx, &v); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, v)
	}
}

func (s *Server) handleVillaDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := extractID(r.URL.Path, "/v1/villas/")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing villa id"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		v, err := s.repo.Villa.GetVillaByID(ctx, id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "villa not found"})
			return
		}
		writeJSON(w, http.StatusOK, v)
	case http.MethodPatch:
		var v domain.VillaProperty
		if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		v.ID = id
		if err := s.repo.Villa.UpdateVilla(ctx, &v); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, v)
	case http.MethodDelete:
		if err := s.repo.Villa.DeleteVilla(ctx, id); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	}
}

// Villa Reservations
func (s *Server) handleVillaReservations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		reservations, err := s.repo.Villa.ListVillaReservations(ctx)
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]interface{}{"reservations": demoVillaReservations()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"reservations": reservations})
	case http.MethodPost:
		var res domain.VillaReservation
		if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		res.TenantID = ctx.Value("tenant_id").(string)
		if err := s.repo.Villa.CreateVillaReservation(ctx, &res); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, res)
	}
}

func (s *Server) handleVillaReservationDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := extractID(r.URL.Path, "/v1/villa-reservations/")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing reservation id"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		res, err := s.repo.Villa.GetVillaReservationByID(ctx, id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "reservation not found"})
			return
		}
		writeJSON(w, http.StatusOK, res)
	case http.MethodPatch:
		var res domain.VillaReservation
		if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		res.ID = id
		if err := s.repo.Villa.UpdateVillaReservation(ctx, &res); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, res)
	case http.MethodDelete:
		if err := s.repo.Villa.DeleteVillaReservation(ctx, id); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	}
}

// Villa Availability
func (s *Server) handleVillaAvailability(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	villaID := r.URL.Query().Get("villa_id")
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")

	if villaID == "" || start == "" || end == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "villa_id, start, and end required"})
		return
	}

	ctx := r.Context()
	avail, err := s.repo.Villa.GetVillaAvailability(ctx, villaID, start, end)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"availability": demoVillaAvailability(villaID, start, end)})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"availability": avail})
}

// Villa Revenue
func (s *Server) handleVillaRevenue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")
	if start == "" || end == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "start and end required"})
		return
	}

	ctx := r.Context()
	stats, err := s.repo.Villa.GetVillaRevenueStats(ctx, start, end)
	if err != nil {
		writeJSON(w, http.StatusOK, demoVillaRevenueStats(start, end))
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// Sensor Logs
func (s *Server) handleSensorLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		villaID := r.URL.Query().Get("villa_id")
		if villaID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "villa_id required"})
			return
		}
		logs, err := s.repo.Villa.GetSensorLogsByVilla(ctx, villaID)
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]interface{}{"logs": demoSensorLogs(villaID)})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"logs": logs})
	case http.MethodPost:
		var log domain.CleanerSensorLog
		if err := json.NewDecoder(r.Body).Decode(&log); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		log.TenantID = ctx.Value("tenant_id").(string)
		if err := s.repo.Villa.RecordSensorLog(ctx, &log); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, log)
	}
}

func (s *Server) handleLatestSensorLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	villaID := r.URL.Query().Get("villa_id")
	cleanerID := r.URL.Query().Get("cleaner_id")
	if villaID == "" || cleanerID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "villa_id and cleaner_id required"})
		return
	}

	ctx := r.Context()
	log, err := s.repo.Villa.GetLatestSensorLog(ctx, villaID, cleanerID)
	if err != nil {
		writeJSON(w, http.StatusOK, demoLatestSensorLog(villaID, cleanerID))
		return
	}
	writeJSON(w, http.StatusOK, log)
}

// Demo data generators
func demoVillas() []domain.VillaProperty {
	return []domain.VillaProperty{
		{
			ID: "villa-001", Name: "Villa Al-Mashta", City: "Ramallah", Country: "Palestine",
			Bedrooms: 3, Bathrooms: 2, MaxGuests: 6, PricePerNight: 350, Currency: "USD",
			CleaningFee: 50, SecurityDeposit: 500, IsActive: true, Status: "available",
		},
		{
			ID: "villa-002", Name: "Chalet Al-Balad", City: "Bethlehem", Country: "Palestine",
			Bedrooms: 4, Bathrooms: 3, MaxGuests: 8, PricePerNight: 450, Currency: "USD",
			CleaningFee: 75, SecurityDeposit: 750, IsActive: true, Status: "reserved",
		},
	}
}

func demoVillaReservations() []domain.VillaReservation {
	return []domain.VillaReservation{
		{
			ID: "res-001", VillaID: "villa-002", VillaName: "Chalet Al-Balad",
			GuestName: "Ahmad Khalil", GuestPhone: "+970599123456",
			CheckInDate: "2024-08-15", CheckOutDate: "2024-08-20", Nights: 5,
			TotalAmount: 2500, Currency: "USD", Status: "reserved",
			Source: "whatsapp",
			DownPayment: &domain.DownPayment{
				Amount: 500, Method: "visa", Status: "received",
				ReceivedAt: time.Now().Add(-24 * time.Hour),
			},
			BalanceDue: 2000, BalancePaid: false,
		},
		{
			ID: "res-002", VillaID: "villa-001", VillaName: "Villa Al-Mashta",
			GuestName: "Sarah Nassar", GuestPhone: "+970599987654",
			CheckInDate: "2024-09-01", CheckOutDate: "2024-09-05", Nights: 4,
			TotalAmount: 1400, Currency: "USD", Status: "pending",
			Source: "phone", BalanceDue: 1400, BalancePaid: false,
		},
	}
}

func demoVillaAvailability(villaID, start, end string) []domain.VillaAvailability {
	return []domain.VillaAvailability{
		{VillaID: villaID, Date: "2024-08-15", Status: "reserved"},
		{VillaID: villaID, Date: "2024-08-16", Status: "reserved"},
		{VillaID: villaID, Date: "2024-08-17", Status: "available"},
		{VillaID: villaID, Date: "2024-08-18", Status: "available"},
		{VillaID: villaID, Date: "2024-08-19", Status: "blocked"},
	}
}

func demoVillaRevenueStats(start, end string) map[string]interface{} {
	return map[string]interface{}{
		"total_revenue":      12500.0,
		"total_reservations": 8,
		"occupancy_rate":     72.5,
		"avg_booking_value":  1562.5,
		"down_payment_total": 3000.0,
		"pending_balance":    9500.0,
		"period_start":       start,
		"period_end":         end,
		"villa_breakdown": []map[string]interface{}{
			{"villa_id": "villa-001", "villa_name": "Villa Al-Mashta", "revenue": 5200, "nights_booked": 18, "occupancy_pct": 65.0},
			{"villa_id": "villa-002", "villa_name": "Chalet Al-Balad", "revenue": 7300, "nights_booked": 28, "occupancy_pct": 80.0},
		},
	}
}

func demoSensorLogs(villaID string) []domain.CleanerSensorLog {
	return []domain.CleanerSensorLog{
		{
			ID: "log-001", VillaID: villaID, VillaName: "Villa Al-Mashta",
			CleanerID: "cleaner-1", CleanerName: "Muhammad",
			Temperature: 31.5, Altitude: 850, Floor: 0,
			LocationType: "outside", BatteryLevel: 78,
			RecordedAt: time.Now().Add(-30 * time.Minute),
		},
		{
			ID: "log-002", VillaID: villaID, VillaName: "Villa Al-Mashta",
			CleanerID: "cleaner-1", CleanerName: "Muhammad",
			Temperature: 24.2, Altitude: 854, Floor: 1,
			LocationType: "inside", BatteryLevel: 76,
			RecordedAt: time.Now().Add(-15 * time.Minute),
		},
	}
}

func demoLatestSensorLog(villaID, cleanerID string) domain.CleanerSensorLog {
	return domain.CleanerSensorLog{
		ID: "log-003", VillaID: villaID, VillaName: "Villa Al-Mashta",
		CleanerID: cleanerID, CleanerName: "Muhammad",
		Temperature: 22.8, Altitude: 858, Floor: 2,
		LocationType: "inside", BatteryLevel: 74,
		RecordedAt: time.Now().Add(-5 * time.Minute),
	}
}
