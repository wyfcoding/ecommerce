package domain

import "time"

// SubscriptionStatus 定义了订阅的生命周期状态。
type SubscriptionStatus int8

const (
	SubscriptionStatusActive   SubscriptionStatus = 1 // 活跃：订阅正在生效中。
	SubscriptionStatusExpired  SubscriptionStatus = 2 // 过期：订阅已到期。
	SubscriptionStatusCanceled SubscriptionStatus = 3 // 取消：订阅已被用户或系统取消。
	SubscriptionStatusPaused   SubscriptionStatus = 4 // 暂停：订阅暂时暂停。
)

// SubscriptionPlan 实体代表一个订阅计划。
type SubscriptionPlan struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       uint64    `json:"price"`
	Duration    int32     `json:"duration"`
	Features    []string  `json:"features"`
	Enabled     bool      `json:"enabled"`
}

// Subscription 实体代表用户的订阅记录。
type Subscription struct {
	ID         uint               `json:"id"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
	UserID     uint64             `json:"user_id"`
	PlanID     uint64             `json:"plan_id"`
	Status     SubscriptionStatus `json:"status"`
	StartDate  time.Time          `json:"start_date"`
	EndDate    time.Time          `json:"end_date"`
	AutoRenew  bool               `json:"auto_renew"`
	CanceledAt *time.Time         `json:"canceled_at"`
}

// IsActive 检查订阅是否当前处于活跃状态。
func (s *Subscription) IsActive() bool {
	return s.Status == SubscriptionStatusActive && time.Now().Before(s.EndDate)
}
