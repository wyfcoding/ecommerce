package domain

import (
	"context"

	"github.com/shopspring/decimal"
)

// FiatAdapter 汇率服务适配器接口
type FiatAdapter interface {
	// GetRate 获取当前汇率
	GetRate(ctx context.Context, from, to string) (decimal.Decimal, error)

	// LockRate 锁定汇率
	LockRate(ctx context.Context, userID, paymentID, from, to string, amount decimal.Decimal) (string, decimal.Decimal, error)

	// VerifyLock 验证汇率锁定
	VerifyLock(ctx context.Context, lockID string) (bool, decimal.Decimal, error)
}
