// Package config provides structured configuration management for the AI inference gateway.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/nexus-platform/ai-platform/inference-gateway/internal/guardrails"
	"github.com/nexus-platform/ai-platform/inference-gateway/internal/modelrouter"
	"github.com/nexus-platform/ai-platform/inference-gateway/internal/rag"
)

// Config holds all service configuration.
type Config struct {
	HTTPPort     string              `json:"http_port"`
	MetricsPort  string              `json:"metrics_port"`
	Models       modelrouter.Config  `json:"models"`
	Guardrails   *guardrails.Config  `json:"guardrails"`
	RAG          rag.Config          `json:"rag"`
	Metrics      MetricsConfig       `json:"metrics"`
	CacheTTL     time.Duration       `json:"cache_ttl"`
}

// GuardrailsConfig configures input/output guardrails.
type GuardrailsConfig = guardrails.Config

// RAGConfig configures retrieval-augmented generation.
type RAGConfig = rag.Config

// MetricsConfig configures observability metrics.
type MetricsConfig struct {
	Enabled bool   `json:"enabled"`
	Port    string `json:"port"`
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	cacheTTL, _ := strconv.Atoi(getEnv("CACHE_TTL_MINUTES", "10"))
	return &Config{
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
		MetricsPort: getEnv("METRICS_PORT", "9090"),
		Models: modelrouter.Config{
			Backends: []modelrouter.BackendConfig{
				{
					ID:     "default-openai",
					Type:   "openai",
					Models: []string{"gpt-4", "gpt-3.5-turbo"},
				},
				{
					ID:     "default-anthropic",
					Type:   "anthropic",
					Models: []string{"claude-3-opus", "claude-3-sonnet"},
				},
			},
		},
		Guardrails: &guardrails.Config{
			BlockedTopics:   []string{"violence", "self-harm", "illegal acts"},
			BlockedPatterns: []string{},
			PIIEntities:     []string{"ssn", "credit_card", "email", "phone"},
			HallucinationRules: guardrails.HallucinationConfig{
				FactualityThreshold:   0.7,
				SourceAlignmentWeight: 0.3,
				SelfConsistencyChecks: 2,
				ConfidenceThreshold:   0.8,
			},
			ToxicityThreshold: 0.5,
		},
		RAG: rag.Config{
			VectorStoreURL: getEnv("VECTOR_STORE_URL", ""),
			DefaultTopK:    5,
			Namespaces:     []string{},
		},
		Metrics: MetricsConfig{
			Enabled: getEnvBool("METRICS_ENABLED", true),
			Port:    getEnv("METRICS_PORT", "9090"),
		},
		CacheTTL: time.Duration(cacheTTL) * time.Minute,
	}, nil
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

// Validate checks that the configuration is complete and correct.
func (c *Config) Validate() error {
	if c.HTTPPort == "" {
		return fmt.Errorf("HTTP_PORT is required")
	}
	if len(c.Models.Backends) == 0 {
		return fmt.Errorf("at least one model backend must be configured")
	}
	return nil
}
