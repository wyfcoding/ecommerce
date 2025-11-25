package application

import (
	"context"
	"errors"
	"github.com/wyfcoding/ecommerce/internal/advanced_coupon/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/advanced_coupon/domain/repository"
	"github.com/wyfcoding/ecommerce/pkg/algorithm"
	"time"

	"log/slog"
)

type AdvancedCouponService struct {
	repo   repository.AdvancedCouponRepository
	logger *slog.Logger
}

func NewAdvancedCouponService(repo repository.AdvancedCouponRepository, logger *slog.Logger) *AdvancedCouponService {
	return &AdvancedCouponService{
		repo:   repo,
		logger: logger,
	}
}

// CreateCoupon 创建优惠券
func (s *AdvancedCouponService) CreateCoupon(ctx context.Context, code string, couponType entity.CouponType, discountValue int64, validFrom, validUntil time.Time, totalQuantity int64) (*entity.Coupon, error) {
	coupon := entity.NewCoupon(code, couponType, discountValue, validFrom, validUntil, totalQuantity)
	if err := s.repo.Save(ctx, coupon); err != nil {
		s.logger.Error("failed to create coupon", "error", err)
		return nil, err
	}
	return coupon, nil
}

// GetCoupon 获取优惠券
func (s *AdvancedCouponService) GetCoupon(ctx context.Context, id uint64) (*entity.Coupon, error) {
	return s.repo.GetByID(ctx, id)
}

// ListCoupons 获取优惠券列表
func (s *AdvancedCouponService) ListCoupons(ctx context.Context, status entity.CouponStatus, page, pageSize int) ([]*entity.Coupon, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, status, offset, pageSize)
}

// UseCoupon 使用优惠券
func (s *AdvancedCouponService) UseCoupon(ctx context.Context, userID uint64, code string, orderID uint64) error {
	coupon, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return err
	}
	if coupon == nil {
		return errors.New("coupon not found")
	}

	if !coupon.IsValid() {
		return errors.New("coupon is invalid or expired")
	}

	// Check per user limit
	usedCount, err := s.repo.CountUsageByUser(ctx, userID, uint64(coupon.ID))
	if err != nil {
		return err
	}
	if usedCount >= coupon.PerUserLimit {
		return errors.New("coupon usage limit exceeded for user")
	}

	// Update coupon usage
	coupon.UsedQuantity++
	if err := s.repo.Save(ctx, coupon); err != nil {
		return err
	}

	// Record usage
	usage := &entity.CouponUsage{
		UserID:   userID,
		CouponID: uint64(coupon.ID),
		OrderID:  orderID,
		Code:     code,
		UsedAt:   time.Now(),
	}
	return s.repo.SaveUsage(ctx, usage)
}

// CalculateBestDiscount 计算最优优惠组合
func (s *AdvancedCouponService) CalculateBestDiscount(ctx context.Context, orderAmount int64, couponIDs []uint64) ([]uint64, int64, int64, error) {
	if len(couponIDs) == 0 {
		return nil, orderAmount, 0, nil
	}

	// Fetch coupons
	var algoCoupons []algorithm.Coupon
	for _, id := range couponIDs {
		coupon, err := s.repo.GetByID(ctx, id)
		if err != nil {
			continue // Skip invalid coupons or handle error
		}
		if coupon == nil || !coupon.IsValid() {
			continue
		}

		// Map to algorithm coupon
		ac := algorithm.Coupon{
			ID:        uint64(coupon.ID),
			Threshold: coupon.MinPurchaseAmount,
			CanStack:  true, // Default to true for now
			Priority:  1,
		}

		switch coupon.Type {
		case entity.CouponTypePercentage:
			ac.Type = algorithm.CouponTypeDiscount
			// DiscountValue is percentage (e.g., 20 for 20% off -> 0.8 rate? No, 20% off means 0.8 factor)
			// Assuming DiscountValue 20 means 20% off.
			// Algorithm expects DiscountRate as factor (0.8 for 20% off)
			// Wait, algorithm comment says: DiscountRate float64 // 折扣率（0.8表示8折）
			// If DiscountValue is 80 (8折), then rate is 0.8.
			// If DiscountValue is 20 (20% off), then rate is 0.8.
			// Let's assume DiscountValue is "percentage off" e.g. 20.
			ac.DiscountRate = 1.0 - float64(coupon.DiscountValue)/100.0
			ac.MaxDiscount = coupon.MaxDiscountAmount
		case entity.CouponTypeFixed:
			ac.Type = algorithm.CouponTypeReduction
			ac.ReductionAmount = coupon.DiscountValue
		case entity.CouponTypeFreeShipping:
			// Treat as cash reduction of 0 for price calc, or specific amount if we knew shipping cost.
			// For now, ignore or treat as small cash reduction?
			// Let's skip free shipping for price optimization or treat as 0 reduction.
			continue
		}

		algoCoupons = append(algoCoupons, ac)
	}

	optimizer := algorithm.NewCouponOptimizer()
	bestCombination, finalPrice, discount := optimizer.OptimalCombination(orderAmount, algoCoupons)

	return bestCombination, finalPrice, discount, nil
}
