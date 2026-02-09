// 变更说明：新增分账支付逻辑，支持订单支付后的资金自动分润，包含商家结算与平台佣金扣除。
// 假设：分账在订单确认收货或担保期结束后触发。支持按比例分账或固定金额分账。
package domain

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// SplitParticipant 分账参与方
type SplitParticipant struct {
	AccountID   string          `json:"account_id"`
	Role        string          `json:"role"` // MERCHANT, PLATFORM, REFERRER
	Amount      decimal.Decimal `json:"amount"`
	Description string          `json:"description"`
}

// SplitPayment 分账任务聚合
type SplitPayment struct {
	SplitID       string              `json:"split_id"`
	PaymentID     string              `json:"payment_id"`
	TotalAmount   decimal.Decimal     `json:"total_amount"`
	Participants  []*SplitParticipant `json:"participants"`
	Status        string              `json:"status"` // PENDING, PROCESSING, SUCCEEDED, FAILED
	CreatedAt     time.Time           `json:"created_at"`
	ExecutedAt    *time.Time          `json:"executed_at"`
	FailureReason string              `json:"failure_reason"`
}

// SplitService 分账计算服务
type SplitService struct {
	PlatformCommissionRate decimal.Decimal // 平台抽佣比例
}

func NewSplitService(rate decimal.Decimal) *SplitService {
	return &SplitService{PlatformCommissionRate: rate}
}

// CalculateSplit 计算分账详情
func (s *SplitService) CalculateSplit(paymentID string, amount decimal.Decimal, merchantAccountID string) *SplitPayment {
	// 1. 计算平台抽佣
	commission := amount.Mul(s.PlatformCommissionRate)

	// 2. 商家所得
	merchantAmount := amount.Sub(commission)

	participants := []*SplitParticipant{
		{
			AccountID:   "PLATFORM_TREASURY",
			Role:        "PLATFORM",
			Amount:      commission,
			Description: "Platform commission fee",
		},
		{
			AccountID:   merchantAccountID,
			Role:        "MERCHANT",
			Amount:      merchantAmount,
			Description: "Merchant settlement",
		},
	}

	return &SplitPayment{
		SplitID:      fmt.Sprintf("SPLIT-%s", paymentID),
		PaymentID:    paymentID,
		TotalAmount:  amount,
		Participants: participants,
		Status:       "PENDING",
		CreatedAt:    time.Now(),
	}
}

// Execute 执行分账标记
func (p *SplitPayment) Execute() {
	p.Status = "PROCESSING"
}

// Complete 标记分账完成
func (p *SplitPayment) Complete() {
	p.Status = "SUCCEEDED"
	now := time.Now()
	p.ExecutedAt = &now
}

// Fail 标记分账失败
func (p *SplitPayment) Fail(reason string) {
	p.Status = "FAILED"
	p.FailureReason = reason
}
