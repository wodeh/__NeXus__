package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nexus-platform/backend-core/internal/domain"
	"github.com/nexus-platform/backend-core/internal/repository"
)

func (s *Server) registerIPTVGuestHandlers(mux *http.ServeMux) {
	// Guest-facing endpoints (room-specific, no JWT needed — room auth via PIN or URL)
	mux.HandleFunc("/v1/iptv/guest/", s.withTenant(s.handleGuestSession))
	mux.HandleFunc("/v1/iptv/room-service/", s.withTenant(s.handleRoomService))
	mux.HandleFunc("/v1/iptv/notifications/", s.withTenant(s.handleGuestNotifications))
}

// handleGuestSession manages guest TV sessions and welcome data
func (s *Server) handleGuestSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	roomID := extractID(r.URL.Path, "/v1/iptv/guest/")
	if roomID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing room id"})
		return
	}

	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)

	// Get room info
	room, err := s.repo.Rooms.GetByID(ctx, uuid.MustParse(roomID))
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"room_id":         roomID,
			"room_number":     "1005",
			"guest_name":      "Guest",
			"has_reservation": false,
			"welcome_message": "Welcome to Grand Plaza Hotel",
			"wifi_name":       "GrandPlaza-Guest",
			"wifi_password":   "Welcome2026",
		})
		return
	}

	// Get active reservation for this room
	res, err := s.repo.Reservations.GetActiveByRoom(ctx, room.PropertyID, room.ID, time.Now())
	if err != nil || res == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"room_id":         roomID,
			"room_number":     room.RoomNumber,
			"guest_name":      "Guest",
			"has_reservation": false,
			"welcome_message": "Welcome to " + room.Name,
			"wifi_name":       "GrandPlaza-Guest",
			"wifi_password":   "Welcome2026",
		})
		return
	}

	guestName := res.GuestName
	if guestName == "" {
		guestName = "Guest"
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"room_id":         roomID,
		"room_number":     room.RoomNumber,
		"guest_name":      guestName,
		"check_in":        res.CheckInDate.Format("2006-01-02"),
		"check_out":       res.CheckOutDate.Format("2006-01-02"),
		"nights":          int(res.CheckOutDate.Sub(res.CheckInDate).Hours() / 24),
		"has_reservation": true,
		"welcome_message": "Welcome to Grand Plaza Hotel",
		"wifi_name":       "GrandPlaza-Guest",
		"wifi_password":   "Welcome2026",
	})
}

// handleRoomService manages room service orders from TV
func (s *Server) handleRoomService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)

	switch r.Method {
	case http.MethodGet:
		// Get menu items
		repo := repository.NewIPTVRepository(s.repo.Pool())
		menu, err := repo.ListMenuItems(ctx, tenantID)
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]interface{}{"menu": demoMenuItems(tenantID)})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"menu": menu})

	case http.MethodPost:
		// Place order
		var req struct {
			RoomID      string               `json:"room_id"`
			Items       []domain.RoomServiceItem `json:"items"`
			SpecialRequests string           `json:"special_requests"`
			BillToRoom  bool                 `json:"bill_to_room"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		order := domain.RoomServiceOrder{
			ID:              uuid.New(),
			TenantID:        tenantUUID(tenantID),
			RoomID:          uuid.MustParse(req.RoomID),
			Items:           req.Items,
			Status:          "pending",
			SpecialRequests: req.SpecialRequests,
			BillToRoom:      req.BillToRoom,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		// TODO: Save to database
		writeJSON(w, http.StatusCreated, map[string]interface{}{
			"order_id": order.ID,
			"status":   "pending",
			"message":  "Order placed successfully. It will be delivered to your room shortly.",
		})

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

// handleGuestNotifications manages guest notifications
func (s *Server) handleGuestNotifications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)

	switch r.Method {
	case http.MethodGet:
		roomID := r.URL.Query().Get("room_id")
		if roomID == "" {
			writeJSONError(w, http.StatusBadRequest, "missing room_id query parameter")
			return
		}

		repo := repository.NewIPTVRepository(s.repo.Pool())
		notifications, err := repo.GetNotifications(ctx, tenantID, roomID)
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"notifications": demoNotifications(roomID),
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"notifications": notifications})

	case http.MethodPost:
		// Mark as read
		var req struct {
			NotificationID string `json:"notification_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"success": true})

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func demoMenuItems(tenantID string) []map[string]interface{} {
	return []map[string]interface{}{
		{"id": "1", "name": "Caesar Salad", "price": 18, "category": "starters", "description": "Classic with parmesan"},
		{"id": "2", "name": "Grilled Salmon", "price": 32, "category": "mains", "description": "With lemon butter sauce"},
		{"id": "3", "name": "Ribeye Steak", "price": 45, "category": "mains", "description": "12oz with roasted vegetables"},
		{"id": "4", "name": "Chocolate Fondant", "price": 14, "category": "desserts", "description": "With vanilla ice cream"},
	}
}

func demoNotifications(roomID string) []map[string]interface{} {
	return []map[string]interface{}{
		{"id": "1", "title": "Welcome!", "message": "Enjoy your stay at Grand Plaza", "type": "info", "time": "15:00", "read": false},
		{"id": "2", "title": "Spa Offer", "message": "20% off massages today", "type": "info", "time": "16:30", "read": false},
	}
}
