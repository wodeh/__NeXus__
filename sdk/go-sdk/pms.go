// sdk/go-sdk/pms.go
package nexus

import (
	"context"
	"fmt"
	"time"
)

type PMSService struct {
	client *Client
}

// Guest operations
func (s *PMSService) GetGuest(ctx context.Context, guestID string) (*Guest, *Response, error) {
	req, err := s.client.newRequest(ctx, "GET", fmt.Sprintf("/guests/%s", guestID), nil)
	if err != nil {
		return nil, nil, err
	}

	var guest Guest
	resp, err := s.client.do(req, &guest)
	if err != nil {
		return nil, resp, err
	}

	return &guest, resp, nil
}

func (s *PMSService) CreateGuest(ctx context.Context, guest *GuestCreateRequest) (*Guest, *Response, error) {
	req, err := s.client.newRequest(ctx, "POST", "/guests", guest)
	if err != nil {
		return nil, nil, err
	}

	var created Guest
	resp, err := s.client.do(req, &created)
	if err != nil {
		return nil, resp, err
	}

	return &created, resp, nil
}

func (s *PMSService) UpdateGuest(ctx context.Context, guestID string, guest *GuestUpdateRequest) (*Guest, *Response, error) {
	req, err := s.client.newRequest(ctx, "PATCH", fmt.Sprintf("/guests/%s", guestID), guest)
	if err != nil {
		return nil, nil, err
	}

	var updated Guest
	resp, err := s.client.do(req, &updated)
	if err != nil {
		return nil, resp, err
	}

	return &updated, resp, nil
}

// Reservation operations
func (s *PMSService) GetReservation(ctx context.Context, reservationID string) (*Reservation, *Response, error) {
	req, err := s.client.newRequest(ctx, "GET", fmt.Sprintf("/reservations/%s", reservationID), nil)
	if err != nil {
		return nil, nil, err
	}

	var reservation Reservation
	resp, err := s.client.do(req, &reservation)
	if err != nil {
		return nil, resp, err
	}

	return &reservation, resp, nil
}

func (s *PMSService) CreateReservation(ctx context.Context, req *ReservationCreateRequest) (*Reservation, *Response, error) {
	r, err := s.client.newRequest(ctx, "POST", "/reservations", req)
	if err != nil {
		return nil, nil, err
	}

	var reservation Reservation
	resp, err := s.client.do(r, &reservation)
	if err != nil {
		return nil, resp, err
	}

	return &reservation, resp, nil
}

func (s *PMSService) CheckIn(ctx context.Context, reservationID string, req *CheckInRequest) (*CheckInResult, *Response, error) {
	r, err := s.client.newRequest(ctx, "POST", fmt.Sprintf("/reservations/%s/checkin", reservationID), req)
	if err != nil {
		return nil, nil, err
	}

	var result CheckInResult
	resp, err := s.client.do(r, &result)
	if err != nil {
		return nil, resp, err
	}

	return &result, resp, nil
}

func (s *PMSService) CheckOut(ctx context.Context, reservationID string, req *CheckOutRequest) (*CheckOutResult, *Response, error) {
	r, err := s.client.newRequest(ctx, "POST", fmt.Sprintf("/reservations/%s/checkout", reservationID), req)
	if err != nil {
		return nil, nil, err
	}

	var result CheckOutResult
	resp, err := s.client.do(r, &result)
	if err != nil {
		return nil, resp, err
	}

	return &result, resp, nil
}

// Room operations
func (s *PMSService) ListRooms(ctx context.Context, propertyID string, opts *ListOptions) ([]*Room, *ListMeta, *Response, error) {
	path := fmt.Sprintf("/properties/%s/rooms", propertyID)
	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, nil, nil, err
	}

	var result struct {
		Data []*Room  `json:"data"`
		Meta ListMeta `json:"meta"`
	}
	resp, err := s.client.do(req, &result)
	if err != nil {
		return nil, nil, resp, err
	}

	return result.Data, &result.Meta, resp, nil
}

func (s *PMSService) UpdateRoomStatus(ctx context.Context, roomID string, status string) (*Room, *Response, error) {
	req, err := s.client.newRequest(ctx, "PATCH", fmt.Sprintf("/rooms/%s", roomID), map[string]string{"status": status})
	if err != nil {
		return nil, nil, err
	}

	var room Room
	resp, err := s.client.do(req, &room)
	if err != nil {
		return nil, resp, err
	}

	return &room, resp, nil
}

// Housekeeping
func (s *PMSService) CreateHousekeepingTask(ctx context.Context, req *HousekeepingTaskRequest) (*HousekeepingTask, *Response, error) {
	r, err := s.client.newRequest(ctx, "POST", "/housekeeping/tasks", req)
	if err != nil {
		return nil, nil, err
	}

	var task HousekeepingTask
	resp, err := s.client.do(r, &task)
	if err != nil {
		return nil, resp, err
	}

	return &task, resp, nil
}

func (s *PMSService) CompleteHousekeepingTask(ctx context.Context, taskID string, req *HousekeepingCompletionRequest) (*HousekeepingTask, *Response, error) {
	r, err := s.client.newRequest(ctx, "POST", fmt.Sprintf("/housekeeping/tasks/%s/complete", taskID), req)
	if err != nil {
		return nil, nil, err
	}

	var task HousekeepingTask
	resp, err := s.client.do(r, &task)
	if err != nil {
		return nil, resp, err
	}

	return &task, resp, nil
}

// Maintenance
func (s *PMSService) CreateMaintenanceRequest(ctx context.Context, req *MaintenanceRequest) (*MaintenanceTicket, *Response, error) {
	r, err := s.client.newRequest(ctx, "POST", "/maintenance/requests", req)
	if err != nil {
		return nil, nil, err
	}

	var ticket MaintenanceTicket
	resp, err := s.client.do(r, &ticket)
	if err != nil {
		return nil, resp, err
	}

	return &ticket, resp, nil
}

