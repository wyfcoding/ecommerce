package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"ecommerce/internal/coupon/model"
	"ecommerce/internal/coupon/repository"
)

// CouponService 定义了优惠券服务的业务逻辑接口
type CouponService interface {
	AssignCouponToUser(ctx context.Context, userID uint, couponCode string) (*model.UserCoupon, error)
	GetUserCoupons(ctx context.Context, userID uint, status model.CouponStatus) ([]model.UserCoupon, error)
	// CalculateDiscount 是给购物车/订单服务调用的核心 gRPC 方法
	CalculateDiscount(ctx context.Context, userID uint, userCouponID uint, orderTotal float64) (float64, error)
	// RedeemCoupon 在订单创建成功后被调用
	RedeemCoupon(ctx context.Context, userID uint, userCouponID uint, orderID uint) error
}

// couponService 是接口的具体实现
type couponService struct {
	repo   repository.CouponRepository
	logger *zap.Logger
}

// NewCouponService 创建一个新的 couponService 实例
func NewCouponService(repo repository.CouponRepository, logger *zap.Logger) CouponService {
	return &couponService{repo: repo, logger: logger}
}

// AssignCouponToUser 用户领取一张优惠券
func (s *couponService) AssignCouponToUser(ctx context.Context, userID uint, couponCode string) (*model.UserCoupon, error) {
	s.logger.Info("Assigning coupon to user", zap.Uint("userID", userID), zap.String("couponCode", couponCode))

	// 1. 查找优惠券定义
	def, err := s.repo.GetDefinitionByCode(ctx, couponCode)
	if err != nil || def == nil {
		return nil, fmt.Errorf("优惠券代码无效")
	}

	// 2. 检查优惠券是否有效和可领取
	now := time.Now()
	if !def.IsActive || now.Before(def.ValidFrom) || now.After(def.ValidTo) {
		return nil, fmt.Errorf("优惠券当前不可用")
	}
	if def.TotalQuantity > 0 && def.IssuedQuantity >= def.TotalQuantity {
		return nil, fmt.Errorf("优惠券已被领完")
	}

	// 3. TODO: 检查用户领取资格和领取次数限制

	// 4. 在事务中增加已发行数量并创建用户优惠券
	userCoupon := &model.UserCoupon{
		UserID:             userID,
		CouponDefinitionID: def.ID,
		Status:             model.StatusUnused,
		AssignedAt:         now,
	}

	if err := s.repo.CreateUserCoupon(ctx, userCoupon); err != nil {
		return nil, err // 假设数据库有唯一性约束防止重复领取
	}
	s.repo.IncrementIssuedQuantity(ctx, def.ID, 1) // 发行量增加

	return userCoupon, nil
}

// GetUserCoupons 获取用户的优惠券列表
func (s *couponService) GetUserCoupons(ctx context.Context, userID uint, status model.CouponStatus) ([]model.UserCoupon, error) {
	return s.repo.ListUserCoupons(ctx, userID, status)
}

// CalculateDiscount 验证优惠券并计算折扣金额
func (s *couponService) CalculateDiscount(ctx context.Context, userID uint, userCouponID uint, orderTotal float64) (float64, error) {
	s.logger.Info("Calculating discount", zap.Uint("userCouponID", userCouponID), zap.Float64("orderTotal", orderTotal))

	// 1. 获取用户优惠券及其定义
	uc, err := s.repo.GetUserCoupon(ctx, userID, userCouponID)
	if err != nil || uc == nil {
		return 0, fmt.Errorf("无效的用户优惠券ID")
	}

	// 2. 验证优惠券的有效性
	if uc.Status != model.StatusUnused {
		return 0, fmt.Errorf("优惠券已被使用或已过期")
	}
	def := uc.CouponDefinition
	now := time.Now()
	if !def.IsActive || now.Before(def.ValidFrom) || now.After(def.ValidTo) {
		return 0, fmt.Errorf("优惠券不在有效期内")
	}
	if orderTotal < def.MinSpend {
		return 0, fmt.Errorf("订单金额未达到最低消费: %.2f", def.MinSpend)
	}

	// 3. 计算折扣金额
	var discount float64
	switch def.DiscountType {
	case model.DiscountTypeFixed:
		discount = def.DiscountValue
	case model.DiscountTypePercentage:
		discount = orderTotal * (def.DiscountValue / 100)
		if def.MaxDiscount > 0 && discount > def.MaxDiscount {
			discount = def.MaxDiscount // 应用最大折扣限制
		}
	default:
		return 0, fmt.Errorf("不支持的优惠类型")
	}

	// 确保折扣不会超过订单总额
	if discount > orderTotal {
		discount = orderTotal
	}

	return discount, nil
}

// RedeemCoupon 核销优惠券
func (s *couponService) RedeemCoupon(ctx context.Context, userID uint, userCouponID uint, orderID uint) error {
	s.logger.Info("Redeeming coupon", zap.Uint("userCouponID", userCouponID), zap.Uint("orderID", orderID))
	// 直接调用 repository 的原子操作
	return s.repo.RedeemUserCoupon(ctx, userID, userCouponID, orderID)
}