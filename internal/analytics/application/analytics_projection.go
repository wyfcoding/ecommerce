// 生成摘要：新增分析读模型投影服务，消费事件后刷新 Redis/ES 读侧。
package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/analytics/domain"
)

// AnalyticsProjectionService 负责将分析事件投影到读模型。
type AnalyticsProjectionService struct {
	repo              domain.AnalyticsRepository
	metricReadRepo    domain.MetricReadRepository
	dashboardReadRepo domain.DashboardReadRepository
	reportReadRepo    domain.ReportReadRepository
	metricSearchRepo  domain.MetricSearchRepository
	logger            *slog.Logger
}

// NewAnalyticsProjectionService 创建投影服务。
func NewAnalyticsProjectionService(
	repo domain.AnalyticsRepository,
	metricReadRepo domain.MetricReadRepository,
	dashboardReadRepo domain.DashboardReadRepository,
	reportReadRepo domain.ReportReadRepository,
	metricSearchRepo domain.MetricSearchRepository,
	logger *slog.Logger,
) *AnalyticsProjectionService {
	return &AnalyticsProjectionService{
		repo:              repo,
		metricReadRepo:    metricReadRepo,
		dashboardReadRepo: dashboardReadRepo,
		reportReadRepo:    reportReadRepo,
		metricSearchRepo:  metricSearchRepo,
		logger:            logger,
	}
}

func (s *AnalyticsProjectionService) OnMetricRecorded(ctx context.Context, event *domain.MetricRecordedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshMetric(ctx, event.MetricID)
}

func (s *AnalyticsProjectionService) OnMetricDeleted(ctx context.Context, event *domain.MetricDeletedEvent) error {
	if event == nil {
		return nil
	}
	if s.metricReadRepo != nil {
		_ = s.metricReadRepo.Delete(ctx, event.MetricID)
	}
	if s.metricSearchRepo != nil {
		_ = s.metricSearchRepo.Delete(ctx, event.MetricID)
	}
	return nil
}

func (s *AnalyticsProjectionService) OnDashboardCreated(ctx context.Context, event *domain.DashboardCreatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshDashboard(ctx, event.DashboardID)
}

func (s *AnalyticsProjectionService) OnDashboardUpdated(ctx context.Context, event *domain.DashboardUpdatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshDashboard(ctx, event.DashboardID)
}

func (s *AnalyticsProjectionService) OnDashboardDeleted(ctx context.Context, event *domain.DashboardDeletedEvent) error {
	if event == nil {
		return nil
	}
	if s.dashboardReadRepo != nil {
		_ = s.dashboardReadRepo.Delete(ctx, event.DashboardID)
	}
	return nil
}

func (s *AnalyticsProjectionService) OnReportCreated(ctx context.Context, event *domain.ReportCreatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshReport(ctx, event.ReportID)
}

func (s *AnalyticsProjectionService) OnReportUpdated(ctx context.Context, event *domain.ReportUpdatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshReport(ctx, event.ReportID)
}

func (s *AnalyticsProjectionService) OnReportPublished(ctx context.Context, event *domain.ReportPublishedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshReport(ctx, event.ReportID)
}

func (s *AnalyticsProjectionService) OnReportDeleted(ctx context.Context, event *domain.ReportDeletedEvent) error {
	if event == nil {
		return nil
	}
	if s.reportReadRepo != nil {
		_ = s.reportReadRepo.Delete(ctx, event.ReportID)
	}
	return nil
}

func (s *AnalyticsProjectionService) refreshMetric(ctx context.Context, metricID uint64) error {
	if s.metricReadRepo == nil && s.metricSearchRepo == nil {
		return nil
	}
	metric, err := s.repo.GetMetric(ctx, metricID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load metric for projection", "metric_id", metricID, "error", err)
		return err
	}
	if metric == nil {
		if s.metricReadRepo != nil {
			_ = s.metricReadRepo.Delete(ctx, metricID)
		}
		if s.metricSearchRepo != nil {
			_ = s.metricSearchRepo.Delete(ctx, metricID)
		}
		return nil
	}
	if s.metricReadRepo != nil {
		if err := s.metricReadRepo.Save(ctx, metric); err != nil {
			s.logger.ErrorContext(ctx, "failed to save metric cache", "metric_id", metricID, "error", err)
			return err
		}
	}
	if s.metricSearchRepo != nil {
		if err := s.metricSearchRepo.Index(ctx, metric); err != nil {
			s.logger.ErrorContext(ctx, "failed to index metric", "metric_id", metricID, "error", err)
			return err
		}
	}
	return nil
}

func (s *AnalyticsProjectionService) refreshDashboard(ctx context.Context, dashboardID uint64) error {
	if s.dashboardReadRepo == nil {
		return nil
	}
	dashboard, err := s.repo.GetDashboard(ctx, dashboardID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load dashboard for projection", "dashboard_id", dashboardID, "error", err)
		return err
	}
	if dashboard == nil {
		_ = s.dashboardReadRepo.Delete(ctx, dashboardID)
		return nil
	}
	if err := s.dashboardReadRepo.Save(ctx, dashboard); err != nil {
		s.logger.ErrorContext(ctx, "failed to save dashboard cache", "dashboard_id", dashboardID, "error", err)
		return err
	}
	return nil
}

func (s *AnalyticsProjectionService) refreshReport(ctx context.Context, reportID uint64) error {
	if s.reportReadRepo == nil {
		return nil
	}
	report, err := s.repo.GetReport(ctx, reportID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load report for projection", "report_id", reportID, "error", err)
		return err
	}
	if report == nil {
		_ = s.reportReadRepo.Delete(ctx, reportID)
		return nil
	}
	if err := s.reportReadRepo.Save(ctx, report); err != nil {
		s.logger.ErrorContext(ctx, "failed to save report cache", "report_id", reportID, "error", err)
		return err
	}
	return nil
}
