// Package domain defines property and room inventory aggregates.
package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ==================== PROPERTY ====================

type PropertyID string

type Property struct {
	ID            PropertyID
	TenantID      string
	Name          string
	Slug          string
	Address       map[string]string
	Timezone      string
	CurrencyCode  string
	ContactEmail  string
	ContactPhone  string
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Version       int
}

func NewProperty(tenantID, name, slug, timezone, currency string) *Property {
	now := time.Now().UTC()
	return &Property{
		ID:           PropertyID(uuid.Must(uuid.NewRandom()).String()),
		TenantID:     tenantID,
		Name:         name,
		Slug:         slug,
		Timezone:     timezone,
		CurrencyCode: currency,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
		Version:      1,
	}
}

// ==================== ROOM TYPE ====================

type RoomTypeID string

type RoomType struct {
	ID            RoomTypeID
	TenantID      string
	PropertyID    string
	Code          string
	Name          string
	Description   string
	BaseOccupancy int
	MaxOccupancy  int
	Amenities     []string
	Images        []string
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Version       int
}

func NewRoomType(tenantID, propertyID, code, name string, baseOcc, maxOcc int) *RoomType {
	now := time.Now().UTC()
	return &RoomType{
		ID:            RoomTypeID(uuid.Must(uuid.NewRandom()).String()),
		TenantID:      tenantID,
		PropertyID:    propertyID,
		Code:          code,
		Name:          name,
		BaseOccupancy: baseOcc,
		MaxOccupancy:  maxOcc,
		IsActive:      true,
		CreatedAt:     now,
		UpdatedAt:     now,
		Version:       1,
	}
}

// ==================== ROOM ====================

type RoomID string

type RoomStatus string

const (
	RoomStatusAvailable   RoomStatus = "available"
	RoomStatusOccupied    RoomStatus = "occupied"
	RoomStatusMaintenance RoomStatus = "maintenance"
	RoomStatusOutOfOrder  RoomStatus = "out_of_order"
)

type HousekeepingStatus string

const (
	HousekeepingClean      HousekeepingStatus = "clean"
	HousekeepingDirty      HousekeepingStatus = "dirty"
	HousekeepingInProgress HousekeepingStatus = "in_progress"
	HousekeepingInspected  HousekeepingStatus = "inspected"
)

type Room struct {
	ID                 RoomID
	TenantID           string
	PropertyID         string
	RoomTypeID         string
	RoomNumber         string
	Floor              string
	Status             RoomStatus
	IsSmoking          bool
	HasAC              bool
	HousekeepingStatus HousekeepingStatus
	Attributes         []string
	SmartLockDeviceID  string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Version            int
}

func NewRoom(tenantID, propertyID, roomTypeID, roomNumber string) *Room {
	now := time.Now().UTC()
	return &Room{
		ID:                 RoomID(uuid.Must(uuid.NewRandom()).String()),
		TenantID:           tenantID,
		PropertyID:         propertyID,
		RoomTypeID:         roomTypeID,
		RoomNumber:         roomNumber,
		Status:             RoomStatusAvailable,
		HasAC:              true,
		HousekeepingStatus: HousekeepingClean,
		CreatedAt:          now,
		UpdatedAt:          now,
		Version:            1,
	}
}

// Assign changes room status to occupied. Idempotent if already occupied.
func (r *Room) Assign() error {
	if r.Status == RoomStatusMaintenance || r.Status == RoomStatusOutOfOrder {
		return fmt.Errorf("cannot assign room %s: status is %s", r.RoomNumber, r.Status)
	}
	r.Status = RoomStatusOccupied
	r.bumpVersion()
	return nil
}

// Vacate changes room status to available.
func (r *Room) Vacate() error {
	if r.Status != RoomStatusOccupied {
		return fmt.Errorf("cannot vacate room %s: not occupied (status=%s)", r.RoomNumber, r.Status)
	}
	r.Status = RoomStatusAvailable
	r.HousekeepingStatus = HousekeepingDirty
	r.bumpVersion()
	return nil
}

// SetMaintenance marks room as under maintenance.
func (r *Room) SetMaintenance() {
	r.Status = RoomStatusMaintenance
	r.bumpVersion()
}

// SetOutOfOrder marks room as out of order.
func (r *Room) SetOutOfOrder() {
	r.Status = RoomStatusOutOfOrder
	r.bumpVersion()
}

func (r *Room) bumpVersion() {
	r.UpdatedAt = time.Now().UTC()
	r.Version++
}

// ==================== RATE PLAN ====================

type RatePlanID string

type RatePlan struct {
	ID                  RatePlanID
	TenantID            string
	PropertyID          string
	RoomTypeID          string
	Code                string
	Name                string
	Description         string
	BaseRate            float64
	CurrencyCode        string
	MinLOS              int
	MaxLOS              int
	AdvanceBookingDays  int
	IsActive            bool
	Restrictions        map[string]interface{}
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Version             int
}

func NewRatePlan(tenantID, propertyID, roomTypeID, code, name string, baseRate float64, currency string) *RatePlan {
	now := time.Now().UTC()
	return &RatePlan{
		ID:           RatePlanID(uuid.Must(uuid.NewRandom()).String()),
		TenantID:     tenantID,
		PropertyID:   propertyID,
		RoomTypeID:   roomTypeID,
		Code:         code,
		Name:         name,
		BaseRate:     baseRate,
		CurrencyCode: currency,
		MinLOS:       1,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
		Version:      1,
	}
}

func (r *RatePlan) bumpVersion() {
	r.UpdatedAt = time.Now().UTC()
	r.Version++
}

// ==================== DAILY RATE ====================

type DailyRateID string

type DailyRate struct {
	ID                DailyRateID
	TenantID          string
	RatePlanID        string
	RateDate          time.Time
	Rate              float64
	Availability      int
	MinStay           int
	CloseToArrival    bool
	CloseToDeparture  bool
	StopSell          bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func NewDailyRate(tenantID, ratePlanID string, rateDate time.Time, rate float64, availability int) *DailyRate {
	now := time.Now().UTC()
	return &DailyRate{
		ID:           DailyRateID(uuid.Must(uuid.NewRandom()).String()),
		TenantID:     tenantID,
		RatePlanID:   ratePlanID,
		RateDate:     rateDate,
		Rate:         rate,
		Availability: availability,
		MinStay:      1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
