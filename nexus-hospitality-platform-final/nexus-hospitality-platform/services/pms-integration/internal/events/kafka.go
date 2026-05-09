package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
)

// KafkaStore implements Store using Kafka as the backing event log.
type KafkaStore struct {
	writer *kafka.Writer
}

// NewKafkaStore creates a Kafka-backed event store.
func NewKafkaStore(brokers []string, topic string) (*KafkaStore, error) {
	if len(brokers) == 0 || topic == "" {
		return nil, fmt.Errorf("brokers and topic are required")
	}
	w := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	return &KafkaStore{writer: w}, nil
}

// Publish serializes a DomainEvent to JSON and writes it to Kafka.
func (s *KafkaStore) Publish(ctx context.Context, event DomainEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	key := []byte(event.AggregateID)
	return s.writer.WriteMessages(ctx, kafka.Message{Key: key, Value: payload})
}

// Close shuts down the store.
func (s *KafkaStore) Close() error {
	return s.writer.Close()
}
