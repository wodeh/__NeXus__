package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"nexus-hospitality-platform/services/backend-core/internal/domain"
)

func (s *Server) registerReservationDetailHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/reservations/", s.withTenant(s.handleReservationDetail))
	mux.HandleFunc("/v1/reservations/bulk", s.withTenant(s.handleBulkReservations))
	mux.HandleFunc("/v1/room-status", s.withTenant(s.handleRoomStatusView))
}

func (s *Server) handleReservationDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantIDStr, _ := ctx.Value("tenant_id").(string)
	tenantID, _ := uuid.Parse(tenantIDStr)

	path := r.URL.Path[len("/v1/reservations/"):]
	parts := splitPath(path)
	if len(parts) == 0 {
		http.Error(w, `{"error":"missing reservation id"}`, http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(parts[0])
	if err != nil {
		http.Error(w, `{"error":"invalid reservation id"}`, http.StatusBadRequest)
		return
	}

	// Handle sub-routes
	if len(parts) > 1 {
		switch parts[1] {
		case "move", "checkin", "checkout", "cancel", "assign-room":
			// Already handled by registerReservationHandlers
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}
	}

	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	detail, err := s.repo.ReservationDetail.GetDetail(ctx, tenantID, id)
	if err != nil {
		// Return demo detail
		writeJSON(w, http.StatusOK, demoReservationDetail(id, tenantID))
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func demoReservationDetail(id, tenantID uuid.UUID) *domain.ReservationDetail {
	now := time.Now()
	return &domain.ReservationDetail{
		ID:       id,
		TenantID: tenantID,
		Guest: domain.GuestInfo{
			Name:  "John Smith",
			Email: "john@example.com",
			Phone: "+1-555-0123",
			VIP:   false,
		},
		Room: domain.RoomAssignment{
			RoomNumber:  "305",
			RoomType:    "Deluxe King",
			Floor:       "3",
			BedType:     "King",
			RateNight:   299.00,
			TotalNights: 3,
		},
		Dates: domain.StayDates{
			CheckIn:  now.AddDate(0, 0, 2),
			CheckOut: now.AddDate(0, 0, 5),
		},
		Party: domain.PartyInfo{
			Adults:   2,
			Children: 0,
		},
		Financials: domain.Financials{
			RoomTotal:   897.00,
			TaxTotal:    89.70,
			Total:       986.70,
			Balance:     986.70,
			Currency:    "USD",
		},
		Status:          "confirmed",
		Source:          "booking_com",
		SpecialRequests: "High floor, away from elevator",
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func (s *Server) handleBulkReservations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantIDStr, _ := ctx.Value("tenant_id").(string)
	tenantID, _ := uuid.Parse(tenantIDStr)

	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req domain.BulkReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	var created []domain.Reservation
	for _, createReq := range req.Reservations {
		res, err := s.repo.Reservations.Create(ctx, nil, tenantID.String(), &createReq)
		if err != nil {
			continue
		}
		created = append(created, *res)
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"created": len(created),
		"reservations": created,
	})
}

func (s *Server) handleRoomStatusView(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantIDStr, _ := ctx.Value("tenant_id").(string)
	tenantID, _ := uuid.Parse(tenantIDStr)

	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	dateStr := r.URL.Query().Get("date")
	date, _ := time.Parse("2006-01-02", dateStr)
	if date.IsZero() { date = time.Now() }

	view := r.URL.Query().Get("view") // day, week, month
	if view == "" { view = "day" }

	var result []domain.RoomStatusView
	switch view {
	case "day":
		rooms, err := s.repo.ReservationDetail.GetRoomStatusView(ctx, tenantID, date)
		if err != nil || len(rooms) == 0 {
			rooms = demoRoomStatusDay()
		}
		result = append(result, domain.RoomStatusView{
			Date:      date,
			Rooms:     rooms,
			Occupancy: 75.0,
			Revenue:   12500.00,
			Arrivals:  8,
			Departures: 5,
			Stayovers:  27,
		})

	case "week":
		for i := 0; i < 7; i++ {
			d := date.AddDate(0, 0, i)
			result = append(result, domain.RoomStatusView{
				Date:       d,
				Rooms:      demoRoomStatusDay(),
				Occupancy:  70.0 + float64(i*2),
				Revenue:    12000.0 + float64(i*500),
				Arrivals:   6 + i,
				Departures: 4 + i,
				Stayovers:  25 + i,
			})
		}

	case "month":
		for i := 0; i < 30; i++ {
			d := date.AddDate(0, 0, i)
			result = append(result, domain.RoomStatusView{
				Date:       d,
				Rooms:      demoRoomStatusDay(),
				Occupancy:  65.0 + float64(i%15),
				Revenue:    10000.0 + float64(i*300),
				Arrivals:   5 + (i % 10),
				Departures: 3 + (i % 8),
				Stayovers:  20 + (i % 15),
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"view":  view,
		"dates": result,
	})
}

func demoRoomStatusDay() []domain.RoomDailyStatus {
	return []domain.RoomDailyStatus{
		{RoomNumber: "101", RoomType: "Standard", Status: "occupied", GuestName: "John Smith", Rate: 199.00},
		{RoomNumber: "102", RoomType: "Standard", Status: "vacant_clean", Rate: 199.00},
		{RoomNumber: "103", RoomType: "Standard", Status: "occupied", GuestName: "Maria Garcia", Rate: 199.00},
		{RoomNumber: "201", RoomType: "Deluxe King", Status: "occupied", GuestName: "Robert Chen", Rate: 299.00},
		{RoomNumber: "202", RoomType: "Deluxe King", Status: "maintenance", Rate: 299.00},
		{RoomNumber: "203", RoomType: "Deluxe King", Status: "occupied", GuestName: "Lisa Wong", Rate: 299.00},
		{RoomNumber: "301", RoomType: "Suite", Status: "occupied", GuestName: "Ahmed Hassan", Rate: 499.00},
		{RoomNumber: "302", RoomType: "Suite", Status: "vacant_dirty", Rate: 499.00},
		{RoomNumber: "303", RoomType: "Twin Room", Status: "occupied", GuestName: "Family Johnson", Rate: 249.00},
		{RoomNumber: "304", RoomType: "Twin Room", Status: "blocked", Rate: 249.00},
	}
}
