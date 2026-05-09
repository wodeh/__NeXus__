//go:build !nvml

// Package gpu provides a CPU-only stub implementation of the GPU manager.
// Use the nvml build tag to compile with NVIDIA NVML support:
//   go build -tags nvml ./...
package gpu

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Manager is a no-op GPU manager for environments without NVIDIA hardware.
type Manager struct {
	devices           []Device
	count             int
	memoryThreshold   float64
	mu                sync.RWMutex
	availableDevices  chan int
	monitoringEnabled bool
}

// Device represents a GPU device (stub).
type Device struct {
	Index       int
	UUID        string
	Name        string
	MemoryTotal uint64
	MemoryUsed  uint64
	Utilization int
	Temperature int
	PowerDraw   float64
	LastUpdated time.Time
}

// Stats aggregates GPU metrics (stub).
type Stats struct {
	AvgUtilization float64
	AvgMemoryUsed  float64
	AvgTemperature float64
	AvailableCount int
}

// NewManager creates a stub GPU manager. count is honored for channel sizing
// but no hardware queries are performed.
func NewManager(count int, memoryThreshold float64) (*Manager, error) {
	if count < 0 {
		count = 0
	}
	m := &Manager{
		devices:          make([]Device, count),
		count:            count,
		memoryThreshold:  memoryThreshold,
		availableDevices: make(chan int, count),
	}
	for i := 0; i < count; i++ {
		m.devices[i] = Device{Index: i, Name: "cpu-stub"}
		m.availableDevices <- i
	}
	slog.Info("gpu manager initialized in stub mode", "count", count)
	return m, nil
}

// Acquire waits for a device to become available. With no devices this always times out.
func (m *Manager) Acquire(ctx context.Context, priority int) (int, error) {
	select {
	case device := <-m.availableDevices:
		return device, nil
	case <-ctx.Done():
		return -1, ctx.Err()
	case <-time.After(30 * time.Second):
		return -1, fmt.Errorf("timeout waiting for GPU (stub mode)")
	}
}

// Release returns a device to the pool.
func (m *Manager) Release(device int) {
	if device < 0 || device >= m.count {
		slog.Warn("invalid device index released", "device", device)
		return
	}
	select {
	case m.availableDevices <- device:
	default:
		slog.Warn("device release failed: channel full", "device", device)
	}
}

// StartMonitoring begins a no-op monitoring loop.
func (m *Manager) StartMonitoring(ctx context.Context) {
	if m.count == 0 {
		return
	}
	m.monitoringEnabled = true
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// no-op: no hardware to query
			}
		}
	}()
}

// GetStats returns zero-value stats.
func (m *Manager) GetStats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	available := len(m.availableDevices)
	return Stats{
		AvgUtilization: 0,
		AvgMemoryUsed:  0,
		AvgTemperature: 0,
		AvailableCount: available,
	}
}

// Shutdown closes the available-devices channel.
func (m *Manager) Shutdown() {
	if m.availableDevices != nil {
		close(m.availableDevices)
	}
}
