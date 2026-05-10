package domain

import (
	"errors"
	"time"
)

// Agent represents a travel agent, OTA, or booking channel.
type Agent struct {
	ID            string
	TenantID      string
	Name          string
	Type          string // ota, travel_agent, corporate, direct
	CommissionPct float64
	ContactName   string
	ContactEmail  string
	ContactPhone  string
	ContractRef   string
	IsActive      bool
	SourceCode    string // for analytics tracking
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// AgentType values.
const (
	AgentTypeOTA         = "ota"
	AgentTypeTravelAgent = "travel_agent"
	AgentTypeCorporate   = "corporate"
	AgentTypeDirect      = "direct"
)

// Validate checks an agent before creation.
func (a *Agent) Validate() error {
	if a.TenantID == "" {
		return errors.New("tenant_id is required")
	}
	if a.Name == "" {
		return errors.New("name is required")
	}
	if a.Type == "" {
		return errors.New("type is required")
	}
	if a.CommissionPct < 0 || a.CommissionPct > 100 {
		return errors.New("commission_pct must be between 0 and 100")
	}
	return nil
}
