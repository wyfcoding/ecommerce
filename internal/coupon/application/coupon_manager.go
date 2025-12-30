package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/coupon/domain"
)

// CouponManager 处理优惠券模块的写操作和核心业务流程。
type CouponManager struct {
	repo   domain.CouponRepository
	logger *slog.Logger
}

// NewCouponManager 创建并返回一个新的 CouponManager 实例。
func NewCouponManager(repo domain.CouponRepository, logger *slog.Logger) *CouponManager {
	return &CouponManager{
		repo:   repo,
		logger: logger,
	}
}

// CreateCoupon 创建一个新的优惠券模板。
func (m *CouponManager) CreateCoupon(ctx context.Context, name, description string, couponType int, discountAmount, minOrderAmount int64) (*domain.Coupon, error) {
	coupon := domain.NewCoupon(name, description, domain.CouponType(couponType), discountAmount, minOrderAmount)
	if err := m.repo.SaveCoupon(ctx, coupon); err != nil {
		m.logger.Error("failed to create coupon", "error", err)
		return nil, err
	}
	return coupon, nil
}

// AcquireCoupon 用户领取优惠券。
func (m *CouponManager) AcquireCoupon(ctx context.Context, userID, couponID uint64) (*domain.UserCoupon, error) {
	// 1. 获取优惠券模板并检查可用性。
	coupon, err := m.repo.GetCoupon(ctx, couponID)
	if err != nil {
		return nil, err
	}
	if coupon == nil {
		return nil, fmt.Errorf("coupon not found")
	}

	if err := coupon.CheckAvailability(); err != nil {
		return nil, err
	}

	// 2. 检查用户是否已领超过限额。
	userCoupons, total, err := m.repo.ListUserCoupons(ctx, userID, "", 0, 1000)
	if err != nil {
		return nil, err
	}
	if total > 0 {
		count := 0
		for _, uc := range userCoupons {
			if uc.CouponID == couponID {
				count++
			}
		}
		if int32(count) >= coupon.UsagePerUser {
			return nil, fmt.Errorf("user has reached the limit for this coupon")
		}
	}

	// 3. 发放优惠券。
	userCoupon := domain.NewUserCoupon(userID, couponID, coupon.CouponNo)
	if err := m.repo.SaveUserCoupon(ctx, userCoupon); err != nil {
		m.logger.Error("failed to save user coupon", "error", err)
		return nil, err
	}

	// 4. 更新优惠券已发行量。
	coupon.Issue(1)
	if err := m.repo.UpdateCoupon(ctx, coupon); err != nil {
		m.logger.Error("failed to update coupon issued count", "error", err)
	}

	return userCoupon, nil
}

// UseCoupon 使用优惠券。
func (m *CouponManager) UseCoupon(ctx context.Context, userCouponID uint64, userID uint64, orderID string) error {
	userCoupon, err := m.repo.GetUserCoupon(ctx, userCouponID)
	if err != nil {
		return err
	}
	if userCoupon == nil {
		return fmt.Errorf("user coupon not found")
	}
	if userCoupon.UserID != userID {
		return fmt.Errorf("permission denied")
	}

	// 转换状态并使用。
	if err := userCoupon.Use(orderID); err != nil {
		return err
	}

	if err := m.repo.UpdateUserCoupon(ctx, userCoupon); err != nil {
		return err
	}

	// 更新模板的使用统计。
	coupon, err := m.repo.GetCoupon(ctx, userCoupon.CouponID)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to get coupon for stats update", "coupon_id", userCoupon.CouponID, "error", err)
	} else if coupon != nil {
		coupon.Use()
		if err := m.repo.UpdateCoupon(ctx, coupon); err != nil {
			m.logger.ErrorContext(ctx, "failed to update coupon stats", "coupon_id", coupon.ID, "error", err)
		}
	}

	return nil
}

// CreateActivity 创建营销活动。
func (m *CouponManager) CreateActivity(ctx context.Context, activity *domain.CouponActivity) error {
	return m.repo.SaveActivity(ctx, activity)
}
