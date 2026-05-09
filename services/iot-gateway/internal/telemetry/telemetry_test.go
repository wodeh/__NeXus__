package telemetry

import (
	"testing"
)

func TestNew(t *testing.T) {
	c := New(Config{BatchSize: 10, FlushInterval: 0})
	if c == nil {
		t.Fatal("expected non-nil collector")
	}
}

func TestEnqueue(t *testing.T) {
	c := New(Config{BatchSize: 100, FlushInterval: 0})
	c.Enqueue("tenant-a", map[string]interface{}{"temp": 22.5})
	if len(c.buffer) != 1 {
		t.Errorf("buffer len = %d, want 1", len(c.buffer))
	}
}

func TestFlush(t *testing.T) {
	c := New(Config{BatchSize: 100, FlushInterval: 0})
	c.Enqueue("tenant-a", map[string]interface{}{"temp": 22.5})
	if err := c.Flush(); err != nil {
		t.Fatalf("Flush error = %v", err)
	}
	if len(c.buffer) != 0 {
		t.Errorf("buffer len after flush = %d, want 0", len(c.buffer))
	}
}

func TestRecordError(t *testing.T) {
	c := New(Config{BatchSize: 100, FlushInterval: 0})
	c.RecordError("tenant-a", "connection_timeout")
	if len(c.buffer) != 1 {
		t.Errorf("buffer len = %d, want 1", len(c.buffer))
	}
}

func TestRecordCommand(t *testing.T) {
	c := New(Config{BatchSize: 100, FlushInterval: 0})
	c.RecordCommand("tenant-a", "dev-1", "turn_on")
	if len(c.buffer) != 1 {
		t.Errorf("buffer len = %d, want 1", len(c.buffer))
	}
}
