package data

import (
	"context"
	"ecommerce/internal/coupon/biz"

	"gorm.io/gorm"
)

// couponRepo is the data layer implementation for CouponRepo.
type couponRepo struct {
	data *Data
	// log  *log.Helper
}

// toBiz converts a data.Coupon model to a biz.Coupon entity.
func (c *Coupon) toBiz() *biz.Coupon {
	if c == nil {
		return nil
	}
	return &biz.Coupon{
		ID:            c.ID,
		Code:          c.Code,
		Description:   c.Description,
		DiscountValue: c.DiscountValue,
		DiscountType:  c.DiscountType,
		ValidFrom:     c.ValidFrom,
		ValidTo:       c.ValidTo,
		IsActive:      c.IsActive,
	}
}

// GetByCode finds a coupon in the database by its code.
func (r *couponRepo) GetByCode(ctx context.Context, code string) (*biz.Coupon, error) {
	var coupon Coupon
	// Note: r.data.db is a placeholder for the actual gorm.DB instance
	// which should be initialized in the `NewData` function in `data.go`.
	if err := r.data.db.WithContext(ctx).Where("code = ?", code).First(&coupon).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, biz.ErrCouponNotFound
		}
		return nil, err
	}

	return coupon.toBiz(), nil
}

// Create saves a new coupon to the database.
func (r *couponRepo) Create(ctx context.Context, b *biz.Coupon) (*biz.Coupon, error) {
	coupon := &Coupon{
		Code:          b.Code,
		Description:   b.Description,
		DiscountValue: b.DiscountValue,
		DiscountType:  b.DiscountType,
		ValidFrom:     b.ValidFrom,
		ValidTo:       b.ValidTo,
		IsActive:      true, // Default to active
	}

	// Note: r.data.db is a placeholder for the actual gorm.DB instance.
	if err := r.data.db.WithContext(ctx).Create(coupon).Error; err != nil {
		return nil, err
	}

	return coupon.toBiz(), nil
}
