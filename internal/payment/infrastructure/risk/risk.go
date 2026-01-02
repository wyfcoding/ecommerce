package risk

import (
	"context"
	"fmt"

	riskv1 "github.com/wyfcoding/ecommerce/goapi/risksecurity/v1"
	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

// RiskServiceImpl 风控服务实现 (gRPC Adapter)
type RiskServiceImpl struct {
	client riskv1.RiskSecurityServiceClient
}

// NewRiskService 创建风控服务实例。
func NewRiskService(client riskv1.RiskSecurityServiceClient) *RiskServiceImpl {
	return &RiskServiceImpl{
		client: client,
	}
}

// CheckPrePayment 支付前回风控检查
func (s *RiskServiceImpl) CheckPrePayment(ctx context.Context, riskCtx *domain.RiskContext) (*domain.RiskResult, error) {
	// 调用远程风控服务
	resp, err := s.client.EvaluateRisk(ctx, &riskv1.EvaluateRiskRequest{
		UserId:        riskCtx.UserID,
		Ip:            riskCtx.IP,
		DeviceId:      riskCtx.DeviceID,
		Amount:        riskCtx.Amount,
		PaymentMethod: riskCtx.PaymentMethod,
		OrderId:       riskCtx.OrderID,
	})
	if err != nil {
		return nil, fmt.Errorf("remote risk check failed: %w", err)
	}

	if resp.Result == nil {
		return nil, fmt.Errorf("remote risk check returned empty result")
	}

	// 映射结果
	result := &domain.RiskResult{
		RuleID:      "REMOTE_RISK",
		Description: fmt.Sprintf("Score: %d", resp.Result.RiskScore),
	}

	// 假设 3=High, 4=Critical (对应 domain.RiskLevel 定义)
	switch resp.Result.RiskLevel {
	case 4: // Critical
		result.Action = domain.RiskActionBlock
		result.Reason = "Critical Risk"
	case 3: // High
		result.Action = domain.RiskActionChallenge
		result.Reason = "High Risk"
	default:
		result.Action = domain.RiskActionPass
		result.Reason = "Risk check passed"
	}

	return result, nil
}

// RecordTransaction 记录交易数据
func (s *RiskServiceImpl) RecordTransaction(ctx context.Context, riskCtx *domain.RiskContext) error {
	// 调用远程风控服务记录行为
	_, err := s.client.RecordUserBehavior(ctx, &riskv1.RecordUserBehaviorRequest{
		UserId:   riskCtx.UserID,
		Ip:       riskCtx.IP,
		DeviceId: riskCtx.DeviceID,
	})
	return err
}
