package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/analytics/domain"
)

// AnalyticsQuery 处理分析模块的查询操作。
type AnalyticsQuery struct {
	repo domain.AnalyticsRepository
}

// NewAnalyticsQuery 创建并返回一个新的 AnalyticsQuery 实例。
func NewAnalyticsQuery(repo domain.AnalyticsRepository) *AnalyticsQuery {
	return &AnalyticsQuery{repo: repo}
}

// GetMetricByID 根据ID获取指标。
func (q *AnalyticsQuery) GetMetricByID(ctx context.Context, id uint64) (*domain.Metric, error) {
	return q.repo.GetMetric(ctx, id)
}

// SearchMetrics 搜索指标。
func (q *AnalyticsQuery) SearchMetrics(ctx context.Context, query *domain.MetricQuery) ([]*domain.Metric, int64, error) {
	return q.repo.ListMetrics(ctx, query)
}

// GetDashboardByID 获取仪表板。
func (q *AnalyticsQuery) GetDashboardByID(ctx context.Context, id uint64) (*domain.Dashboard, error) {
	return q.repo.GetDashboard(ctx, id)
}

// ListUserDashboards 列出用户的仪表板。
func (q *AnalyticsQuery) ListUserDashboards(ctx context.Context, userID uint64, offset, limit int) ([]*domain.Dashboard, int64, error) {
	return q.repo.ListDashboards(ctx, userID, offset, limit)
}

// GetReportByID 获取报告。
func (q *AnalyticsQuery) GetReportByID(ctx context.Context, id uint64) (*domain.Report, error) {
	return q.repo.GetReport(ctx, id)
}

// ListUserReports 列出用户的报告。
func (q *AnalyticsQuery) ListUserReports(ctx context.Context, userID uint64, offset, limit int) ([]*domain.Report, int64, error) {
	return q.repo.ListReports(ctx, userID, offset, limit)
}
