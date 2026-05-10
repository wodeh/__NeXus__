// Package events provides Kafka-backed event publishing for PMS integration.
package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

// KafkaEventStore adapts the PMS Store interface to a local Kafka writer.
// It satisfies the events.Store interface used by the saga orchestrator.
type KafkaEventStore struct {
	writer *kafka.Writer
	mapper *DomainEventMapper
}

// NewKafkaEventStore creates a Kafka-backed event store adapter.
func NewKafkaEventStore(brokers []string, topic string, mapper *DomainEventMapper) (*KafkaEventStore, error) {
	if len(brokers) == 0 || topic == "" {
		return nil, fmt.Errorf("brokers and topic are required")
	}
	w := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	return &KafkaEventStore{writer: w, mapper: mapper}, nil
}

// Publish converts a PMS DomainEvent to a transport Event and publishes it to Kafka.
func (s *KafkaEventStore) Publish(ctx context.Context, event DomainEvent) error {
	transportEvent, err := s.mapper.Map(event)
	if err != nil {
		return fmt.Errorf("map event: %w", err)
	}

	payload, err := json.Marshal(transportEvent)
	if err != nil {
		return fmt.Errorf("marshal transport event: %w", err)
	}

	key := []byte(event.AggregateID)
	if err := s.writer.WriteMessages(ctx, kafka.Message{Key: key, Value: payload}); err != nil {
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
	return s.writer.Close()
}
