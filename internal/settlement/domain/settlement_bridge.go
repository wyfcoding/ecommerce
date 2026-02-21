// Package domain 资金结算桥接领域模型
package domain

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// SettlementStatus uses the definition in settlement.go

// SettlementBridge 结算桥接聚合根
type SettlementBridge struct {
	ID            string           `json:"id"`
	PaymentID     string           `json:"payment_id"`
	OrderID       string           `json:"order_id"`
	Amount        decimal.Decimal  `json:"amount"`
	Currency      string           `json:"currency"`
	FromAccount   string           `json:"from_account"` // 电商资金账户
	ToAccount     string           `json:"to_account"`   // 金融国库账户
	Status        SettlementStatus `json:"status"`
	CreatedAt     time.Time        `json:"created_at"`
	ProcessedAt   *time.Time       `json:"processed_at"`
	CompletedAt   *time.Time       `json:"completed_at"`
	FailureReason string           `json:"failure_reason"`
	RetryCount    int              `json:"retry_count"`
}

// NewSettlementBridge 创建结算桥接记录
func NewSettlementBridge(paymentID, orderID string, amount decimal.Decimal, currency string) *SettlementBridge {
	return &SettlementBridge{
		ID:         paymentID, // 使用支付ID作为桥接记录ID
		PaymentID:  paymentID,
		OrderID:    orderID,
		Amount:     amount,
		Currency:   currency,
		Status:     StatusPending,
		CreatedAt:  time.Now(),
		RetryCount: 0,
	}
}

// StartProcessing 开始处理
func (s *SettlementBridge) StartProcessing() {
	s.Status = StatusProcessing
	now := time.Now()
	s.ProcessedAt = &now
}

// Complete 完成结算
func (s *SettlementBridge) Complete() {
	s.Status = StatusCompleted
	now := time.Now()
	s.CompletedAt = &now
}

// Fail 结算失败
func (s *SettlementBridge) Fail(reason string) {
	s.Status = StatusFailed
	s.FailureReason = reason
	s.RetryCount++
}

// CanRetry 是否可以重试
func (s *SettlementBridge) CanRetry() bool {
	return s.Status == StatusFailed && s.RetryCount < 3
}

// SettlementBridgeRepository 结算桥接仓储接口
type SettlementBridgeRepository interface {
	Save(ctx context.Context, settlement *SettlementBridge) error
	GetByPaymentID(ctx context.Context, paymentID string) (*SettlementBridge, error)
	ListPending(ctx context.Context, limit int) ([]*SettlementBridge, error)
}

// TreasuryService 金融国库服务接口 (External)
type TreasuryService interface {
	DepositToTreasury(ctx context.Context, accountID string, amount decimal.Decimal, currency, refID string) error
}

// PaymentService 电商支付服务接口 (Internal)
type PaymentService interface {
	GetPaymentDetails(ctx context.Context, paymentID string) (*PaymentDetails, error)
}

type PaymentDetails struct {
	PaymentID string
	OrderID   string
	Amount    decimal.Decimal
	Currency  string
	Status    string
}
