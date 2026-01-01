package application

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
	analyticsv1 "github.com/wyfcoding/ecommerce/goapi/analytics/v1"
	"github.com/wyfcoding/ecommerce/internal/analytics/domain"
	accountv1 "github.com/wyfcoding/financialtrading/goapi/account/v1"
	positionv1 "github.com/wyfcoding/financialtrading/goapi/position/v1"
)

// AnalyticsQuery 处理分析模块的查询操作。
type AnalyticsQuery struct {
	repo        domain.AnalyticsRepository
	accountCli  accountv1.AccountServiceClient
	positionCli positionv1.PositionServiceClient
}

// NewAnalyticsQuery 创建并返回一个新的 AnalyticsQuery 实例。
func NewAnalyticsQuery(repo domain.AnalyticsRepository) *AnalyticsQuery {
	return &AnalyticsQuery{repo: repo}
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
	resp := &analyticsv1.UnifiedWealthDashboardResponse{
		UserId: userID,
	}

	// 1. 获取零售总支出 (从 Repo 聚合销售指标，使用维度进行用户过滤)
	retailSales, _, _ := q.repo.ListMetrics(ctx, &domain.MetricQuery{
		MetricType:   domain.MetricTypeSales,
		Dimension:    "user",
		DimensionVal: fmt.Sprintf("%d", userID),
	})
	var totalRetail float64
	for _, m := range retailSales {
		totalRetail += m.Value
	}
	resp.TotalRetailSpending = totalRetail

	// 2. 获取交易数据 (如果客户端已注入)
	if q.accountCli != nil && q.positionCli != nil {
		userIDStr := fmt.Sprintf("%d", userID)

		// 2.1 获取盈亏统计
		posSummary, err := q.positionCli.GetPositionSummary(ctx, &positionv1.GetPositionSummaryRequest{
			UserId: userIDStr,
		})
		if err == nil {
			unrealized, _ := decimal.NewFromString(posSummary.TotalUnrealizedPnl)
			realized, _ := decimal.NewFromString(posSummary.TotalRealizedPnl)
			resp.TotalTradingPnl, _ = unrealized.Add(realized).Float64()
		}

		// 2.2 获取现金余额
		balance, err := q.accountCli.GetBalance(ctx, &accountv1.GetBalanceRequest{
			UserId: userIDStr,
		})
		if err == nil {
			bal, _ := decimal.NewFromString(balance.Balance)
			resp.CashBalance, _ = bal.Float64()
		}
	}

	// 3. 计算总资产
	resp.TotalEquity = resp.CashBalance + resp.TotalTradingPnl // 简化计算：现金 + 浮盈亏

	// 4. 计算分布
	if resp.TotalEquity > 0 {
		resp.AssetDistribution = []*analyticsv1.AssetDistribution{
			{
				AssetType:  "TRADING_CASH",
				Amount:     resp.CashBalance,
				Percentage: (resp.CashBalance / resp.TotalEquity) * 100,
			},
			{
				AssetType:  "TRADING_PNL",
				Amount:     resp.TotalTradingPnl,
				Percentage: (resp.TotalTradingPnl / resp.TotalEquity) * 100,
			},
		}
	}

	return resp, nil
}
