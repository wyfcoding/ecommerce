package domain

import "time"

const (
	MemberAccountUpdatedEventType     = "loyalty.account.updated"
	PointsTransactionCreatedEventType = "loyalty.points.transaction.created"
	MemberBenefitSavedEventType       = "loyalty.benefit.saved"
	MemberBenefitDeletedEventType     = "loyalty.benefit.deleted"
)

// MemberAccountUpdatedEvent 会员账户更新事件。
type MemberAccountUpdatedEvent struct {
	UserID    uint64    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}

// PointsTransactionCreatedEvent 积分交易创建事件。
type PointsTransactionCreatedEvent struct {
	TransactionID uint64    `json:"transaction_id"`
	UserID        uint64    `json:"user_id"`
	Timestamp     time.Time `json:"timestamp"`
}

// MemberBenefitSavedEvent 会员权益保存事件。
type MemberBenefitSavedEvent struct {
	BenefitID uint64      `json:"benefit_id"`
	Level     MemberLevel `json:"level"`
	Timestamp time.Time   `json:"timestamp"`
}

// MemberBenefitDeletedEvent 会员权益删除事件。
type MemberBenefitDeletedEvent struct {
	BenefitID uint64      `json:"benefit_id"`
	Level     MemberLevel `json:"level"`
	Timestamp time.Time   `json:"timestamp"`
}
