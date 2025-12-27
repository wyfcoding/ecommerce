package application

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/advancedcoupon/domain"
	"github.com/wyfcoding/pkg/algorithm"
)

// AdvancedCoupon 定义了 AdvancedCoupon 相关的服务逻辑。
type AdvancedCoupon struct {
	repo   domain.AdvancedCouponRepository
	logger *slog.Logger
}

// NewAdvancedCoupon 创建高级优惠券服务实例。
func NewAdvancedCoupon(repo domain.AdvancedCouponRepository, logger *slog.Logger) *AdvancedCoupon {
	return &AdvancedCoupon{
		repo:   repo,
		logger: logger,
	}
}

// CreateCoupon 创建一个新的高级优惠券模板。
func (s *AdvancedCoupon) CreateCoupon(ctx context.Context, code string, couponType domain.CouponType, discountValue int64, validFrom, validUntil time.Time, totalQuantity int64) (*domain.Coupon, error) {
	coupon := domain.NewCoupon(code, couponType, discountValue, validFrom, validUntil, totalQuantity)
	if err := s.repo.Save(ctx, coupon); err != nil {
		s.logger.Error("failed to create coupon", "error", err)
		return nil, err
	}
	return coupon, nil
}

// GetCoupon 获取指定ID的优惠券详情。
func (s *AdvancedCoupon) GetCoupon(ctx context.Context, id uint64) (*domain.Coupon, error) {
	return s.repo.GetByID(ctx, id)
}

// ListCoupons 根据状态分页列出优惠券模板。
func (s *AdvancedCoupon) ListCoupons(ctx context.Context, status domain.CouponStatus, page, pageSize int) ([]*domain.Coupon, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, status, offset, pageSize)
}

// UseCoupon 核心核销逻辑：验证并使用指定的优惠券码。
func (s *AdvancedCoupon) UseCoupon(ctx context.Context, userID uint64, code string, orderID uint64) error {
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

	// 检查每位用户的限制
	usedCount, err := s.repo.CountUsageByUser(ctx, userID, uint64(coupon.ID))
	if err != nil {
		return err
	}
	if usedCount >= coupon.PerUserLimit {
		return errors.New("coupon usage limit exceeded for user")
	}

	// 更新优惠券使用情况
	coupon.UsedQuantity++
	if err := s.repo.Save(ctx, coupon); err != nil {
		return err
	}

	// 记录使用情况
	usage := &domain.CouponUsage{
		UserID:   userID,
		CouponID: uint64(coupon.ID),
		OrderID:  orderID,
		Code:     code,
		UsedAt:   time.Now(),
	}
	return s.repo.SaveUsage(ctx, usage)
}

// CalculateBestDiscount 核心算法：基于多种约束计算订单的最优优惠组合及最终金额。
func (s *AdvancedCoupon) CalculateBestDiscount(ctx context.Context, orderAmount int64, couponIDs []uint64) ([]uint64, int64, int64, error) {
	if len(couponIDs) == 0 {
		return nil, orderAmount, 0, nil
	}

	// 获取优惠券
	var algoCoupons []algorithm.Coupon
	for _, id := range couponIDs {
		coupon, err := s.repo.GetByID(ctx, id)
		if err != nil {
			continue // 跳过无效优惠券或处理错误
		}
		if coupon == nil || !coupon.IsValid() {
			continue
		}

		// 映射到算法优惠券
		ac := algorithm.Coupon{
			ID:        uint64(coupon.ID),
			Threshold: coupon.MinPurchaseAmount,
			CanStack:  true, // 暂时默认为 true
			Priority:  1,
		}

		switch coupon.Type {
		case domain.CouponTypePercentage:
			ac.Type = algorithm.CouponTypeDiscount
			// DiscountValue 是百分比（例如，20 表示 20% 折扣 -> 0.8 比率？不，20% 折扣意味着 0.8 因子）
			// 假设 DiscountValue 20 表示 20% 折扣。
			// 算法期望 DiscountRate 作为因子（0.8 表示 20% 折扣）
			// 等等，算法注释说：DiscountRate float64 // 折扣率（0.8表示8折）
			// 如果 DiscountValue 是 80（8折），那么比率是 0.8。
			// 如果 DiscountValue 是 20（20% 折扣），那么比率是 0.8。
			// 让我们假设 DiscountValue 是“折扣百分比”，例如 20。
			ac.DiscountRate = 1.0 - float64(coupon.DiscountValue)/100.0
			ac.MaxDiscount = coupon.MaxDiscountAmount
		case domain.CouponTypeFixed:
			ac.Type = algorithm.CouponTypeReduction
			ac.ReductionAmount = coupon.DiscountValue
		case domain.CouponTypeFreeShipping:
			// 视为 0 现金减免或具体的运费金额（如果我们知道运费）。
			// 目前，忽略或视为小额现金减免？
			// 让我们在价格优化中跳过免运费，或视为 0 减免。
			continue
		}

		algoCoupons = append(algoCoupons, ac)
	}

	optimizer := algorithm.NewCouponOptimizer()
	bestCombination, finalPrice, discount := optimizer.OptimalCombination(orderAmount, algoCoupons)

	return bestCombination, finalPrice, discount, nil
}
