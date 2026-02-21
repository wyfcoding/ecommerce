// 变更说明：
// 新增促销引擎服务，统一管理满减/满折/买赠/阶梯价/组合优惠等促销规则。
// 该服务是电商系统的核心缺失模块，之前促销逻辑散落在 coupon/marketing/pricing 等多个服务中。
// 促销引擎采用规则引擎模式，支持多规则叠加、互斥、优先级排序。
package domain

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/shopspring/decimal"
)

// 促销模块业务错误定义。
var (
	ErrPromotionNotFound     = errors.New("promotion not found")
	ErrPromotionExpired      = errors.New("promotion has expired")
	ErrPromotionNotStarted   = errors.New("promotion has not started")
	ErrPromotionDisabled     = errors.New("promotion is disabled")
	ErrPromotionConflict     = errors.New("promotion conflicts with existing rules")
	ErrConditionNotMet       = errors.New("promotion condition not met")
	ErrPromotionUsageLimitExceeded = errors.New("promotion usage limit exceeded")
)

// PromotionType 促销类型枚举。
type PromotionType string

const (
	PromotionTypeFullReduction PromotionType = "FULL_REDUCTION" // 满减：满 X 元减 Y 元。
	PromotionTypeFullDiscount  PromotionType = "FULL_DISCOUNT"  // 满折：满 X 元打 Y 折。
	PromotionTypeBuyNGetM      PromotionType = "BUY_N_GET_M"    // 买赠：买 N 件赠 M 件。
	PromotionTypeTieredPrice   PromotionType = "TIERED_PRICE"   // 阶梯价：买越多越便宜。
	PromotionTypeBundlePrice   PromotionType = "BUNDLE_PRICE"   // 组合价：指定商品组合特价。
	PromotionTypeSecondHalf    PromotionType = "SECOND_HALF"    // 第二件半价。
	PromotionTypeFixedPrice    PromotionType = "FIXED_PRICE"    // 一口价：N 件 X 元。
	PromotionTypeFreeShipping  PromotionType = "FREE_SHIPPING"  // 包邮：满 X 元免运费。
	PromotionTypeGift          PromotionType = "GIFT"           // 赠品：满足条件赠送指定商品。
)

// PromotionStatus 促销状态。
type PromotionStatus string

const (
	PromotionStatusDraft    PromotionStatus = "DRAFT"    // 草稿。
	PromotionStatusActive   PromotionStatus = "ACTIVE"   // 生效中。
	PromotionStatusPaused   PromotionStatus = "PAUSED"   // 已暂停。
	PromotionStatusExpired  PromotionStatus = "EXPIRED"  // 已过期。
	PromotionStatusDisabled PromotionStatus = "DISABLED" // 已禁用。
)

// PromotionScope 促销适用范围。
type PromotionScope string

const (
	PromotionScopeAll      PromotionScope = "ALL"      // 全场通用。
	PromotionScopeCategory PromotionScope = "CATEGORY" // 指定品类。
	PromotionScopeBrand    PromotionScope = "BRAND"    // 指定品牌。
	PromotionScopeProduct  PromotionScope = "PRODUCT"  // 指定商品。
	PromotionScopeSKU      PromotionScope = "SKU"      // 指定 SKU。
	PromotionScopeMerchant PromotionScope = "MERCHANT" // 指定商户。
)

// StackMode 叠加模式。
type StackMode string

const (
	StackModeExclusive StackMode = "EXCLUSIVE" // 互斥：不可与其他促销叠加。
	StackModeStackable StackMode = "STACKABLE" // 可叠加：可与其他促销同时生效。
	StackModeBest      StackMode = "BEST"      // 最优：自动选择最优惠的促销。
)

