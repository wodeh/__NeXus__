// Package events provides Kafka-backed event publishing for PMS integration.
package events

import (
	"context"
	"fmt"
	"log/slog"

	bcEvents "github.com/nexus-platform/backend-core/internal/events"
)

// KafkaEventStore adapts the PMS Store interface to the backend-core ProducerManager.
// It satisfies the events.Store interface used by the saga orchestrator.
type KafkaEventStore struct {
	producer *bcEvents.ProducerManager
	mapper   *DomainEventMapper
}

// NewKafkaEventStore creates a Kafka-backed event store adapter.
func NewKafkaEventStore(producer *bcEvents.ProducerManager, mapper *DomainEventMapper) *KafkaEventStore {
	return &KafkaEventStore{producer: producer, mapper: mapper}
}

// Publish converts a PMS DomainEvent to a transport Event and publishes it to Kafka.
func (s *KafkaEventStore) Publish(ctx context.Context, event DomainEvent) error {
	transportEvent, err := s.mapper.Map(event)
	if err != nil {
		return fmt.Errorf("map event: %w", err)
	}

	if err := s.producer.Publish(ctx, transportEvent); err != nil {
		return fmt.Errorf("publish event: %w", err)
	}

	slog.Info("event published",
		slog.String("type", event.Type),
		slog.String("tenant", event.TenantID),
		slog.String("correlation_id", event.CorrelationID),
	)
	return nil
}

// Close gracefully shuts down the producer.
func (s *KafkaEventStore) Close() error {
	return s.producer.Close()
}
