package config

import "github.com/prometheus/client_golang/prometheus"

// ConfigMetrics instruments config resolution with Prometheus counters.
type ConfigMetrics struct {
	resolvedTotal *prometheus.CounterVec
}

// NewConfigMetrics creates config metrics. If reg is nil, metrics are not registered.
func NewConfigMetrics(reg prometheus.Registerer) *ConfigMetrics {
	m := &ConfigMetrics{
		resolvedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{Name: "config_resolved_total", Help: "Total config resolutions"},
			[]string{"key", "source"},
		),
	}
	if reg != nil {
		reg.MustRegister(m.resolvedTotal)
	}
	return m
}

// IncResolved increments the resolution counter.
func (m *ConfigMetrics) IncResolved(key, source string) {
	m.resolvedTotal.WithLabelValues(key, source).Inc()
}