// Promotion 促销活动聚合根。
// 并发控制策略：乐观锁 (Version) + 分布式锁（创建/修改时）。
type Promotion struct {
	ID          uint64          `json:"id"`
	Name        string          `json:"name"`           // 促销名称。
	Description string          `json:"description"`    // 促销描述。
	Type        PromotionType   `json:"type"`           // 促销类型。
	Status      PromotionStatus `json:"status"`         // 促销状态。
	Scope       PromotionScope  `json:"scope"`          // 适用范围。
	StackMode   StackMode       `json:"stack_mode"`     // 叠加模式。
	Priority    int32           `json:"priority"`       // 优先级（数值越大优先级越高）。
	StartTime   time.Time       `json:"start_time"`     // 生效开始时间。
	EndTime     time.Time       `json:"end_time"`       // 生效结束时间。
	Rules       []*PromotionRule `json:"rules"`          // 促销规则列表（阶梯规则）。
	ScopeIDs    []uint64        `json:"scope_ids"`      // 适用范围 ID 列表。
	ExcludeIDs  []uint64        `json:"exclude_ids"`    // 排除的商品/品类 ID 列表。
	MerchantID  uint64          `json:"merchant_id"`    // 所属商户 ID（0 表示平台级）。
	UsageLimit  int64           `json:"usage_limit"`    // 总使用次数限制（0 表示不限）。
	UsedCount   int64           `json:"used_count"`     // 已使用次数。
	UserLimit   int32           `json:"user_limit"`     // 每用户使用次数限制。
	Label       string          `json:"label"`          // 促销标签（如"满200减30"）。
	Version     int64           `json:"version"`        // 乐观锁版本号。
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// PromotionRule 促销规则（支持阶梯规则）。
// 例如满减：满100减10、满200减30、满500减100。
type PromotionRule struct {
	ID          uint64          `json:"id"`
	PromotionID uint64          `json:"promotion_id"`
	Threshold   decimal.Decimal `json:"threshold"`    // 门槛值（金额或数量）。
	Discount    decimal.Decimal `json:"discount"`     // 优惠值（减免金额/折扣率/赠品数量）。
	GiftSKUID   uint64          `json:"gift_sku_id"`  // 赠品 SKU ID（买赠/赠品类型使用）。
	GiftQty     int32           `json:"gift_qty"`     // 赠品数量。
	FixedPrice  decimal.Decimal `json:"fixed_price"`  // 一口价金额。
	SortOrder   int32           `json:"sort_order"`   // 排序（阶梯规则从低到高）。
}

// IsActive 判断促销是否在有效期内且状态为生效。
func (p *Promotion) IsActive(now time.Time) bool {
	return p.Status == PromotionStatusActive &&
		!now.Before(p.StartTime) &&
		!now.After(p.EndTime)
}

// HasUsageQuota 判断促销是否还有使用额度。
func (p *Promotion) HasUsageQuota() bool {
	if p.UsageLimit <= 0 {
		return true
	}
	return p.UsedCount < p.UsageLimit
}

// IncrementUsage 增加使用次数。
func (p *Promotion) IncrementUsage() error {
	if p.UsageLimit > 0 && p.UsedCount >= p.UsageLimit {
		return ErrPromotionUsageLimitExceeded
	}
	p.UsedCount++
	return nil
}

// MatchRule 根据订单金额/数量匹配最优规则（阶梯匹配，取最高档）。
func (p *Promotion) MatchRule(amount decimal.Decimal, qty int32) *PromotionRule {
	if len(p.Rules) == 0 {
		return nil
	}

	// 按门槛从高到低排序。
	sorted := make([]*PromotionRule, len(p.Rules))
	copy(sorted, p.Rules)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Threshold.GreaterThan(sorted[j].Threshold)
	})

	for _, rule := range sorted {
		switch p.Type {
		case PromotionTypeTieredPrice, PromotionTypeBuyNGetM, PromotionTypeFixedPrice:
			// 按数量匹配。
			if decimal.NewFromInt32(qty).GreaterThanOrEqual(rule.Threshold) {
				return rule
			}
		default:
			// 按金额匹配。
			if amount.GreaterThanOrEqual(rule.Threshold) {
				return rule
			}
		}
	}
	return nil
}

