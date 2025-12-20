package application

import (
	"context"
	"errors"
	"time"

	"github.com/wyfcoding/ecommerce/internal/advanced_coupon/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/advanced_coupon/domain/repository"
	"github.com/wyfcoding/pkg/algorithm"

	"log/slog"
)

// AdvancedCouponService 定义了 AdvancedCoupon 相关的服务逻辑。
type AdvancedCouponService struct {
	repo   repository.AdvancedCouponRepository
	logger *slog.Logger
}

// NewAdvancedCouponService 定义了 NewAdvancedCoupon 相关的服务逻辑。
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
		case entity.CouponTypePercentage:
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
		case entity.CouponTypeFixed:
			ac.Type = algorithm.CouponTypeReduction
			ac.ReductionAmount = coupon.DiscountValue
		case entity.CouponTypeFreeShipping:
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
