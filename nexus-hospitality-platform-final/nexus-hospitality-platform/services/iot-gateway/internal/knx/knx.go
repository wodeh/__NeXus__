// Package knx implements a KNX protocol handler for the IoT gateway.
package knx

import (
	"context"
	"log/slog"

	"github.com/nexus-platform/iot-gateway/internal/config"
)

// Handler manages KNX device communication.
type Handler struct {
	cfg config.KNXConfig
}

// NewHandler creates a KNX protocol handler.
func NewHandler(cfg config.KNXConfig) *Handler {
	return &Handler{cfg: cfg}
}

// Start begins listening for KNX devices.
func (h *Handler) Start(ctx context.Context) error {
	if !h.cfg.Enabled {
		slog.Info("knx handler disabled")
		return nil
	}
	slog.Info("knx handler started", slog.String("gateway", h.cfg.GatewayIP))
	<-ctx.Done()
	return ctx.Err()
}

// Stop halts the KNX handler.
func (h *Handler) Stop() error {
	slog.Info("knx handler stopped")
	return nil
}

// RegisterDevice registers a KNX device.
func (h *Handler) RegisterDevice(deviceID string, props map[string]interface{}) error {
	return nil
}

// SendCommand sends a command to a KNX device.
func (h *Handler) SendCommand(deviceID string, action string, params map[string]interface{}) error {
	return nil
}

// GetProtocol returns the protocol type.
func (h *Handler) GetProtocol() string {
	return "knx"
}