// CalculateDiscount 计算促销优惠金额。
// amount: 参与促销的商品总金额。
// qty: 参与促销的商品总数量。
// 返回优惠金额（正数表示减免）。
func (p *Promotion) CalculateDiscount(amount decimal.Decimal, qty int32) decimal.Decimal {
	rule := p.MatchRule(amount, qty)
	if rule == nil {
		return decimal.Zero
	}

	switch p.Type {
	case PromotionTypeFullReduction:
		// 满减：直接减去优惠金额。
		return rule.Discount

	case PromotionTypeFullDiscount:
		// 满折：折扣率为 0.8 表示打八折，优惠 = 原价 * (1 - 折扣率)。
		discountRate := decimal.NewFromInt(1).Sub(rule.Discount)
		return amount.Mul(discountRate).Round(0)

	case PromotionTypeSecondHalf:
		// 第二件半价：每两件中第二件半价。
		pairs := int32(qty / 2)
		if pairs <= 0 {
			return decimal.Zero
		}
		avgPrice := amount.Div(decimal.NewFromInt32(qty))
		return avgPrice.Mul(decimal.NewFromFloat(0.5)).Mul(decimal.NewFromInt32(pairs)).Round(0)

	case PromotionTypeFixedPrice:
		// 一口价：N 件 X 元，优惠 = 原价 - 一口价。
		if decimal.NewFromInt32(qty).GreaterThanOrEqual(rule.Threshold) {
			groups := decimal.NewFromInt32(qty).Div(rule.Threshold).IntPart()
			remainder := qty - int32(groups)*int32(rule.Threshold.IntPart())
			avgPrice := amount.Div(decimal.NewFromInt32(qty))
			groupOriginal := avgPrice.Mul(rule.Threshold).Mul(decimal.NewFromInt(groups))
			groupFixed := rule.FixedPrice.Mul(decimal.NewFromInt(groups))
			_ = remainder
			return groupOriginal.Sub(groupFixed).Round(0)
		}
		return decimal.Zero

	case PromotionTypeFreeShipping:
		// 包邮：返回运费减免标记金额（由调用方处理）。
		return rule.Discount

	default:
		return decimal.Zero
	}
}

// CartItem 购物车商品项（用于促销计算的输入）。
type CartItem struct {
	ProductID  uint64          `json:"product_id"`
	SkuID      uint64          `json:"sku_id"`
	CategoryID uint64          `json:"category_id"`
	BrandID    uint64          `json:"brand_id"`
	MerchantID uint64          `json:"merchant_id"`
	Price      decimal.Decimal `json:"price"`     // 单价（分）。
	Quantity   int32           `json:"quantity"`   // 数量。
}

// TotalAmount 计算商品项总金额。
func (c *CartItem) TotalAmount() decimal.Decimal {
	return c.Price.Mul(decimal.NewFromInt32(c.Quantity))
}

// PromotionResult 促销计算结果。
type PromotionResult struct {
	PromotionID   uint64          `json:"promotion_id"`
	PromotionName string          `json:"promotion_name"`
	PromotionType PromotionType   `json:"promotion_type"`
	Label         string          `json:"label"`          // 促销标签。
	Discount      decimal.Decimal `json:"discount"`       // 优惠金额。
	GiftSKUID     uint64          `json:"gift_sku_id"`    // 赠品 SKU（如有）。
	GiftQty       int32           `json:"gift_qty"`       // 赠品数量。
	AppliedItems  []uint64        `json:"applied_items"`  // 参与促销的商品 SKU 列表。
}

// PromotionCalculation 促销计算总结果。
type PromotionCalculation struct {
	OriginalAmount decimal.Decimal    `json:"original_amount"` // 原始总金额。
	TotalDiscount  decimal.Decimal    `json:"total_discount"`  // 总优惠金额。
	FinalAmount    decimal.Decimal    `json:"final_amount"`    // 最终金额。
	Promotions     []*PromotionResult `json:"promotions"`      // 命中的促销列表。
	FreeShipping   bool               `json:"free_shipping"`   // 是否包邮。
}

// PromotionRepository 促销仓储接口。
type PromotionRepository interface {
	// Save 保存促销活动。
	Save(ctx context.Context, promotion *Promotion) error
	// GetByID 根据 ID 获取促销活动。
	GetByID(ctx context.Context, id uint64) (*Promotion, error)
	// ListActive 获取当前生效的促销活动列表。
	ListActive(ctx context.Context, now time.Time) ([]*Promotion, error)
	// ListByScope 根据适用范围获取促销活动。
	ListByScope(ctx context.Context, scope PromotionScope, scopeIDs []uint64, now time.Time) ([]*Promotion, error)
	// IncrementUsage 原子增加使用次数。
	IncrementUsage(ctx context.Context, id uint64) error
	// GetUserUsageCount 获取用户对某促销的使用次数。
	GetUserUsageCount(ctx context.Context, promotionID, userID uint64) (int32, error)
}

// PromotionReadRepository 促销读模型仓储接口。
type PromotionReadRepository interface {
	// ListActiveByProduct 获取指定商品可用的促销列表。
	ListActiveByProduct(ctx context.Context, productID uint64, now time.Time) ([]*Promotion, error)
	// ListActiveByCategory 获取指定品类可用的促销列表。
	ListActiveByCategory(ctx context.Context, categoryID uint64, now time.Time) ([]*Promotion, error)
	// ListActiveByMerchant 获取指定商户可用的促销列表。
	ListActiveByMerchant(ctx context.Context, merchantID uint64, now time.Time) ([]*Promotion, error)
}
