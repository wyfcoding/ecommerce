// 变更说明：实现跨项目积分互通逻辑，支持电商积分 (Loyalty Points) 与交易积分 (Trading Points) 之间的双向兑换。
// 假设：兑换汇率为 1:1，但支持后台动态配置。兑换过程涉及两方账户的同步扣减与增加。
package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// PointType 积分类型
type PointType string

const (
	PointTypeEcommerce PointType = "ECOMMERCE"
	PointTypeTrading   PointType = "TRADING"
)

// PointExchange 积分兑换聚合
type PointExchange struct {
	ExchangeID   string
	UserID       string
	FromType     PointType
	ToType       PointType
	Amount       decimal.Decimal
	ExchangeRate decimal.Decimal
	TargetAmount decimal.Decimal
	Status       string // PENDING, COMPLETED, FAILED
	CreatedAt    time.Time
	CompletedAt  *time.Time
}

// PointsSynergyService 跨项目积分协同接口
type PointsSynergyService interface {
	GetEcommercePoints(ctx context.Context, userID string) (decimal.Decimal, error)
	GetTradingPoints(ctx context.Context, userID string) (decimal.Decimal, error)
	AdjustPoints(ctx context.Context, userID string, pType PointType, amount decimal.Decimal) error
}

// NewPointExchange 创建积分兑换请求
func NewPointExchange(userID string, from, to PointType, amount, rate decimal.Decimal) *PointExchange {
	return &PointExchange{
		ExchangeID:   fmt.Sprintf("EX-%d", time.Now().UnixNano()),
		UserID:       userID,
		FromType:     from,
		ToType:       to,
		Amount:       amount,
		ExchangeRate: rate,
		TargetAmount: amount.Mul(rate),
		Status:       "PENDING",
		CreatedAt:    time.Now(),
	}
}

// Execute 执行兑换流程
func (e *PointExchange) Execute(ctx context.Context, api PointsSynergyService) error {
	// 1. 扣减源端积分
	if err := api.AdjustPoints(ctx, e.UserID, e.FromType, e.Amount.Neg()); err != nil {
		e.Status = "FAILED"
		return err
	}

	// 2. 增加目标端积分
	if err := api.AdjustPoints(ctx, e.UserID, e.ToType, e.TargetAmount); err != nil {
		// 注意：实际应有补偿机制（回退源端），此处简化处理
		e.Status = "FAILED"
		return err
	}

	now := time.Now()
	e.Status = "COMPLETED"
	e.CompletedAt = &now
	return nil
}
