package model

import "time"

// OrderItemInfo is a simplified model of order item information for coupon calculation.
type OrderItemInfo struct {
	SkuID      uint64
	SpuID      uint64
	CategoryID uint64
	Price      uint64 // Unit price of the product (in cents)
	Quantity   uint32
}

// RuleSet defines the specific rules of a coupon
type RuleSet struct {
	Threshold    uint64 // Threshold for discount (in cents)
	Discount     uint64 // Discount amount (in cents) or discount value (e.g., 88 for 12% off)
	MaxDeduction uint64 // Maximum deduction for a discount coupon (in cents)
}

// CouponTemplate is the domain model for a coupon template
type CouponTemplate struct {
	ID                  uint64
	Title               string
	Type                int8
	ScopeType           int8
	ScopeIDs            []uint64
	Rules               RuleSet
	TotalQuantity       uint
	IssuedQuantity      uint
	PerUserLimit        uint8
	ValidityType        int8
	ValidFrom           *time.Time
	ValidTo             *time.Time
	ValidDaysAfterClaim uint
	Status              int8
}

// UserCoupon is the domain model for a user coupon
type UserCoupon struct {
	ID         uint64
	TemplateID uint64
	UserID     uint64
	CouponCode string
	Status     int8
	ClaimedAt  time.Time
	ValidFrom  time.Time
	ValidTo    time.Time
}

// Promotion is the domain model for a promotion
type Promotion struct {
	ID             uint64
	Name           string
	Type           int8 // 1: Limited-time discount, 2: Tiered discount, 3: Gift with purchase
	Description    string
	StartTime      *time.Time
	EndTime        *time.Time
	Status         *int8    // 1: In progress, 2: Ended, 3: Not started, 4: Disabled
	ProductIDs     []uint64 // Associated product ID
	DiscountValue  uint64   // Discount value or tiered discount amount
	MinOrderAmount uint64   // Minimum order amount (for tiered discount)
}
