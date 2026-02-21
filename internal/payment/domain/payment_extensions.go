// Package domain 支付增强逻辑
package domain

import (
	"time"
)

// FXLock 汇率锁定机制
type FXLock struct {
	LockID    string
	Rate      float64
	ExpiresAt time.Time
}

// BankCard 银行卡管理模型
type BankCard struct {
	CardID   string
	UserID   uint64
	BankName string
	MaskedNo string // 138****5678
	Status   string // VERIFIED, BIN_FAILED
}
