// Package config provides structured configuration for the backend core API.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all backend-core configuration.
type Config struct {
	HTTPPort     string   `json:"http_port"`
	MetricsPort  string   `json:"metrics_port"`
	DatabaseURL  string   `json:"database_url"`
	RedisURL     string   `json:"redis_url"`
	KafkaBrokers []string `json:"kafka_brokers"`
	KafkaTopic   string   `json:"kafka_topic"`
	Environment  string   `json:"environment"`
	LogLevel     string   `json:"log_level"`
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	brokers := os.Getenv("KAFKA_BROKERS")
	var brokerList []string
	if brokers != "" {
		brokerList = strings.Split(brokers, ",")
	}

	return &Config{
		HTTPPort:     getEnv("HTTP_PORT", "8080"),
		MetricsPort:  getEnv("METRICS_PORT", "9090"),
		DatabaseURL:  getEnv("DATABASE_URL", ""),
		RedisURL:     getEnv("REDIS_URL", ""),
		KafkaBrokers: brokerList,
		KafkaTopic:   getEnv("KAFKA_TOPIC", "nexus-events"),
		Environment:  getEnv("ENVIRONMENT", "development"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
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
