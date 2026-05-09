package events

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ProducerManager publishes tenant-aware events to Kafka with retry, idempotency, and tracing.
type ProducerManager struct {
	writer      *kafka.Writer
	serializer  Serializer
	idempotency IdempotencyStore
	metrics     *EventMetrics
	tracer      trace.Tracer
	source      string

	mu     sync.RWMutex
	closed bool
}

// ProducerConfig configures the producer manager.
type ProducerConfig struct {
	Brokers     []string
	Topic       string
	Source      string
	Serializer  Serializer
	Idempotency IdempotencyStore
	Metrics     *EventMetrics
}

// NewProducerManager creates a high-level event producer.
func NewProducerManager(cfg ProducerConfig) (*ProducerManager, error) {
	if len(cfg.Brokers) == 0 || cfg.Topic == "" {
		return nil, fmt.Errorf("brokers and topic are required")
	}
	if cfg.Serializer == nil {
		cfg.Serializer = NewJSONSerializer()
	}
	if cfg.Idempotency == nil {
		cfg.Idempotency = NewInMemoryIdempotencyStore(24 * time.Hour)
	}
	if cfg.Metrics == nil {
		cfg.Metrics = NewEventMetrics(nil)
	}

	w := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		MaxAttempts:  3,
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
	}

	return &ProducerManager{
		writer:      w,
		serializer:  cfg.Serializer,
		idempotency: cfg.Idempotency,
		metrics:     cfg.Metrics,
		tracer:      otel.Tracer("event-producer"),
		source:      cfg.Source,
	}, nil
}

// Publish sends an event to Kafka with at-least-once delivery semantics.
func (p *ProducerManager) Publish(ctx context.Context, event Event) error {
	ctx, span := p.tracer.Start(ctx, "event.publish",
		trace.WithAttributes(
			attribute.String("event.type", event.Type),
			attribute.String("event.id", event.ID),
			attribute.String("tenant.id", event.TenantID),
		))
	defer span.End()

	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return fmt.Errorf("producer closed")
	}
	p.mu.RUnlock()

	// Idempotency guard
	seen, err := p.idempotency.IsProcessed(ctx, event.ID)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("idempotency check: %w", err)
	}
	if seen {
		p.metrics.IncPublished(event.Type, "deduplicated")
		return ErrDuplicateEvent
	}

	payload, err := p.serializer.Serialize(event)
	if err != nil {
		span.RecordError(err)
		p.metrics.IncPublished(event.Type, "serialize_error")
		return fmt.Errorf("serialize: %w", err)
	}

	key := []byte(event.TenantID)
	msg := kafka.Message{
		Key:   key,
		Value: payload,
		Headers: []kafka.Header{
			{Key: "x-tenant-id", Value: []byte(event.TenantID)},
			{Key: "x-correlation-id", Value: []byte(event.CorrelationID)},
			{Key: "x-event-type", Value: []byte(event.Type)},
			{Key: "x-event-version", Value: []byte(event.Version)},
		},
	}

	// Retry with exponential backoff for transient errors.
	b := backoff.WithMaxRetries(
		backoff.NewExponentialBackOff(
			backoff.WithInitialInterval(100*time.Millisecond),
			backoff.WithMaxInterval(2*time.Second),
			backoff.WithMaxElapsedTime(10*time.Second),
		),
		3,
	)

	err = backoff.RetryNotify(
		func() error {
			return p.writer.WriteMessages(ctx, msg)
		},
		b,
		func(err error, d time.Duration) {
			slog.Warn("kafka publish retry", slog.String("event.id", event.ID), slog.Duration("after", d))
			p.metrics.IncPublished(event.Type, "retry")
		},
	)
	if err != nil {
		span.RecordError(err)
		p.metrics.IncPublished(event.Type, "error")
		return fmt.Errorf("kafka write: %w", err)
	}

	if markErr := p.idempotency.MarkProcessed(ctx, event.ID, 24*time.Hour); markErr != nil {
		slog.Warn("failed to mark event processed", slog.String("event.id", event.ID), slog.String("error", markErr.Error()))
	}

	p.metrics.IncPublished(event.Type, "success")
	p.metrics.ObservePublishLatency(event.Type, time.Since(event.Timestamp).Seconds())
	return nil
}

// Close gracefully shuts down the producer.
func (p *ProducerManager) Close() error {
	p.mu.Lock()
	p.closed = true
	p.mu.Unlock()
	return p.writer.Close()
}
