//go:build nvml

package gpu

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

type Manager struct {
	devices          []Device
	count            int
	memoryThreshold  float64
	mu               sync.RWMutex
	availableDevices chan int
	monitoringEnabled bool
}

type Device struct {
	Index        int
	UUID         string
	Name         string
	MemoryTotal  uint64
	MemoryUsed   uint64
	Utilization  int
	Temperature  int
	PowerDraw    float64
	LastUpdated  time.Time
}

type Stats struct {
	AvgUtilization float64
	AvgMemoryUsed  float64
	AvgTemperature float64
	AvailableCount int
}

func NewManager(count int, memoryThreshold float64) (*Manager, error) {
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		return nil, fmt.Errorf("nvml init failed: %v", nvml.ErrorString(ret))
	}

	deviceCount, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		nvml.Shutdown()
		return nil, fmt.Errorf("failed to get device count: %v", nvml.ErrorString(ret))
	}

	if count > deviceCount {
		count = deviceCount
	}

	m := &Manager{
		devices:          make([]Device, count),
		count:            count,
		memoryThreshold:  memoryThreshold,
		availableDevices: make(chan int, count),
	}

	for i := 0; i < count; i++ {
		dev, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			log.Printf("Failed to get device %d: %v", i, nvml.ErrorString(ret))
			continue
		}

		uuid, ret := dev.GetUUID()
		if ret != nvml.SUCCESS {
			log.Printf("Failed to get UUID for device %d: %v", i, nvml.ErrorString(ret))
		}

		name, ret := dev.GetName()
		if ret != nvml.SUCCESS {
			log.Printf("Failed to get name for device %d: %v", i, nvml.ErrorString(ret))
		}

		memInfo, ret := dev.GetMemoryInfo()
		if ret != nvml.SUCCESS {
			log.Printf("Failed to get memory info for device %d: %v", i, nvml.ErrorString(ret))
		}

		m.devices[i] = Device{
			Index:       i,
			UUID:        uuid,
			Name:        name,
			MemoryTotal: memInfo.Total,
		}

		m.availableDevices <- i
	}

	return m, nil
}

func (m *Manager) Acquire(ctx context.Context, priority int) (int, error) {
	select {
	case device := <-m.availableDevices:
		return device, nil
	case <-ctx.Done():
		return -1, ctx.Err()
	case <-time.After(30 * time.Second):
		return -1, fmt.Errorf("timeout waiting for GPU")
	}
}

func (m *Manager) Release(device int) {
	if device < 0 || device >= m.count {
		log.Printf("Invalid device index: %d", device)
		return
	}
	select {
	case m.availableDevices <- device:
	default:
		log.Printf("Device %d release failed: channel full", device)
	}
}

func (m *Manager) StartMonitoring(ctx context.Context) {
	m.monitoringEnabled = true
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.updateStats()
			}
		}
	}()
}

func (m *Manager) updateStats() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := 0; i < m.count; i++ {
		dev, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			continue
		}

		utilization, ret := dev.GetUtilizationRates()
		if ret == nvml.SUCCESS {
			m.devices[i].Utilization = int(utilization.Gpu)
		}

		memInfo, ret := dev.GetMemoryInfo()
		if ret == nvml.SUCCESS {
			m.devices[i].MemoryUsed = memInfo.Used
		}

		temp, ret := dev.GetTemperature(nvml.TEMPERATURE_GPU)
		if ret == nvml.SUCCESS {
			m.devices[i].Temperature = int(temp)
		}

		power, ret := dev.GetPowerUsage()
		if ret == nvml.SUCCESS {
			m.devices[i].PowerDraw = float64(power) / 1000.0
		}

		m.devices[i].LastUpdated = time.Now()
	}
}

func (m *Manager) GetStats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var totalUtil, totalMemUsed, totalTemp float64
	available := len(m.availableDevices)

	for _, dev := range m.devices {
		totalUtil += float64(dev.Utilization)
		if dev.MemoryTotal > 0 {
			totalMemUsed += float64(dev.MemoryUsed) / float64(dev.MemoryTotal) * 100
		}
		totalTemp += float64(dev.Temperature)
	}

	count := float64(m.count)
	return Stats{
		AvgUtilization: totalUtil / count,
		AvgMemoryUsed:  totalMemUsed / count,
		AvgTemperature: totalTemp / count,
		AvailableCount: available,
	}
}

func (m *Manager) Shutdown() {
	nvml.Shutdown()
	close(m.availableDevices)
}
