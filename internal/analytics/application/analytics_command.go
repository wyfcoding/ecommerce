package application

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/wyfcoding/ecommerce/internal/analytics/domain"
	"github.com/wyfcoding/pkg/algorithm/graph"
	"github.com/wyfcoding/pkg/algorithm/sim"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/messagequeue"
	"github.com/wyfcoding/pkg/redis"
)

// AnalyticsCommandService 处理分析模块的写操作和业务逻辑。
// 引入树状数组（FenwickTree）用于实时、高频的订单金额和数量统计。
type AnalyticsCommandService struct {
	repo        domain.AnalyticsRepository
	publisher   messagequeue.EventPublisher
	logger      *slog.Logger
	redisClient redis.Client
	idGenerator idgen.Generator
	gmvStats    *graph.FenwickTree            // 用于统计 24 小时内每一分钟的 GMV
	orderStats  *graph.FenwickTree            // 用于统计 24 小时内每一分钟的订单数
	sampler     *sim.ReservoirSampler[string] // 用于对原始事件流进行抽样
	samplerMu   sync.Mutex
}

// NewAnalyticsCommandService 创建并返回一个新的 AnalyticsCommandService 实例。
func NewAnalyticsCommandService(repo domain.AnalyticsRepository, publisher messagequeue.EventPublisher, idGenerator idgen.Generator, redisClient redis.Client, logger *slog.Logger) *AnalyticsCommandService {
	return &AnalyticsCommandService{
		repo:        repo,
		publisher:   publisher,
		logger:      logger,
		redisClient: redisClient,
		idGenerator: idGenerator,
		gmvStats:    graph.NewFenwickTree(1440), // 一天 1440 分钟
		orderStats:  graph.NewFenwickTree(1440),
		sampler:     sim.NewReservoirSampler[string](1000), // 采样 1000 条
	}
}

// LogAndSampleEvent 记录事件并进行抽样
func (m *AnalyticsCommandService) LogAndSampleEvent(event string) {
	m.samplerMu.Lock()
	defer m.samplerMu.Unlock()
	m.sampler.Observe(event)
}

// GetEventSamples 获取当前的事件样本
func (m *AnalyticsCommandService) GetEventSamples() []string {
	m.samplerMu.Lock()
	defer m.samplerMu.Unlock()
	return m.sampler.GetSamples()
}

// TrackUserVisit 追踪用户访问 (统计 DAU)
func (m *AnalyticsCommandService) TrackUserVisit(ctx context.Context, userID uint64) {
	today := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("analytics:uv:%s", today)

	if err := m.redisClient.PFAdd(ctx, key, userID).Err(); err != nil {
		m.logger.ErrorContext(ctx, "failed to track user visit", "user_id", userID, "error", err)
	}
}

// GetDailyUV 获取今日去重访问量
func (m *AnalyticsCommandService) GetDailyUV(ctx context.Context) (int64, error) {
	today := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("analytics:uv:%s", today)

	return m.redisClient.PFCount(ctx, key).Result()
}

// TrackRealtimeOrder 实时追踪订单数据。
// 利用树状数组的 O(log N) 更新特性，支持极高并发的实时统计。
func (m *AnalyticsCommandService) TrackRealtimeOrder(ctx context.Context, amount int64, timestamp time.Time) {
	minute := timestamp.Hour()*60 + timestamp.Minute()

	m.gmvStats.Update(minute, amount)
	m.orderStats.Update(minute, 1)

	m.logger.DebugContext(ctx, "realtime order tracked", "minute", minute, "amount", amount)
}

// GetHourlyStats 获取指定小时的聚合统计数据。
// 利用树状数组的区间查询 O(log N) 特性，快速获取结果。
func (m *AnalyticsCommandService) GetHourlyStats(_ context.Context, hour int) (int64, int64) {
	if hour < 0 || hour > 23 {
		return 0, 0
	}

	startMin := hour * 60
	endMin := startMin + 59

	totalGMV := m.gmvStats.RangeQuery(startMin, endMin)
	totalOrders := m.orderStats.RangeQuery(startMin, endMin)

	return totalGMV, totalOrders
}

