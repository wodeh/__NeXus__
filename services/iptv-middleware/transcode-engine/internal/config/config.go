// Package config provides structured configuration for the IPTV transcode engine.
package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all transcode engine configuration.
type Config struct {
	HTTPPort    string        `json:"http_port"`
	MetricsPort string        `json:"metrics_port"`
	HealthPort  string        `json:"health_port"`
	Workers     int           `json:"workers"`
	QueueSize   int           `json:"queue_size"`
	GPU         GPUConfig     `json:"gpu"`
	Metrics     MetricsConfig `json:"metrics"`
	NVEncPreset string        `json:"nvenc_preset"`
}

// GPUConfig configures GPU acceleration.
type GPUConfig struct {
	Enabled         bool    `json:"enabled"`
	Count           int     `json:"count"`
	MemoryThreshold float64 `json:"memory_threshold"`
}

// MetricsConfig configures observability metrics.
type MetricsConfig struct {
	Enabled bool   `json:"enabled"`
	Port    string `json:"port"`
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	gpuEnabled := getEnvBool("GPU_ENABLED", true)
	gpuCount, _ := strconv.Atoi(getEnv("GPU_COUNT", "1"))
	if !gpuEnabled {
		gpuCount = 0
	}
	workers, _ := strconv.Atoi(getEnv("TRANSCODE_WORKERS", "4"))
	queueSize, _ := strconv.Atoi(getEnv("QUEUE_SIZE", "100"))
	memThreshold, _ := strconv.ParseFloat(getEnv("GPU_MEMORY_THRESHOLD", "0.9"), 64)

	return &Config{
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
		MetricsPort: getEnv("METRICS_PORT", "9090"),
		HealthPort:  getEnv("HEALTH_PORT", "8081"),
		Workers:     workers,
		QueueSize:   queueSize,
		GPU: GPUConfig{
			Enabled:         gpuEnabled,
			Count:           gpuCount,
			MemoryThreshold: memThreshold,
		},
		Metrics: MetricsConfig{
			Enabled: getEnvBool("METRICS_ENABLED", true),
			Port:    getEnv("METRICS_PORT", "9090"),
		},
		NVEncPreset: getEnv("NVENC_PRESET", "p4"),
	}, nil
}

// Validate checks configuration completeness.
func (c *Config) Validate() error {
	if c.HTTPPort == "" {
		return fmt.Errorf("HTTP_PORT is required")
	}
	if c.Workers <= 0 {
		return fmt.Errorf("workers must be > 0")
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
