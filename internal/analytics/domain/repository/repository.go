package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/analytics/domain/entity"
	"time"
)

// AnalyticsRepository 分析服务仓储接口
type AnalyticsRepository interface {
	// Metric methods
	CreateMetric(ctx context.Context, metric *entity.Metric) error
	GetMetric(ctx context.Context, id uint64) (*entity.Metric, error)
	ListMetrics(ctx context.Context, query *MetricQuery) ([]*entity.Metric, int64, error)
	DeleteMetric(ctx context.Context, id uint64) error

	// Dashboard methods
	CreateDashboard(ctx context.Context, dashboard *entity.Dashboard) error
	GetDashboard(ctx context.Context, id uint64) (*entity.Dashboard, error)
	ListDashboards(ctx context.Context, userID uint64, offset, limit int) ([]*entity.Dashboard, int64, error)
	UpdateDashboard(ctx context.Context, dashboard *entity.Dashboard) error
	DeleteDashboard(ctx context.Context, id uint64) error

	// Report methods
	CreateReport(ctx context.Context, report *entity.Report) error
	GetReport(ctx context.Context, id uint64) (*entity.Report, error)
	ListReports(ctx context.Context, userID uint64, offset, limit int) ([]*entity.Report, int64, error)
	UpdateReport(ctx context.Context, report *entity.Report) error
	DeleteReport(ctx context.Context, id uint64) error
}

// MetricQuery 指标查询条件
type MetricQuery struct {
	MetricType  entity.MetricType
	Granularity entity.TimeGranularity
	Dimension   string
	StartTime   time.Time
	EndTime     time.Time
	Page        int
	PageSize    int
}
