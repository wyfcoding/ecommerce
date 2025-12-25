package application

import (
	"context"
	"fmt"
	"time"

	"github.com/wyfcoding/ecommerce/internal/analytics/domain"
	"github.com/wyfcoding/pkg/idgen"
)

// Analytics 结构体定义了数据分析相关的应用服务 (外观模式)。
// 它协调 AnalyticsManager 和 AnalyticsQuery 处理指标、仪表板、报告等业务。
type Analytics struct {
	manager     *AnalyticsManager
	query       *AnalyticsQuery
	idGenerator idgen.Generator
}

// NewAnalytics 创建并返回一个新的 Analytics 实例。
func NewAnalytics(manager *AnalyticsManager, query *AnalyticsQuery, idGenerator idgen.Generator) *Analytics {
	return &Analytics{
		manager:     manager,
		query:       query,
		idGenerator: idGenerator,
	}
}

// RecordMetric 记录一个业务指标数据。
func (s *Analytics) RecordMetric(ctx context.Context, metricType domain.MetricType, name string, value float64, granularity domain.TimeGranularity, dimension, dimensionVal string) error {
	metric := domain.NewMetric(metricType, name, value, granularity)
	metric.Dimension = dimension
	metric.DimensionVal = dimensionVal
	return s.manager.LogMetric(ctx, metric)
}

// QueryMetrics 查询符合条件的指标数据。
func (s *Analytics) QueryMetrics(ctx context.Context, query *domain.MetricQuery) ([]*domain.Metric, int64, error) {
	return s.query.SearchMetrics(ctx, query)
}

// CreateDashboard 创建一个新的仪表板。
func (s *Analytics) CreateDashboard(ctx context.Context, name, description string, userID uint64) (*domain.Dashboard, error) {
	dashboard := domain.NewDashboard(name, description, userID)
	if err := s.manager.CreateDashboard(ctx, dashboard); err != nil {
		return nil, err
	}
	return dashboard, nil
}

// GetDashboard 获取指定ID的仪表板详情。
func (s *Analytics) GetDashboard(ctx context.Context, id uint64) (*domain.Dashboard, error) {
	return s.query.GetDashboardByID(ctx, id)
}

// AddMetricToDashboard 将一个指标添加到指定的仪表板。
func (s *Analytics) AddMetricToDashboard(ctx context.Context, dashboardID uint64, metricType domain.MetricType, title, chartType string) error {
	dashboard, err := s.query.GetDashboardByID(ctx, dashboardID)
	if err != nil {
		return err
	}
	if dashboard == nil {
		return fmt.Errorf("dashboard not found")
	}

	metric := &domain.DashboardMetric{
		DashboardID: dashboardID,
		MetricType:  metricType,
		Title:       title,
		ChartType:   chartType,
	}
	dashboard.AddMetric(metric)
	return s.manager.UpdateDashboard(ctx, dashboard)
}

// PublishDashboard 发布指定ID的仪表板。
func (s *Analytics) PublishDashboard(ctx context.Context, id uint64) error {
	dashboard, err := s.query.GetDashboardByID(ctx, id)
	if err != nil {
		return err
	}
	if dashboard == nil {
		return fmt.Errorf("dashboard not found")
	}

	dashboard.Publish()
	return s.manager.UpdateDashboard(ctx, dashboard)
}

// CreateReport 创建一个新的数据报告。
func (s *Analytics) CreateReport(ctx context.Context, title, description string, userID uint64, reportType string) (*domain.Report, error) {
	reportNo := fmt.Sprintf("RPT%d", s.idGenerator.Generate())
	report := domain.NewReport(reportNo, title, description, userID, reportType)
	if err := s.manager.SaveReport(ctx, report); err != nil {
		return nil, err
	}
	return report, nil
}

// GetReport 获取指定ID的数据报告详情。
func (s *Analytics) GetReport(ctx context.Context, id uint64) (*domain.Report, error) {
	return s.query.GetReportByID(ctx, id)
}

