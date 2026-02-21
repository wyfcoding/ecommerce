package risk

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	riskv1 "github.com/wyfcoding/ecommerce/go-api/risk/v1"
	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

// ServiceImpl 风控服务实现 (gRPC Adapter)
type ServiceImpl struct {
	client riskv1.RiskServiceClient
}

// NewRiskService 创建风控服务实例。
func NewRiskService(client riskv1.RiskServiceClient) *ServiceImpl {
	return &ServiceImpl{
		client: client,
	}
}

// CheckPrePayment 支付前回风控检查
func (s *ServiceImpl) CheckPrePayment(ctx context.Context, riskCtx *domain.RiskContext) (*domain.RiskResult, error) {
	// 调用远程风控服务
	resp, err := s.client.EvaluateRisk(ctx, &riskv1.EvaluateRiskRequest{
		UserId:     fmt.Sprintf("%d", riskCtx.UserID),
		IpAddress:  riskCtx.IP,
		DeviceId:   riskCtx.DeviceID,
		ActionType: "PAYMENT",
		Context: map[string]string{
			"amount":         strconv.FormatInt(riskCtx.Amount, 10),
			"payment_method": riskCtx.PaymentMethod,
			"order_id":       strconv.FormatUint(riskCtx.OrderID, 10),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("remote risk check failed: %w", err)
	}

	// 映射结果
	result := &domain.RiskResult{
		RuleID:      "REMOTE_RISK",
		Description: resp.Reason,
	}

	switch resp.RiskLevel {
	case "CRITICAL":
		result.Action = domain.RiskActionBlock
		result.Reason = "Critical Risk"
	case "HIGH":
		result.Action = domain.RiskActionChallenge
		result.Reason = "High Risk"
	default:
		result.Action = domain.RiskActionPass
		result.Reason = "Risk check passed"
	}

	return result, nil
}

// RecordTransaction 记录交易数据
func (s *ServiceImpl) RecordTransaction(ctx context.Context, riskCtx *domain.RiskContext) error {
	// 调用远程风控服务记录行为
	meta := map[string]string{
		"ip":        riskCtx.IP,
		"device_id": riskCtx.DeviceID,
	}
	metaJSON, _ := json.Marshal(meta)

	_, err := s.client.RecordUserBehavior(ctx, &riskv1.RecordUserBehaviorRequest{
		UserId:       fmt.Sprintf("%d", riskCtx.UserID),
		BehaviorType: "PAYMENT_RECORD",
		Metadata:     string(metaJSON),
	})
	return err
}
