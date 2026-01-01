package domain

import (
	"context"
	"time" // 导入时间包，用于查询条件。
)

// AnalyticsRepository 是分析模块的仓储接口。
// 它定义了对指标、仪表板和报告进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type AnalyticsRepository interface {
	// --- Metric methods ---

	// CreateMetric 在数据存储中创建一个新的指标实体。
	// ctx: 上下文。
	// metric: 待创建的指标实体。
	CreateMetric(ctx context.Context, metric *Metric) error
	// GetMetric 根据ID获取指标实体。
	GetMetric(ctx context.Context, id uint64) (*Metric, error)
	// ListMetrics 列出所有指标实体，支持通过查询条件进行过滤和分页。
	ListMetrics(ctx context.Context, query *MetricQuery) ([]*Metric, int64, error)
	// DeleteMetric 根据ID删除指标实体。
	DeleteMetric(ctx context.Context, id uint64) error

	// --- Dashboard methods ---

	// CreateDashboard 在数据存储中创建一个新的仪表板实体。
	CreateDashboard(ctx context.Context, dashboard *Dashboard) error
	// GetDashboard 根据ID获取仪表板实体，并预加载其关联的指标和过滤器。
	GetDashboard(ctx context.Context, id uint64) (*Dashboard, error)
	// ListDashboards 列出指定用户ID的仪表板实体，支持分页。
	ListDashboards(ctx context.Context, userID uint64, offset, limit int) ([]*Dashboard, int64, error)
	// UpdateDashboard 更新仪表板实体的信息。
	UpdateDashboard(ctx context.Context, dashboard *Dashboard) error
	// DeleteDashboard 根据ID删除仪表板实体。
	DeleteDashboard(ctx context.Context, id uint64) error

	// --- Report methods ---

	// CreateReport 在数据存储中创建一个新的报告实体。
	CreateReport(ctx context.Context, report *Report) error
	// GetReport 根据ID获取报告实体，并预加载其关联的指标。
	GetReport(ctx context.Context, id uint64) (*Report, error)
	// ListReports 列出指定用户ID的报告实体，支持分页。
	ListReports(ctx context.Context, userID uint64, offset, limit int) ([]*Report, int64, error)
	// UpdateReport 更新报告实体的信息。
	UpdateReport(ctx context.Context, report *Report) error
	// DeleteReport 根据ID删除报告实体。
	DeleteReport(ctx context.Context, id uint64) error

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
