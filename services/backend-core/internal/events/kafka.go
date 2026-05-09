// Package events provides Kafka producer and consumer wrappers for backend-core.
package events

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
)

// Producer wraps kafka-go Writer for tenant-aware event publishing.
type Producer struct {
	writer *kafka.Writer
}

// NewProducer creates a Kafka producer.
func NewProducer(brokers []string, topic string) (*Producer, error) {
	if len(brokers) == 0 || topic == "" {
		return nil, fmt.Errorf("brokers and topic are required")
	}
	w := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	return &Producer{writer: w}, nil
}

// Publish sends a message to Kafka.
func (p *Producer) Publish(ctx context.Context, key, value []byte) error {
	return p.writer.WriteMessages(ctx, kafka.Message{Key: key, Value: value})
}

// Close shuts down the producer.
func (p *Producer) Close() error {
	return p.writer.Close()
}

// Consumer wraps kafka-go Reader.
type Consumer struct {
	reader *kafka.Reader
}

// NewConsumer creates a Kafka consumer.
func NewConsumer(brokers []string, topic, groupID string) (*Consumer, error) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})
	return &Consumer{reader: r}, nil
}

// Fetch reads the next message.
func (c *Consumer) Fetch(ctx context.Context) (kafka.Message, error) {
	return c.reader.ReadMessage(ctx)
}

// Close shuts down the consumer.
func (c *Consumer) Close() error {
	return c.reader.Close()
}
