// Package zigbee implements a Zigbee protocol handler for the IoT gateway.
package zigbee

import (
	"context"
	"log/slog"

	"github.com/nexus-platform/iot-gateway/internal/config"
)

// Handler manages Zigbee device communication.
type Handler struct {
	cfg config.ZigbeeConfig
}

// NewHandler creates a Zigbee protocol handler.
func NewHandler(cfg config.ZigbeeConfig) *Handler {
	return &Handler{cfg: cfg}
}

// Start begins listening for Zigbee devices.
func (h *Handler) Start(ctx context.Context) error {
	if !h.cfg.Enabled {
		slog.Info("zigbee handler disabled")
		return nil
	}
	slog.Info("zigbee handler started", slog.String("port", h.cfg.SerialPort))
	<-ctx.Done()
	return ctx.Err()
}

// Stop halts the Zigbee handler.
func (h *Handler) Stop() error {
	slog.Info("zigbee handler stopped")
	return nil
}

// RegisterDevice registers a Zigbee device.
func (h *Handler) RegisterDevice(deviceID string, props map[string]interface{}) error {
	return nil
}

// SendCommand sends a command to a Zigbee device.
func (h *Handler) SendCommand(deviceID string, action string, params map[string]interface{}) error {
	return nil
}

// GetProtocol returns the protocol type.
func (h *Handler) GetProtocol() string {
	return "zigbee"
}
