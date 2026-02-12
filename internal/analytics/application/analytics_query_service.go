package application

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"slices"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	analyticsv1 "github.com/wyfcoding/ecommerce/go-api/analytics/v1"
	"github.com/wyfcoding/ecommerce/internal/analytics/domain"
	accountv1 "github.com/wyfcoding/financialtrading/go-api/account/v1"
	positionv1 "github.com/wyfcoding/financialtrading/go-api/position/v1"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/tracing"
)

// AnalyticsQueryService 处理分析模块的查询操作。
type AnalyticsQueryService struct {
	repo              domain.AnalyticsRepository
	metricReadRepo    domain.MetricReadRepository
	dashboardReadRepo domain.DashboardReadRepository
	reportReadRepo    domain.ReportReadRepository
	metricSearchRepo  domain.MetricSearchRepository
	redis             redis.UniversalClient
	accountCli        accountv1.AccountServiceClient
	positionCli       positionv1.PositionServiceClient
	logger            *slog.Logger
}

// NewAnalyticsQueryService 创建并返回一个新的 AnalyticsQueryService 实例。
func NewAnalyticsQueryService(
	repo domain.AnalyticsRepository,
	metricReadRepo domain.MetricReadRepository,
	dashboardReadRepo domain.DashboardReadRepository,
	reportReadRepo domain.ReportReadRepository,
	metricSearchRepo domain.MetricSearchRepository,
	redis redis.UniversalClient,
	logger *slog.Logger,
) *AnalyticsQueryService {
	return &AnalyticsQueryService{
		repo:              repo,
		metricReadRepo:    metricReadRepo,
		dashboardReadRepo: dashboardReadRepo,
		reportReadRepo:    reportReadRepo,
		metricSearchRepo:  metricSearchRepo,
		redis:             redis,
		logger:            logger,
	}
}

func (q *AnalyticsQueryService) SetFinancialClients(accCli accountv1.AccountServiceClient, posCli positionv1.PositionServiceClient) {
	q.accountCli = accCli
	q.positionCli = posCli
}

// GetRealtimeVisitors 获取实时访客数据（基于 Redis HyperLogLog 并集统计）。
func (q *AnalyticsQueryService) GetRealtimeVisitors(ctx context.Context) (int64, []string, error) {
	now := time.Now()
	keys := make([]string, 0, 5)
	for i := range 5 {
		t := now.Add(-time.Duration(i) * time.Minute)
		keys = append(keys, fmt.Sprintf("analytics:uv:%s", t.Format("200601021504")))
	}

	count, err := q.redis.PFCount(ctx, keys...).Result()
	if err != nil {
		return 0, nil, err
	}

	pages, err := q.repo.GetActivePages(ctx, 10)
	if err != nil {
		return count, nil, nil
	}

	return count, pages, nil
}

// GetMetricByID 根据ID获取指标。
func (q *AnalyticsQueryService) GetMetricByID(ctx context.Context, id uint64) (*domain.Metric, error) {
	if q.metricReadRepo != nil {
		if cached, err := q.metricReadRepo.GetByID(ctx, id); err == nil && cached != nil {
			return cached, nil
		}
	}

	metric, err := q.repo.GetMetric(ctx, id)
	if err != nil {
		return nil, err
	}
	if metric != nil && q.metricReadRepo != nil {
		_ = q.metricReadRepo.Save(ctx, metric)
	}
	return metric, nil
}

// SearchMetrics 搜索指标。
func (q *AnalyticsQueryService) SearchMetrics(ctx context.Context, query *domain.MetricQuery) ([]*domain.Metric, int64, error) {
	page := 1
	pageSize := 10
	if query != nil {
		if query.Page > 0 {
			page = query.Page
		}
		if query.PageSize > 0 {
			pageSize = query.PageSize
		}
	}
	offset := (page - 1) * pageSize

	if q.metricSearchRepo != nil {
		list, total, err := q.metricSearchRepo.Search(ctx, query, offset, pageSize)
		if err == nil {
			return list, total, nil
		}
		q.logger.WarnContext(ctx, "metric search fallback to mysql", "error", err)
	}
	return q.repo.ListMetrics(ctx, query)
}

// GetDashboardByID 获取仪表板。
func (q *AnalyticsQueryService) GetDashboardByID(ctx context.Context, id uint64) (*domain.Dashboard, error) {
	if q.dashboardReadRepo != nil {
		if cached, err := q.dashboardReadRepo.GetByID(ctx, id); err == nil && cached != nil {
			return cached, nil
		}
	}

	dashboard, err := q.repo.GetDashboard(ctx, id)
	if err != nil {
		return nil, err
	}
	if dashboard != nil && q.dashboardReadRepo != nil {
		_ = q.dashboardReadRepo.Save(ctx, dashboard)
	}
	return dashboard, nil
}

