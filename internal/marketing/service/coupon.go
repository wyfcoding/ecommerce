package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ecommerce/internal/marketing/model"
	"ecommerce/internal/marketing/repository"

	"go.uber.org/zap"
)

type CouponService struct {
	repo repository.CouponRepo
	Log  *zap.SugaredLogger // 添加日志器
}

func NewCouponService(repo repository.CouponRepo, logger *zap.SugaredLogger) *CouponService {
	return &CouponService{repo: repo, Log: logger}
}

// CreateCouponTemplate 实现了创建优惠券模板的业务逻辑
func (s *CouponService) CreateCouponTemplate(ctx context.Context, template *model.CouponTemplate) (*model.CouponTemplate, error) {
	// 1. 业务规则校验
	if template.Title == "" {
		return nil, errors.New("title cannot be empty")
	}

	// 校验有效期
	if template.ValidityType == 1 { // 固定时间
		if template.ValidFrom == nil || template.ValidTo == nil || template.ValidTo.Before(*template.ValidFrom) {
			return nil, errors.New("invalid fixed validity period")
		}
	} else if template.ValidityType == 2 { // 领取后有效
		if template.ValidDaysAfterClaim == 0 {
			return nil, errors.New("valid days after claim cannot be empty")
		}
	} else {
		return nil, errors.New("invalid validity period type")
	}

	// 校验满减规则
	if template.Type == 1 { // 满减券
		if template.Rules.Threshold <= template.Rules.Discount {
			return nil, errors.New("discount amount of a tiered discount coupon must be less than the threshold")
		}
	}

	// ... 其他规则校验 ...

	// 2. 调用 repo 进行持久化
	return s.repo.CreateTemplate(ctx, template)
}

// ClaimCoupon 实现了用户领取优惠券的业务逻辑
func (s *CouponService) ClaimCoupon(ctx context.Context, userID, templateID uint64) (*model.UserCoupon, error) {
	// 在这里可以增加一些前置校验，例如通过缓存判断活动是否有效，以减少数据库压力
	// ...

	// 核心逻辑委托给 repo 层的事务方法
	return s.repo.ClaimCoupon(ctx, userID, templateID)
}

// CalculateDiscount 实现了计算优惠金额的核心业务逻辑
func (s *CouponService) CalculateDiscount(ctx context.Context, userID uint64, couponCode string, items []*model.OrderItemInfo) (uint64, error) {
	// 1. 获取优惠券实例并进行基础校验
	userCoupon, err := s.repo.GetUserCouponByCode(ctx, userID, couponCode)
	if err != nil {
		return 0, errors.New("invalid coupon code")
	}
	if userCoupon.Status != 1 { // 1-未使用
		return 0, errors.New("coupon has been used or has expired")
	}
	if time.Now().Before(userCoupon.ValidFrom) || time.Now().After(userCoupon.ValidTo) {
		return 0, errors.New("coupon is not within the validity period")
	}

	// 2. 获取优惠券模板以了解规则
	template, err := s.repo.GetTemplateByID(ctx, userCoupon.TemplateID)
	if err != nil {
		return 0, errors.New("coupon template not found")
	}

	// 3. 根据使用范围 (scope) 筛选出适用的商品，并计算总价
	var applicableTotal uint64 = 0
	scopeIDsMap := make(map[uint64]struct{}, len(template.ScopeIDs))
	for _, id := range template.ScopeIDs {
		scopeIDsMap[id] = struct{}{}
	}

	for _, item := range items {
		isApplicable := false
		switch template.ScopeType {
		case 1: // 全场通用
			isApplicable = true
		case 2: // 指定分类
			if _, ok := scopeIDsMap[item.CategoryID]; ok {
				isApplicable = true
			}
		case 3: // 指定SPU
			if _, ok := scopeIDsMap[item.SpuID]; ok {
				isApplicable = true
			}
		}
		if isApplicable {
			applicableTotal += item.Price * uint64(item.Quantity)
		}
	}
	if applicableTotal == 0 {
		return 0, errors.New("this coupon is not applicable to any selected product")
	}

	// 4. 根据优惠券类型 (type) 和规则 (rules) 计算优惠金额
	var discountAmount uint64 = 0
	rules := template.Rules
	switch template.Type {
	case 1: // 满减券
		if applicableTotal >= rules.Threshold {
			discountAmount = rules.Discount
		} else {
			return 0, fmt.Errorf("order total does not meet the threshold of ¥%.2f", float64(rules.Threshold)/100.0)
		}
	case 2: // 折扣券
		if applicableTotal >= rules.Threshold { // 折扣券也可能有门槛
			rawDiscount := float64(applicableTotal) * (100.0 - float64(rules.Discount)) / 100.0
			discountAmount = uint64(rawDiscount)
			if rules.MaxDeduction > 0 && discountAmount > rules.MaxDeduction {
				discountAmount = rules.MaxDeduction // 最高优惠限额
			}
		} else {
			return 0, fmt.Errorf("order total does not meet the discount threshold of ¥%.2f", float64(rules.Threshold)/100.0)
		}
	case 3: // 无门槛券
		discountAmount = rules.Discount
	default:
		return 0, fmt.Errorf("unknown coupon type")
	}

	// 最终校验优惠金额不能超过适用商品总价
	if discountAmount > applicableTotal {
		discountAmount = applicableTotal
	}

	return discountAmount, nil
}
