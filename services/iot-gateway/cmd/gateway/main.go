package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/nexus-platform/iot-gateway/internal/bacnet"
	"github.com/nexus-platform/iot-gateway/internal/config"
	"github.com/nexus-platform/iot-gateway/internal/health"
	"github.com/nexus-platform/iot-gateway/internal/knx"
	"github.com/nexus-platform/iot-gateway/internal/metrics"
	"github.com/nexus-platform/iot-gateway/internal/modbus"
	"github.com/nexus-platform/iot-gateway/internal/registry"
	"github.com/nexus-platform/iot-gateway/internal/telemetry"
	"github.com/nexus-platform/iot-gateway/internal/zigbee"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("iot-gateway")

type ProtocolType string

const (
	ProtocolMQTT    ProtocolType = "mqtt"
	ProtocolCoAP    ProtocolType = "coap"
	ProtocolZigbee  ProtocolType = "zigbee"
	ProtocolZWave   ProtocolType = "zwave"
	ProtocolBACnet  ProtocolType = "bacnet"
	ProtocolModbus  ProtocolType = "modbus"
	ProtocolKNX     ProtocolType = "knx"
	ProtocolBLE     ProtocolType = "ble"
	ProtocolLoRaWAN ProtocolType = "lorawan"
)

type Device struct {
	ID              string                 `json:"id"`
	Protocol        ProtocolType           `json:"protocol"`
	HardwareID      string                 `json:"hardware_id"`
	DeviceType      string                 `json:"device_type"`
	Manufacturer    string                 `json:"manufacturer"`
	Model           string                 `json:"model"`
	FirmwareVersion string                 `json:"firmware_version"`
	RoomID          string                 `json:"room_id"`
	PropertyID      string                 `json:"property_id"`
	TenantID        string                 `json:"tenant_id"`
	Region          string                 `json:"region"`
	Capabilities    []Capability           `json:"capabilities"`
	State           map[string]interface{} `json:"state"`
	LastSeen        time.Time              `json:"last_seen"`
	RegisteredAt    time.Time              `json:"registered_at"`
	OTAStatus       *OTAStatus             `json:"ota_status,omitempty"`
	SecurityProfile SecurityProfile        `json:"security_profile"`
	NetworkInfo     NetworkInfo            `json:"network_info"`
}

type Capability struct {
	Name       string              `json:"name"`
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Actions    []string            `json:"actions"`
	Events     []string            `json:"events"`
}

type Property struct {
	Type     string      `json:"type"`
	Unit     string      `json:"unit,omitempty"`
	Min      *float64    `json:"min,omitempty"`
	Max      *float64    `json:"max,omitempty"`
	ReadOnly bool        `json:"read_only"`
	Value    interface{} `json:"value,omitempty"`
}

type SecurityProfile struct {
	AuthMethod    string    `json:"auth_method"`
	CertificateID string    `json:"certificate_id,omitempty"`
	KeyRotation   time.Time `json:"key_rotation"`
	TrustZone     string    `json:"trust_zone"`
}

type NetworkInfo struct {
	IPAddress      string `json:"ip_address,omitempty"`
	MACAddress     string `json:"mac_address,omitempty"`
	VLANID         int    `json:"vlan_id,omitempty"`
	SignalStrength int    `json:"signal_strength,omitempty"`
	GatewayID      string `json:"gateway_id"`
}

type OTAStatus struct {
	CurrentVersion string    `json:"current_version"`
	TargetVersion  string    `json:"target_version"`
	Status         string    `json:"status"`
	Progress       int       `json:"progress"`
	LastAttempt    time.Time `json:"last_attempt"`
}

type IoTGateway struct {
	config             *config.Config
	mqttClient         mqtt.Client
	deviceRegistry     *registry.DeviceRegistry
	telemetryCollector *telemetry.Collector
	protocolHandlers   map[ProtocolType]ProtocolHandler
	mu                 sync.RWMutex
	activeDevices      map[string]*DeviceSession
	cancelFunc         context.CancelFunc
	wg                 sync.WaitGroup
	metrics            *metrics.Collector
	healthSrv          *health.Server
}

type DeviceSession struct {
	Device    *Device
	LastPing  time.Time
	Handler   ProtocolHandler
	ControlCh chan ControlMessage
}

type ControlMessage struct {
	Action    string                 `json:"action"`
	Params    map[string]interface{} `json:"params"`
	Timestamp time.Time              `json:"timestamp"`
}

