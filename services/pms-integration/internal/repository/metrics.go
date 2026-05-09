// Package repository provides Prometheus metrics for repository operations.
package repository

import (
	"github.com/prometheus/client_golang/prometheus"
)

// RepositoryMetrics instruments repository operations with Prometheus counters and histograms.
type RepositoryMetrics struct {
	queryTotal    *prometheus.CounterVec
	queryErrors   *prometheus.CounterVec
	queryDuration *prometheus.HistogramVec
}

// NewRepositoryMetrics creates repository metrics.
func NewRepositoryMetrics() *RepositoryMetrics {
	m := &RepositoryMetrics{
		queryTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{Name: "repository_queries_total", Help: "Total repository queries"},
			[]string{"entity", "operation"},
		),
		queryErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{Name: "repository_query_errors_total", Help: "Total repository query errors"},
			[]string{"entity", "operation", "error_type"},
		),
		queryDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{Name: "repository_query_duration_seconds", Help: "Repository query duration"},
			[]string{"entity", "operation"},
		),
	}
	// Register with default registry; caller should ensure no duplicate registration.
	_ = prometheus.DefaultRegisterer.Register(m.queryTotal)
	_ = prometheus.DefaultRegisterer.Register(m.queryErrors)
	_ = prometheus.DefaultRegisterer.Register(m.queryDuration)
	return m
}

// IncQuery increments the query counter.
func (m *RepositoryMetrics) IncQuery(entity, operation string) {
	m.queryTotal.WithLabelValues(entity, operation).Inc()
}

// IncError increments the error counter.
func (m *RepositoryMetrics) IncError(entity, operation, errorType string) {
	m.queryErrors.WithLabelValues(entity, operation, errorType).Inc()
}

// ObserveDuration records query duration.
func (m *RepositoryMetrics) ObserveDuration(entity, operation string, seconds float64) {
	m.queryDuration.WithLabelValues(entity, operation).Observe(seconds)
}
