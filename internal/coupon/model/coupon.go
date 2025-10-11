package biz

import (
	"context"
	"errors"
	"time"
)

// ErrCouponNotFound is a specific error for when a coupon is not found.
var ErrCouponNotFound = errors.New("coupon not found")

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

// CouponRepo defines the data storage interface for coupons.
// The business layer depends on this interface, not on a concrete data implementation.
type CouponRepo interface {
	GetByCode(ctx context.Context, code string) (*Coupon, error)
	Create(ctx context.Context, coupon *Coupon) (*Coupon, error)
	// ... other data access methods
}

// CouponUsecase is the use case for coupon-related operations.
// It orchestrates the business logic.
type CouponUsecase struct {
	repo CouponRepo
	// You can also inject other dependencies like a logger
}

// NewCouponUsecase creates a new CouponUsecase.
func NewCouponUsecase(repo CouponRepo) *CouponUsecase {
	return &CouponUsecase{repo: repo}
}

// GetByCode retrieves a coupon by its code.
func (uc *CouponUsecase) GetByCode(ctx context.Context, code string) (*Coupon, error) {
	// Here you can add more business logic, e.g., validation, logging, etc.
	return uc.repo.GetByCode(ctx, code)
}

// CreateCoupon creates a new coupon.
func (uc *CouponUsecase) CreateCoupon(ctx context.Context, coupon *Coupon) (*Coupon, error) {
	// Here you can add business logic before creating, e.g., validating the dates, discount value, etc.
	return uc.repo.Create(ctx, coupon)
}
