package domain

import (
	"time"
)

// CouponCreatedEvent 优惠券模板创建事件
type CouponCreatedEvent struct {
	CouponID       uint      `json:"coupon_id"`
	CouponNo       string    `json:"coupon_no"`
	Name           string    `json:"name"`
	DiscountAmount int64     `json:"discount_amount"`
	Timestamp      time.Time `json:"timestamp"`
}

// CouponIssuedEvent 优惠券发放事件
type CouponIssuedEvent struct {
	UserCouponID uint      `json:"user_coupon_id"`
	UserID       uint64    `json:"user_id"`
	CouponID     uint64    `json:"coupon_id"`
	CouponNo     string    `json:"coupon_no"`
	Timestamp    time.Time `json:"timestamp"`
}

// CouponUsedEvent 优惠券核销事件
type CouponUsedEvent struct {
	UserCouponID uint      `json:"user_coupon_id"`
	UserID       uint64    `json:"user_id"`
	OrderID      string    `json:"order_id"`
	Timestamp    time.Time `json:"timestamp"`
}

// CouponExpiredEvent 优惠券过期事件
type CouponExpiredEvent struct {
	UserCouponID uint      `json:"user_coupon_id"`
	Timestamp    time.Time `json:"timestamp"`
}
