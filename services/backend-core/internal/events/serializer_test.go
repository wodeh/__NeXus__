package events

import (
	"testing"
	"time"
)

func TestJSONSerializer_RoundTrip(t *testing.T) {
	s := NewJSONSerializer()
	event := Event{
		ID:            "evt-test-1",
		Type:          "test.event",
		Version:       "1.0.0",
		TenantID:      "tenant-a",
		CorrelationID: "corr-1",
		Timestamp:     time.Now().UTC(),
		Source:        "backend-core",
		Payload:       []byte(`{"key":"value"}`),
	}

	data, err := s.Serialize(event)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}

	got, err := s.Deserialize(data)
	if err != nil {
		t.Fatalf("deserialize: %v", err)
	}

	if got.ID != event.ID {
		t.Fatalf("expected ID %s, got %s", event.ID, got.ID)
	}
	if got.Type != event.Type {
		t.Fatalf("expected Type %s, got %s", event.Type, got.Type)
	}
}

func TestJSONSerializer_InvalidData(t *testing.T) {
	s := NewJSONSerializer()
	_, err := s.Deserialize([]byte("not-json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
