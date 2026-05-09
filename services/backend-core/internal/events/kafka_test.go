package events

import (
	"testing"
)

func TestNewProducer_MissingBrokers(t *testing.T) {
	_, err := NewProducer([]string{}, "test-topic")
	if err == nil {
		t.Fatal("expected error when brokers are empty")
	}
}

func TestNewProducer_MissingTopic(t *testing.T) {
	_, err := NewProducer([]string{"localhost:9092"}, "")
	if err == nil {
		t.Fatal("expected error when topic is empty")
	}
}

func TestNewConsumer_OK(t *testing.T) {
	c, err := NewConsumer([]string{"localhost:9092"}, "test-topic", "test-group")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected consumer, got nil")
	}
	_ = c.Close()
}
