package application

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	analyticsv1 "github.com/wyfcoding/ecommerce/goapi/analytics/v1"
	"github.com/wyfcoding/ecommerce/internal/analytics/domain"
	accountv1 "github.com/wyfcoding/financialtrading/go-api/account/v1"
	positionv1 "github.com/wyfcoding/financialtrading/go-api/position/v1"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/tracing"
)

// AnalyticsQuery 处理分析模块的查询操作。
type AnalyticsQuery struct {
	repo        domain.AnalyticsRepository
	redis       redis.UniversalClient
	accountCli  accountv1.AccountServiceClient
	positionCli positionv1.PositionServiceClient
}

// NewAnalyticsQuery 创建并返回一个新的 AnalyticsQuery 实例。
func NewAnalyticsQuery(repo domain.AnalyticsRepository, redis redis.UniversalClient) *AnalyticsQuery {
	return &AnalyticsQuery{
		repo:  repo,
		redis: redis,
	}
}

// GetRealtimeVisitors 获取实时访客数据（基于 Redis HyperLogLog 并集统计）。
func (q *AnalyticsQuery) GetRealtimeVisitors(ctx context.Context) (int64, []string, error) {
	// 完整版实现：统计过去 5 分钟的去重活跃用户 (UV)
	now := time.Now()
	keys := make([]string, 0, 5)
	for i := range 5 {
		// 生成每一分钟的 Key: analytics:uv:YYYYMMDDHHMM
		t := now.Add(-time.Duration(i) * time.Minute)
		keys = append(keys, fmt.Sprintf("analytics:uv:%s", t.Format("200601021504")))
	}

	// 利用 Redis PFCount 的多 Key 并集功能，自动完成去重计数
	count, err := q.redis.PFCount(ctx, keys...).Result()
	if err != nil {
		return 0, nil, err
	}

	// 获取热门页面（从 DB 聚合最近的 Event 指标）
	pages, err := q.repo.GetActivePages(ctx, 10)
	if err != nil {
		return count, nil, nil
	}

	return count, pages, nil
}

func (q *AnalyticsQuery) SetFinancialClients(accCli accountv1.AccountServiceClient, posCli positionv1.PositionServiceClient) {
	q.accountCli = accCli
	q.positionCli = posCli
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

// GetUnifiedWealthDashboard 整合零售和交易数据。
func (q *AnalyticsQuery) GetUnifiedWealthDashboard(ctx context.Context, userID uint64) (*analyticsv1.UnifiedWealthDashboardResponse, error) {
	ctx, span := tracing.Tracer().Start(ctx, "AnalyticsQuery.GetUnifiedWealthDashboard")
	defer span.End()

	logging.Info(ctx, "fetching unified wealth dashboard", "user_id", userID)

	resp := &analyticsv1.UnifiedWealthDashboardResponse{
		UserId: userID,
	}

	// 1. 获取零售支出
	retailMetrics, _, err := q.repo.ListMetrics(ctx, &domain.MetricQuery{
		Dimension:    "user",
		DimensionVal: fmt.Sprintf("%d", userID),
	})
	if err != nil {
		logging.Error(ctx, "failed to list retail metrics", "user_id", userID, "error", err)
		// 零售指标失败不应阻止显示交易侧数据，默认设为 0
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

	// 2. 获取交易侧资产 (具备韧性检查)
	if q.accountCli != nil && q.positionCli != nil {
		userIDStr := fmt.Sprintf("%d", userID)

		// 使用 Context 超时控制跨服务调用
		rpcCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		// 2.1 获取盈亏统计 (gRPC) - 改用 GetPositions 并手动聚合
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

		// 2.2 获取现金余额 (gRPC)
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

	// 3. 计算总资产 (使用 Decimal 确保金融计算准确)
	totalDebt := totalSpending + pendingSettlement
	equityDec := decimal.NewFromFloat(resp.CashBalance).
		Add(decimal.NewFromFloat(resp.TotalTradingPnl)).
		Sub(decimal.NewFromFloat(totalDebt))

	resp.TotalEquity, _ = equityDec.Float64()

	// 4. 计算多维资产分布
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
