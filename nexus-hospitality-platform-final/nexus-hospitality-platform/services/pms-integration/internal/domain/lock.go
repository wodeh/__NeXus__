// Package domain defines smart lock aggregates for Orbita Tech integration.
package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ==================== SMART LOCK ====================

type SmartLockID string

type KeyStatus string

const (
	KeyStatusActive    KeyStatus = "active"
	KeyStatusExpired   KeyStatus = "expired"
	KeyStatusRevoked   KeyStatus = "revoked"
	KeyStatusPending   KeyStatus = "pending"
	KeyStatusSuspended KeyStatus = "suspended"
)

type SmartLock struct {
	ID              SmartLockID
	TenantID        string
	PropertyID      string
	RoomID          string
	DeviceID        string
	DeviceModel     string
	FirmwareVersion string
	BatteryLevel    int
	IsOnline        bool
	LastSeenAt      time.Time
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Version         int
}

func NewSmartLock(tenantID, propertyID, roomID, deviceID string) *SmartLock {
	now := time.Now().UTC()
	return &SmartLock{
		ID:         SmartLockID(uuid.Must(uuid.NewRandom()).String()),
		TenantID:   tenantID,
		PropertyID: propertyID,
		RoomID:     roomID,
		DeviceID:   deviceID,
		IsOnline:   true,
		IsActive:   true,
		BatteryLevel: 100,
		CreatedAt:  now,
		UpdatedAt:  now,
		Version:    1,
	}
}

// ==================== DIGITAL KEY ====================

type DigitalKeyID string

type DigitalKey struct {
	ID             DigitalKeyID
	TenantID       string
	SmartLockID    string
	ReservationID  string
	GuestID        string
	GuestEmail     string
	GuestPhone     string
	KeyCode        string
	PinCode        string
	NfcToken       string
	BleToken       string
	Status         KeyStatus
	ValidFrom      time.Time
	ValidUntil     time.Time
	MaxUses        int
	UsesRemaining  int
	LastUsedAt     *time.Time
	IssuedBy       string
	IssuedAt       time.Time
	RevokedAt      *time.Time
	RevokedBy      string
	RevokeReason   string
	Metadata       map[string]interface{}
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Version        int
}

func NewDigitalKey(tenantID, smartLockID, reservationID, guestID string, validFrom, validUntil time.Time) *DigitalKey {
	now := time.Now().UTC()
	return &DigitalKey{
		ID:            DigitalKeyID(uuid.Must(uuid.NewRandom()).String()),
		TenantID:      tenantID,
		SmartLockID:   smartLockID,
		ReservationID: reservationID,
		GuestID:       guestID,
		Status:        KeyStatusPending,
		ValidFrom:     validFrom,
		ValidUntil:    validUntil,
		MaxUses:       999,
		UsesRemaining: 999,
		IssuedBy:      "system",
		IssuedAt:      now,
		CreatedAt:     now,
		UpdatedAt:     now,
		Version:       1,
	}
}

func (k *DigitalKey) Activate() {
	k.Status = KeyStatusActive
	k.UpdatedAt = time.Now().UTC()
}

func (k *DigitalKey) Revoke(by, reason string) {
	now := time.Now().UTC()
	k.Status = KeyStatusRevoked
	k.RevokedAt = &now
	k.RevokedBy = by
	k.RevokeReason = reason
	k.UpdatedAt = now
}

func (k *DigitalKey) RecordUse() error {
	if k.Status != KeyStatusActive {
		return fmt.Errorf("key %s is not active (status=%s)", k.ID, k.Status)
	}
	if k.UsesRemaining <= 0 {
		return fmt.Errorf("key %s has no uses remaining", k.ID)
	}
	now := time.Now().UTC()
	k.UsesRemaining--
	k.LastUsedAt = &now
	k.UpdatedAt = now
	return nil
}

func (k *DigitalKey) IsValid() bool {
	now := time.Now().UTC()
	return k.Status == KeyStatusActive &&
		now.After(k.ValidFrom) && now.Before(k.ValidUntil) &&
		k.UsesRemaining > 0
}

// ==================== LOCK EVENT ====================

type LockEventID string

type LockEventType string

const (
	LockEventTypeUnlock      LockEventType = "unlock"
	LockEventTypeLock        LockEventType = "lock"
	LockEventTypeTamper      LockEventType = "tamper"
	LockEventTypeLowBattery  LockEventType = "low_battery"
	LockEventTypeOffline     LockEventType = "offline"
	LockEventTypeOnline      LockEventType = "online"
	LockEventTypeKeyUsed      LockEventType = "key_used"
	LockEventTypeMasterUsed   LockEventType = "master_used"
	LockEventTypePinFailed    LockEventType = "pin_failed"
)

type LockEvent struct {
	ID           LockEventID
	TenantID     string
	SmartLockID  string
	EventType    LockEventType
	DigitalKeyID *string
	KeyCode      string
	PinCode      string
	UserID       string
	Method       string
	Success      bool
	BatteryLevel int
	Metadata     map[string]interface{}
	OccurredAt   time.Time
	CreatedAt    time.Time
}

func NewLockEvent(tenantID, smartLockID string, eventType LockEventType) *LockEvent {
	now := time.Now().UTC()
	return &LockEvent{
		ID:          LockEventID(uuid.Must(uuid.NewRandom()).String()),
		TenantID:    tenantID,
		SmartLockID: smartLockID,
		EventType:   eventType,
		OccurredAt:  now,
		CreatedAt:   now,
	}
}
