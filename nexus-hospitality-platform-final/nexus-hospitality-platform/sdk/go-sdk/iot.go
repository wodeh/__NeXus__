// sdk/go-sdk/iot.go
package nexus

import (
	"context"
	"fmt"
)

type IoTService struct {
	client *Client
}

func (s *IoTService) RegisterDevice(ctx context.Context, req *DeviceRegistrationRequest) (*Device, *Response, error) {
	r, err := s.client.newRequest(ctx, "POST", "/iot/devices", req)
	if err != nil {
		return nil, nil, err
	}

	var device Device
	resp, err := s.client.do(r, &device)
	if err != nil {
		return nil, resp, err
	}

	return &device, resp, nil
}

func (s *IoTService) GetDevice(ctx context.Context, deviceID string) (*Device, *Response, error) {
	req, err := s.client.newRequest(ctx, "GET", fmt.Sprintf("/iot/devices/%s", deviceID), nil)
	if err != nil {
		return nil, nil, err
	}

	var device Device
	resp, err := s.client.do(req, &device)
	if err != nil {
		return nil, resp, err
	}

	return &device, resp, nil
}

func (s *IoTService) SendCommand(ctx context.Context, deviceID string, command *DeviceCommand) (*CommandResult, *Response, error) {
	r, err := s.client.newRequest(ctx, "POST", fmt.Sprintf("/iot/devices/%s/commands", deviceID), command)
	if err != nil {
		return nil, nil, err
	}

	var result CommandResult
	resp, err := s.client.do(r, &result)
	if err != nil {
		return nil, resp, err
	}

	return &result, resp, nil
}

func (s *IoTService) GetTelemetry(ctx context.Context, deviceID string, params *TelemetryQuery) (*TelemetryData, *Response, error) {
	path := fmt.Sprintf("/iot/devices/%s/telemetry", deviceID)
	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, nil, err
	}

	var data TelemetryData
	resp, err := s.client.do(req, &data)
	if err != nil {
		return nil, resp, err
	}

	return &data, resp, nil
}

type Device struct {
	ID              string                 `json:"id"`
	Protocol        string                 `json:"protocol"`
	HardwareID      string                 `json:"hardware_id"`
	DeviceType      string                 `json:"device_type"`
	Manufacturer    string                 `json:"manufacturer"`
	Model           string                 `json:"model"`
	RoomID          string                 `json:"room_id"`
	PropertyID      string                 `json:"property_id"`
	Capabilities    []string               `json:"capabilities"`
	State           map[string]interface{} `json:"state"`
	Status          string                 `json:"status"`
	LastSeen        string                 `json:"last_seen"`
}

type DeviceRegistrationRequest struct {
	Protocol        string                 `json:"protocol"`
	HardwareID      string                 `json:"hardware_id"`
	DeviceType      string                 `json:"device_type"`
	Manufacturer    string                 `json:"manufacturer"`
	Model           string                 `json:"model"`
	RoomID          string                 `json:"room_id"`
	PropertyID      string                 `json:"property_id"`
	Capabilities    []string               `json:"capabilities"`
}

type DeviceCommand struct {
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params,omitempty"`
}

type CommandResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	NewState  map[string]interface{} `json:"new_state,omitempty"`
}

type TelemetryQuery struct {
	StartTime string `json:"start_time,omitempty"`
	EndTime   string `json:"end_time,omitempty"`
	Interval  string `json:"interval,omitempty"` // 1m, 5m, 1h, 1d
}

type TelemetryData struct {
	DeviceID    string                   `json:"device_id"`
	DataPoints  []map[string]interface{} `json:"data_points"`
}
