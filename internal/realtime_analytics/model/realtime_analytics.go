package model

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// AggregatedMetric represents a real-time aggregated metric in the business logic layer.
type AggregatedMetric struct {
	MetricName string
	Value      float64
	Timestamp  time.Time
	Labels     map[string]string
}

// RealtimeSalesMetrics represents real-time sales metrics.
type RealtimeSalesMetrics struct {
	CurrentGmv    uint64
	CurrentOrders uint32
	ActiveUsers   uint32
	LastUpdated   *timestamppb.Timestamp
}

// RealtimeUserActivity represents real-time user activity metrics.
type RealtimeUserActivity struct {
	OnlineUsers        uint32
	NewUsersLastHour   uint32
	PageViewsPerMinute map[string]uint32
}
