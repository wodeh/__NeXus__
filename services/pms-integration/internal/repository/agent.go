package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nexus-platform/pms-integration/internal/domain"
)

// AgentRepository manages travel agents and OTA partners.
type AgentRepository struct {
	mu     sync.RWMutex
	agents map[string]*domain.Agent
	seq    int64
}

// NewAgentRepository creates a new agent repository.
func NewAgentRepository() *AgentRepository {
	return &AgentRepository{
		agents: make(map[string]*domain.Agent),
		seq:    100,
	}
}

// ListByTenant returns all agents for a tenant, optionally filtered by type.
func (r *AgentRepository) ListByTenant(ctx context.Context, tenant, agentType string, limit, offset int) ([]*domain.Agent, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*domain.Agent
	for _, a := range r.agents {
		if a.TenantID != tenant {
			continue
		}
		if agentType != "" && a.Type != agentType {
			continue
		}
		results = append(results, a)
	}
	total := len(results)
	if offset >= total {
		return []*domain.Agent{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return results[offset:end], total, nil
}

// GetByID returns a specific agent.
func (r *AgentRepository) GetByID(ctx context.Context, tenant, id string) (*domain.Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	a, ok := r.agents[id]
	if !ok || a.TenantID != tenant {
		return nil, fmt.Errorf("agent not found")
	}
	return a, nil
}

// Create stores a new agent.
func (r *AgentRepository) Create(ctx context.Context, a *domain.Agent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.seq++
	a.ID = fmt.Sprintf("ag-%d", r.seq)
	a.CreatedAt = time.Now()
	a.UpdatedAt = a.CreatedAt
	r.agents[a.ID] = a
	return nil
}

// Update modifies an existing agent.
func (r *AgentRepository) Update(ctx context.Context, tenant, id string, updates *domain.Agent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	a, ok := r.agents[id]
	if !ok || a.TenantID != tenant {
		return fmt.Errorf("agent not found")
	}
	if updates.Name != "" {
		a.Name = updates.Name
	}
	if updates.CommissionPct >= 0 {
		a.CommissionPct = updates.CommissionPct
	}
	if updates.ContactName != "" {
		a.ContactName = updates.ContactName
	}
	a.IsActive = updates.IsActive
	a.UpdatedAt = time.Now()
	return nil
}

// Delete removes an agent.
func (r *AgentRepository) Delete(ctx context.Context, tenant, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	a, ok := r.agents[id]
	if !ok || a.TenantID != tenant {
		return fmt.Errorf("agent not found")
	}
	delete(r.agents, id)
	return nil
}

// SeedAgents pre-populates demo data.
func (r *AgentRepository) SeedAgents() {
	r.Create(context.Background(), &domain.Agent{
		TenantID:      "demo",
		Name:          "Booking.com",
		Type:          domain.AgentTypeOTA,
		CommissionPct:   15.0,
		ContactName:   "Partner Support",
		ContactEmail:  "partners@booking.com",
		ContactPhone:  "+1-800-BOOKING",
		ContractRef:   "CNT-2026-001",
		IsActive:      true,
		SourceCode:    "BKG",
	})
	r.Create(context.Background(), &domain.Agent{
		TenantID:      "demo",
		Name:          "Expedia",
		Type:          domain.AgentTypeOTA,
		CommissionPct:   18.0,
		ContactName:   "Account Manager",
		ContactEmail:  "hotel@expedia.com",
		ContactPhone:  "+1-800-EXPEDIA",
		ContractRef:   "CNT-2026-002",
		IsActive:      true,
		SourceCode:    "EXP",
	})
	r.Create(context.Background(), &domain.Agent{
		TenantID:      "demo",
		Name:          "Virtuoso Travel",
		Type:          domain.AgentTypeTravelAgent,
		CommissionPct:   10.0,
		ContactName:   "Jane Smith",
		ContactEmail:  "jane@virtuoso.com",
		ContactPhone:  "+1-555-0300",
		ContractRef:   "CNT-2026-003",
		IsActive:      true,
		SourceCode:    "VRT",
	})
	r.Create(context.Background(), &domain.Agent{
		TenantID:      "demo",
		Name:          "TechCorp Corporate",
		Type:          domain.AgentTypeCorporate,
		CommissionPct:   0.0,
		ContactName:   "Mike Chen",
		ContactEmail:  "mike@techcorp.com",
		ContactPhone:  "+1-555-0200",
		ContractRef:   "CNT-2026-004",
		IsActive:      true,
		SourceCode:    "TCP",
	})
}
