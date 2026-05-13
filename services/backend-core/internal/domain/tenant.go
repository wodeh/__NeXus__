package domain

import (
	"github.com/google/uuid"
	"time"
)

// Tenant represents a hotel/villa property tenant in the system.
type Tenant struct {
	ID         uuid.UUID
	ExternalID string
	Name       string
	Region     string
	Tier       string
	Config     []byte
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
