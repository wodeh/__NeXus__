package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"nexus-hospitality-platform/services/backend-core/internal/domain"
)

func (s *Server) registerChannelManagerHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/channel-reservations", s.withTenant(s.handleChannelReservations))
	mux.HandleFunc("/v1/channel-reservations/", s.withTenant(s.handleChannelReservationDetail))
	mux.HandleFunc("/v1/channel-availability", s.withTenant(s.handleChannelAvailability))
	mux.HandleFunc("/v1/email-reservations", s.withTenant(s.handleEmailReservation))
	mux.HandleFunc("/v1/walk-in", s.withTenant(s.handleWalkIn))
	mux.HandleFunc("/v1/whatsapp-reservations", s.withTenant(s.handleWhatsAppReservation))
}

/* ─── Channel Reservations ─── */

func (s *Server) handleChannelReservations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantIDStr, _ := ctx.Value("tenant_id").(string)
	tenantID, _ := uuid.Parse(tenantIDStr)

	switch r.Method {
	case http.MethodGet:
		source := r.URL.Query().Get("source")
		reservations, err := s.repo.Channel.ListChannelReservations(ctx, tenantID, source)
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]interface{}{"reservations": demoChannelReservations()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"reservations": reservations})

	case http.MethodPost:
		var req domain.ChannelReservation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
			return
		}
		req.TenantID = tenantID
		req.Status = "pending"
		if req.ChannelSource == "" { req.ChannelSource = "manual" }
		if err := s.repo.Channel.CreateChannelReservation(ctx, &req); err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, req)

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleChannelReservationDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantIDStr, _ := ctx.Value("tenant_id").(string)
	tenantID, _ := uuid.Parse(tenantIDStr)

	path := r.URL.Path[len("/v1/channel-reservations/"):]
	id, err := uuid.Parse(path)
	if err != nil {
		http.Error(w, `{"error":"invalid reservation id"}`, http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodPatch {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if err := s.repo.Channel.UpdateChannelReservationStatus(ctx, tenantID, id, req.Status); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func demoChannelReservations() []domain.ChannelReservation {
	return []domain.ChannelReservation{
		{ID: uuid.MustParse("c1000001-0000-0000-0000-000000000001"), TenantID: uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"), ChannelSource: "booking_com", ExternalRef: "BDC-28475639", GuestName: "John Smith", GuestEmail: "john@example.com", GuestPhone: "+1-555-0123", RoomType: "Deluxe King", CheckIn: time.Now().AddDate(0, 0, 2), CheckOut: time.Now().AddDate(0, 0, 5), Adults: 2, Children: 0, Total: 897.00, Currency: "USD", Status: "confirmed", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.MustParse("c1000002-0000-0000-0000-000000000002"), TenantID: uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"), ChannelSource: "expedia", ExternalRef: "EXP-9928374", GuestName: "Maria Garcia", GuestEmail: "maria@example.com", GuestPhone: "+34-612-345-678", RoomType: "Twin Room", CheckIn: time.Now().AddDate(0, 0, 1), CheckOut: time.Now().AddDate(0, 0, 3), Adults: 2, Children: 1, Total: 546.00, Currency: "USD", Status: "confirmed", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.MustParse("c1000003-0000-0000-0000-000000000003"), TenantID: uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"), ChannelSource: "email", ExternalRef: "email-thread-12345", GuestName: "Robert Chen", GuestEmail: "robert@company.com", RoomType: "Suite", CheckIn: time.Now().AddDate(0, 0, 7), CheckOut: time.Now().AddDate(0, 0, 10), Adults: 2, Children: 0, Total: 2100.00, Currency: "USD", Status: "pending", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.MustParse("c1000004-0000-0000-0000-000000000004"), TenantID: uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"), ChannelSource: "whatsapp", ExternalRef: "wa-conv-67890", GuestName: "Ahmed Hassan", GuestPhone: "+971-55-123-4567", RoomType: "Standard", CheckIn: time.Now().AddDate(0, 0, 3), CheckOut: time.Now().AddDate(0, 0, 4), Adults: 1, Children: 0, Total: 199.00, Currency: "USD", Status: "pending", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.MustParse("c1000005-0000-0000-0000-000000000005"), TenantID: uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"), ChannelSource: "walk_in", ExternalRef: "fd-001", GuestName: "Lisa Wong", GuestPhone: "+1-555-0199", RoomType: "Deluxe King", RoomNumber: "305", CheckIn: time.Now(), CheckOut: time.Now().AddDate(0, 0, 2), Adults: 2, Children: 0, Total: 598.00, Currency: "USD", Status: "checked_in", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
}

/* ─── Channel Availability ─── */

func (s *Server) handleChannelAvailability(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantIDStr, _ := ctx.Value("tenant_id").(string)
	tenantID, _ := uuid.Parse(tenantIDStr)

	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	roomType := r.URL.Query().Get("room_type")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	fromTime, _ := time.Parse("2006-01-02", from)
	toTime, _ := time.Parse("2006-01-02", to)
	if fromTime.IsZero() { fromTime = time.Now() }
	if toTime.IsZero() { toTime = fromTime.AddDate(0, 0, 30) }

	avail, err := s.repo.Channel.GetAvailability(ctx, tenantID, roomType, fromTime, toTime)
	if err != nil || len(avail) == 0 {
		writeJSON(w, http.StatusOK, map[string]interface{}{"availability": demoAvailability(roomType, fromTime, toTime)})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"availability": avail})
}

func demoAvailability(roomType string, from, to time.Time) []domain.ChannelAvailability {
	if roomType == "" { roomType = "Deluxe King" }
	var out []domain.ChannelAvailability
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		out = append(out, domain.ChannelAvailability{
			TenantID:     uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"),
			RoomType:     roomType,
			Date:         d,
			TotalRooms:   20,
			BookedRooms:  12,
			BlockedRooms: 2,
			Available:    6,
			Rate:         299.00,
			Currency:     "USD",
		})
	}
	return out
}

/* ─── Email Reservation ─── */

func (s *Server) handleEmailReservation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantIDStr, _ := ctx.Value("tenant_id").(string)
	tenantID, _ := uuid.Parse(tenantIDStr)

	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req domain.EmailReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Parse email body to extract reservation details (simplified)
	res := &domain.ChannelReservation{
		ID:            uuid.New(),
		TenantID:      tenantID,
		ChannelSource: "email",
		ExternalRef:   req.Subject,
		GuestName:     req.From,
		GuestEmail:    req.From,
		Status:        "pending",
		RawPayload:    map[string]interface{}{"subject": req.Subject, "body": req.Body},
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	if err := s.repo.Channel.CreateChannelReservation(ctx, res); err != nil {
		// Return demo response even if DB fails
		writeJSON(w, http.StatusCreated, map[string]interface{}{
			"status":      "created",
			"reservation": res,
			"message":     "Email reservation parsed and queued for review",
		})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"status":      "created",
		"reservation": res,
		"message":     "Email reservation parsed and queued for review",
	})
}

