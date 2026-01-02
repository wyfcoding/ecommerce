package application

import (
	"context"
	"sync"
	"time"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
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
	metrics     map[string]*ChannelMetrics // key: channel_code
	mu          sync.RWMutex
}

func NewRoutingEngine(repo domain.ChannelRepository) *RoutingEngine {
	return &RoutingEngine{
		channelRepo: repo,
		metrics:     make(map[string]*ChannelMetrics),
	}
}

// SelectBestChannel 根据实时健康度和费率选择最优网关
func (e *RoutingEngine) SelectBestChannel(ctx context.Context, amount int64, method string) (domain.GatewayType, *domain.ChannelConfig) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// 1. 获取所有可用渠道
	var channelType domain.ChannelType
	switch method {
	case "alipay": channelType = domain.ChannelTypeAlipay
	case "wechat": channelType = domain.ChannelTypeWechat
	default: channelType = domain.ChannelTypeStripe
	}

	channels, err := e.channelRepo.ListEnabledByType(ctx, channelType)
	if err != nil || len(channels) == 0 {
		return domain.GatewayTypeMock, nil
	}

	// 2. 智能评估得分 (Score = Weight_Priority * 0.4 + Weight_SuccessRate * 0.6)
	var bestChan *domain.ChannelConfig
	maxScore := -1.0

	for _, ch := range channels {
		metrics, ok := e.metrics[ch.Code]
		successRate := 1.0 // 默认新渠道 100%
		if ok && (metrics.SuccessCount + metrics.FailureCount) > 0 {
			successRate = float64(metrics.SuccessCount) / float64(metrics.SuccessCount + metrics.FailureCount)
		}

		// 如果延迟过高 (> 5s)，大幅降分
		latencyPenalty := 1.0
		if ok && metrics.LastLatency > 5*time.Second {
			latencyPenalty = 0.5
		}

		score := (float64(ch.Priority) / 100.0 * 0.4) + (successRate * 0.6 * latencyPenalty)
		
		if score > maxScore {
			maxScore = score
			bestChan = ch
		}
	}

	if bestChan == nil {
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