// ListUserDashboards 列出用户的仪表板。
func (q *AnalyticsQueryService) ListUserDashboards(ctx context.Context, userID uint64, offset, limit int) ([]*domain.Dashboard, int64, error) {
	return q.repo.ListDashboards(ctx, userID, offset, limit)
}

// GetReportByID 获取报告。
func (q *AnalyticsQueryService) GetReportByID(ctx context.Context, id uint64) (*domain.Report, error) {
	if q.reportReadRepo != nil {
		if cached, err := q.reportReadRepo.GetByID(ctx, id); err == nil && cached != nil {
			return cached, nil
		}
	}

	report, err := q.repo.GetReport(ctx, id)
	if err != nil {
		return nil, err
	}
	if report != nil && q.reportReadRepo != nil {
		_ = q.reportReadRepo.Save(ctx, report)
	}
	return report, nil
}

// ListUserReports 列出用户的报告。
func (q *AnalyticsQueryService) ListUserReports(ctx context.Context, userID uint64, offset, limit int) ([]*domain.Report, int64, error) {
	return q.repo.ListReports(ctx, userID, offset, limit)
}

// GetUserActivityReport 获取用户活动概览报告。
func (q *AnalyticsQueryService) GetUserActivityReport(ctx context.Context, startTime, endTime time.Time) (map[string]any, error) {
	query := &domain.MetricQuery{
		MetricType: domain.MetricTypeActiveUsers,
		StartTime:  startTime,
		EndTime:    endTime,
	}
	metrics, _, err := q.SearchMetrics(ctx, query)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"active_users": len(metrics),
	}, nil
}

// GetProductPerformanceReport 获取商品销售性能分析报告。
func (q *AnalyticsQueryService) GetProductPerformanceReport(ctx context.Context, startTime, endTime time.Time) (map[string]any, error) {
	query := &domain.MetricQuery{
		MetricType: domain.MetricTypeSales,
		StartTime:  startTime,
		EndTime:    endTime,
	}

	metrics, _, err := q.SearchMetrics(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search sales metrics: %w", err)
	}

	productSales := make(map[string]float64)
	for _, m := range metrics {
		if m.Dimension == "product" {
			productSales[m.DimensionVal] += m.Value
		}
	}

	type productStat struct {
		Product string  `json:"product"`
		Sales   float64 `json:"sales"`
	}
	stats := make([]productStat, 0, len(productSales))
	for p, sales := range productSales {
		stats = append(stats, productStat{Product: p, Sales: sales})
	}

	slices.SortFunc(stats, func(a, b productStat) int {
		if a.Sales > b.Sales {
			return -1
		}
		if a.Sales < b.Sales {
			return 1
		}
		return 0
	})
	topN := min(len(stats), 10)

	return map[string]any{
		"total_sales":  len(metrics),
		"top_products": stats[:topN],
	}, nil
}

// GetConversionFunnelReport 获取用户转化漏斗分析报告。
func (q *AnalyticsQueryService) GetConversionFunnelReport(ctx context.Context, startTime, endTime time.Time) (map[string]any, error) {
	steps := []struct {
		Name string
		Type domain.MetricType
	}{
		{"Page View", domain.MetricTypePageViews},
		{"Add to Cart", "add_to_cart"},
		{"Order Placed", domain.MetricTypeOrders},
	}

	funnelData := make([]map[string]any, 0, len(steps))
	var prevCount float64

	for i, step := range steps {
		query := &domain.MetricQuery{
			MetricType: step.Type,
			StartTime:  startTime,
			EndTime:    endTime,
		}

		metrics, _, err := q.SearchMetrics(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("failed to search metrics for step %s: %w", step.Name, err)
		}

		count := 0.0
		for _, m := range metrics {
			count += m.Value
		}

		rate := 0.0
		if i > 0 && prevCount > 0 {
			rate = (count / prevCount) * 100
		} else if i == 0 {
			rate = 100.0
		}

		funnelData = append(funnelData, map[string]any{
			"step":            step.Name,
			"count":           count,
			"conversion_rate": fmt.Sprintf("%.2f%%", rate),
		})
		prevCount = count
	}

	return map[string]any{"funnel": funnelData}, nil
}

