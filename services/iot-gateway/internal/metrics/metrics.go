// Package metrics provides tenant-aware Prometheus metrics for the IoT gateway.
package metrics

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Collector aggregates service-level Prometheus metrics with tenant labels.
type Collector struct {
	registry    *prometheus.Registry
	counters    map[string]*prometheus.CounterVec
	histograms  map[string]*prometheus.HistogramVec
	gauges      map[string]*prometheus.GaugeVec
	mu          sync.RWMutex
	enabled     bool
	metricsPort string
	server      *http.Server
}

// Config configures the metrics collector.
type Config struct {
	Enabled bool   `json:"enabled"`
	Port    string `json:"port"`
}

// NewCollector creates a metrics collector.
func NewCollector(cfg Config) *Collector {
	c := &Collector{
		registry:    prometheus.NewRegistry(),
		counters:    make(map[string]*prometheus.CounterVec),
		histograms:  make(map[string]*prometheus.HistogramVec),
		gauges:      make(map[string]*prometheus.GaugeVec),
		enabled:     cfg.Enabled,
		metricsPort: cfg.Port,
	}
	if !c.enabled {
		return c
	}
	c.registerCounter("iot_messages_received_total", "Total MQTT messages received", []string{"tenant_id", "topic"})
	c.registerCounter("iot_commands_sent_total", "Total commands sent to devices", []string{"tenant_id", "device_type"})
	c.registerCounter("iot_device_registrations_total", "Total device registrations", []string{"tenant_id", "protocol"})
	c.registerGauge("iot_active_devices", "Active device connections", []string{"tenant_id"})
	c.registerCounter("iot_telemetry_batches_total", "Total telemetry batches flushed", []string{"tenant_id"})
	go c.serve()
	return c
}

func (c *Collector) registerCounter(name, help string, labels []string) {
	vec := prometheus.NewCounterVec(prometheus.CounterOpts{Name: name, Help: help}, labels)
	c.registry.MustRegister(vec)
	c.counters[name] = vec
}

func (c *Collector) registerHistogram(name, help string, labels []string) {
	vec := prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: name, Help: help, Buckets: prometheus.DefBuckets}, labels)
	c.registry.MustRegister(vec)
	c.histograms[name] = vec
}

func (c *Collector) registerGauge(name, help string, labels []string) {
	vec := prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: name, Help: help}, labels)
	c.registry.MustRegister(vec)
	c.gauges[name] = vec
}

// IncrementCounter increments a named counter.
func (c *Collector) IncrementCounter(name string, labelValues ...string) {
	if !c.enabled {
		return
	}
	c.mu.RLock()
	vec, ok := c.counters[name]
	c.mu.RUnlock()
	if !ok {
		return
	}
	vec.WithLabelValues(labelValues...).Inc()
}

// HistogramObserve records a histogram observation.
func (c *Collector) HistogramObserve(name string, value float64, labelValues ...string) {
	if !c.enabled {
		return
	}
	c.mu.RLock()
	vec, ok := c.histograms[name]
	c.mu.RUnlock()
	if !ok {
		return
	}
	vec.WithLabelValues(labelValues...).Observe(value)
}

// GaugeSet sets a gauge value.
func (c *Collector) GaugeSet(name string, value float64, labelValues ...string) {
	if !c.enabled {
		return
	}
	c.mu.Lock()
	vec, ok := c.gauges[name]
	if !ok {
		vec = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: name, Help: name}, []string{"tenant_id"})
		c.registry.MustRegister(vec)
		c.gauges[name] = vec
	}
	c.mu.Unlock()
	vec.WithLabelValues(labelValues...).Set(value)
}

func (c *Collector) serve() {
	if c.metricsPort == "" {
		return
	}
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(c.registry, promhttp.HandlerOpts{}))
	c.server = &http.Server{Addr: ":" + c.metricsPort, Handler: mux}
	slog.Info("metrics server starting", slog.String("port", c.metricsPort))
	if err := c.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("metrics server failed", slog.String("error", err.Error()))
	}
}

// Shutdown gracefully stops the metrics HTTP server.
func (c *Collector) Shutdown(ctx context.Context) error {
	if c.server != nil {
		return c.server.Shutdown(ctx)
	}
	return nil
}
