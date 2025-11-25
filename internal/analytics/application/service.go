package application

import (
	"context"
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/analytics/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/analytics/domain/repository"
	"github.com/wyfcoding/ecommerce/pkg/idgen"

	"log/slog"
)

type AnalyticsService struct {
	repo        repository.AnalyticsRepository
	idGenerator idgen.Generator
	logger      *slog.Logger
}

func NewAnalyticsService(repo repository.AnalyticsRepository, idGenerator idgen.Generator, logger *slog.Logger) *AnalyticsService {
	return &AnalyticsService{
		repo:        repo,
		idGenerator: idGenerator,
		logger:      logger,
	}
}

// RecordMetric 记录指标
func (s *AnalyticsService) RecordMetric(ctx context.Context, metricType entity.MetricType, name string, value float64, granularity entity.TimeGranularity, dimension, dimensionVal string) error {
	metric := entity.NewMetric(metricType, name, value, granularity)
	metric.Dimension = dimension
	metric.DimensionVal = dimensionVal

	if err := s.repo.CreateMetric(ctx, metric); err != nil {
		s.logger.Error("failed to create metric", "error", err)
		return err
	}
	return nil
}

// QueryMetrics 查询指标
func (s *AnalyticsService) QueryMetrics(ctx context.Context, query *repository.MetricQuery) ([]*entity.Metric, int64, error) {
	return s.repo.ListMetrics(ctx, query)
}

// CreateDashboard 创建仪表板
func (s *AnalyticsService) CreateDashboard(ctx context.Context, name, description string, userID uint64) (*entity.Dashboard, error) {
	dashboard := entity.NewDashboard(name, description, userID)
	if err := s.repo.CreateDashboard(ctx, dashboard); err != nil {
		s.logger.Error("failed to create dashboard", "error", err)
		return nil, err
	}
	return dashboard, nil
}

// GetDashboard 获取仪表板详情
func (s *AnalyticsService) GetDashboard(ctx context.Context, id uint64) (*entity.Dashboard, error) {
	return s.repo.GetDashboard(ctx, id)
}

// AddMetricToDashboard 添加指标到仪表板
func (s *AnalyticsService) AddMetricToDashboard(ctx context.Context, dashboardID uint64, metricType entity.MetricType, title, chartType string) error {
	dashboard, err := s.repo.GetDashboard(ctx, dashboardID)
	if err != nil {
		return err
	}

	metric := &entity.DashboardMetric{
		DashboardID: dashboardID,
		MetricType:  metricType,
		Title:       title,
		ChartType:   chartType,
	}
	dashboard.AddMetric(metric)

	return s.repo.UpdateDashboard(ctx, dashboard)
}

// PublishDashboard 发布仪表板
func (s *AnalyticsService) PublishDashboard(ctx context.Context, id uint64) error {
	dashboard, err := s.repo.GetDashboard(ctx, id)
	if err != nil {
		return err
	}

	dashboard.Publish()
	return s.repo.UpdateDashboard(ctx, dashboard)
}

// CreateReport 创建报告
func (s *AnalyticsService) CreateReport(ctx context.Context, title, description string, userID uint64, reportType string) (*entity.Report, error) {
	reportNo := fmt.Sprintf("RPT%d", s.idGenerator.Generate())
	report := entity.NewReport(reportNo, title, description, userID, reportType)

	if err := s.repo.CreateReport(ctx, report); err != nil {
		s.logger.Error("failed to create report", "error", err)
		return nil, err
	}
	return report, nil
}

// GetReport 获取报告详情
func (s *AnalyticsService) GetReport(ctx context.Context, id uint64) (*entity.Report, error) {
	return s.repo.GetReport(ctx, id)
}

// PublishReport 发布报告
func (s *AnalyticsService) PublishReport(ctx context.Context, id uint64) error {
	report, err := s.repo.GetReport(ctx, id)
	if err != nil {
		return err
	}

	report.Publish()
	return s.repo.UpdateReport(ctx, report)
}

// ListDashboards 获取仪表板列表
func (s *AnalyticsService) ListDashboards(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.Dashboard, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListDashboards(ctx, userID, offset, pageSize)
}

// ListReports 获取报告列表
func (s *AnalyticsService) ListReports(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.Report, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListReports(ctx, userID, offset, pageSize)
}
