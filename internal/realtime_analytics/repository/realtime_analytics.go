package repository

import (
	"context"

	"ecommerce/internal/realtime_analytics/model"
)

// RealtimeAnalyticsRepo defines the interface for real-time analytics data access.
type RealtimeAnalyticsRepo interface {
	SaveMetric(ctx context.Context, metric *model.AggregatedMetric) error
	GetMetric(ctx context.Context, metricName string) (*model.AggregatedMetric, error)
}