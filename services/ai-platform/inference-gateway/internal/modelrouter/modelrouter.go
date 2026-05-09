// Package modelrouter implements tenant-aware model backend selection.
package modelrouter

import (
	"fmt"
	"sync"
)

// Router selects appropriate inference backends based on model ID and tenant.
type Router struct {
	backends map[string]*Backend
	models   map[string]string // modelID -> backendID
	mu       sync.RWMutex
}

// Backend represents a model inference provider.
type Backend struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"` // openai, anthropic, local, custom
	Endpoint string   `json:"endpoint"`
	APIKey   string   `json:"api_key"`
	Models   []string `json:"models"`
}

// Config defines available backends.
type Config struct {
	Backends []BackendConfig `json:"backends"`
}

// BackendConfig is the raw configuration for a backend.
type BackendConfig struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Endpoint string   `json:"endpoint"`
	APIKey   string   `json:"api_key"`
	Models   []string `json:"models"`
}

// New creates a model router from configuration.
func New(cfg Config) (*Router, error) {
	r := &Router{
		backends: make(map[string]*Backend),
		models:   make(map[string]string),
	}
	for _, bc := range cfg.Backends {
		b := &Backend{
			ID:       bc.ID,
			Type:     bc.Type,
			Endpoint: bc.Endpoint,
			APIKey:   bc.APIKey,
			Models:   bc.Models,
		}
		r.backends[b.ID] = b
		for _, m := range b.Models {
			r.models[m] = b.ID
		}
	}
	return r, nil
}

// SelectBackend returns the appropriate backend for a model ID and tenant.
func (r *Router) SelectBackend(modelID, tenantID string) (*Backend, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	backendID, ok := r.models[modelID]
	if !ok {
		for _, b := range r.backends {
			return b, nil
		}
		return nil, fmt.Errorf("no backends available")
	}
	b, ok := r.backends[backendID]
	if !ok {
		return nil, fmt.Errorf("backend %s not found", backendID)
	}
	return b, nil
}

// RegisterBackend dynamically adds a backend at runtime.
func (r *Router) RegisterBackend(b *Backend) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.backends[b.ID] = b
	for _, m := range b.Models {
		r.models[m] = b.ID
	}
}
