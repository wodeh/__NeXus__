package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	os.Clearenv()
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HTTPPort != "8080" {
		t.Errorf("HTTPPort = %q, want %q", cfg.HTTPPort, "8080")
	}
	if cfg.MetricsPort != "9090" {
		t.Errorf("MetricsPort = %q, want %q", cfg.MetricsPort, "9090")
	}
	if cfg.Environment != "development" {
		t.Errorf("Environment = %q, want %q", cfg.Environment, "development")
	}
}

func TestLoad_FromEnv(t *testing.T) {
	os.Clearenv()
	os.Setenv("HTTP_PORT", "9000")
	os.Setenv("LOG_LEVEL", "debug")
	defer os.Clearenv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HTTPPort != "9000" {
		t.Errorf("HTTPPort = %q, want %q", cfg.HTTPPort, "9000")
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "debug")
	}
}

func TestValidate_MissingHTTPPort(t *testing.T) {
	cfg := &Config{HTTPPort: ""}
	if err := cfg.Validate(); err == nil {
		t.Error("Validate() expected error for empty HTTPPort")
	}
}

func TestValidate_OK(t *testing.T) {
	cfg := &Config{HTTPPort: "8080"}
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() unexpected error = %v", err)
	}
}
