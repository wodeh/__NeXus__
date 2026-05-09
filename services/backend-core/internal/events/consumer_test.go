package events

import (
	"context"
	"testing"
)

func TestNewConsumerManager_MissingBrokers(t *testing.T) {
	_, err := NewConsumerManager(ConsumerConfig{
		Brokers: []string{},
		Topic:   "test",
		GroupID: "test-group",
		Handler: HandlerFunc(func(ctx context.Context, event Event) error { return nil }),
	})
	if err == nil {
		t.Fatal("expected error for missing brokers")
	}
}

func TestNewConsumerManager_MissingHandler(t *testing.T) {
	_, err := NewConsumerManager(ConsumerConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "test",
		GroupID: "test-group",
	})
	if err == nil {
		t.Fatal("expected error for missing handler")
	}
}
