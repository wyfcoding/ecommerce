package biz

import (
	"context"

	"go.uber.org/zap"
)

// --- Domain Models ---

// OrderItemInfo is a simplified model of order item information for coupon calculation.
type OrderItemInfo struct {
	SkuID      uint64
	SpuID      uint64
	CategoryID uint64
	Price      uint64 // Unit price of the product (in cents)
	Quantity   uint32
}

// RuleSet defines the specific rules of a coupon
type RuleSet struct {
	Threshold    uint64     // Threshold for discount (in cents)
	Discount     uint64     // Discount amount (in cents) or discount value (e.g., 88 for 12% off)
	MaxDeduction uint64     // Maximum deduction for a discount coupon (in cents)
}

// CouponTemplate is the domain model for a coupon template
type CouponTemplate struct {
	ID                  uint64
	Title               string
	Type                int8
	ScopeType           int8
	ScopeIDs            []uint64
	Rules               RuleSet
	TotalQuantity       uint
	IssuedQuantity      uint
	PerUserLimit        uint8
	ValidityType        int8
	ValidFrom           *time.Time
	ValidTo             *time.Time
	ValidDaysAfterClaim uint
	Status              int8
}

// UserCoupon is the domain model for a user coupon
type UserCoupon struct {
	ID         uint64
	TemplateID uint64
	UserID     uint64
	CouponCode string
	Status     int8
	ClaimedAt  time.Time
	ValidFrom  time.Time
	ValidTo    time.Time
}

// Promotion is the domain model for a promotion
type Promotion struct {
	ID             uint64
	Name           string
	Type           int8 // 1: Limited-time discount, 2: Tiered discount, 3: Gift with purchase
	Description    string
	StartTime      *time.Time
	EndTime        *time.Time
	Status         *int8 // 1: In progress, 2: Ended, 3: Not started, 4: Disabled
	ProductIDs     []uint64 // Associated product ID
	DiscountValue  uint64 // Discount value or tiered discount amount
	MinOrderAmount uint64 // Minimum order amount (for tiered discount)
}

// --- Repo Interface ---

type CouponRepo interface {
	CreateTemplate(ctx context.Context, template *CouponTemplate) (*CouponTemplate, error)
	ClaimCoupon(ctx context.Context, userID, templateID uint64) (*UserCoupon, error)
	GetUserCouponByCode(ctx context.Context, userID uint64, code string) (*UserCoupon, error)
	GetTemplateByID(ctx context.Context, templateID uint64) (*CouponTemplate, error)
	ListUserCoupons(ctx context.Context, userID uint64, status int8) ([]*UserCoupon, error)
	UpdateUserCouponStatus(ctx context.Context, userCouponID uint64, newStatus int8, orderID *uint64) error
}

type PromotionRepo interface {
	CreatePromotion(ctx context.Context, promotion *Promotion) (*Promotion, error)
	UpdatePromotion(ctx context.Context, promotion *Promotion) (*Promotion, error)
	DeletePromotion(ctx context.Context, id uint64) error
	GetPromotion(ctx context.Context, id uint64) (*Promotion, error)
	ListPromotions(ctx context.Context, pageSize, pageNum uint32, name *string, promoType *uint32, status *uint32) ([]*Promotion, uint64, error)
}

type CouponUsecase struct {
	repo CouponRepo
	Log  *zap.SugaredLogger // 添加日志器
}

func NewCouponUsecase(repo CouponRepo, logger *zap.SugaredLogger) *CouponUsecase {
	return &CouponUsecase{repo: repo, Log: logger}
}

// CreateCouponTemplate 实现了创建优惠券模板的业务逻辑
func (uc *CouponUsecase) CreateCouponTemplate(ctx context.Context, template *CouponTemplate) (*CouponTemplate, error) {
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
	return uc.repo.CreateTemplate(ctx, template)
}

// ClaimCoupon 实现了用户领取优惠券的业务逻辑
func (uc *CouponUsecase) ClaimCoupon(ctx context.Context, userID, templateID uint64) (*UserCoupon, error) {
	// 在这里可以增加一些前置校验，例如通过缓存判断活动是否有效，以减少数据库压力
	// ...

	// 核心逻辑委托给 repo 层的事务方法
	return uc.repo.ClaimCoupon(ctx, userID, templateID)
}

// CalculateDiscount 实现了计算优惠金额的核心业务逻辑
func (uc *CouponUsecase) CalculateDiscount(ctx context.Context, userID uint64, couponCode string, items []*OrderItemInfo) (uint64, error) {
	// 1. 获取优惠券实例并进行基础校验
	userCoupon, err := uc.repo.GetUserCouponByCode(ctx, userID, couponCode)
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
	template, err := uc.repo.GetTemplateByID(ctx, userCoupon.TemplateID)
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

// PromotionUsecase 是促销用例
type PromotionUsecase struct {
	repo PromotionRepo
	Log  *zap.SugaredLogger
}

// NewPromotionUsecase 创建一个新的 PromotionUsecase 实例
func NewPromotionUsecase(repo PromotionRepo, logger *zap.SugaredLogger) *PromotionUsecase {
	return &PromotionUsecase{repo: repo, Log: logger}
}

// CreatePromotion 实现了创建促销的业务逻辑
func (uc *PromotionUsecase) CreatePromotion(ctx context.Context, promotion *Promotion) (*Promotion, error) {
	// 业务规则校验
	if promotion.Name == "" {
		return nil, errors.New("promotion name cannot be empty")
	}
	if promotion.StartTime == nil || promotion.EndTime == nil || promotion.EndTime.Before(*promotion.StartTime) {
		return nil, errors.New("invalid promotion time range")
	}
	// ... 其他规则校验

	return uc.repo.CreatePromotion(ctx, promotion)
}

// UpdatePromotion 实现了更新促销的业务逻辑
func (uc *PromotionUsecase) UpdatePromotion(ctx context.Context, promotion *Promotion) (*Promotion, error) {
	// 业务规则校验
	if promotion.ID == 0 {
		return nil, errors.New("ID is required to update a promotion")
	}
	// ... 其他规则校验

	return uc.repo.UpdatePromotion(ctx, promotion)
}

// DeletePromotion 实现了删除促销的业务逻辑
func (uc *PromotionUsecase) DeletePromotion(ctx context.Context, id uint64) error {
	return uc.repo.DeletePromotion(ctx, id)
}

// GetPromotion 实现了获取促销详情的业务逻辑
func (uc *PromotionUsecase) GetPromotion(ctx context.Context, id uint64) (*Promotion, error) {
	return uc.repo.GetPromotion(ctx, id)
}

// ListPromotions 实现了获取促销列表的业务逻辑
func (uc *PromotionUsecase) ListPromotions(ctx context.Context, pageSize, pageNum uint32, name *string, promoType *uint32, status *uint32) ([]*Promotion, uint64, error) {
	return uc.repo.ListPromotions(ctx, pageSize, pageNum, name, promoType, status)
}