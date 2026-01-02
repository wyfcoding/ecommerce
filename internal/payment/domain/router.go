package domain

import (
	"context"
	"errors"
	"math"
	"sort"

	"github.com/shopspring/decimal"
)

// RouteContext 路由上下文：包含决策所需的所有信息
type RouteContext struct {
	UserID        uint64
	Amount        decimal.Decimal
	Currency      string
	ClientIP      string
	PaymentMethod string // CARD, WALLET, BANK_TRANSFER
	CardBin       string // 信用卡前6位，用于识别发卡行/卡组
	Platform      string // iOS, Android, Web
}

// ChannelScore 渠道评分结果
type ChannelScore struct {
	ChannelCode   string
	Score         float64 // 分数越高越好
	EstimatedCost decimal.Decimal
	Reason        string
}

// RouterStrategy 路由策略接口
type RouterStrategy interface {
	Name() string
	// SelectOptimalChannel 返回最佳渠道代码
	SelectOptimalChannel(ctx context.Context, routeCtx *RouteContext, candidates []*ChannelConfig) (*ChannelConfig, error)
}

// SmartRouter 智能路由引擎
type SmartRouter struct {
	strategies []RouterStrategy
}

func NewSmartRouter() *SmartRouter {
	return &SmartRouter{
		strategies: []RouterStrategy{
			&AvailabilityFirstStrategy{}, // 默认策略：可用性优先
			&CostBasedStrategy{},         // 进阶策略：成本优先
		},
	}
}

// Route 根据配置和上下文选择最佳渠道
func (r *SmartRouter) Route(ctx context.Context, routeCtx *RouteContext, candidates []*ChannelConfig, strategyName string) (*ChannelConfig, error) {
	if len(candidates) == 0 {
		return nil, errors.New("no available channels")
	}

	var strategy RouterStrategy
	for _, s := range r.strategies {
		if s.Name() == strategyName {
			strategy = s
			break
		}
	}
	// Fallback to first strategy if not found
	if strategy == nil && len(r.strategies) > 0 {
		strategy = r.strategies[0]
	}

	return strategy.SelectOptimalChannel(ctx, routeCtx, candidates)
}

// --- Concrete Strategies ---

// AvailabilityFirstStrategy 可用性优先策略
// 逻辑：优先选择“Enabled”且“Priority”最高的渠道。
// 这是最基础的逻辑，通常用于系统冷启动或降级模式。
type AvailabilityFirstStrategy struct{}

func (s *AvailabilityFirstStrategy) Name() string { return "AVAILABILITY_FIRST" }

func (s *AvailabilityFirstStrategy) SelectOptimalChannel(ctx context.Context, routeCtx *RouteContext, candidates []*ChannelConfig) (*ChannelConfig, error) {
	// 过滤已禁用的 (假设 candidates 已经经过基础过滤，这里再保险一次)
	active := make([]*ChannelConfig, 0)
	for _, c := range candidates {
		if c.Enabled {
			active = append(active, c)
		}
	}

	if len(active) == 0 {
		return nil, errors.New("no active channels available")
	}

	// 按 Priority 降序排序 (Priority 值越大优先级越高)
	sort.Slice(active, func(i, j int) bool {
		return active[i].Priority > active[j].Priority
	})

	return active[0], nil
}

// CostBasedStrategy 成本优先策略
// 逻辑：计算每笔交易的预估费率 (RatePercent + FixedFee)，选择最便宜的。
// 这是一个典型的“复杂业务逻辑”，直接影响公司利润。
type CostBasedStrategy struct{}

func (s *CostBasedStrategy) Name() string { return "COST_BASED" }

func (s *CostBasedStrategy) SelectOptimalChannel(ctx context.Context, routeCtx *RouteContext, candidates []*ChannelConfig) (*ChannelConfig, error) {
	if len(candidates) == 0 {
		return nil, errors.New("no candidates")
	}

	var bestChannel *ChannelConfig
	minCost := decimal.NewFromFloat(math.MaxFloat64)

	for _, c := range candidates {
		if !c.Enabled {
			continue
		}

		// 计算预估成本
		// Cost = Amount * RatePercent + FixedFee (假设 FixedFee 存在于 ConfigJSON 解析后的结构中)
		// 这里简化模型，仅使用 RatePercent
		rate := decimal.NewFromFloat(c.RatePercent).Div(decimal.NewFromInt(100))
		estimatedCost := routeCtx.Amount.Mul(rate)

		// 假设还有复杂的 Tiered Pricing (阶梯费率)，这里是逻辑扩展点

		if estimatedCost.LessThan(minCost) {
			minCost = estimatedCost
			bestChannel = c
		} else if estimatedCost.Equal(minCost) {
			// 成本相同，比较 Priority
			if bestChannel == nil || c.Priority > bestChannel.Priority {
				bestChannel = c
			}
		}
	}

	if bestChannel == nil {
		return nil, errors.New("no suitable channel found after cost analysis")
	}

	return bestChannel, nil
}

// SuccessRateWeightedStrategy 成功率加权策略 (高级)
// 逻辑：结合成功率和成本。 Score = w1 * (1/Cost) + w2 * SuccessRate
// 成功率通常来自 Prometheus 监控或 Redis 实时统计
type SuccessRateWeightedStrategy struct {
	// 依赖 Metrics 服务获取实时成功率
}

func (s *SuccessRateWeightedStrategy) Name() string { return "SUCCESS_RATE_WEIGHTED" }

func (s *SuccessRateWeightedStrategy) SelectOptimalChannel(ctx context.Context, routeCtx *RouteContext, candidates []*ChannelConfig) (*ChannelConfig, error) {
	// 这是一个占位符，展示顶级系统会怎么做：
	// 1. GetSuccessRate(channelID, timeWindow="5m")
	// 2. Normalize inputs
	// 3. Calculate Score
	return nil, errors.New("not implemented yet")
}
