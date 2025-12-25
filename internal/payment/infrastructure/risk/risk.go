package risk

import (
	"context"
	"fmt"
	"time"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

// RiskServiceImpl 风控服务实现
type RiskServiceImpl struct {
	// TODO: 替换为 Redis 客户端
	counters map[string]int
}

// NewRiskService 定义了 NewRisk 相关的服务逻辑。
func NewRiskService() *RiskServiceImpl {
	return &RiskServiceImpl{
		counters: make(map[string]int),
	}
}

// CheckPrePayment 支付前回风控检查
func (s *RiskServiceImpl) CheckPrePayment(ctx context.Context, riskCtx *domain.RiskContext) (*domain.RiskResult, error) {
	// 规则 1: 大额检查 (> 50000 CNY)
	if riskCtx.Amount > 5000000 { // 50000.00 分
		return &domain.RiskResult{
			Action:      domain.RiskActionChallenge,
			Reason:      "Large Amount",
			RuleID:      "RULE_LARGE_AMOUNT",
			Description: "Transaction amount exceeds limit",
		}, nil
	}

	// 规则 2: 频控 (User ID 维度)
	key := fmt.Sprintf("user_velocity:%d", riskCtx.UserID)
	if s.counters[key] > 5 {
		return &domain.RiskResult{
			Action:      domain.RiskActionBlock,
			Reason:      "Velocity Limit",
			RuleID:      "RULE_USER_VELOCITY",
			Description: "Too many transactions in short period",
		}, nil
	}

	return &domain.RiskResult{
		Action: domain.RiskActionPass,
	}, nil
}

// RecordTransaction 记录交易数据
func (s *RiskServiceImpl) RecordTransaction(ctx context.Context, riskCtx *domain.RiskContext) error {
	// 模拟 Redis INCR
	key := fmt.Sprintf("user_velocity:%d", riskCtx.UserID)
	s.counters[key]++

	// 模拟过期清理
	go func() {
		time.Sleep(1 * time.Hour)
		s.counters[key]--
	}()

	return nil
}
