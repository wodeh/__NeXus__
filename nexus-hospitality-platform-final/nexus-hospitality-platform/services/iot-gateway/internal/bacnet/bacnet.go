// Package bacnet implements a BACnet protocol handler for the IoT gateway.
package bacnet

import (
	"context"
	"log/slog"

	"github.com/nexus-platform/iot-gateway/internal/config"
)

// Handler manages BACnet device communication.
type Handler struct {
	cfg config.BACnetConfig
}

// NewHandler creates a BACnet protocol handler.
func NewHandler(cfg config.BACnetConfig) *Handler {
	return &Handler{cfg: cfg}
}

// Start begins listening for BACnet devices.
func (h *Handler) Start(ctx context.Context) error {
	if !h.cfg.Enabled {
		slog.Info("bacnet handler disabled")
		return nil
	}
	slog.Info("bacnet handler started", slog.Int("port", h.cfg.Port))
	<-ctx.Done()
	return ctx.Err()
}

// Stop halts the BACnet handler.
func (h *Handler) Stop() error {
	slog.Info("bacnet handler stopped")
	return nil
}

// RegisterDevice registers a BACnet device.
func (h *Handler) RegisterDevice(deviceID string, props map[string]interface{}) error {
	return nil
}

// SendCommand sends a command to a BACnet device.
func (h *Handler) SendCommand(deviceID string, action string, params map[string]interface{}) error {
	return nil
}

// GetProtocol returns the protocol type.
func (h *Handler) GetProtocol() string {
	return "bacnet"
}
