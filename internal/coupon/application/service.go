package application

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/coupon/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/coupon/domain/repository"
	"fmt"
	"time"

	"log/slog"
)

type CouponService struct {
	repo   repository.CouponRepository
	logger *slog.Logger
}

func NewCouponService(repo repository.CouponRepository, logger *slog.Logger) *CouponService {
	return &CouponService{
		repo:   repo,
		logger: logger,
	}
}

// CreateCoupon 创建优惠券
func (s *CouponService) CreateCoupon(ctx context.Context, name, description string, couponType entity.CouponType, discountAmount, minOrderAmount int64) (*entity.Coupon, error) {
	coupon := entity.NewCoupon(name, description, couponType, discountAmount, minOrderAmount)
	if err := s.repo.SaveCoupon(ctx, coupon); err != nil {
		s.logger.Error("failed to create coupon", "error", err)
		return nil, err
	}
	return coupon, nil
}

// ActivateCoupon 激活优惠券
func (s *CouponService) ActivateCoupon(ctx context.Context, id uint64) error {
	coupon, err := s.repo.GetCoupon(ctx, id)
	if err != nil {
		return err
	}

	if err := coupon.Activate(); err != nil {
		return err
	}

	return s.repo.UpdateCoupon(ctx, coupon)
}

// IssueCoupon 发放优惠券给用户
func (s *CouponService) IssueCoupon(ctx context.Context, userID, couponID uint64) (*entity.UserCoupon, error) {
	coupon, err := s.repo.GetCoupon(ctx, couponID)
	if err != nil {
		return nil, err
	}

	if err := coupon.CheckAvailability(); err != nil {
		return nil, err
	}

	// Check user limit
	userCoupons, _, err := s.repo.ListUserCoupons(ctx, userID, "", 0, 1000) // Simple check, optimize for production
	if err != nil {
		return nil, err
	}

	count := 0
	for _, uc := range userCoupons {
		if uc.CouponID == couponID {
			count++
		}
	}
	if int32(count) >= coupon.UsagePerUser {
		return nil, fmt.Errorf("user usage limit reached")
	}

	coupon.Issue(1)
	if err := s.repo.UpdateCoupon(ctx, coupon); err != nil {
		return nil, err
	}

	userCoupon := entity.NewUserCoupon(userID, couponID, coupon.CouponNo)
	if err := s.repo.SaveUserCoupon(ctx, userCoupon); err != nil {
		return nil, err
	}

	return userCoupon, nil
}

// UseCoupon 使用优惠券
func (s *CouponService) UseCoupon(ctx context.Context, userCouponID uint64, orderID string) error {
	userCoupon, err := s.repo.GetUserCoupon(ctx, userCouponID)
	if err != nil {
		return err
	}

	if err := userCoupon.Use(orderID); err != nil {
		return err
	}

	if err := s.repo.UpdateUserCoupon(ctx, userCoupon); err != nil {
		return err
	}

	coupon, err := s.repo.GetCoupon(ctx, userCoupon.CouponID)
	if err != nil {
		return err
	}
	coupon.Use()
	return s.repo.UpdateCoupon(ctx, coupon)
}

// ListCoupons 获取优惠券列表
func (s *CouponService) ListCoupons(ctx context.Context, page, pageSize int) ([]*entity.Coupon, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListCoupons(ctx, 0, offset, pageSize)
}

// ListUserCoupons 获取用户优惠券列表
func (s *CouponService) ListUserCoupons(ctx context.Context, userID uint64, status string, page, pageSize int) ([]*entity.UserCoupon, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListUserCoupons(ctx, userID, status, offset, pageSize)
}

// CreateActivity 创建活动
func (s *CouponService) CreateActivity(ctx context.Context, name, description string, startTime, endTime time.Time, couponIDs []uint64) (*entity.CouponActivity, error) {
	activity := entity.NewCouponActivity(name, description, startTime, endTime, couponIDs)
	if err := s.repo.SaveActivity(ctx, activity); err != nil {
		s.logger.Error("failed to create activity", "error", err)
		return nil, err
	}
	return activity, nil
}

// ListActiveActivities 获取进行中的活动
func (s *CouponService) ListActiveActivities(ctx context.Context) ([]*entity.CouponActivity, error) {
	return s.repo.ListActiveActivities(ctx, time.Now())
}