type ProtocolHandler interface {
	Start(ctx context.Context) error
	Stop() error
	RegisterDevice(deviceID string, props map[string]interface{}) error
	SendCommand(deviceID string, action string, params map[string]interface{}) error
	GetProtocol() string
}

func NewIoTGateway(cfg *config.Config) (*IoTGateway, error) {
	reg, err := registry.New(cfg.Registry)
	if err != nil {
		return nil, fmt.Errorf("registry init failed: %w", err)
	}

	tel := telemetry.New(cfg.Telemetry)

	mqttOpts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTT.BrokerURL).
		SetClientID(fmt.Sprintf("nexus-iot-gateway-%s", uuid.New().String()[:8])).
		SetUsername(cfg.MQTT.Username).
		SetPassword(cfg.MQTT.Password).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(5 * time.Second).
		SetKeepAlive(30 * time.Second).
		SetCleanSession(false).
		SetOrderMatters(false).
		SetDefaultPublishHandler(messageHandler).
		SetOnConnectHandler(connectHandler).
		SetConnectionLostHandler(connectionLostHandler)

	if cfg.MQTT.TLS.Enabled {
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS13,
		}
		if cfg.MQTT.TLS.CAFile != "" {
			caCert, err := os.ReadFile(cfg.MQTT.TLS.CAFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA file: %w", err)
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig.RootCAs = caCertPool
		}
		if cfg.MQTT.TLS.CertFile != "" && cfg.MQTT.TLS.KeyFile != "" {
			cert, err := tls.LoadX509KeyPair(cfg.MQTT.TLS.CertFile, cfg.MQTT.TLS.KeyFile)
			if err != nil {
				return nil, fmt.Errorf("failed to load client cert: %w", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
		mqttOpts.SetTLSConfig(tlsConfig)
	}

	mqttClient := mqtt.NewClient(mqttOpts)

	metricsCollector := metrics.NewCollector(metrics.Config{Enabled: true, Port: cfg.MetricsPort})
	healthServer := health.NewServer(cfg.HealthPort)

	gw := &IoTGateway{
		config:             cfg,
		mqttClient:         mqttClient,
		deviceRegistry:     reg,
		telemetryCollector: tel,
		protocolHandlers:   make(map[ProtocolType]ProtocolHandler),
		activeDevices:      make(map[string]*DeviceSession),
		metrics:            metricsCollector,
		healthSrv:          healthServer,
	}

	gw.protocolHandlers[ProtocolZigbee] = zigbee.NewHandler(cfg.Zigbee)
	gw.protocolHandlers[ProtocolBACnet] = bacnet.NewHandler(cfg.BACnet)
	gw.protocolHandlers[ProtocolModbus] = modbus.NewHandler(cfg.Modbus)
	gw.protocolHandlers[ProtocolKNX] = knx.NewHandler(cfg.KNX)

	return gw, nil
}

func (gw *IoTGateway) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	gw.cancelFunc = cancel

	if err := gw.healthSrv.Start(); err != nil {
		return fmt.Errorf("health server start failed: %w", err)
	}

	if token := gw.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("mqtt connection failed: %w", token.Error())
	}

	controlTopics := []string{
		"nexus/commands/+/+/+",
		"nexus/ota/+/+",
		"nexus/config/+/+",
	}
	for _, topic := range controlTopics {
		if token := gw.mqttClient.Subscribe(topic, 2, gw.controlMessageHandler); token.Wait() && token.Error() != nil {
			log.Printf("Failed to subscribe to %s: %v", topic, token.Error())
		}
	}

	for proto, handler := range gw.protocolHandlers {
		gw.wg.Add(1)
		go func(p ProtocolType, h ProtocolHandler) {
			defer gw.wg.Done()
			if err := h.Start(ctx); err != nil {
				log.Printf("Protocol handler %s failed: %v", p, err)
			}
		}(proto, handler)
	}

	gw.wg.Add(1)
	go gw.heartbeatMonitor(ctx)
	gw.wg.Add(1)
	go gw.telemetryExporter(ctx)
	gw.wg.Add(1)
	go gw.otaManager(ctx)

	log.Println("IoT Gateway started successfully")
	return nil
}

