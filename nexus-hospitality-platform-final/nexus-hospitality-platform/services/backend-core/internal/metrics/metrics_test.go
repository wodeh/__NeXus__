package metrics

import (
	"context"
	"testing"
	"time"
)

func TestNewCollector_Disabled(t *testing.T) {
	c := NewCollector(Config{Enabled: false})
	if c == nil {
		t.Fatal("NewCollector returned nil")
	}
	// Should not panic
	c.IncrementCounter("http_requests_total", "tenant-a")
	c.HistogramObserve("http_request_duration_seconds", 0.1, "tenant-a")
	c.GaugeSet("active_connections", 5, "tenant-a")
}

func TestNewCollector_Enabled(t *testing.T) {
	c := NewCollector(Config{Enabled: true, Port: "0"})
	if c == nil {
		t.Fatal("NewCollector returned nil")
	}
	c.IncrementCounter("http_requests_total", "tenant-a")
	c.HistogramObserve("http_request_duration_seconds", 0.1, "tenant-a")
	c.GaugeSet("active_connections", 5, "tenant-a")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_ = c.Shutdown(ctx)
}

func TestGaugeSet_DynamicRegistration(t *testing.T) {
	c := NewCollector(Config{Enabled: true, Port: "0"})
	c.GaugeSet("custom_gauge", 42, "tenant-b")
	// Second call should not panic
	c.GaugeSet("custom_gauge", 43, "tenant-b")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_ = c.Shutdown(ctx)
}
