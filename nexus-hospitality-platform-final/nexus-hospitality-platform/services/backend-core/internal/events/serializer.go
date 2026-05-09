package events

import (
	"encoding/json"
	"fmt"
)

// Serializer converts Events to and from Kafka message payloads.
type Serializer interface {
	Serialize(event Event) ([]byte, error)
	Deserialize(data []byte) (Event, error)
}

// JSONSerializer implements Serializer using JSON.
type JSONSerializer struct{}

// NewJSONSerializer creates a JSON serializer.
func NewJSONSerializer() *JSONSerializer {
	return &JSONSerializer{}
}

// Serialize encodes an Event to JSON bytes.
func (s *JSONSerializer) Serialize(event Event) ([]byte, error) {
	return json.Marshal(event)
}

// Deserialize decodes JSON bytes into an Event.
func (s *JSONSerializer) Deserialize(data []byte) (Event, error) {
	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		return Event{}, fmt.Errorf("unmarshal event: %w", err)
	}
	return event, nil
}
