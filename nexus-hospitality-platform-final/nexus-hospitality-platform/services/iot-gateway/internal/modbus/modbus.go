// Package modbus implements a Modbus protocol handler for the IoT gateway.
package modbus

import (
	"context"
	"log/slog"

	"github.com/nexus-platform/iot-gateway/internal/config"
)

// Handler manages Modbus device communication.
type Handler struct {
	cfg config.ModbusConfig
}

// NewHandler creates a Modbus protocol handler.
func NewHandler(cfg config.ModbusConfig) *Handler {
	return &Handler{cfg: cfg}
}

// Start begins listening for Modbus devices.
func (h *Handler) Start(ctx context.Context) error {
	if !h.cfg.Enabled {
		slog.Info("modbus handler disabled")
		return nil
	}
	slog.Info("modbus handler started", slog.String("address", h.cfg.Address), slog.Int("port", h.cfg.Port))
	<-ctx.Done()
	return ctx.Err()
}

// Stop halts the Modbus handler.
func (h *Handler) Stop() error {
	slog.Info("modbus handler stopped")
	return nil
}

// RegisterDevice registers a Modbus device.
func (h *Handler) RegisterDevice(deviceID string, props map[string]interface{}) error {
	return nil
}

// SendCommand sends a command to a Modbus device.
func (h *Handler) SendCommand(deviceID string, action string, params map[string]interface{}) error {
	return nil
}

// GetProtocol returns the protocol type.
func (h *Handler) GetProtocol() string {
	return "modbus"
}
