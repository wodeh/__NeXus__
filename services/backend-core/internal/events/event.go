package events

import (
	"encoding/json"
	"time"
)

// Event is the canonical envelope for all domain events in the Nexus platform.
// It carries tenant isolation, distributed tracing, and schema versioning metadata.
type Event struct {
	ID            string          `json:"id"`
	Type          string          `json:"type"`
	Version       string          `json:"version"`
	TenantID      string          `json:"tenant_id"`
	CorrelationID string          `json:"correlation_id"`
	CausationID   string          `json:"causation_id"`
	Timestamp     time.Time       `json:"timestamp"`
	Source        string          `json:"source"`
	Payload       json.RawMessage `json:"payload"`
}

// NewEvent constructs an Event with required metadata.
func NewEvent(eventType, version, tenantID, correlationID, source string, payload interface{}) (Event, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return Event{}, err
	}
	return Event{
		ID:            generateID(),
		Type:          eventType,
		Version:       version,
		TenantID:      tenantID,
		CorrelationID: correlationID,
		Timestamp:     time.Now().UTC(),
		Source:        source,
		Payload:       raw,
	}, nil
}

func generateID() string {
	// Use a simple UUID-like generation; replace with github.com/google/uuid if preferred.
	// For now we avoid extra deps in this package.
	return "evt_" + time.Now().UTC().Format("20060102T150405.000000000") + randomSuffix()
}

func randomSuffix() string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