// Types
type Guest struct {
	ID              string    `json:"id"`
	FirstName       string    `json:"first_name"`
	LastName        string    `json:"last_name"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	DateOfBirth     *string   `json:"date_of_birth,omitempty"`
	Nationality     string    `json:"nationality,omitempty"`
	Language        string    `json:"language,omitempty"`
	Preferences     map[string]interface{} `json:"preferences,omitempty"`
	LoyaltyMemberID *string   `json:"loyalty_member_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type GuestCreateRequest struct {
	FirstName   string                 `json:"first_name"`
	LastName    string                 `json:"last_name"`
	Email       string                 `json:"email,omitempty"`
	Phone       string                 `json:"phone,omitempty"`
	Nationality string                 `json:"nationality,omitempty"`
	Language    string                 `json:"language,omitempty"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
}

type GuestUpdateRequest struct {
	FirstName   string                 `json:"first_name,omitempty"`
	LastName    string                 `json:"last_name,omitempty"`
	Email       string                 `json:"email,omitempty"`
	Phone       string                 `json:"phone,omitempty"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
}

type Reservation struct {
	ID               string    `json:"id"`
	ConfirmationNumber string  `json:"confirmation_number"`
	GuestID          string    `json:"guest_id"`
	RoomID           *string   `json:"room_id,omitempty"`
	PropertyID       string    `json:"property_id"`
	Status           string    `json:"status"`
	CheckInDate      string    `json:"check_in_date"`
	CheckOutDate     string    `json:"check_out_date"`
	NumAdults        int       `json:"num_adults"`
	NumChildren      int       `json:"num_children"`
	TotalAmount      float64   `json:"total_amount"`
	CurrencyCode     string    `json:"currency_code"`
	BookingSource    string    `json:"booking_source"`
	SpecialRequests  string    `json:"special_requests,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type ReservationCreateRequest struct {
	GuestID         string  `json:"guest_id"`
	PropertyID      string  `json:"property_id"`
	CheckInDate     string  `json:"check_in_date"`
	CheckOutDate    string  `json:"check_out_date"`
	NumAdults       int     `json:"num_adults"`
	NumChildren     int     `json:"num_children,omitempty"`
	RoomTypeID      string  `json:"room_type_id,omitempty"`
	RatePlanID      string  `json:"rate_plan_id,omitempty"`
	SpecialRequests string  `json:"special_requests,omitempty"`
}

type CheckInRequest struct {
	RoomID      string   `json:"room_id,omitempty"`
	KeyCardID   string   `json:"key_card_id,omitempty"`
	Preferences []string `json:"preferences,omitempty"`
}

type CheckInResult struct {
	ReservationID string    `json:"reservation_id"`
	RoomID        string    `json:"room_id"`
	KeyCardID     string    `json:"key_card_id"`
	CheckedInAt   time.Time `json:"checked_in_at"`
}

type CheckOutRequest struct {
	FolioCharges []map[string]interface{} `json:"folio_charges,omitempty"`
}

type CheckOutResult struct {
	ReservationID string    `json:"reservation_id"`
	CheckedOutAt  time.Time `json:"checked_out_at"`
	Balance       float64   `json:"balance"`
}

type Room struct {
	ID              string    `json:"id"`
	RoomNumber      string    `json:"room_number"`
	PropertyID      string    `json:"property_id"`
	RoomTypeID      string    `json:"room_type_id"`
	Status          string    `json:"status"`
	HousekeepingStatus string `json:"housekeeping_status,omitempty"`
	Floor           int       `json:"floor,omitempty"`
	IsAccessible    bool      `json:"is_accessible,omitempty"`
	IPTVDeviceID    string    `json:"iptv_device_id,omitempty"`
	IoTGatewayID    string    `json:"iot_gateway_id,omitempty"`
}

type HousekeepingTask struct {
	ID           string    `json:"id"`
	RoomID       string    `json:"room_id"`
	TaskType     string    `json:"task_type"`
	Priority     string    `json:"priority"`
	Status       string    `json:"status"`
	AssignedTo   *string   `json:"assigned_to,omitempty"`
	ScheduledAt  time.Time `json:"scheduled_at"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	DurationMins *int      `json:"duration_minutes,omitempty"`
	Score        *int      `json:"score,omitempty"`
}

type HousekeepingTaskRequest struct {
	RoomID      string   `json:"room_id"`
	TaskType    string   `json:"task_type"`
	Priority    string   `json:"priority,omitempty"`
	AssignedTo  string   `json:"assigned_to,omitempty"`
	ScheduledAt time.Time `json:"scheduled_at"`
	Checklist   []string `json:"checklist,omitempty"`
}

type HousekeepingCompletionRequest struct {
	CompletedItems []string `json:"completed_items"`
	Issues         []string `json:"issues,omitempty"`
	Notes          string   `json:"notes,omitempty"`
}

type MaintenanceRequest struct {
	RoomID      string `json:"room_id"`
	Category    string `json:"category"`
	Priority    string `json:"priority"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ReportedBy  string `json:"reported_by,omitempty"`
}

type MaintenanceTicket struct {
	ID           string    `json:"id"`
	RoomID       string    `json:"room_id"`
	Category     string    `json:"category"`
	Priority     string    `json:"priority"`
	Status       string    `json:"status"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	ReportedAt   time.Time `json:"reported_at"`
	AssignedTo   *string   `json:"assigned_to,omitempty"`
	ScheduledDate *string  `json:"scheduled_date,omitempty"`
}