func (gw *IoTGateway) controlMessageHandler(client mqtt.Client, msg mqtt.Message) {
	ctx, span := tracer.Start(context.Background(), "control-message",
		trace.WithAttributes(attribute.String("topic", msg.Topic())))
	defer span.End()

	var cmd struct {
		DeviceID string                 `json:"device_id"`
		Action   string                 `json:"action"`
		Params   map[string]interface{} `json:"params"`
		TenantID string                 `json:"tenant_id"`
	}

	if err := json.Unmarshal(msg.Payload(), &cmd); err != nil {
		log.Printf("Failed to unmarshal control message: %v", err)
		return
	}

	span.SetAttributes(
		attribute.String("device.id", cmd.DeviceID),
		attribute.String("action", cmd.Action),
	)

	gw.mu.RLock()
	session, exists := gw.activeDevices[cmd.DeviceID]
	gw.mu.RUnlock()

	if !exists {
		log.Printf("Device %s not found for command", cmd.DeviceID)
		return
	}

	if err := session.Handler.SendCommand(cmd.DeviceID, cmd.Action, cmd.Params); err != nil {
		log.Printf("Failed to send command to %s: %v", cmd.DeviceID, err)
		gw.telemetryCollector.RecordError(cmd.TenantID, "command_delivery_failed")
		return
	}

	gw.telemetryCollector.RecordCommand(cmd.TenantID, cmd.DeviceID, cmd.Action)
	gw.metrics.IncrementCounter("iot_commands_sent_total", cmd.TenantID, session.Device.DeviceType)
	log.Printf("Command %s sent to device %s", cmd.Action, cmd.DeviceID)
}

func (gw *IoTGateway) heartbeatMonitor(ctx context.Context) {
	defer gw.wg.Done()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			gw.mu.Lock()
			for id, session := range gw.activeDevices {
				if now.Sub(session.LastPing) > 5*time.Minute {
					log.Printf("Device %s timed out, removing", id)
					delete(gw.activeDevices, id)
					gw.deviceRegistry.UpdateStatus(id, "offline")
					gw.metrics.GaugeSet("iot_active_devices", float64(len(gw.activeDevices)), session.Device.TenantID)
				}
			}
			gw.mu.Unlock()
		}
	}
}

func (gw *IoTGateway) telemetryExporter(ctx context.Context) {
	defer gw.wg.Done()
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			gw.telemetryCollector.Flush()
			return
		case <-ticker.C:
			if err := gw.telemetryCollector.Flush(); err != nil {
				log.Printf("Telemetry flush failed: %v", err)
			} else {
				gw.metrics.IncrementCounter("iot_telemetry_batches_total", "aggregate", "batch")
			}
		}
	}
}

func (gw *IoTGateway) otaManager(ctx context.Context) {
	defer gw.wg.Done()
}

func (gw *IoTGateway) RegisterDevice(device *Device) error {
	ctx, span := tracer.Start(context.Background(), "register-device",
		trace.WithAttributes(
			attribute.String("device.id", device.ID),
			attribute.String("device.type", device.DeviceType),
			attribute.String("device.protocol", string(device.Protocol)),
		))
	defer span.End()

	if err := gw.validateDeviceIdentity(device); err != nil {
		return fmt.Errorf("identity validation failed: %w", err)
	}

	if err := gw.verifyTenantIsolation(device); err != nil {
		return fmt.Errorf("tenant isolation violation: %w", err)
	}

	device.RegisteredAt = time.Now()
	device.LastSeen = time.Now()
	if err := gw.deviceRegistry.Register(ctx, device); err != nil {
		return fmt.Errorf("registry registration failed: %w", err)
	}

	handler, exists := gw.protocolHandlers[device.Protocol]
	if !exists {
		return fmt.Errorf("unsupported protocol: %s", device.Protocol)
	}

	if err := handler.RegisterDevice(device.ID, map[string]interface{}{
		"hardware_id":  device.HardwareID,
		"device_type":  device.DeviceType,
		"protocol":     string(device.Protocol),
		"room_id":      device.RoomID,
		"property_id":  device.PropertyID,
		"tenant_id":    device.TenantID,
		"manufacturer": device.Manufacturer,
		"model":        device.Model,
	}); err != nil {
		return fmt.Errorf("protocol registration failed: %w", err)
	}

	session := &DeviceSession{
		Device:    device,
		LastPing:  time.Now(),
		Handler:   handler,
		ControlCh: make(chan ControlMessage, 100),
	}

	gw.mu.Lock()
	gw.activeDevices[device.ID] = session
	activeCount := len(gw.activeDevices)
	gw.mu.Unlock()

	gw.metrics.IncrementCounter("iot_device_registrations_total", device.TenantID, string(device.Protocol))
	gw.metrics.GaugeSet("iot_active_devices", float64(activeCount), device.TenantID)

	event := map[string]interface{}{
		"event":       "device_registered",
		"device_id":   device.ID,
		"tenant_id":   device.TenantID,
		"property_id": device.PropertyID,
		"room_id":     device.RoomID,
		"timestamp":   time.Now().UTC(),
	}
	eventJSON, _ := json.Marshal(event)
	gw.mqttClient.Publish(fmt.Sprintf("nexus/events/%s/%s", device.TenantID, device.PropertyID), 1, false, eventJSON)

	log.Printf("Device %s registered (type: %s, room: %s)", device.ID, device.DeviceType, device.RoomID)
	return nil
}

