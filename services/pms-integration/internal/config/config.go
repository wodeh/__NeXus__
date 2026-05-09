// Package config provides structured configuration for the PMS integration service.
package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds all PMS integration configuration.
type Config struct {
	HTTPPort     string   `json:"http_port"`
	MetricsPort  string   `json:"metrics_port"`
	DatabaseURL  string   `json:"database_url"`
	EventBusURL  string   `json:"event_bus_url"`
	KafkaBrokers []string `json:"kafka_brokers"`
	KafkaTopic   string   `json:"kafka_topic"`
	KafkaGroupID string   `json:"kafka_group_id"`
	KafkaDLQTopic string  `json:"kafka_dlq_topic"`
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = os.Getenv("EVENT_BUS_URL")
		if brokers != "" {
			// Handle kafka:// prefix
			brokers = strings.TrimPrefix(brokers, "kafka://")
		}
	}

	var brokerList []string
	if brokers != "" {
		brokerList = strings.Split(brokers, ",")
	}

	return &Config{
		HTTPPort:      getEnv("HTTP_PORT", "8080"),
		MetricsPort:   getEnv("METRICS_PORT", "9090"),
		DatabaseURL:   getEnv("DATABASE_URL", ""),
		EventBusURL:   getEnv("EVENT_BUS_URL", ""),
		KafkaBrokers:  brokerList,
		KafkaTopic:    getEnv("KAFKA_TOPIC", "pms-events"),
		KafkaGroupID:  getEnv("KAFKA_GROUP_ID", "pms-integration"),
		KafkaDLQTopic: getEnv("KAFKA_DLQ_TOPIC", "pms-events-dlq"),
	}, nil
}

// Validate checks configuration completeness.
func (c *Config) Validate() error {
	if c.HTTPPort == "" {
		return fmt.Errorf("HTTP_PORT is required")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
