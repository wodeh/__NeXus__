package domain

import (
	"time"

	"github.com/google/uuid"
)

// Agent represents an OTA, travel agent, or corporate partner.
type Agent struct {
	ID            uuid.UUID `json:"id"`
	TenantID      uuid.UUID `json:"tenant_id"`
	Name          string    `json:"name"`
	Type          string    `json:"type"`           // ota, travel_agent, corporate, direct
	CommissionPct int       `json:"commission_pct"` // 0-100
	ContactName   string    `json:"contact_name"`
	ContactEmail  string    `json:"contact_email"`
	ContactPhone  string    `json:"contact_phone"`
	ContractRef   string    `json:"contract_ref"`
	IsActive      bool      `json:"is_active"`
	SourceCode    string    `json:"source_code"`    // short code for reservation source mapping
	Config        map[string]interface{} `json:"config,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// AgentCreateRequest is the payload for creating an agent.
type AgentCreateRequest struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	CommissionPct int    `json:"commission_pct"`
	ContactName   string `json:"contact_name"`
	ContactEmail  string `json:"contact_email"`
	ContactPhone  string `json:"contact_phone"`
	ContractRef   string `json:"contract_ref"`
	SourceCode    string `json:"source_code"`
}

// AgentUpdateRequest is the payload for updating an agent.
type AgentUpdateRequest struct {
	Name          string `json:"name,omitempty"`
	Type          string `json:"type,omitempty"`
	CommissionPct *int   `json:"commission_pct,omitempty"`
	ContactName   string `json:"contact_name,omitempty"`
	ContactEmail  string `json:"contact_email,omitempty"`
	ContactPhone  string `json:"contact_phone,omitempty"`
	ContractRef   string `json:"contract_ref,omitempty"`
	IsActive      *bool  `json:"is_active,omitempty"`
	SourceCode    string `json:"source_code,omitempty"`
}
