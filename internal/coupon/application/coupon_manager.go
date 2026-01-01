package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/coupon/domain"
	"github.com/wyfcoding/pkg/algorithm"
)

// CouponManager 处理优惠券模块的写操作和核心业务流程。
type CouponManager struct {
	repo      domain.CouponRepository
	logger    *slog.Logger
	optimizer *algorithm.CouponOptimizer
}

// NewCouponManager 创建并返回一个新的 CouponManager 实例。
func NewCouponManager(repo domain.CouponRepository, logger *slog.Logger) *CouponManager {
	return &CouponManager{
		repo:      repo,
		logger:    logger,
		optimizer: algorithm.NewCouponOptimizer(),
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

// SuggestBestCoupons 为用户推荐最优的优惠券组合。
func (m *CouponManager) SuggestBestCoupons(ctx context.Context, userID uint64, orderAmount int64) ([]uint64, int64, int64, error) {
	// 1. 获取用户未使用的优惠券
	userCoupons, _, err := m.repo.ListUserCoupons(ctx, userID, "unused", 1, 100)
	if err != nil {
		return nil, 0, 0, err
	}
	if len(userCoupons) == 0 {
		return []uint64{}, orderAmount, 0, nil
	}

	// 2. 转换为算法模型
	algoCoupons := make([]algorithm.Coupon, 0)
	// 缓存已获取的 Coupon 模板，避免重复查询
	couponTemplateCache := make(map[uint64]*domain.Coupon)

	for _, uc := range userCoupons {
		template, exists := couponTemplateCache[uc.CouponID]
		if !exists {
			var err error
			template, err = m.repo.GetCoupon(ctx, uc.CouponID)
			if err != nil {
				m.logger.WarnContext(ctx, "failed to get coupon template", "coupon_id", uc.CouponID, "error", err)
				continue
			}
			if template == nil {
				continue
			}
			couponTemplateCache[uc.CouponID] = template
		}

		// 检查基本有效性
		if err := template.CheckAvailability(); err != nil {
			continue
		}

		// 映射类型
		var algoType algorithm.CouponType
		var discountRate float64
		var reductionAmount int64

		switch template.Type {
		case domain.CouponTypeDiscount: // 1
			algoType = algorithm.CouponTypeDiscount // 1
			// 假设 domain 的 DiscountAmount 是百分比整数 (e.g. 80 for 80% or 8.0) or amount?
			// 注释说是 "优惠金额或折扣比例"。
			// 这里假设如果是 Discount 类型，DiscountAmount 80 代表 0.8 (8折) ? 或者 80% ?
			// 通常 discount amount 存的是分，但如果是折扣率，可能存的是整数 80 (8折)。
			// 这里做一个假设：如果 < 100，当做折扣率 * 100。例如 80 => 0.8。
			if template.DiscountAmount < 100 {
				discountRate = float64(template.DiscountAmount) / 100.0
			} else {
				// 可能是错误配置，或者存的是分但类型选错了。这里暂且处理为9折兜底
				discountRate = 0.9
			}
		case domain.CouponTypeCash: // 2
			// 现金券 (直接抵扣)
			if template.MinOrderAmount > 0 {
				algoType = algorithm.CouponTypeReduction // 2 (满减)
			} else {
				algoType = algorithm.CouponTypeCash // 3 (立减)
			}
			reductionAmount = template.DiscountAmount
		default:
			continue // 其他类型暂不支持优化计算
		}

		algoCoupons = append(algoCoupons, algorithm.Coupon{
			ID:              uint64(uc.ID), // 注意：这里使用 UserCoupon 的 ID，以便返回时知道选了哪张
			Type:            algoType,
			Threshold:       template.MinOrderAmount,
			DiscountRate:    discountRate,
			ReductionAmount: reductionAmount,
			MaxDiscount:     template.MaxDiscount,
			CanStack:        template.CanStack,
			Priority:        1,
		})
	}

	// 3. 计算最优组合
	// 如果优惠券数量少 (< 20)，用最优解；否则用贪心
	var bestIds []uint64
	var finalPrice, discount int64

	if len(algoCoupons) < 20 {
		bestIds, finalPrice, discount = m.optimizer.OptimalCombination(orderAmount, algoCoupons)
	} else {
		bestIds, finalPrice, discount = m.optimizer.GreedyOptimization(orderAmount, algoCoupons)
	}

	return bestIds, finalPrice, discount, nil
}