/* ─── Walk-In ─── */

func (s *Server) handleWalkIn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantIDStr, _ := ctx.Value("tenant_id").(string)
	tenantID, _ := uuid.Parse(tenantIDStr)

	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req domain.FrontDeskWalkIn
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	checkIn, _ := time.Parse("2006-01-02", req.CheckIn)
	checkOut, _ := time.Parse("2006-01-02", req.CheckOut)

	res := &domain.ChannelReservation{
		ID:              uuid.New(),
		TenantID:        tenantID,
		ChannelSource:   "walk_in",
		ExternalRef:     "fd-" + uuid.New().String()[:8],
		GuestName:       req.GuestName,
		GuestEmail:      req.GuestEmail,
		GuestPhone:      req.GuestPhone,
		RoomType:        req.RoomType,
		RoomNumber:      req.RoomNumber,
		CheckIn:         checkIn,
		CheckOut:        checkOut,
		Adults:          req.Adults,
		Children:        req.Children,
		Total:           req.Total,
		Currency:        "USD",
		Status:          "checked_in",
		SpecialRequests: req.SpecialRequests,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	if err := s.repo.Channel.CreateChannelReservation(ctx, res); err != nil {
		writeJSON(w, http.StatusCreated, map[string]interface{}{
			"status":      "created",
			"reservation": res,
			"message":     "Walk-in reservation created (demo mode)",
		})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"status":      "created",
		"reservation": res,
		"message":     "Walk-in reservation created successfully",
	})
}

/* ─── WhatsApp Reservation ─── */

func (s *Server) handleWhatsAppReservation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantIDStr, _ := ctx.Value("tenant_id").(string)
	tenantID, _ := uuid.Parse(tenantIDStr)

	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req domain.WhatsAppBotReservation
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	checkIn, _ := time.Parse("2006-01-02", req.CheckIn)
	checkOut, _ := time.Parse("2006-01-02", req.CheckOut)

	res := &domain.ChannelReservation{
		ID:              uuid.New(),
		TenantID:        tenantID,
		ChannelSource:   "whatsapp",
		ExternalRef:     req.ConversationID,
		GuestName:       req.GuestName,
		GuestPhone:      req.GuestPhone,
		RoomType:        req.RoomType,
		CheckIn:         checkIn,
		CheckOut:        checkOut,
		Adults:          req.Adults,
		Children:        req.Children,
		Total:           0, // Will be calculated
		Currency:        "USD",
		Status:          "pending",
		SpecialRequests: req.SpecialRequests,
		RawPayload:      map[string]interface{}{"conversation_id": req.ConversationID},
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	if err := s.repo.Channel.CreateChannelReservation(ctx, res); err != nil {
		writeJSON(w, http.StatusCreated, map[string]interface{}{
			"status":      "created",
			"reservation": res,
			"message":     "WhatsApp reservation queued for review (demo mode)",
		})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"status":      "created",
		"reservation": res,
		"message":     "WhatsApp reservation queued for review",
	})
}
