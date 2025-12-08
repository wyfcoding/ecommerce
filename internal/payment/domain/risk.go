package domain

import (
	"context"
)

// RiskAction 定义风控动作
type RiskAction string

const (
	RiskActionPass      RiskAction = "PASS"      // 通过
	RiskActionBlock     RiskAction = "BLOCK"     // 拦截
	RiskActionChallenge RiskAction = "CHALLENGE" // 需要验证（如3DS/短信验证码）
)

// RiskResult 风控检查结果
type RiskResult struct {
	Action      RiskAction // 动作
	Reason      string     // 原因 (e.g., "Velocity Limit Exceeded")
	RuleID      string     // 触发的规则ID
	Description string     // 详细描述
}

// RiskContext 风控上下文
type RiskContext struct {
	UserID        uint64
	IP            string
	DeviceID      string
	Amount        int64
	PaymentMethod string
	OrderID       uint64
}

// RiskService 风控服务接口
type RiskService interface {
	// CheckPrePayment 支付前风控检查
	CheckPrePayment(ctx context.Context, riskCtx *RiskContext) (*RiskResult, error)

	// RecordTransaction 记录交易（用于风控统计，如频控）
	RecordTransaction(ctx context.Context, riskCtx *RiskContext) error
}