// PublishReport 发布指定ID的数据报告。
func (s *Analytics) PublishReport(ctx context.Context, id uint64) error {
	report, err := s.query.GetReportByID(ctx, id)
	if err != nil {
		return err
	}
	if report == nil {
		return fmt.Errorf("report not found")
	}

	report.Publish()
	return s.manager.SaveReport(ctx, report)
}

// ListDashboards 获取用户的仪表板列表。
func (s *Analytics) ListDashboards(ctx context.Context, userID uint64, page, pageSize int) ([]*domain.Dashboard, int64, error) {
	offset := (page - 1) * pageSize
	return s.query.ListUserDashboards(ctx, userID, offset, pageSize)
}

// ListReports 获取用户的数据报告列表。
func (s *Analytics) ListReports(ctx context.Context, userID uint64, page, pageSize int) ([]*domain.Report, int64, error) {
	offset := (page - 1) * pageSize
	return s.query.ListUserReports(ctx, userID, offset, pageSize)
}

// UpdateDashboard 更新仪表板的基础信息。
func (s *Analytics) UpdateDashboard(ctx context.Context, id uint64, name, description string) (*domain.Dashboard, error) {
	dashboard, err := s.query.GetDashboardByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if dashboard == nil {
		return nil, fmt.Errorf("dashboard not found")
	}

	if name != "" {
		dashboard.Name = name
	}
	if description != "" {
		dashboard.Description = description
	}

	if err := s.manager.UpdateDashboard(ctx, dashboard); err != nil {
		return nil, err
	}
	return dashboard, nil
}

// DeleteDashboard 删除指定的仪表板。
func (s *Analytics) DeleteDashboard(ctx context.Context, id uint64) error {
	return s.manager.DeleteDashboard(ctx, id)
}

// UpdateReport 更新报告的基础信息。
func (s *Analytics) UpdateReport(ctx context.Context, id uint64, title, description string) (*domain.Report, error) {
	report, err := s.query.GetReportByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if report == nil {
		return nil, fmt.Errorf("report not found")
	}

	if title != "" {
		report.Title = title
	}
	if description != "" {
		report.Description = description
	}

	if err := s.manager.SaveReport(ctx, report); err != nil {
		return nil, err
	}
	return report, nil
}

// DeleteReport 删除指定的数据报告。
func (s *Analytics) DeleteReport(ctx context.Context, id uint64) error {
	return s.manager.DeleteReport(ctx, id)
}

// GetUserActivityReport 获取用户活动概览报告。
func (s *Analytics) GetUserActivityReport(ctx context.Context, startTime, endTime time.Time) (map[string]any, error) {
	query := &domain.MetricQuery{
		MetricType: domain.MetricTypeActiveUsers,
		StartTime:  startTime,
		EndTime:    endTime,
	}
	metrics, _, err := s.query.SearchMetrics(ctx, query)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"active_users": len(metrics),
	}, nil
}

// GetProductPerformanceReport 获取商品销售性能分析报告。
func (s *Analytics) GetProductPerformanceReport(ctx context.Context, startTime, endTime time.Time) (map[string]any, error) {
	return map[string]any{"top_products": []string{}}, nil
}

// GetConversionFunnelReport 获取用户转化漏斗分析报告。
func (s *Analytics) GetConversionFunnelReport(ctx context.Context, startTime, endTime time.Time) (map[string]any, error) {
	return map[string]any{"funnel_steps": []string{}}, nil
}

// GetCustomReport 获取自定义报告数据。
func (s *Analytics) GetCustomReport(ctx context.Context, reportID uint64, startTime, endTime time.Time) (map[string]any, error) {
	return map[string]any{"custom_data": "data"}, nil
}

// GetUserBehaviorPath 获取用户的行为路径追踪数据。
func (s *Analytics) GetUserBehaviorPath(ctx context.Context, userID uint64, startTime, endTime time.Time) (map[string]any, error) {
	return map[string]any{"path": []string{}}, nil
}

// GetUserSegments 获取用户细分群体分析数据。
func (s *Analytics) GetUserSegments(ctx context.Context) (map[string]any, error) {
	return map[string]any{"segments": []string{}}, nil
}
