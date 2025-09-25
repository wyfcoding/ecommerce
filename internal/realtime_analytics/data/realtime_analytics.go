package data

import (
	"context"
	"ecommerce/internal/realtime_analytics/biz"
	"ecommerce/internal/realtime_analytics/data/model"
	"fmt"
	"sync"
	"time"
)

// In-memory store for simplicity, simulating a real-time database
var (
	metricsStore = make(map[string]model.AggregatedMetric)
	metricsMutex sync.RWMutex
)

type realtimeAnalyticsRepo struct {
	data *Data // Placeholder for common data dependencies if any
}

// NewRealtimeAnalyticsRepo creates a new RealtimeAnalyticsRepo.
func NewRealtimeAnalyticsRepo(data *Data) biz.RealtimeAnalyticsRepo {
	return &realtimeAnalyticsRepo{data: data}
}

// SaveMetric saves an aggregated metric.
func (r *realtimeAnalyticsRepo) SaveMetric(ctx context.Context, metric *biz.AggregatedMetric) error {
	metricsMutex.Lock()
	defer metricsMutex.Unlock()

	metricsStore[metric.MetricName] = model.AggregatedMetric{
		MetricName: metric.MetricName,
		Value:      metric.Value,
		Timestamp:  metric.Timestamp,
		Labels:     metric.Labels,
	}
	fmt.Printf("Metric saved: %s = %.2f at %s\n", metric.MetricName, metric.Value, metric.Timestamp.Format(time.RFC3339))
	return nil
}

// GetMetric retrieves an aggregated metric by name.
func (r *realtimeAnalyticsRepo) GetMetric(ctx context.Context, metricName string) (*biz.AggregatedMetric, error) {
	metricsMutex.RLock()
	defer metricsMutex.RUnlock()

	metric, ok := metricsStore[metricName]
	if !ok {
		return nil, nil // Metric not found
	}
	return &biz.AggregatedMetric{
		MetricName: metric.MetricName,
		Value:      metric.Value,
		Timestamp:  metric.Timestamp,
		Labels:     metric.Labels,
	},
	nil
}

