// Package events defines the domain event and command bus interfaces for PMS integration.
package events

import (
	"context"
	"time"
)

// Store publishes domain events to the event bus.
type Store interface {
	Publish(ctx context.Context, event DomainEvent) error
}

// CommandBus sends commands to downstream services and awaits responses.
type CommandBus interface {
	Send(ctx context.Context, command Command) (*CommandResponse, error)
}

// DomainEvent represents a business event in the PMS domain.
type DomainEvent struct {
	Type          string                 `json:"type"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	TenantID      string                 `json:"tenant_id"`
	CorrelationID string                 `json:"correlation_id"`
	OccurredAt    time.Time              `json:"occurred_at"`
	Payload       map[string]interface{} `json:"payload"`
}

// Command represents an imperative instruction to a downstream service.
type Command struct {
	Type          string                 `json:"type"`
	Service       string                 `json:"service"`
	SagaID        string                 `json:"saga_id"`
	StepIndex     int                    `json:"step_index"`
	CorrelationID string                 `json:"correlation_id"`
	TenantID      string                 `json:"tenant_id"`
	Payload       map[string]interface{} `json:"payload"`
}

// CommandResponse carries the result of a command execution.
type CommandResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data"`
	Error   string                 `json:"error"`
}
