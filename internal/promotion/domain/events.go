// 变更说明：
// 新增促销领域事件定义。
package domain

import "time"

// 促销领域事件类型常量。
const (
	PromotionCreatedEventType  = "promotion.created"
	PromotionActivatedEventType = "promotion.activated"
	PromotionPausedEventType   = "promotion.paused"
	PromotionExpiredEventType  = "promotion.expired"
	PromotionUsedEventType     = "promotion.used"
)

// PromotionCreatedEvent 促销创建事件。
type PromotionCreatedEvent struct {
	PromotionID uint64        `json:"promotion_id"`
	Name        string        `json:"name"`
	Type        PromotionType `json:"type"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	Timestamp   time.Time     `json:"timestamp"`
}

// PromotionUsedEvent 促销使用事件。
type PromotionUsedEvent struct {
	PromotionID uint64 `json:"promotion_id"`
	OrderID     uint64 `json:"order_id"`
	UserID      uint64 `json:"user_id"`
	DiscountAmt int64  `json:"discount_amt"` // 优惠金额（分）。
	Timestamp   time.Time `json:"timestamp"`
}
