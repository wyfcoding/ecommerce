package application

import (
	"context"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wyfcoding/ecommerce/internal/payment/domain"
	"github.com/wyfcoding/pkg/utils/ctxutil"
)

// ChannelMetrics 实时通道指标
type ChannelMetrics struct {
	SuccessCount uint64
	FailureCount uint64
	LastLatency  time.Duration
	LastUpdated  time.Time
}

// RoutingEngine 智能路由引擎
type RoutingEngine struct {
	channelRepo domain.ChannelRepository
	smartRouter *domain.SmartRouter
	metrics     map[string]*ChannelMetrics // key: channel_code
	mu          sync.RWMutex
}

func NewRoutingEngine(repo domain.ChannelRepository) *RoutingEngine {
	return &RoutingEngine{
		channelRepo: repo,
		smartRouter: domain.NewSmartRouter(),
		metrics:     make(map[string]*ChannelMetrics),
	}
}

// SelectBestChannel 根据实时健康度和费率选择最优网关
func (e *RoutingEngine) SelectBestChannel(ctx context.Context, amount int64, method string) (domain.GatewayType, *domain.ChannelConfig) {
	e.mu.RLock()
	// 1. 获取所有可用渠道
	var channelType domain.ChannelType
	switch method {
	case "alipay":
		channelType = domain.ChannelTypeAlipay
	case "wechat":
		channelType = domain.ChannelTypeWechat
	default:
		channelType = domain.ChannelTypeStripe
	}
	e.mu.RUnlock()

	channels, err := e.channelRepo.ListEnabledByType(ctx, channelType)
	if err != nil || len(channels) == 0 {
		return domain.GatewayTypeMock, nil
	}

	// 2. 构造路由上下文
	routeCtx := &domain.RouteContext{
		Amount:        decimal.NewFromInt(amount),
		Currency:      "CNY",
		PaymentMethod: method,
		ClientIP:      ctxutil.GetIP(ctx),
	}

	// 3. 使用 SmartRouter 进行决策
	// 策略选择：默认使用成本优先 (COST_BASED)，如果失败则降级到可用性优先
	bestChan, err := e.smartRouter.Route(ctx, routeCtx, channels, "COST_BASED")
	if err != nil {
		// 降级策略
		bestChan, err = e.smartRouter.Route(ctx, routeCtx, channels, "AVAILABILITY_FIRST")
	}

	if err != nil || bestChan == nil {
		return domain.GatewayTypeMock, nil
	}

	return domain.GatewayType(bestChan.Type), bestChan
}

// RecordResult 异步上报执行结果，用于持续优化路由
func (e *RoutingEngine) RecordResult(channelCode string, success bool, latency time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()

	m, ok := e.metrics[channelCode]
	if !ok {
		m = &ChannelMetrics{}
		e.metrics[channelCode] = m
	}

	if success {
		m.SuccessCount++
	} else {
		m.FailureCount++
	}
	m.LastLatency = latency
	m.LastUpdated = time.Now()
}
