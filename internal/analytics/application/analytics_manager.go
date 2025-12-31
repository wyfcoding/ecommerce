package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/analytics/domain"
	"github.com/wyfcoding/pkg/algorithm"
	pkgredis "github.com/wyfcoding/pkg/redis"
	"github.com/redis/go-redis/v9"
)

// AnalyticsManager 处理分析模块的写操作和业务逻辑。
// 引入树状数组（FenwickTree）用于实时、高频的订单金额和数量统计。
type AnalyticsManager struct {
	repo        domain.AnalyticsRepository
	logger      *slog.Logger
	redisClient *redis.Client
	gmvStats    *algorithm.FenwickTree // 用于统计 24 小时内每一分钟的 GMV
	orderStats  *algorithm.FenwickTree // 用于统计 24 小时内每一分钟的订单数
}

// NewAnalyticsManager 创建并返回一个新的 AnalyticsManager 实例。
func NewAnalyticsManager(repo domain.AnalyticsRepository, redisClient *redis.Client, logger *slog.Logger) *AnalyticsManager {
	return &AnalyticsManager{
		repo:        repo,
		logger:      logger,
		redisClient: redisClient,
		gmvStats:    algorithm.NewFenwickTree(1440), // 一天 1440 分钟
		orderStats:  algorithm.NewFenwickTree(1440),
	}
}

// TrackUserVisit 追踪用户访问 (统计 DAU)
func (m *AnalyticsManager) TrackUserVisit(ctx context.Context, userID uint64) {
	today := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("analytics:uv:%s", today)
	
	// 使用封装的 PFAdd 统计基数
	if err := pkgredis.PFAdd(ctx, m.redisClient, key, userID); err != nil {
		m.logger.ErrorContext(ctx, "failed to track user visit", "user_id", userID, "error", err)
	}
}

// GetDailyUV 获取今日去重访问量
func (m *AnalyticsManager) GetDailyUV(ctx context.Context) (int64, error) {
	today := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("analytics:uv:%s", today)
	
	return pkgredis.PFCount(ctx, m.redisClient, key)
}

// TrackRealtimeOrder 实时追踪订单数据。
// 利用树状数组的 O(log N) 更新特性，支持极高并发的实时统计。
func (m *AnalyticsManager) TrackRealtimeOrder(ctx context.Context, amount int64, timestamp time.Time) {
	// 计算当前时间在一天中的分钟偏移量 (0-1439)
	minute := timestamp.Hour()*60 + timestamp.Minute()

	m.gmvStats.Update(minute, amount)
	m.orderStats.Update(minute, 1)

	m.logger.DebugContext(ctx, "realtime order tracked", "minute", minute, "amount", amount)
}

// GetHourlyStats 获取指定小时的聚合统计数据。
// 利用树状数组的区间查询 O(log N) 特性，快速获取结果。
func (m *AnalyticsManager) GetHourlyStats(ctx context.Context, hour int) (int64, int64) {
	if hour < 0 || hour > 23 {
		return 0, 0
	}

	startMin := hour * 60
	endMin := startMin + 59

	totalGMV := m.gmvStats.RangeQuery(startMin, endMin)
	totalOrders := m.orderStats.RangeQuery(startMin, endMin)

	return totalGMV, totalOrders
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
