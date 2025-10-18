package model

import "time"

// Coupon represents a coupon entity in the business layer.
type Coupon struct {
	ID            uint
	Code          string
	Description   string
	DiscountValue float64
	DiscountType  string
	ValidFrom     time.Time
	ValidTo       time.Time
	IsActive      bool
}