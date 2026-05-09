// Package registry provides tenant-isolated device registration and state management.
package registry

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Device represents a registered IoT device.
type Device struct {
	ID              string                 `json:"id"`
	Protocol        string                 `json:"protocol"`
	HardwareID      string                 `json:"hardware_id"`
	DeviceType      string                 `json:"device_type"`
	Manufacturer    string                 `json:"manufacturer"`
	Model           string                 `json:"model"`
	FirmwareVersion string                 `json:"firmware_version"`
	RoomID          string                 `json:"room_id"`
	PropertyID      string                 `json:"property_id"`
	TenantID        string                 `json:"tenant_id"`
	Region          string                 `json:"region"`
	State           map[string]interface{} `json:"state"`
	Status          string                 `json:"status"`
	LastSeen        time.Time              `json:"last_seen"`
	RegisteredAt    time.Time              `json:"registered_at"`
}

// DeviceRegistry maintains an in-memory registry with tenant isolation.
type DeviceRegistry struct {
	devices map[string]*Device
	mu      sync.RWMutex
}

// New creates a device registry.
func New(cfg interface{}) (*DeviceRegistry, error) {
	return &DeviceRegistry{
		devices: make(map[string]*Device),
	}, nil
}

// Register adds a device to the registry.
func (r *DeviceRegistry) Register(ctx context.Context, device *Device) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.devices[device.ID]; exists {
		return fmt.Errorf("device %s already registered", device.ID)
	}
	device.RegisteredAt = time.Now()
	device.Status = "online"
	r.devices[device.ID] = device
	return nil
}

// UpdateStatus updates the operational status of a device.
func (r *DeviceRegistry) UpdateStatus(deviceID, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	dev, ok := r.devices[deviceID]
	if !ok {
		return fmt.Errorf("device %s not found", deviceID)
	}
	dev.Status = status
	dev.LastSeen = time.Now()
	return nil
}

// UpdateState updates the state snapshot of a device.
func (r *DeviceRegistry) UpdateState(deviceID string, state map[string]interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	dev, ok := r.devices[deviceID]
	if !ok {
		return fmt.Errorf("device %s not found", deviceID)
	}
	dev.State = state
	dev.LastSeen = time.Now()
	return nil
}

// Get retrieves a device by ID.
func (r *DeviceRegistry) Get(deviceID string) (*Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	dev, ok := r.devices[deviceID]
	if !ok {
		return nil, fmt.Errorf("device %s not found", deviceID)
	}
	return dev, nil
}

// ListByTenant returns all devices for a given tenant.
func (r *DeviceRegistry) ListByTenant(tenantID string) []*Device {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*Device
	for _, d := range r.devices {
		if d.TenantID == tenantID {
			out = append(out, d)
		}
	}
	return out
}