// GetCustomReport 获取自定义报告数据。
func (q *AnalyticsQueryService) GetCustomReport(_ context.Context, reportID uint64, startTime, endTime time.Time) (map[string]any, error) {
	return map[string]any{"custom_data": "data"}, nil
}

// GetUserBehaviorPath 获取用户的行为路径追踪数据。
func (q *AnalyticsQueryService) GetUserBehaviorPath(_ context.Context, userID uint64, startTime, endTime time.Time) (map[string]any, error) {
	return map[string]any{"path": []string{}}, nil
}

// GetUserSegments 获取用户细分群体分析数据。
func (q *AnalyticsQueryService) GetUserSegments(ctx context.Context) (map[string]any, error) {
	return map[string]any{"segments": []string{}}, nil
}

// GetUnifiedWealthDashboard 整合零售和交易数据。
func (q *AnalyticsQueryService) GetUnifiedWealthDashboard(ctx context.Context, userID uint64) (*analyticsv1.UnifiedWealthDashboardResponse, error) {
	ctx, span := tracing.Tracer().Start(ctx, "AnalyticsQuery.GetUnifiedWealthDashboard")
	defer span.End()

	logging.Info(ctx, "fetching unified wealth dashboard", "user_id", userID)

	resp := &analyticsv1.UnifiedWealthDashboardResponse{
		UserId: userID,
	}

	retailMetrics, _, err := q.repo.ListMetrics(ctx, &domain.MetricQuery{
		Dimension:    "user",
		DimensionVal: fmt.Sprintf("%d", userID),
	})
	if err != nil {
		logging.Error(ctx, "failed to list retail metrics", "user_id", userID, "error", err)
	}

	var (
		totalSpending     float64
		pendingSettlement float64
	)

	for _, m := range retailMetrics {
		switch m.MetricType {
		case domain.MetricTypeSales:
			totalSpending += m.Value
		case "PENDING_SETTLEMENT":
			pendingSettlement += m.Value
		}
	}
	resp.TotalRetailSpending = totalSpending

	if q.accountCli != nil && q.positionCli != nil {
		userIDStr := fmt.Sprintf("%d", userID)

		rpcCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		posResp, err := q.positionCli.GetPositions(rpcCtx, &positionv1.GetPositionsRequest{
			UserId:   userIDStr,
			PageSize: 1000,
		})
		if err != nil {
			logging.Warn(ctx, "failed to fetch positions for summary", "user_id", userID, "error", err)
		} else {
			var totalPnl decimal.Decimal
			for _, p := range posResp.GetPositions() {
				unrealized, _ := decimal.NewFromString(p.UnrealizedPnl)
				realized, _ := decimal.NewFromString(p.RealizedPnl)
				totalPnl = totalPnl.Add(unrealized).Add(realized)
			}
			resp.TotalTradingPnl, _ = totalPnl.Float64()
		}

		balance, err := q.accountCli.GetBalance(rpcCtx, &accountv1.GetBalanceRequest{
			UserId: userIDStr,
		})
		if err != nil {
			logging.Warn(ctx, "failed to fetch trading balance", "user_id", userID, "error", err)
		} else {
			bal, err := decimal.NewFromString(balance.Balance)
			if err == nil {
				resp.CashBalance, _ = bal.Float64()
			}
		}
	} else {
		logging.Warn(ctx, "financial clients not configured, skipping trading data", "user_id", userID)
	}

	totalDebt := totalSpending + pendingSettlement
	equityDec := decimal.NewFromFloat(resp.CashBalance).
		Add(decimal.NewFromFloat(resp.TotalTradingPnl)).
		Sub(decimal.NewFromFloat(totalDebt))

	resp.TotalEquity, _ = equityDec.Float64()

	if resp.TotalEquity != 0 {
		total := decimal.NewFromFloat(math.Abs(resp.TotalEquity))

		addAsset := func(assetType string, amount float64) {
			if amount == 0 {
				return
			}
			pct, _ := decimal.NewFromFloat(math.Abs(amount)).Div(total).Mul(decimal.NewFromInt(100)).Float64()
			resp.AssetDistribution = append(resp.AssetDistribution, &analyticsv1.AssetDistribution{
				AssetType:  assetType,
				Amount:     amount,
				Percentage: pct,
			})
		}

		addAsset("TRADING_CASH", resp.CashBalance)
		addAsset("TRADING_PNL", resp.TotalTradingPnl)
		addAsset("RETAIL_DEBT", -totalDebt)
	}

	logging.Info(ctx, "unified wealth dashboard fetched successfully", "user_id", userID, "total_equity", resp.TotalEquity)
	return resp, nil
}
