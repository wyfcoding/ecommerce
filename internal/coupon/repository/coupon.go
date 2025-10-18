package repository

import (
	"context"

	"ecommerce/internal/coupon/model"
)

// CouponRepo defines the data storage interface for coupons.
// The business layer depends on this interface, not on a concrete data implementation.
type CouponRepo interface {
	GetByCode(ctx context.Context, code string) (*model.Coupon, error)
	Create(ctx context.Context, coupon *model.Coupon) (*model.Coupon, error)
	// ... other data access methods
}