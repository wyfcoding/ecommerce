package model

import "time"

// PriceRule represents a pricing rule in the business logic layer.
type PriceRule struct {
	ID        uint
	ProductID uint64
	SKUID     uint64
	RuleType  string
	Price     uint64 // Price in cents
	ValidFrom *time.Time
	ValidTo   *time.Time
	Priority  int32
}

// Discount represents a discount applied in the business logic layer.
type Discount struct {
	ID            uint
	DiscountType  string
	DiscountID    uint64
	Amount        uint64 // Discount amount in cents
	AppliedTo     string
	AppliedItemID uint64
}

// SkuPriceInfo represents SKU price information for calculation.
type SkuPriceInfo struct {
	SkuID         uint64
	OriginalPrice uint64 // Original price in cents
	Quantity      uint32
}