// RecordMetric 记录一个业务指标数据。
func (m *AnalyticsCommandService) RecordMetric(ctx context.Context, metricType domain.MetricType, name string, value float64, granularity domain.TimeGranularity, dimension, dimensionVal string) error {
	metric := domain.NewMetric(metricType, name, value, granularity)
	metric.Dimension = dimension
	metric.DimensionVal = dimensionVal

	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveMetricInTx(ctx, tx, metric); err != nil {
			m.logger.ErrorContext(ctx, "failed to create metric", "error", err, "type", metric.MetricType)
			return err
		}
		if m.publisher == nil {
			return nil
		}
		event := &domain.MetricRecordedEvent{
			MetricID:  uint64(metric.ID),
			Timestamp: time.Now(),
		}
		return m.publisher.PublishInTx(ctx, tx, domain.MetricRecordedEventType, fmt.Sprintf("%d", metric.ID), event)
	})
}

// DeleteMetric 删除指标。
func (m *AnalyticsCommandService) DeleteMetric(ctx context.Context, id uint64) error {
	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.DeleteMetricInTx(ctx, tx, id); err != nil {
			return err
		}
		if m.publisher == nil {
			return nil
		}
		event := &domain.MetricDeletedEvent{
			MetricID:  id,
			Timestamp: time.Now(),
		}
		return m.publisher.PublishInTx(ctx, tx, domain.MetricDeletedEventType, fmt.Sprintf("%d", id), event)
	})
}

// CreateDashboard 创建仪表板。
func (m *AnalyticsCommandService) CreateDashboard(ctx context.Context, name, description string, userID uint64) (*domain.Dashboard, error) {
	dashboard := domain.NewDashboard(name, description, userID)

	err := m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveDashboardInTx(ctx, tx, dashboard); err != nil {
			m.logger.ErrorContext(ctx, "failed to create dashboard", "error", err, "name", dashboard.Name)
			return err
		}
		if m.publisher == nil {
			return nil
		}
		event := &domain.DashboardCreatedEvent{
			DashboardID: uint64(dashboard.ID),
			UserID:      dashboard.UserID,
			Timestamp:   time.Now(),
		}
		return m.publisher.PublishInTx(ctx, tx, domain.DashboardCreatedEventType, fmt.Sprintf("%d", dashboard.ID), event)
	})
	if err != nil {
		return nil, err
	}
	return dashboard, nil
}

// AddMetricToDashboard 将一个指标添加到指定的仪表板。
func (m *AnalyticsCommandService) AddMetricToDashboard(ctx context.Context, dashboardID uint64, metricType domain.MetricType, title, chartType string) error {
	dashboard, err := m.repo.GetDashboard(ctx, dashboardID)
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

	return m.updateDashboard(ctx, dashboard)
}

// UpdateDashboard 更新仪表板。
func (m *AnalyticsCommandService) UpdateDashboard(ctx context.Context, id uint64, name, description string) (*domain.Dashboard, error) {
	dashboard, err := m.repo.GetDashboard(ctx, id)
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

	if err := m.updateDashboard(ctx, dashboard); err != nil {
		return nil, err
	}
	return dashboard, nil
}

// PublishDashboard 发布仪表板，将其状态设为公开。
func (m *AnalyticsCommandService) PublishDashboard(ctx context.Context, id uint64) error {
	dashboard, err := m.repo.GetDashboard(ctx, id)
	if err != nil {
		return err
	}
	if dashboard == nil {
		return fmt.Errorf("dashboard not found")
	}
	dashboard.Publish()
	return m.updateDashboard(ctx, dashboard)
}

// DeleteDashboard 删除仪表板。
func (m *AnalyticsCommandService) DeleteDashboard(ctx context.Context, id uint64) error {
	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.DeleteDashboardInTx(ctx, tx, id); err != nil {
			m.logger.ErrorContext(ctx, "failed to delete dashboard", "error", err, "id", id)
			return err
		}
		if m.publisher == nil {
			return nil
		}
		event := &domain.DashboardDeletedEvent{
			DashboardID: id,
			Timestamp:   time.Now(),
		}
		return m.publisher.PublishInTx(ctx, tx, domain.DashboardDeletedEventType, fmt.Sprintf("%d", id), event)
	})
}

