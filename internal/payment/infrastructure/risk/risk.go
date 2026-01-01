package risk

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

// RiskServiceImpl 风控服务实现
type RiskServiceImpl struct {
	redisClient *redis.Client
}

// NewRiskService 创建风控服务实例。
func NewRiskService(redisClient *redis.Client) *RiskServiceImpl {
	return &RiskServiceImpl{
		redisClient: redisClient,
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
	key := fmt.Sprintf("payment:risk:user_velocity:%d", riskCtx.UserID)
	val, err := s.redisClient.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get risk counter: %w", err)
	}

	if val > 5 {
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

// RecordTransaction 记录交易数据，增加 Redis 计数器
func (s *RiskServiceImpl) RecordTransaction(ctx context.Context, riskCtx *domain.RiskContext) error {
	key := fmt.Sprintf("payment:risk:user_velocity:%d", riskCtx.UserID)
	pipe := s.redisClient.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, 1*time.Hour)
	_, err := pipe.Exec(ctx)
	return err
}
