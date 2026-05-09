package events

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// EventMetrics instruments the event bus with Prometheus counters and histograms.
type EventMetrics struct {
	publishedTotal   *prometheus.CounterVec
	publishedLatency *prometheus.HistogramVec
	consumedTotal    *prometheus.CounterVec
	consumeLatency   *prometheus.HistogramVec
	mu               sync.RWMutex
}

// NewEventMetrics creates event bus metrics. If reg is nil, metrics are not registered.
func NewEventMetrics(reg prometheus.Registerer) *EventMetrics {
	m := &EventMetrics{
		publishedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{Name: "event_published_total", Help: "Total events published"},
			[]string{"event_type", "status"},
		),
		publishedLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{Name: "event_publish_latency_seconds", Help: "Event publish latency"},
			[]string{"event_type"},
		),
		consumedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{Name: "event_consumed_total", Help: "Total events consumed"},
			[]string{"event_type", "status"},
		),
		consumeLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{Name: "event_consume_latency_seconds", Help: "Event consume latency"},
			[]string{"event_type"},
		),
	}
	if reg != nil {
		reg.MustRegister(m.publishedTotal, m.publishedLatency, m.consumedTotal, m.consumeLatency)
	}
	return m
}

// IncPublished increments the published counter.
func (m *EventMetrics) IncPublished(eventType, status string) {
	m.publishedTotal.WithLabelValues(eventType, status).Inc()
}

// ObservePublishLatency records publish latency.
func (m *EventMetrics) ObservePublishLatency(eventType string, seconds float64) {
	m.publishedLatency.WithLabelValues(eventType).Observe(seconds)
}

// IncConsumed increments the consumed counter.
func (m *EventMetrics) IncConsumed(eventType, status string) {
	m.consumedTotal.WithLabelValues(eventType, status).Inc()
}

// ObserveConsumeLatency records consume latency.
func (m *EventMetrics) ObserveConsumeLatency(eventType string, seconds float64) {
	m.consumeLatency.WithLabelValues(eventType).Observe(seconds)
}
