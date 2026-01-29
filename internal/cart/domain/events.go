package domain

import (
	"time"
)

// CartItemAddedEvent 商品加入购物车事件
type CartItemAddedEvent struct {
	UserID    uint64    `json:"user_id"`
	ProductID uint64    `json:"product_id"`
	SkuID     uint64    `json:"sku_id"`
	Quantity  int32     `json:"quantity"`
	Timestamp time.Time `json:"timestamp"`
}

// CartItemUpdatedEvent 购物车项更新事件
type CartItemUpdatedEvent struct {
	UserID    uint64    `json:"user_id"`
	SkuID     uint64    `json:"sku_id"`
	Quantity  int32     `json:"quantity"`
	Timestamp time.Time `json:"timestamp"`
}

// CartItemRemovedEvent 购物车项移除事件
type CartItemRemovedEvent struct {
	UserID    uint64    `json:"user_id"`
	SkuIDs    []uint64  `json:"sku_ids"`
	Timestamp time.Time `json:"timestamp"`
}

// CartClearedEvent 购物车清空事件
type CartClearedEvent struct {
	UserID    uint64    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}

// CartsMergedEvent 购物车合并事件
type CartsMergedEvent struct {
	SourceUserID uint64    `json:"source_user_id"`
	TargetUserID uint64    `json:"target_user_id"`
	Timestamp    time.Time `json:"timestamp"`
}

// CouponAppliedEvent 优惠券应用事件
type CouponAppliedEvent struct {
	UserID     uint64    `json:"user_id"`
	CouponCode string    `json:"coupon_code"`
	Timestamp  time.Time `json:"timestamp"`
}
