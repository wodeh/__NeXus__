package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nexus-platform/backend-core/internal/db"
	"github.com/nexus-platform/backend-core/internal/domain"
)

// AgentRepository provides agent data access.
type AgentRepository struct {
	pool *db.Pool
}

// NewAgentRepository creates an agent repository.
func NewAgentRepository(pool *db.Pool) *AgentRepository {
	return &AgentRepository{pool: pool}
}

// List returns all agents for a tenant.
func (r *AgentRepository) List(ctx context.Context, tenantID string) ([]domain.Agent, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, name, type, commission_pct, contact_name, contact_email, contact_phone, contract_ref, is_active, source_code, config, created_at, updated_at
		FROM agents
		WHERE tenant_id = $1
		ORDER BY name
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list agents: %w", err)
	}
	defer rows.Close()

	var agents []domain.Agent
	for rows.Next() {
		var a domain.Agent
		var configJSON []byte
		if err := rows.Scan(
			&a.ID, &a.TenantID, &a.Name, &a.Type, &a.CommissionPct, &a.ContactName, &a.ContactEmail, &a.ContactPhone, &a.ContractRef, &a.IsActive, &a.SourceCode, &configJSON, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan agent: %w", err)
		}
		if len(configJSON) > 0 {
			_ = json.Unmarshal(configJSON, &a.Config)
		}
		agents = append(agents, a)
	}
	return agents, rows.Err()
}

// GetByID returns a single agent.
func (r *AgentRepository) GetByID(ctx context.Context, tenantID string, id uuid.UUID) (*domain.Agent, error) {
	var a domain.Agent
	var configJSON []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, name, type, commission_pct, contact_name, contact_email, contact_phone, contract_ref, is_active, source_code, config, created_at, updated_at
		FROM agents
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, id).Scan(
		&a.ID, &a.TenantID, &a.Name, &a.Type, &a.CommissionPct, &a.ContactName, &a.ContactEmail, &a.ContactPhone, &a.ContractRef, &a.IsActive, &a.SourceCode, &configJSON, &a.CreatedAt, &a.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("agent not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get agent: %w", err)
	}
	if len(configJSON) > 0 {
		_ = json.Unmarshal(configJSON, &a.Config)
	}
	return &a, nil
}

// Create creates a new agent.
func (r *AgentRepository) Create(ctx context.Context, tenantID string, req *domain.AgentCreateRequest) (*domain.Agent, error) {
	id := uuid.New()
	now := time.Now().UTC()
	var configJSON []byte

	_, err := r.pool.Exec(ctx, `
		INSERT INTO agents (id, tenant_id, name, type, commission_pct, contact_name, contact_email, contact_phone, contract_ref, is_active, source_code, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $13)
	`, id, tenantID, req.Name, req.Type, req.CommissionPct, req.ContactName, req.ContactEmail, req.ContactPhone, req.ContractRef, true, req.SourceCode, configJSON, now)
	if err != nil {
		return nil, fmt.Errorf("create agent: %w", err)
	}

	return &domain.Agent{
		ID: id, TenantID: uuid.MustParse(tenantID), Name: req.Name, Type: req.Type,
		CommissionPct: req.CommissionPct, ContactName: req.ContactName, ContactEmail: req.ContactEmail,
		ContactPhone: req.ContactPhone, ContractRef: req.ContractRef, IsActive: true, SourceCode: req.SourceCode,
		CreatedAt: now, UpdatedAt: now,
	}, nil
}

// Update updates an agent.
func (r *AgentRepository) Update(ctx context.Context, tenantID string, id uuid.UUID, req *domain.AgentUpdateRequest) (*domain.Agent, error) {
	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
		UPDATE agents SET
			name = COALESCE(NULLIF($3, ''), name),
			type = COALESCE(NULLIF($4, ''), type),
			commission_pct = COALESCE($5, commission_pct),
			contact_name = COALESCE(NULLIF($6, ''), contact_name),
			contact_email = COALESCE(NULLIF($7, ''), contact_email),
			contact_phone = COALESCE(NULLIF($8, ''), contact_phone),
			contract_ref = COALESCE(NULLIF($9, ''), contract_ref),
			is_active = COALESCE($10, is_active),
			source_code = COALESCE(NULLIF($11, ''), source_code),
			updated_at = $12
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, id, req.Name, req.Type, req.CommissionPct, req.ContactName, req.ContactEmail, req.ContactPhone, req.ContractRef, req.IsActive, req.SourceCode, now)
	if err != nil {
		return nil, fmt.Errorf("update agent: %w", err)
	}
	return r.GetByID(ctx, tenantID, id)
}

// Delete removes an agent.
func (r *AgentRepository) Delete(ctx context.Context, tenantID string, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM agents WHERE tenant_id = $1 AND id = $2`, tenantID, id)
	if err != nil {
		return fmt.Errorf("delete agent: %w", err)
	}
	return nil
}
