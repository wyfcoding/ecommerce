package metrics

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds a custom Prometheus registry and provides methods to create metrics.
type Metrics struct {
	registry *prometheus.Registry
}

// NewMetrics creates a new Metrics instance with its own registry.
func NewMetrics(serviceName string) *Metrics {
	registry := prometheus.NewRegistry()
	registry.MustRegister(prometheus.NewGoCollector())                                       // Collect default Go metrics
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{})) // Collect process metrics

	slog.Info("Metrics registry created for service", "service", serviceName)

	return &Metrics{registry: registry}
}

// NewCounterVec creates, registers, and returns a new CounterVec.
func (m *Metrics) NewCounterVec(opts prometheus.CounterOpts, labelNames []string) *prometheus.CounterVec {
	counter := prometheus.NewCounterVec(opts, labelNames)
	m.registry.MustRegister(counter)
	return counter
}

// NewGaugeVec creates, registers, and returns a new GaugeVec.
func (m *Metrics) NewGaugeVec(opts prometheus.GaugeOpts, labelNames []string) *prometheus.GaugeVec {
	gauge := prometheus.NewGaugeVec(opts, labelNames)
	m.registry.MustRegister(gauge)
	return gauge
}

// NewHistogramVec creates, registers, and returns a new HistogramVec.
func (m *Metrics) NewHistogramVec(opts prometheus.HistogramOpts, labelNames []string) *prometheus.HistogramVec {
	histogram := prometheus.NewHistogramVec(opts, labelNames)
	m.registry.MustRegister(histogram)
	return histogram
}

// ExposeHttp starts a new HTTP server to expose the metrics from the registry.
// It returns a cleanup function to gracefully shut down the server.
func (m *Metrics) ExposeHttp(port string) func() {
	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{}),
	}

	go func() {
		slog.Info("Metrics server listening", "port", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("failed to start metrics server", "error", err)
		}
	}()

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		slog.Info("shutting down metrics server...")
		if err := httpServer.Shutdown(ctx); err != nil {
			slog.Error("failed to gracefully shutdown metrics server", "error", err)
		}
	}

	return cleanup
}
