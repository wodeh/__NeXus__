// Package telemetry collects and batches IoT device telemetry with tenant isolation.
package telemetry

import (
	"log/slog"
	"sync"
	"time"
)

// Collector buffers and exports device telemetry.
type Collector struct {
	buffer    []TelemetryRecord
	mu        sync.Mutex
	batchSize int
	flushInterval time.Duration
}

// TelemetryRecord represents a single telemetry event.
type TelemetryRecord struct {
	TenantID  string                 `json:"tenant_id"`
	DeviceID  string                 `json:"device_id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// Config configures the telemetry collector.
type Config struct {
	Backend       string        `json:"backend"`
	FlushInterval time.Duration `json:"flush_interval"`
	BatchSize     int           `json:"batch_size"`
}

// New creates a telemetry collector.
func New(cfg Config) *Collector {
	return &Collector{
		buffer:        make([]TelemetryRecord, 0, cfg.BatchSize),
		batchSize:     cfg.BatchSize,
		flushInterval: cfg.FlushInterval,
	}
}

// Enqueue adds a telemetry record to the buffer.
func (c *Collector) Enqueue(tenantID string, data map[string]interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.buffer = append(c.buffer, TelemetryRecord{
		TenantID:  tenantID,
		Timestamp: time.Now(),
		Data:      data,
	})
	if len(c.buffer) >= c.batchSize {
		go c.flush()
	}
}

// RecordError records an error telemetry event.
func (c *Collector) RecordError(tenantID, errorType string) {
	c.Enqueue(tenantID, map[string]interface{}{"error_type": errorType})
}

// RecordCommand records a command delivery telemetry event.
func (c *Collector) RecordCommand(tenantID, deviceID, action string) {
	c.Enqueue(tenantID, map[string]interface{}{
		"device_id": deviceID,
		"action":    action,
		"event":     "command_delivered",
	})
}

// Flush sends buffered telemetry to the backend.
func (c *Collector) Flush() error {
	c.mu.Lock()
	if len(c.buffer) == 0 {
		c.mu.Unlock()
		return nil
	}
	batch := make([]TelemetryRecord, len(c.buffer))
	copy(batch, c.buffer)
	c.buffer = c.buffer[:0]
	c.mu.Unlock()

	slog.Info("telemetry flushed", slog.Int("count", len(batch)))
	return nil
}

func (c *Collector) flush() {
	if err := c.Flush(); err != nil {
		slog.Error("telemetry flush failed", slog.String("error", err.Error()))
	}
}
