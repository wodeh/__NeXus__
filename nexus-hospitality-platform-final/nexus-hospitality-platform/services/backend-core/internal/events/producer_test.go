package events

import (
	"testing"
)

func TestNewProducerManager_MissingBrokers(t *testing.T) {
	_, err := NewProducerManager(ProducerConfig{
		Brokers: []string{},
		Topic:   "test",
		Source:  "test",
	})
	if err == nil {
		t.Fatal("expected error for missing brokers")
	}
}

func TestNewProducerManager_MissingTopic(t *testing.T) {
	_, err := NewProducerManager(ProducerConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "",
		Source:  "test",
	})
	if err == nil {
		t.Fatal("expected error for missing topic")
	}
}
