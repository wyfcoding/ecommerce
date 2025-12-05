package application

import (
	"context"
	"fmt"
	"time"

	"github.com/wyfcoding/ecommerce/internal/analytics/domain/entity"     // 导入分析模块的领域实体。
	"github.com/wyfcoding/ecommerce/internal/analytics/domain/repository" // 导入分析模块的仓储接口。
	"github.com/wyfcoding/ecommerce/pkg/idgen"                            // 导入ID生成器接口。

	"log/slog" // 导入结构化日志库。
)

// AnalyticsService 结构体定义了数据分析相关的应用服务。
// 它协调领域层和基础设施层，处理指标记录、仪表板管理、报告生成等业务逻辑。
type AnalyticsService struct {
	repo        repository.AnalyticsRepository // 依赖AnalyticsRepository接口，用于数据持久化操作。
	idGenerator idgen.Generator                // 依赖ID生成器接口，用于生成唯一的报告编号。
	logger      *slog.Logger                   // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewAnalyticsService 创建并返回一个新的 AnalyticsService 实例。
func NewAnalyticsService(repo repository.AnalyticsRepository, idGenerator idgen.Generator, logger *slog.Logger) *AnalyticsService {
	return &AnalyticsService{
		repo:        repo,
		idGenerator: idGenerator,
		logger:      logger,
	}
}

// RecordMetric 记录一个业务指标数据。
// ctx: 上下文。
// metricType: 指标类型（例如，计数器、仪表盘）。
// name: 指标名称。
// value: 指标值。
// granularity: 时间粒度（例如，小时、天）。
// dimension, dimensionVal: 指标的维度和维度值（例如，Dimension="region", DimensionVal="east"）。
// 返回可能发生的错误。
func (s *AnalyticsService) RecordMetric(ctx context.Context, metricType entity.MetricType, name string, value float64, granularity entity.TimeGranularity, dimension, dimensionVal string) error {
	metric := entity.NewMetric(metricType, name, value, granularity) // 创建Metric实体。
	metric.Dimension = dimension
	metric.DimensionVal = dimensionVal

	// 通过仓储接口保存指标。
	if err := s.repo.CreateMetric(ctx, metric); err != nil {
		s.logger.ErrorContext(ctx, "failed to create metric", "name", name, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "metric recorded successfully", "name", name, "value", value)
	return nil
}

// QueryMetrics 查询符合条件的指标数据。
// ctx: 上下文。
// query: 包含过滤条件和分页参数的查询对象。
// 返回指标列表、总数和可能发生的错误。
func (s *AnalyticsService) QueryMetrics(ctx context.Context, query *repository.MetricQuery) ([]*entity.Metric, int64, error) {
	return s.repo.ListMetrics(ctx, query)
}

// CreateDashboard 创建一个新的仪表板。
// ctx: 上下文。
// name: 仪表板名称。
// description: 仪表板描述。
// userID: 创建仪表板的用户ID。
// 返回创建成功的Dashboard实体和可能发生的错误。
func (s *AnalyticsService) CreateDashboard(ctx context.Context, name, description string, userID uint64) (*entity.Dashboard, error) {
	dashboard := entity.NewDashboard(name, description, userID) // 创建Dashboard实体。
	if err := s.repo.CreateDashboard(ctx, dashboard); err != nil {
		s.logger.ErrorContext(ctx, "failed to create dashboard", "name", name, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "dashboard created successfully", "dashboard_id", dashboard.ID, "name", name)
	return dashboard, nil
}

// GetDashboard 获取指定ID的仪表板详情。
// ctx: 上下文。
// id: 仪表板ID。
// 返回Dashboard实体和可能发生的错误。
func (s *AnalyticsService) GetDashboard(ctx context.Context, id uint64) (*entity.Dashboard, error) {
	return s.repo.GetDashboard(ctx, id)
}

// AddMetricToDashboard 将一个指标添加到指定的仪表板。
// ctx: 上下文。
// dashboardID: 仪表板ID。
// metricType: 要添加的指标类型。
// title: 指标的标题。
// chartType: 指标的图表类型。
// 返回可能发生的错误。
func (s *AnalyticsService) AddMetricToDashboard(ctx context.Context, dashboardID uint64, metricType entity.MetricType, title, chartType string) error {
	// 获取仪表板实体。
	dashboard, err := s.repo.GetDashboard(ctx, dashboardID)
	if err != nil {
		return err
	}

	// 创建DashboardMetric实体。
	metric := &entity.DashboardMetric{
		DashboardID: dashboardID,
		MetricType:  metricType,
		Title:       title,
		ChartType:   chartType,
	}
	// 调用实体方法将指标添加到仪表板。
	dashboard.AddMetric(metric)

	// 更新数据库中的仪表板。
	return s.repo.UpdateDashboard(ctx, dashboard)
}

// PublishDashboard 发布指定ID的仪表板。
// ctx: 上下文。
// id: 仪表板ID。
// 返回可能发生的错误。
func (s *AnalyticsService) PublishDashboard(ctx context.Context, id uint64) error {
	// 获取仪表板实体。
	dashboard, err := s.repo.GetDashboard(ctx, id)
	if err != nil {
		return err
	}

	// 调用实体方法发布仪表板。
	dashboard.Publish()
	// 更新数据库中的仪表板状态。
	return s.repo.UpdateDashboard(ctx, dashboard)
}

// CreateReport 创建一个新的数据报告。
// ctx: 上下文。
// title: 报告标题。
// description: 报告描述。
// userID: 创建报告的用户ID。
// reportType: 报告类型。
// 返回创建成功的Report实体和可能发生的错误。
func (s *AnalyticsService) CreateReport(ctx context.Context, title, description string, userID uint64, reportType string) (*entity.Report, error) {
	// 生成唯一的报告编号。
	reportNo := fmt.Sprintf("RPT%d", s.idGenerator.Generate())
	// 创建Report实体。
	report := entity.NewReport(reportNo, title, description, userID, reportType)

	// 通过仓储接口保存报告。
	if err := s.repo.CreateReport(ctx, report); err != nil {
		s.logger.ErrorContext(ctx, "failed to create report", "title", title, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "report created successfully", "report_id", report.ID, "title", title)
	return report, nil
}

// GetReport 获取指定ID的数据报告详情。
// ctx: 上下文。
// id: 报告ID。
// 返回Report实体和可能发生的错误。
func (s *AnalyticsService) GetReport(ctx context.Context, id uint64) (*entity.Report, error) {
	return s.repo.GetReport(ctx, id)
}

// PublishReport 发布指定ID的数据报告。
// ctx: 上下文。
// id: 报告ID。
// 返回可能发生的错误。
func (s *AnalyticsService) PublishReport(ctx context.Context, id uint64) error {
	// 获取报告实体。
	report, err := s.repo.GetReport(ctx, id)
	if err != nil {
		return err
	}

	// 调用实体方法发布报告。
	report.Publish()
	// 更新数据库中的报告状态。
	return s.repo.UpdateReport(ctx, report)
}

// ListDashboards 获取仪表板列表，支持分页和用户过滤。
// ctx: 上下文。
// userID: 筛选仪表板的用户ID。
// page, pageSize: 分页参数。
// 返回Dashboard列表、总数和可能发生的错误。
func (s *AnalyticsService) ListDashboards(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.Dashboard, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListDashboards(ctx, userID, offset, pageSize)
}

// ListReports 获取数据报告列表，支持分页和用户过滤。
// ctx: 上下文。
// userID: 筛选报告的用户ID。
// page, pageSize: 分页参数。
// 返回Report列表、总数和可能发生的错误。
func (s *AnalyticsService) ListReports(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.Report, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListReports(ctx, userID, offset, pageSize)
}

// UpdateDashboard 更新仪表板信息。
func (s *AnalyticsService) UpdateDashboard(ctx context.Context, id uint64, name, description string) (*entity.Dashboard, error) {
	dashboard, err := s.repo.GetDashboard(ctx, id)
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

	if err := s.repo.UpdateDashboard(ctx, dashboard); err != nil {
		return nil, err
	}
	return dashboard, nil
}

// DeleteDashboard 删除仪表板。
func (s *AnalyticsService) DeleteDashboard(ctx context.Context, id uint64) error {
	return s.repo.DeleteDashboard(ctx, id)
}

// UpdateReport 更新报告信息。
func (s *AnalyticsService) UpdateReport(ctx context.Context, id uint64, title, description string) (*entity.Report, error) {
	report, err := s.repo.GetReport(ctx, id)
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

	if err := s.repo.UpdateReport(ctx, report); err != nil {
		return nil, err
	}
	return report, nil
}

// DeleteReport 删除报告。
func (s *AnalyticsService) DeleteReport(ctx context.Context, id uint64) error {
	return s.repo.DeleteReport(ctx, id)
}

// GetUserActivityReport 获取用户活动报告。
// 这里是一个简化的实现，实际可能需要复杂的聚合查询。
func (s *AnalyticsService) GetUserActivityReport(ctx context.Context, startTime, endTime time.Time) (map[string]interface{}, error) {
	// 示例：查询活跃用户数
	query := &repository.MetricQuery{
		MetricType: entity.MetricTypeActiveUsers,
		StartTime:  startTime,
		EndTime:    endTime,
	}
	metrics, _, err := s.repo.ListMetrics(ctx, query)
	if err != nil {
		return nil, err
	}

	// 聚合逻辑...
	return map[string]interface{}{
		"active_users": len(metrics), // 仅作为示例
	}, nil
}

// GetProductPerformanceReport 获取产品性能报告。
func (s *AnalyticsService) GetProductPerformanceReport(ctx context.Context, startTime, endTime time.Time) (map[string]interface{}, error) {
	// 示例逻辑
	return map[string]interface{}{
		"top_products": []string{},
	}, nil
}

// GetConversionFunnelReport 获取转化漏斗报告。
func (s *AnalyticsService) GetConversionFunnelReport(ctx context.Context, startTime, endTime time.Time) (map[string]interface{}, error) {
	// 示例逻辑
	return map[string]interface{}{
		"funnel_steps": []string{},
	}, nil
}

// GetCustomReport 获取自定义报告。
func (s *AnalyticsService) GetCustomReport(ctx context.Context, reportID uint64, startTime, endTime time.Time) (map[string]interface{}, error) {
	// 示例逻辑
	return map[string]interface{}{
		"custom_data": "data",
	}, nil
}

// GetUserBehaviorPath 获取用户行为路径。
func (s *AnalyticsService) GetUserBehaviorPath(ctx context.Context, userID uint64, startTime, endTime time.Time) (map[string]interface{}, error) {
	// 示例逻辑
	return map[string]interface{}{
		"path": []string{},
	}, nil
}

// GetUserSegments 获取用户细分。
func (s *AnalyticsService) GetUserSegments(ctx context.Context) (map[string]interface{}, error) {
	// 示例逻辑
	return map[string]interface{}{
		"segments": []string{},
	}, nil
}
