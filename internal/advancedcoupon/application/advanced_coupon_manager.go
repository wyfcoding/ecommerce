package application

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/advancedcoupon/domain"
)

// AdvancedCouponManager 处理所有高级优惠券相关的写入操作（Commands）。
type AdvancedCouponManager struct {
	repo   domain.AdvancedCouponRepository
	logger *slog.Logger
}

// NewAdvancedCouponManager 构造函数。
func NewAdvancedCouponManager(repo domain.AdvancedCouponRepository, logger *slog.Logger) *AdvancedCouponManager {
	return &AdvancedCouponManager{
		repo:   repo,
		logger: logger,
	}
}

// CreateCoupon 创建一个新的高级优惠券模板。
func (m *AdvancedCouponManager) CreateCoupon(ctx context.Context, code string, couponType domain.CouponType, discountValue int64, validFrom, validUntil time.Time, totalQuantity int64) (*domain.Coupon, error) {
	coupon := domain.NewCoupon(code, couponType, discountValue, validFrom, validUntil, totalQuantity)
	if err := m.repo.Save(ctx, coupon); err != nil {
		m.logger.Error("failed to create coupon", "error", err)
		return nil, err
	}
	return coupon, nil
}

// UseCoupon 核心核销逻辑：验证并使用指定的优惠券码。
func (m *AdvancedCouponManager) UseCoupon(ctx context.Context, userID uint64, code string, orderID uint64) error {
	return m.repo.Transaction(ctx, func(txCtx context.Context) error {
		// 1. 获取优惠券信息
		coupon, err := m.repo.GetByCode(txCtx, code)
		if err != nil {
			return err
		}
		if coupon == nil {
			return errors.New("coupon not found")
		}

		// 2. 基础校验
		if !coupon.IsValid() {
			return errors.New("coupon is invalid or expired")
		}

		// 3. 用户限领校验
		usedCount, err := m.repo.CountUsageByUser(txCtx, userID, uint64(coupon.ID))
		if err != nil {
			return err
		}
		if usedCount >= coupon.PerUserLimit {
			return errors.New("coupon usage limit exceeded for user")
		}

		// 4. 扣减库存
		if err := m.repo.IncrementUsage(txCtx, uint64(coupon.ID)); err != nil {
			m.logger.Warn("failed to increment usage", "code", code, "error", err)
			return err
		}

		// 5. 记录使用日志
		usage := &domain.CouponUsage{
			UserID:   userID,
			CouponID: uint64(coupon.ID),
			OrderID:  orderID,
			Code:     code,
			UsedAt:   time.Now(),
		}
		if err := m.repo.SaveUsage(txCtx, usage); err != nil {
			return err
		}

		m.logger.InfoContext(ctx, "coupon used successfully", "user_id", userID, "code", code, "order_id", orderID)
		return nil
	})
}
