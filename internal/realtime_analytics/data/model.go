package data

import (
	"time"
)

// AggregatedMetric represents a real-time aggregated metric.
type AggregatedMetric struct {
	MetricName string            `json:"metric_name"` // e.g., "current_gmv", "active_users"
	Value      float64           `json:"value"`
	Timestamp  time.Time         `json:"timestamp"`
	Labels     map[string]string `json:"labels,omitempty"` // e.g., {"region": "north"}
}