// CreateReport 创建一个新的数据报告。
func (m *AnalyticsCommandService) CreateReport(ctx context.Context, title, description string, userID uint64, reportType string) (*domain.Report, error) {
	reportNo := fmt.Sprintf("RPT%d", m.idGenerator.Generate())
	report := domain.NewReport(reportNo, title, description, userID, reportType)

	err := m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveReportInTx(ctx, tx, report); err != nil {
			m.logger.ErrorContext(ctx, "failed to create report", "error", err, "no", report.ReportNo)
			return err
		}
		if m.publisher == nil {
			return nil
		}
		event := &domain.ReportCreatedEvent{
			ReportID:  uint64(report.ID),
			UserID:    report.UserID,
			Timestamp: time.Now(),
		}
		return m.publisher.PublishInTx(ctx, tx, domain.ReportCreatedEventType, fmt.Sprintf("%d", report.ID), event)
	})
	if err != nil {
		return nil, err
	}
	return report, nil
}

// UpdateReport 更新报告的基础信息。
func (m *AnalyticsCommandService) UpdateReport(ctx context.Context, id uint64, title, description string) (*domain.Report, error) {
	report, err := m.repo.GetReport(ctx, id)
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

	if err := m.saveReport(ctx, report, domain.ReportUpdatedEventType); err != nil {
		return nil, err
	}
	return report, nil
}

// PublishReport 发布报告。
func (m *AnalyticsCommandService) PublishReport(ctx context.Context, id uint64) error {
	report, err := m.repo.GetReport(ctx, id)
	if err != nil {
		return err
	}
	if report == nil {
		return fmt.Errorf("report not found")
	}
	report.Publish()
	return m.saveReport(ctx, report, domain.ReportPublishedEventType)
}

// DeleteReport 删除报告。
func (m *AnalyticsCommandService) DeleteReport(ctx context.Context, id uint64) error {
	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.DeleteReportInTx(ctx, tx, id); err != nil {
			m.logger.ErrorContext(ctx, "failed to delete report", "error", err, "id", id)
			return err
		}
		if m.publisher == nil {
			return nil
		}
		event := &domain.ReportDeletedEvent{
			ReportID:  id,
			Timestamp: time.Now(),
		}
		return m.publisher.PublishInTx(ctx, tx, domain.ReportDeletedEventType, fmt.Sprintf("%d", id), event)
	})
}

// --- internal helpers ---

func (m *AnalyticsCommandService) updateDashboard(ctx context.Context, dashboard *domain.Dashboard) error {
	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveDashboardInTx(ctx, tx, dashboard); err != nil {
			m.logger.ErrorContext(ctx, "failed to update dashboard", "error", err, "id", dashboard.ID)
			return err
		}
		if m.publisher == nil {
			return nil
		}
		event := &domain.DashboardUpdatedEvent{
			DashboardID: uint64(dashboard.ID),
			UserID:      dashboard.UserID,
			Timestamp:   time.Now(),
		}
		return m.publisher.PublishInTx(ctx, tx, domain.DashboardUpdatedEventType, fmt.Sprintf("%d", dashboard.ID), event)
	})
}

func (m *AnalyticsCommandService) saveReport(ctx context.Context, report *domain.Report, eventType string) error {
	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveReportInTx(ctx, tx, report); err != nil {
			m.logger.ErrorContext(ctx, "failed to save report", "error", err, "id", report.ID)
			return err
		}
		if m.publisher == nil {
			return nil
		}
		var event any
		switch eventType {
		case domain.ReportPublishedEventType:
			event = &domain.ReportPublishedEvent{
				ReportID:  uint64(report.ID),
				UserID:    report.UserID,
				Timestamp: time.Now(),
			}
		case domain.ReportUpdatedEventType:
			event = &domain.ReportUpdatedEvent{
				ReportID:  uint64(report.ID),
				UserID:    report.UserID,
				Timestamp: time.Now(),
			}
		default:
			event = &domain.ReportUpdatedEvent{
				ReportID:  uint64(report.ID),
				UserID:    report.UserID,
				Timestamp: time.Now(),
			}
			eventType = domain.ReportUpdatedEventType
		}
		return m.publisher.PublishInTx(ctx, tx, eventType, fmt.Sprintf("%d", report.ID), event)
	})
}
