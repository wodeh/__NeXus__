package config

import (
	"os"
	"testing"
	"time"
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
	if cfg.CacheTTL != 10*time.Minute {
		t.Errorf("CacheTTL = %v, want %v", cfg.CacheTTL, 10*time.Minute)
	}
	if len(cfg.Models.Backends) == 0 {
		t.Error("expected at least one backend")
	}
}

func TestValidate_MissingHTTPPort(t *testing.T) {
	cfg := &Config{HTTPPort: ""}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for empty HTTPPort")
	}
}

func TestValidate_MissingBackends(t *testing.T) {
	cfg := &Config{HTTPPort: "8080", Models: ModelsConfig{Backends: []BackendConfig{}}}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for empty backends")
	}
}
