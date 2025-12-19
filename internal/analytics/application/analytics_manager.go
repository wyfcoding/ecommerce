package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/analytics/domain"
)

// AnalyticsManager 处理分析模块的写操作和业务逻辑。
type AnalyticsManager struct {
	repo   domain.AnalyticsRepository
	logger *slog.Logger
}

// NewAnalyticsManager 创建并返回一个新的 AnalyticsManager 实例。
func NewAnalyticsManager(repo domain.AnalyticsRepository, logger *slog.Logger) *AnalyticsManager {
	return &AnalyticsManager{
		repo:   repo,
		logger: logger,
	}
}

// LogMetric 记录一个指标。
func (m *AnalyticsManager) LogMetric(ctx context.Context, metric *domain.Metric) error {
	if err := m.repo.CreateMetric(ctx, metric); err != nil {
		m.logger.Error("failed to create metric", "error", err, "type", metric.MetricType)
		return err
	}
	return nil
}

// CreateDashboard 创建仪表板。
func (m *AnalyticsManager) CreateDashboard(ctx context.Context, dashboard *domain.Dashboard) error {
	if err := m.repo.CreateDashboard(ctx, dashboard); err != nil {
		m.logger.Error("failed to create dashboard", "error", err, "name", dashboard.Name)
		return err
	}
	return nil
}

// UpdateDashboard 更新仪表板。
func (m *AnalyticsManager) UpdateDashboard(ctx context.Context, dashboard *domain.Dashboard) error {
	if err := m.repo.UpdateDashboard(ctx, dashboard); err != nil {
		m.logger.Error("failed to update dashboard", "error", err, "id", dashboard.ID)
		return err
	}
	return nil
}

// DeleteDashboard 删除仪表板。
func (m *AnalyticsManager) DeleteDashboard(ctx context.Context, id uint64) error {
	if err := m.repo.DeleteDashboard(ctx, id); err != nil {
		m.logger.Error("failed to delete dashboard", "error", err, "id", id)
		return err
	}
	return nil
}

// SaveReport 保存分析报告。
func (m *AnalyticsManager) SaveReport(ctx context.Context, report *domain.Report) error {
	if report.ID == 0 {
		if err := m.repo.CreateReport(ctx, report); err != nil {
			m.logger.Error("failed to create report", "error", err, "no", report.ReportNo)
			return err
		}
	} else {
		if err := m.repo.UpdateReport(ctx, report); err != nil {
			m.logger.Error("failed to update report", "error", err, "id", report.ID)
			return err
		}
	}
	return nil
}

// DeleteReport 删除报告。
func (m *AnalyticsManager) DeleteReport(ctx context.Context, id uint64) error {
	if err := m.repo.DeleteReport(ctx, id); err != nil {
		m.logger.Error("failed to delete report", "error", err, "id", id)
		return err
	}
	return nil
}
