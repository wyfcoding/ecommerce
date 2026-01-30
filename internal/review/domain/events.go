package domain

import "time"

// ReviewCreatedEvent 评论创建事件。
type ReviewCreatedEvent struct {
	ReviewID  uint64    `json:"review_id"`
	UserID    uint64    `json:"user_id"`
	ProductID uint64    `json:"product_id"`
	Rating    int32     `json:"rating"`
	Timestamp time.Time `json:"timestamp"`
}

// ReviewUpdatedEvent 评论更新事件。
type ReviewUpdatedEvent struct {
	ReviewID  uint64    `json:"review_id"`
	Rating    int32     `json:"rating"`
	Timestamp time.Time `json:"timestamp"`
}

// ReviewDeletedEvent 评论删除事件。
type ReviewDeletedEvent struct {
	ReviewID  uint64    `json:"review_id"`
	Timestamp time.Time `json:"timestamp"`
}
