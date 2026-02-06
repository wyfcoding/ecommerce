package domain

import (
	"context"
	"time"
)

// LogisticsSearchRepository 定义物流搜索仓储接口（Elasticsearch）。
type LogisticsSearchRepository interface {
	Index(ctx context.Context, logistics *Logistics) error
	Delete(ctx context.Context, id uint64) error
	Search(ctx context.Context, orderID *uint64, trackingNo, carrier *string, status *LogisticsStatus, offset, limit int, startTime, endTime *time.Time, sortBy string) ([]*Logistics, int64, error)
	FindByTrackingNo(ctx context.Context, trackingNo string) (*Logistics, error)
}
