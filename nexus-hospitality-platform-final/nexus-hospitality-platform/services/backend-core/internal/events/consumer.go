package events

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Handler processes a single domain event.
type Handler interface {
	Handle(ctx context.Context, event Event) error
}

// HandlerFunc is an adapter to allow ordinary functions as handlers.
type HandlerFunc func(ctx context.Context, event Event) error

// Handle implements Handler.
func (f HandlerFunc) Handle(ctx context.Context, event Event) error {
	return f(ctx, event)
}

// ConsumerManager consumes events from Kafka with idempotency, DLQ, tracing, and metrics.
type ConsumerManager struct {
	reader      *kafka.Reader
	handler     Handler
	serializer  Serializer
	idempotency IdempotencyStore
	metrics     *EventMetrics
	tracer      trace.Tracer
	dlqWriter   *kafka.Writer

	wg     sync.WaitGroup
	cancel context.CancelFunc
	mu     sync.RWMutex
	closed bool
}

// ConsumerConfig configures the consumer manager.
type ConsumerConfig struct {
	Brokers     []string
	Topic       string
	GroupID     string
	Handler     Handler
	Serializer  Serializer
	Idempotency IdempotencyStore
	Metrics     *EventMetrics
	DLQTopic    string // optional; if empty, DLQ is disabled
}

// NewConsumerManager creates a high-level event consumer.
func NewConsumerManager(cfg ConsumerConfig) (*ConsumerManager, error) {
	if len(cfg.Brokers) == 0 || cfg.Topic == "" || cfg.GroupID == "" {
		return nil, fmt.Errorf("brokers, topic, and group_id are required")
	}
	if cfg.Handler == nil {
		return nil, fmt.Errorf("handler is required")
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

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          cfg.Topic,
		GroupID:        cfg.GroupID,
		MinBytes:       1,
		MaxBytes:       10e6, // 10MB
		CommitInterval: 1 * time.Second,
	})

	c := &ConsumerManager{
		reader:      r,
		handler:     cfg.Handler,
		serializer:  cfg.Serializer,
		idempotency: cfg.Idempotency,
		metrics:     cfg.Metrics,
		tracer:      otel.Tracer("event-consumer"),
	}

	if cfg.DLQTopic != "" {
		c.dlqWriter = &kafka.Writer{
			Addr:     kafka.TCP(cfg.Brokers...),
			Topic:    cfg.DLQTopic,
			Balancer: &kafka.LeastBytes{},
		}
	}

	return c, nil
}

// Start begins consuming messages in a background goroutine.
func (c *ConsumerManager) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	c.wg.Add(1)
	go c.run(ctx)
	slog.Info("event consumer started", slog.String("topic", c.reader.Config().Topic), slog.String("group", c.reader.Config().GroupID))
}

func (c *ConsumerManager) run(ctx context.Context) {
	defer c.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			slog.Error("fetch message failed", slog.String("error", err.Error()))
			c.metrics.IncConsumed("unknown", "fetch_error")
			continue
		}

		c.processMessage(ctx, msg)
	}
}

func (c *ConsumerManager) processMessage(ctx context.Context, msg kafka.Message) {
	start := time.Now()
	eventType := "unknown"
	defer func() {
		duration := time.Since(start).Seconds()
		c.metrics.ObserveConsumeLatency(eventType, duration)
	}()

	// Deserialize
	event, err := c.serializer.Deserialize(msg.Value)
	if err != nil {
		slog.Error("deserialize failed", slog.String("error", err.Error()))
		c.metrics.IncConsumed(eventType, "deserialize_error")
		_ = c.sendToDLQ(ctx, msg, "deserialize_error")
		return
	}
	eventType = event.Type

	// Tracing
	ctx, span := c.tracer.Start(ctx, "event.consume",
		trace.WithAttributes(
			attribute.String("event.type", event.Type),
			attribute.String("event.id", event.ID),
			attribute.String("tenant.id", event.TenantID),
		),
	)
	defer span.End()

	// Idempotency
	seen, err := c.idempotency.IsProcessed(ctx, event.ID)
	if err != nil {
		span.RecordError(err)
		slog.Error("idempotency check failed", slog.String("event.id", event.ID), slog.String("error", err.Error()))
		c.metrics.IncConsumed(eventType, "idempotency_error")
		return
	}
	if seen {
		c.metrics.IncConsumed(eventType, "deduplicated")
		return
	}

	// Handler with retry
	const maxRetries = 3
	var handlerErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		handlerErr = c.handler.Handle(ctx, event)
		if handlerErr == nil {
			break
		}
		c.metrics.IncConsumed(eventType, "retry")
		slog.Warn("handler retry", slog.String("event.id", event.ID), slog.Int("attempt", attempt+1))
		time.Sleep(time.Duration(attempt+1) * 200 * time.Millisecond)
	}

	if handlerErr != nil {
		span.RecordError(handlerErr)
		slog.Error("handler failed permanently", slog.String("event.id", event.ID), slog.String("error", handlerErr.Error()))
		c.metrics.IncConsumed(eventType, "error")
		_ = c.sendToDLQ(ctx, msg, handlerErr.Error())
		return
	}

	// Mark processed
	if err := c.idempotency.MarkProcessed(ctx, event.ID, 24*time.Hour); err != nil {
		slog.Warn("failed to mark consumed event", slog.String("event.id", event.ID), slog.String("error", err.Error()))
	}

	// Commit offset
	if err := c.reader.CommitMessages(ctx, msg); err != nil {
		slog.Error("commit failed", slog.String("event.id", event.ID), slog.String("error", err.Error()))
		c.metrics.IncConsumed(eventType, "commit_error")
		return
	}

	c.metrics.IncConsumed(eventType, "success")
}

func (c *ConsumerManager) sendToDLQ(ctx context.Context, msg kafka.Message, reason string) error {
	if c.dlqWriter == nil {
		return nil
	}
	msg.Headers = append(msg.Headers, kafka.Header{Key: "x-dlq-reason", Value: []byte(reason)})
	msg.Headers = append(msg.Headers, kafka.Header{Key: "x-dlq-timestamp", Value: []byte(time.Now().UTC().Format(time.RFC3339))})
	if err := c.dlqWriter.WriteMessages(ctx, msg); err != nil {
		slog.Error("dlq write failed", slog.String("error", err.Error()))
		return err
	}
	return nil
}

// Shutdown gracefully stops the consumer with a timeout.
func (c *ConsumerManager) Shutdown(timeout time.Duration) error {
	c.mu.Lock()
	c.closed = true
	c.mu.Unlock()

	if c.cancel != nil {
		c.cancel()
	}

	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		slog.Info("consumer shutdown complete")
	case <-time.After(timeout):
		slog.Warn("consumer shutdown timed out")
	}

	if err := c.reader.Close(); err != nil {
		return fmt.Errorf("close reader: %w", err)
	}
	if c.dlqWriter != nil {
		_ = c.dlqWriter.Close()
	}
	return nil
}
