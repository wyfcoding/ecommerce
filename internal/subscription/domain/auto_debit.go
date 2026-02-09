// 变更说明：新增订阅自动扣款与续期逻辑，支持订阅到期前的自动支付尝试。
// 假设：系统在订阅到期前 24 小时触发自动扣款尝试。扣款成功后延长 EndDate，失败则记录重试或标记待过期。
package domain

import (
	"fmt"
	"time"
)

// DebitStatus 扣款状态
type DebitStatus string

const (
	DebitStatusPending   DebitStatus = "PENDING"
	DebitStatusSucceeded DebitStatus = "SUCCEEDED"
	DebitStatusFailed    DebitStatus = "FAILED"
)

// AutoDebitRecord 自动扣款记录
type AutoDebitRecord struct {
	RecordID       string
	SubscriptionID uint
	Amount         uint64
	AttemptCount   int
	LastAttempt    time.Time
	Status         DebitStatus
	ErrorMessage   string
}

// BillingService 计费服务接口
type BillingService interface {
	ExecuteAutoDebit(subscriptionID uint, amount uint64) error
}

// Renew 续期逻辑
func (s *Subscription) Renew(duration int32) {
	s.EndDate = s.EndDate.AddDate(0, 0, int(duration))
	s.Status = SubscriptionStatusActive
	s.UpdatedAt = time.Now()
}

// CreateAutoDebitRecord 创建扣款记录
func (s *Subscription) CreateAutoDebitRecord(amount uint64) *AutoDebitRecord {
	return &AutoDebitRecord{
		RecordID:       fmt.Sprintf("ADB-%d-%d", s.ID, time.Now().Unix()),
		SubscriptionID: s.ID,
		Amount:         amount,
		AttemptCount:   0,
		Status:         DebitStatusPending,
		LastAttempt:    time.Now(),
	}
}

// HandleAutoRenew 自动续期处理流程
func (s *Subscription) HandleAutoRenew(billing BillingService, amount uint64) (*AutoDebitRecord, error) {
	if !s.AutoRenew {
		return nil, fmt.Errorf("auto renew not enabled")
	}

	record := s.CreateAutoDebitRecord(amount)
	record.AttemptCount++
	record.LastAttempt = time.Now()

	err := billing.ExecuteAutoDebit(s.ID, amount)
	if err != nil {
		record.Status = DebitStatusFailed
		record.ErrorMessage = err.Error()
		return record, err
	}

	record.Status = DebitStatusSucceeded
	// 假设每期续 30 天，实际应从 Plan 中获取
	s.Renew(30)

	return record, nil
}
