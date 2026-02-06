package domain

import (
	"context"
	"time"
)

// AnalyticsRepository 是分析模块的仓储接口。
// 它定义了对指标、仪表板和报告进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type AnalyticsRepository interface {
	// 事务支持
	BeginTx(ctx context.Context) any
	CommitTx(tx any) error
	RollbackTx(tx any) error
	WithTx(ctx context.Context, fn func(tx any) error) error

	// --- Metric methods ---
	SaveMetric(ctx context.Context, metric *Metric) error
	SaveMetricInTx(ctx context.Context, tx any, metric *Metric) error
	GetMetric(ctx context.Context, id uint64) (*Metric, error)
	ListMetrics(ctx context.Context, query *MetricQuery) ([]*Metric, int64, error)
	DeleteMetric(ctx context.Context, id uint64) error
	DeleteMetricInTx(ctx context.Context, tx any, id uint64) error

	// --- Dashboard methods ---
	SaveDashboard(ctx context.Context, dashboard *Dashboard) error
	SaveDashboardInTx(ctx context.Context, tx any, dashboard *Dashboard) error
	GetDashboard(ctx context.Context, id uint64) (*Dashboard, error)
	ListDashboards(ctx context.Context, userID uint64, offset, limit int) ([]*Dashboard, int64, error)
	DeleteDashboard(ctx context.Context, id uint64) error
	DeleteDashboardInTx(ctx context.Context, tx any, id uint64) error

	// --- Report methods ---
	SaveReport(ctx context.Context, report *Report) error
	SaveReportInTx(ctx context.Context, tx any, report *Report) error
	GetReport(ctx context.Context, id uint64) (*Report, error)
	ListReports(ctx context.Context, userID uint64, offset, limit int) ([]*Report, int64, error)
	DeleteReport(ctx context.Context, id uint64) error
	DeleteReportInTx(ctx context.Context, tx any, id uint64) error

	// GetActivePages 获取最近活跃页面。
	GetActivePages(ctx context.Context, limit int) ([]string, error)
}

// MetricQuery 结构体定义了查询指标的条件。
// 它用于在仓储层进行数据过滤和分页。
type MetricQuery struct {
	MetricType   MetricType      // 根据指标类型过滤。
	Granularity  TimeGranularity // 根据时间粒度过滤。
	Dimension    string          // 根据维度过滤。
	DimensionVal string          // 根据维度值过滤。
	StartTime    time.Time       // 查询的起始时间。
	EndTime      time.Time       // 查询的结束时间。
	Page         int             // 页码，用于分页。
	PageSize     int             // 每页数量，用于分页。
}
