// 变更说明：新增自动扣款记录仓储接口和自动续费任务调度逻辑
package domain

import (
	"context"
	"errors"
	"time"
)

// AutoDebitRepository 自动扣款记录仓储接口
type AutoDebitRepository interface {
	Save(ctx context.Context, record *AutoDebitRecord) error
	Get(ctx context.Context, recordID string) (*AutoDebitRecord, error)
	GetBySubscriptionID(ctx context.Context, subscriptionID uint, limit int) ([]*AutoDebitRecord, error)
	GetPendingRecords(ctx context.Context, limit int) ([]*AutoDebitRecord, error)
	UpdateStatus(ctx context.Context, recordID string, status DebitStatus, errorMsg string) error
}

// RenewalReminder 续费提醒
type RenewalReminder struct {
	SubscriptionID uint
	UserID         uint64
	PlanName       string
	EndDate        time.Time
	DaysRemaining  int
	ReminderType   ReminderType
}

// ReminderType 提醒类型
type ReminderType string

const (
	ReminderType7Days  ReminderType = "7_DAYS"
	ReminderType3Days  ReminderType = "3_DAYS"
	ReminderType1Day   ReminderType = "1_DAY"
	ReminderTypeExpired ReminderType = "EXPIRED"
)

// RenewalReminderRepository 续费提醒仓储接口
type RenewalReminderRepository interface {
	Save(ctx context.Context, reminder *RenewalReminder) error
	GetSentReminders(ctx context.Context, subscriptionID uint, reminderType ReminderType) (bool, error)
	MarkAsSent(ctx context.Context, subscriptionID uint, reminderType ReminderType) error
}

// SubscriptionExpiringEvent 订阅即将过期事件
type SubscriptionExpiringEvent struct {
	SubscriptionID uint64    `json:"subscription_id"`
	UserID         uint64    `json:"user_id"`
	PlanID         uint64    `json:"plan_id"`
	EndDate        time.Time `json:"end_date"`
	DaysRemaining  int       `json:"days_remaining"`
	Timestamp      time.Time `json:"timestamp"`
}

// AutoDebitExecutedEvent 自动扣款执行事件
type AutoDebitExecutedEvent struct {
	RecordID       string      `json:"record_id"`
	SubscriptionID uint        `json:"subscription_id"`
	UserID         uint64      `json:"user_id"`
	Amount         uint64      `json:"amount"`
	Status         DebitStatus `json:"status"`
	AttemptCount   int         `json:"attempt_count"`
	Timestamp      time.Time   `json:"timestamp"`
}

const (
	SubscriptionExpiringEventType = "subscription.expiring"
	AutoDebitExecutedEventType    = "subscription.auto_debit.executed"
)

// PaymentGateway 支付网关接口
type PaymentGateway interface {
	ChargeUser(ctx context.Context, userID uint64, amount uint64, description string) (transactionID string, err error)
	Refund(ctx context.Context, transactionID string, amount uint64) error
}

// NotificationSender 通知发送接口
type NotificationSender interface {
	SendRenewalReminder(ctx context.Context, userID uint64, reminder *RenewalReminder) error
	SendAutoDebitResult(ctx context.Context, userID uint64, success bool, amount uint64) error
}

// ShouldSendReminder 判断是否应该发送提醒
func (s *Subscription) ShouldSendReminder(now time.Time) (bool, ReminderType) {
	if s.Status != SubscriptionStatusActive {
		return false, ""
	}
	
	daysRemaining := int(time.Until(s.EndDate).Hours() / 24)
	
	switch {
	case daysRemaining <= 0:
		return true, ReminderTypeExpired
	case daysRemaining <= 1:
		return true, ReminderType1Day
	case daysRemaining <= 3:
		return true, ReminderType3Days
	case daysRemaining <= 7:
		return true, ReminderType7Days
	default:
		return false, ""
	}
}

// NeedsAutoDebit 判断是否需要自动扣款
func (s *Subscription) NeedsAutoDebit(now time.Time) bool {
	if !s.AutoRenew {
		return false
	}
	if s.Status != SubscriptionStatusActive {
		return false
	}
	
	hoursUntilExpiry := time.Until(s.EndDate).Hours()
	return hoursUntilExpiry > 0 && hoursUntilExpiry <= 24
}

// Pause 暂停订阅
func (s *Subscription) Pause(reason string) error {
	if s.Status != SubscriptionStatusActive {
		return ErrInvalidSubscriptionStatus
	}
	s.Status = SubscriptionStatusPaused
	s.UpdatedAt = time.Now()
	return nil
}

// Resume 恢复订阅
func (s *Subscription) Resume() error {
	if s.Status != SubscriptionStatusPaused {
		return ErrInvalidSubscriptionStatus
	}
	if time.Now().After(s.EndDate) {
		return ErrSubscriptionExpired
	}
	s.Status = SubscriptionStatusActive
	s.UpdatedAt = time.Now()
	return nil
}

// 错误定义
var (
	ErrInvalidSubscriptionStatus = errors.New("invalid subscription status for this operation")
	ErrSubscriptionExpired       = errors.New("subscription has expired")
)
