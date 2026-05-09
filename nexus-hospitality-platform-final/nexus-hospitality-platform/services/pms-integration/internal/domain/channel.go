// Package domain defines channel manager aggregates for OTA integrations.
package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ==================== CHANNEL INTEGRATION ====================

type ChannelIntegrationID string

type ChannelType string

const (
	ChannelBookingCom ChannelType = "booking_com"
	ChannelExpedia    ChannelType = "expedia"
	ChannelAirbnb     ChannelType = "airbnb"
	ChannelAgoda      ChannelType = "agoda"
	ChannelTripCom    ChannelType = "trip_com"
	ChannelDirect     ChannelType = "direct"
)

type ChannelStatus string

const (
	ChannelStatusActive    ChannelStatus = "active"
	ChannelStatusInactive  ChannelStatus = "inactive"
	ChannelStatusError     ChannelStatus = "error"
	ChannelStatusSyncing   ChannelStatus = "syncing"
)

type ChannelIntegration struct {
	ID             ChannelIntegrationID
	TenantID       string
	PropertyID     string
	ChannelType    ChannelType
	Name           string
	APIKey         string // encrypted
	APISecret      string // encrypted
	HotelID        string // OTA-side hotel identifier
	RatePlanMap    map[string]string // local rate_plan_id -> OTA rate plan ID
	RoomTypeMap    map[string]string // local room_type_id -> OTA room type ID
	CommissionPct  float64
	IsActive       bool
	Status         ChannelStatus
	LastSyncAt     *time.Time
	LastError      string
	LastErrorAt    *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewChannelIntegration(tenantID, propertyID string, channelType ChannelType, name string) *ChannelIntegration {
	now := time.Now().UTC()
	return &ChannelIntegration{
		ID:          ChannelIntegrationID(uuid.Must(uuid.NewRandom()).String()),
		TenantID:    tenantID,
		PropertyID:  propertyID,
		ChannelType: channelType,
		Name:        name,
		RatePlanMap: make(map[string]string),
		RoomTypeMap: make(map[string]string),
		CommissionPct: 15.0, // default OTA commission
		IsActive:    true,
		Status:      ChannelStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// ==================== CHANNEL RESERVATION ====================

type ChannelReservationID string

type ChannelReservation struct {
	ID                ChannelReservationID
	TenantID          string
	ChannelIntegrationID string
	OTAConfirmationNumber string
	ReservationID     *string // local reservation ID once mapped
	Status            string
	GuestName         string
	GuestEmail        string
	GuestPhone        string
	RoomTypeID        string
	RatePlanID        string
	CheckInDate       time.Time
	CheckOutDate      time.Time
	Adults            int
	Children          int
	TotalAmount       float64
	CurrencyCode      string
	CommissionAmount  float64
	RawData           string // original OTA payload
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func NewChannelReservation(tenantID, channelIntegrationID, otaConf string) *ChannelReservation {
	now := time.Now().UTC()
	return &ChannelReservation{
		ID:                ChannelReservationID(uuid.Must(uuid.NewRandom()).String()),
		TenantID:          tenantID,
		ChannelIntegrationID: channelIntegrationID,
		OTAConfirmationNumber: otaConf,
		Status:            "pending",
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

// ==================== AVAILABILITY PUSH LOG ====================

type AvailabilityPushLogID string

type AvailabilityPushLog struct {
	ID                AvailabilityPushLogID
	TenantID          string
	ChannelIntegrationID string
	RoomTypeID        string
	Date              time.Time
	AvailableRooms    int
	Status            string // success, failed
	ErrorMessage      string
	CreatedAt         time.Time
}

func NewAvailabilityPushLog(tenantID, channelIntegrationID, roomTypeID string, date time.Time, availableRooms int) *AvailabilityPushLog {
	return &AvailabilityPushLog{
		ID:                AvailabilityPushLogID(uuid.Must(uuid.NewRandom()).String()),
		TenantID:          tenantID,
		ChannelIntegrationID: channelIntegrationID,
		RoomTypeID:        roomTypeID,
		Date:              date,
		AvailableRooms:    availableRooms,
		Status:            "pending",
		CreatedAt:         time.Now().UTC(),
	}
}

// Commission calculation
func (c *ChannelIntegration) CalculateCommission(totalAmount float64) float64 {
	return totalAmount * (c.CommissionPct / 100.0)
}

// ValidateChannelType ensures the channel type is supported
func ValidateChannelType(ct string) (ChannelType, error) {
	switch ct {
	case "booking_com":
		return ChannelBookingCom, nil
	case "expedia":
		return ChannelExpedia, nil
	case "airbnb":
		return ChannelAirbnb, nil
	case "agoda":
		return ChannelAgoda, nil
	case "trip_com":
		return ChannelTripCom, nil
	case "direct":
		return ChannelDirect, nil
	default:
		return "", fmt.Errorf("unsupported channel type: %s", ct)
	}
}
