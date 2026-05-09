// Package pipeline provides transcode pipeline management and job orchestration.
package pipeline

import (
	"fmt"
	"sync"
)

// Manager coordinates transcoding pipelines across multiple workers.
type Manager struct {
	mu       sync.RWMutex
	pipelines map[string]*Pipeline
}

// Pipeline represents a single transcoding workflow.
type Pipeline struct {
	ID       string            `json:"id"`
	StreamID string            `json:"stream_id"`
	Status   string            `json:"status"`
	TenantID string            `json:"tenant_id"`
	Stages   []Stage           `json:"stages"`
}

// Stage represents a pipeline stage.
type Stage struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// NewManager creates a pipeline manager.
func NewManager() *Manager {
	return &Manager{
		pipelines: make(map[string]*Pipeline),
	}
}

// Create initializes a new pipeline.
func (m *Manager) Create(id, streamID, tenantID string) (*Pipeline, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.pipelines[id]; exists {
		return nil, fmt.Errorf("pipeline %s already exists", id)
	}
	p := &Pipeline{
		ID:       id,
		StreamID: streamID,
		TenantID: tenantID,
		Status:   "created",
		Stages:   []Stage{},
	}
	m.pipelines[id] = p
	return p, nil
}

// Get retrieves a pipeline by ID.
func (m *Manager) Get(id string) (*Pipeline, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.pipelines[id]
	if !ok {
		return nil, fmt.Errorf("pipeline %s not found", id)
	}
	return p, nil
}

// ListByTenant returns pipelines for a tenant.
func (m *Manager) ListByTenant(tenantID string) []*Pipeline {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []*Pipeline
	for _, p := range m.pipelines {
		if p.TenantID == tenantID {
			out = append(out, p)
		}
	}
	return out
}

// UpdateStatus updates pipeline status.
func (m *Manager) UpdateStatus(id, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.pipelines[id]
	if !ok {
		return fmt.Errorf("pipeline %s not found", id)
	}
	p.Status = status
	return nil
}
