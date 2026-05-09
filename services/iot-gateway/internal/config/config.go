// Package config provides structured configuration for the IoT gateway.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all IoT gateway configuration.
type Config struct {
	HTTPPort    string          `json:"http_port"`
	MetricsPort string          `json:"metrics_port"`
	HealthPort  string          `json:"health_port"`
	MQTT        MQTTConfig      `json:"mqtt"`
	Registry    RegistryConfig  `json:"registry"`
	Telemetry   TelemetryConfig `json:"telemetry"`
	Zigbee      ZigbeeConfig    `json:"zigbee"`
	BACnet      BACnetConfig    `json:"bacnet"`
	Modbus      ModbusConfig    `json:"modbus"`
	KNX         KNXConfig       `json:"knx"`
}

// MQTTConfig configures the MQTT broker connection.
type MQTTConfig struct {
	BrokerURL string    `json:"broker_url"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	TLS       TLSConfig `json:"tls"`
}

// TLSConfig holds mTLS configuration.
type TLSConfig struct {
	Enabled  bool   `json:"enabled"`
	CAFile   string `json:"ca_file"`
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
}

// RegistryConfig configures the device registry.
type RegistryConfig struct {
	Backend string `json:"backend"` // postgres, redis, memory
	DSN     string `json:"dsn"`
}

// TelemetryConfig configures telemetry collection.
type TelemetryConfig struct {
	Backend      string        `json:"backend"`
	FlushInterval time.Duration `json:"flush_interval"`
	BatchSize    int           `json:"batch_size"`
}

// ZigbeeConfig configures Zigbee protocol handler.
type ZigbeeConfig struct {
	Enabled    bool   `json:"enabled"`
	SerialPort string `json:"serial_port"`
	BaudRate   int    `json:"baud_rate"`
}

// BACnetConfig configures BACnet protocol handler.
type BACnetConfig struct {
	Enabled bool   `json:"enabled"`
	Port    int    `json:"port"`
	Network int    `json:"network"`
}

// ModbusConfig configures Modbus protocol handler.
type ModbusConfig struct {
	Enabled  bool   `json:"enabled"`
	Mode     string `json:"mode"` // tcp, rtu
	Address  string `json:"address"`
	Port     int    `json:"port"`
}

// KNXConfig configures KNX protocol handler.
type KNXConfig struct {
	Enabled   bool   `json:"enabled"`
	GatewayIP string `json:"gateway_ip"`
	Port      int    `json:"port"`
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	return &Config{
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
		MetricsPort: getEnv("METRICS_PORT", "9090"),
		HealthPort:  getEnv("HEALTH_PORT", "8081"),
		MQTT: MQTTConfig{
			BrokerURL: getEnv("MQTT_BROKER_URL", "tcp://localhost:1883"),
			Username:  getEnv("MQTT_USERNAME", ""),
			Password:  getEnv("MQTT_PASSWORD", ""),
			TLS: TLSConfig{
				Enabled:  getEnvBool("MQTT_TLS_ENABLED", false),
				CAFile:   getEnv("MQTT_TLS_CA_FILE", ""),
				CertFile: getEnv("MQTT_TLS_CERT_FILE", ""),
				KeyFile:  getEnv("MQTT_TLS_KEY_FILE", ""),
			},
		},
		Registry: RegistryConfig{
			Backend: getEnv("REGISTRY_BACKEND", "memory"),
			DSN:     getEnv("REGISTRY_DSN", ""),
		},
		Telemetry: TelemetryConfig{
			Backend:       getEnv("TELEMETRY_BACKEND", "memory"),
			FlushInterval: time.Duration(getEnvInt("TELEMETRY_FLUSH_INTERVAL_SECONDS", 10)) * time.Second,
			BatchSize:     getEnvInt("TELEMETRY_BATCH_SIZE", 100),
		},
		Zigbee: ZigbeeConfig{
			Enabled:    getEnvBool("ZIGBEE_ENABLED", false),
			SerialPort: getEnv("ZIGBEE_SERIAL_PORT", "/dev/ttyUSB0"),
			BaudRate:   getEnvInt("ZIGBEE_BAUD_RATE", 115200),
		},
		BACnet: BACnetConfig{
			Enabled: getEnvBool("BACNET_ENABLED", false),
			Port:    getEnvInt("BACNET_PORT", 47808),
			Network: getEnvInt("BACNET_NETWORK", 0),
		},
		Modbus: ModbusConfig{
			Enabled: getEnvBool("MODBUS_ENABLED", false),
			Mode:    getEnv("MODBUS_MODE", "tcp"),
			Address: getEnv("MODBUS_ADDRESS", "localhost"),
			Port:    getEnvInt("MODBUS_PORT", 502),
		},
		KNX: KNXConfig{
			Enabled:   getEnvBool("KNX_ENABLED", false),
			GatewayIP: getEnv("KNX_GATEWAY_IP", ""),
			Port:      getEnvInt("KNX_PORT", 3671),
		},
	}, nil
}

// Validate checks configuration completeness.
func (c *Config) Validate() error {
	if c.HTTPPort == "" {
		return fmt.Errorf("HTTP_PORT is required")
	}
	if c.MQTT.BrokerURL == "" {
		return fmt.Errorf("MQTT_BROKER_URL is required")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}