func (gw *IoTGateway) validateDeviceIdentity(device *Device) error {
	if device.HardwareID == "" {
		return fmt.Errorf("hardware_id is required")
	}
	if device.SecurityProfile.AuthMethod == "certificate" && device.SecurityProfile.CertificateID == "" {
		return fmt.Errorf("certificate_id required for certificate-based authentication")
	}
	if device.SecurityProfile.AuthMethod == "" {
		return fmt.Errorf("security_profile.auth_method is required")
	}
	return nil
}

func (gw *IoTGateway) verifyTenantIsolation(device *Device) error {
	if device.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if len(device.TenantID) < 3 {
		return fmt.Errorf("invalid tenant_id format")
	}
	return nil
}

func (gw *IoTGateway) ProcessTelemetry(deviceID string, data map[string]interface{}) error {
	gw.mu.RLock()
	session, exists := gw.activeDevices[deviceID]
	gw.mu.RUnlock()

	if !exists {
		return fmt.Errorf("device %s not found", deviceID)
	}

	session.LastPing = time.Now()

	enriched := map[string]interface{}{
		"device_id":   deviceID,
		"tenant_id":   session.Device.TenantID,
		"property_id": session.Device.PropertyID,
		"room_id":     session.Device.RoomID,
		"timestamp":   time.Now().UTC(),
		"data":        data,
	}

	switch session.Device.DeviceType {
	case "thermostat":
		gw.processHVACData(session.Device, data)
	case "occupancy_sensor":
		gw.processOccupancyData(session.Device, data)
	case "light":
		gw.processLightingData(session.Device, data)
	case "lock":
		gw.processSecurityData(session.Device, data)
	}

	gw.telemetryCollector.Enqueue(session.Device.TenantID, enriched)
	session.Device.State = data
	gw.deviceRegistry.UpdateState(deviceID, data)

	return nil
}

func (gw *IoTGateway) processHVACData(device *Device, data map[string]interface{}) {
	slog.Info("processing hvac telemetry",
		slog.String("tenant_id", device.TenantID),
		slog.String("device_id", device.ID),
		slog.String("device_type", device.DeviceType),
	)
}
func (gw *IoTGateway) processOccupancyData(device *Device, data map[string]interface{}) {
	slog.Info("processing occupancy telemetry",
		slog.String("tenant_id", device.TenantID),
		slog.String("device_id", device.ID),
		slog.String("device_type", device.DeviceType),
	)
}
func (gw *IoTGateway) processLightingData(device *Device, data map[string]interface{}) {
	slog.Info("processing lighting telemetry",
		slog.String("tenant_id", device.TenantID),
		slog.String("device_id", device.ID),
		slog.String("device_type", device.DeviceType),
	)
}
func (gw *IoTGateway) processSecurityData(device *Device, data map[string]interface{}) {
	slog.Info("processing security telemetry",
		slog.String("tenant_id", device.TenantID),
		slog.String("device_id", device.ID),
		slog.String("device_type", device.DeviceType),
	)
}

func (gw *IoTGateway) Shutdown() error {
	log.Println("Shutting down IoT Gateway...")
	gw.healthSrv.SetReady(false)
	gw.cancelFunc()
	gw.mqttClient.Disconnect(5000)

	for _, handler := range gw.protocolHandlers {
		handler.Stop()
	}

	gw.wg.Wait()
	gw.healthSrv.Stop()
	_ = gw.metrics.Shutdown(context.Background())
	log.Println("IoT Gateway shutdown complete")
	return nil
}

var messageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received message on topic: %s", msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("Connected to MQTT broker")
}

var connectionLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("Connection lost: %v", err)
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	gw, err := NewIoTGateway(cfg)
	if err != nil {
		log.Fatalf("Failed to create gateway: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := gw.Start(ctx); err != nil {
		log.Fatalf("Failed to start gateway: %v", err)
	}

	<-ctx.Done()
	stop()

	if err := gw.Shutdown(); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
}
