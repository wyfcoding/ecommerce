package application

import (
	"context"
	"ecommerce/internal/advanced_coupon/domain/entity"
	"ecommerce/internal/advanced_coupon/domain/repository"
	"errors"
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
