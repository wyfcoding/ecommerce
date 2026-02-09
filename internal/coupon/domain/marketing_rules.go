// 变更说明：新增营销规则增强功能，支持阶梯满减、买赠活动、限购规则、会员专属价。
// 假设：阶梯满减按最高档位计算，买赠活动赠品不计入限购。
package domain

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// --- 促销规则类型 ---

// PromotionRuleType 促销规则类型
type PromotionRuleType int

const (
	RuleTieredDiscount PromotionRuleType = 1  // 阶梯满减
	RuleBuyGetFree     PromotionRuleType = 2  // 买赠
	RulePurchaseLimit  PromotionRuleType = 3  // 限购
	RuleMemberPrice    PromotionRuleType = 4  // 会员价
	RuleNewUserPrice   PromotionRuleType = 5  // 新人价
	RuleBundlePrice    PromotionRuleType = 6  // 套餐价
	RuleFlashSale      PromotionRuleType = 7  // 秒杀
	RuleGroupBuy       PromotionRuleType = 8  // 拼团
	RuleCashback       PromotionRuleType = 9  // 返现
	RuleReferralReward PromotionRuleType = 10 // 分享返利
)

// --- 促销规则状态 ---

// PromotionRuleStatus 促销规则状态
type PromotionRuleStatus int

const (
	RuleStatusDraft   PromotionRuleStatus = 1 // 草稿
	RuleStatusPending PromotionRuleStatus = 2 // 待生效
	RuleStatusActive  PromotionRuleStatus = 3 // 生效中
	RuleStatusPaused  PromotionRuleStatus = 4 // 已暂停
	RuleStatusExpired PromotionRuleStatus = 5 // 已过期
	RuleStatusDeleted PromotionRuleStatus = 6 // 已删除
)

// --- 阶梯满减规则 ---

// TieredDiscountRule 阶梯满减规则
type TieredDiscountRule struct {
	ID          uint64              `json:"id"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Tiers       []*DiscountTier     `json:"tiers"`       // 阶梯配置
	ApplyScope  *ApplyScope         `json:"apply_scope"` // 适用范围
	StartTime   time.Time           `json:"start_time"`
	EndTime     time.Time           `json:"end_time"`
	Status      PromotionRuleStatus `json:"status"`
	StackType   string              `json:"stack_type"`  // EXCLUSIVE/STACKABLE 是否可叠加
	Priority    int                 `json:"priority"`    // 优先级
	UsageLimit  int32               `json:"usage_limit"` // 总使用次数限制
	UsedCount   int32               `json:"used_count"`  // 已使用次数
}

// DiscountTier 满减阶梯
type DiscountTier struct {
	Threshold     int64  `json:"threshold"`      // 满足金额（分）
	DiscountType  string `json:"discount_type"`  // AMOUNT/PERCENT 减免类型
	DiscountValue int64  `json:"discount_value"` // 减免值（分或百分比*100）
	MaxDiscount   int64  `json:"max_discount"`   // 最大减免（百分比时用）
}

// ApplyScope 适用范围
type ApplyScope struct {
	Type        string   `json:"type"`         // ALL/CATEGORY/PRODUCT/MERCHANT
	IncludeIDs  []uint64 `json:"include_ids"`  // 包含的ID
	ExcludeIDs  []uint64 `json:"exclude_ids"`  // 排除的ID
	MemberLevel []string `json:"member_level"` // 适用会员等级
}

// NewTieredDiscountRule 创建阶梯满减规则
func NewTieredDiscountRule(name, description string, startTime, endTime time.Time) *TieredDiscountRule {
	return &TieredDiscountRule{
		Name:        name,
		Description: description,
		Tiers:       make([]*DiscountTier, 0),
		StartTime:   startTime,
		EndTime:     endTime,
		Status:      RuleStatusDraft,
		StackType:   "EXCLUSIVE",
		Priority:    0,
	}
}

// AddTier 添加阶梯
func (r *TieredDiscountRule) AddTier(threshold, discountValue, maxDiscount int64, discountType string) {
	tier := &DiscountTier{
		Threshold:     threshold,
		DiscountType:  discountType,
		DiscountValue: discountValue,
		MaxDiscount:   maxDiscount,
	}
	r.Tiers = append(r.Tiers, tier)
	// 按阈值排序
	sort.Slice(r.Tiers, func(i, j int) bool {
		return r.Tiers[i].Threshold < r.Tiers[j].Threshold
	})
}

// CalculateDiscount 计算满减金额
func (r *TieredDiscountRule) CalculateDiscount(orderAmount int64) int64 {
	if !r.IsActive() {
		return 0
	}

	// 找到适用的最高阶梯
	var applicableTier *DiscountTier
	for i := len(r.Tiers) - 1; i >= 0; i-- {
		if orderAmount >= r.Tiers[i].Threshold {
			applicableTier = r.Tiers[i]
			break
		}
	}

	if applicableTier == nil {
		return 0
	}

	var discount int64
	if applicableTier.DiscountType == "AMOUNT" {
		discount = applicableTier.DiscountValue
	} else { // PERCENT
		discount = orderAmount * applicableTier.DiscountValue / 10000 // 百分比*100
		if applicableTier.MaxDiscount > 0 && discount > applicableTier.MaxDiscount {
			discount = applicableTier.MaxDiscount
		}
	}

	return discount
}

// IsActive 检查规则是否生效
func (r *TieredDiscountRule) IsActive() bool {
	now := time.Now()
	return r.Status == RuleStatusActive && now.After(r.StartTime) && now.Before(r.EndTime)
}

// --- 买赠规则 ---

// BuyGetRule 买赠规则
type BuyGetRule struct {
	ID             uint64              `json:"id"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
	Name           string              `json:"name"`
	Description    string              `json:"description"`
	BuyCondition   *BuyCondition       `json:"buy_condition"` // 购买条件
	GiftItems      []*GiftItem         `json:"gift_items"`    // 赠品列表
	ApplyScope     *ApplyScope         `json:"apply_scope"`
	StartTime      time.Time           `json:"start_time"`
	EndTime        time.Time           `json:"end_time"`
	Status         PromotionRuleStatus `json:"status"`
	TotalGiftStock int32               `json:"total_gift_stock"` // 赠品总库存
	UsedGiftStock  int32               `json:"used_gift_stock"`  // 已送出赠品数
	PerUserLimit   int32               `json:"per_user_limit"`   // 每人限领
}

// BuyCondition 购买条件
type BuyCondition struct {
	Type      string `json:"type"`      // QUANTITY/AMOUNT 按数量还是金额
	Threshold int64  `json:"threshold"` // 阈值（件数或金额）
	SkuID     uint64 `json:"sku_id"`    // 指定SKU（0表示不限）
}

// GiftItem 赠品
type GiftItem struct {
	SkuID       uint64 `json:"sku_id"`
	ProductID   uint64 `json:"product_id"`
	ProductName string `json:"product_name"`
	Quantity    int32  `json:"quantity"`   // 赠送数量
	Stock       int32  `json:"stock"`      // 赠品库存
	UsedStock   int32  `json:"used_stock"` // 已送出库存
}

// NewBuyGetRule 创建买赠规则
func NewBuyGetRule(name, description string, startTime, endTime time.Time) *BuyGetRule {
	return &BuyGetRule{
		Name:        name,
		Description: description,
		GiftItems:   make([]*GiftItem, 0),
		StartTime:   startTime,
		EndTime:     endTime,
		Status:      RuleStatusDraft,
	}
}

// SetBuyCondition 设置购买条件
func (r *BuyGetRule) SetBuyCondition(condType string, threshold int64, skuID uint64) {
	r.BuyCondition = &BuyCondition{
		Type:      condType,
		Threshold: threshold,
		SkuID:     skuID,
	}
}

// AddGiftItem 添加赠品
func (r *BuyGetRule) AddGiftItem(skuID, productID uint64, productName string, quantity, stock int32) {
	gift := &GiftItem{
		SkuID:       skuID,
		ProductID:   productID,
		ProductName: productName,
		Quantity:    quantity,
		Stock:       stock,
	}
	r.GiftItems = append(r.GiftItems, gift)
	r.TotalGiftStock += stock
}

// CheckEligibility 检查是否满足赠送条件
func (r *BuyGetRule) CheckEligibility(items []*OrderItemForPromotion) bool {
	if !r.IsActive() {
		return false
	}
	if r.BuyCondition == nil {
		return false
	}

	var total int64
	for _, item := range items {
		if r.BuyCondition.SkuID == 0 || item.SkuID == r.BuyCondition.SkuID {
			if r.BuyCondition.Type == "QUANTITY" {
				total += int64(item.Quantity)
			} else {
				total += item.Amount
			}
		}
	}

	return total >= r.BuyCondition.Threshold
}

// GetAvailableGifts 获取可用赠品
func (r *BuyGetRule) GetAvailableGifts() []*GiftItem {
	var available []*GiftItem
	for _, gift := range r.GiftItems {
		if gift.Stock-gift.UsedStock > 0 {
			available = append(available, gift)
		}
	}
	return available
}

// IsActive 检查规则是否生效
func (r *BuyGetRule) IsActive() bool {
	now := time.Now()
	return r.Status == RuleStatusActive && now.After(r.StartTime) && now.Before(r.EndTime)
}

// --- 限购规则 ---

// PurchaseLimitRule 限购规则
type PurchaseLimitRule struct {
	ID          uint64              `json:"id"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	LimitType   string              `json:"limit_type"` // PER_ORDER/PER_DAY/PER_ACTIVITY/TOTAL
	LimitQty    int32               `json:"limit_qty"`  // 限购数量
	ApplyScope  *ApplyScope         `json:"apply_scope"`
	StartTime   time.Time           `json:"start_time"`
	EndTime     time.Time           `json:"end_time"`
	Status      PromotionRuleStatus `json:"status"`
}

// NewPurchaseLimitRule 创建限购规则
func NewPurchaseLimitRule(name, description, limitType string, limitQty int32, startTime, endTime time.Time) *PurchaseLimitRule {
	return &PurchaseLimitRule{
		Name:        name,
		Description: description,
		LimitType:   limitType,
		LimitQty:    limitQty,
		StartTime:   startTime,
		EndTime:     endTime,
		Status:      RuleStatusDraft,
	}
}

// CheckLimit 检查限购
func (r *PurchaseLimitRule) CheckLimit(ctx context.Context, userID uint64, skuID uint64, requestQty int32, purchaseHistory PurchaseHistoryService) error {
	if !r.IsActive() {
		return nil // 规则未生效，不限购
	}

	var purchasedQty int32
	var err error

	switch r.LimitType {
	case "PER_ORDER":
		// 单笔订单限购，直接检查请求数量
		if requestQty > r.LimitQty {
			return fmt.Errorf("exceeds per order limit: max=%d, requested=%d", r.LimitQty, requestQty)
		}
		return nil
	case "PER_DAY":
		purchasedQty, err = purchaseHistory.GetTodayPurchased(ctx, userID, skuID)
	case "PER_ACTIVITY":
		purchasedQty, err = purchaseHistory.GetActivityPurchased(ctx, userID, skuID, r.ID)
	case "TOTAL":
		purchasedQty, err = purchaseHistory.GetTotalPurchased(ctx, userID, skuID)
	}

	if err != nil {
		return err
	}

	if purchasedQty+requestQty > r.LimitQty {
		remaining := r.LimitQty - purchasedQty
		return fmt.Errorf("exceeds purchase limit: max=%d, purchased=%d, remaining=%d", r.LimitQty, purchasedQty, remaining)
	}

	return nil
}

// IsActive 检查规则是否生效
func (r *PurchaseLimitRule) IsActive() bool {
	now := time.Now()
	return r.Status == RuleStatusActive && now.After(r.StartTime) && now.Before(r.EndTime)
}

// PurchaseHistoryService 购买历史服务接口
type PurchaseHistoryService interface {
	GetTodayPurchased(ctx context.Context, userID, skuID uint64) (int32, error)
	GetActivityPurchased(ctx context.Context, userID, skuID, activityID uint64) (int32, error)
	GetTotalPurchased(ctx context.Context, userID, skuID uint64) (int32, error)
}

// --- 会员专属价 ---

// MemberPriceRule 会员价规则
type MemberPriceRule struct {
	ID          uint64              `json:"id"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Prices      []*MemberLevelPrice `json:"prices"` // 各等级价格
	ApplyScope  *ApplyScope         `json:"apply_scope"`
	StartTime   time.Time           `json:"start_time"`
	EndTime     time.Time           `json:"end_time"`
	Status      PromotionRuleStatus `json:"status"`
}

// MemberLevelPrice 会员等级价格
type MemberLevelPrice struct {
	MemberLevel  string `json:"member_level"`  // 会员等级
	PriceType    string `json:"price_type"`    // FIXED/DISCOUNT 固定价/折扣
	Price        int64  `json:"price"`         // 固定价（分）
	DiscountRate int    `json:"discount_rate"` // 折扣率（百分比）
}

// NewMemberPriceRule 创建会员价规则
func NewMemberPriceRule(name, description string, startTime, endTime time.Time) *MemberPriceRule {
	return &MemberPriceRule{
		Name:        name,
		Description: description,
		Prices:      make([]*MemberLevelPrice, 0),
		StartTime:   startTime,
		EndTime:     endTime,
		Status:      RuleStatusDraft,
	}
}

// AddMemberPrice 添加会员价格
func (r *MemberPriceRule) AddMemberPrice(level, priceType string, price int64, discountRate int) {
	r.Prices = append(r.Prices, &MemberLevelPrice{
		MemberLevel:  level,
		PriceType:    priceType,
		Price:        price,
		DiscountRate: discountRate,
	})
}

// GetMemberPrice 获取会员价
func (r *MemberPriceRule) GetMemberPrice(originalPrice int64, memberLevel string) int64 {
	if !r.IsActive() {
		return originalPrice
	}

	for _, levelPrice := range r.Prices {
		if levelPrice.MemberLevel == memberLevel {
			if levelPrice.PriceType == "FIXED" {
				return levelPrice.Price
			}
			// 折扣计算
			return originalPrice * int64(levelPrice.DiscountRate) / 100
		}
	}

	return originalPrice
}

// IsActive 检查规则是否生效
func (r *MemberPriceRule) IsActive() bool {
	now := time.Now()
	return r.Status == RuleStatusActive && now.After(r.StartTime) && now.Before(r.EndTime)
}

// --- 新人专享价 ---

// NewUserPriceRule 新人价规则
type NewUserPriceRule struct {
	ID           uint64              `json:"id"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
	Name         string              `json:"name"`
	Description  string              `json:"description"`
	PriceType    string              `json:"price_type"`    // FIXED/DISCOUNT
	Price        int64               `json:"price"`         // 固定价（分）
	DiscountRate int                 `json:"discount_rate"` // 折扣率（百分比）
	NewUserDays  int                 `json:"new_user_days"` // 新用户定义（注册N天内）
	MaxUseTimes  int32               `json:"max_use_times"` // 最多使用次数
	ApplyScope   *ApplyScope         `json:"apply_scope"`
	StartTime    time.Time           `json:"start_time"`
	EndTime      time.Time           `json:"end_time"`
	Status       PromotionRuleStatus `json:"status"`
}

// NewNewUserPriceRule 创建新人价规则
func NewNewUserPriceRule(name, description string, newUserDays int, maxUseTimes int32, startTime, endTime time.Time) *NewUserPriceRule {
	return &NewUserPriceRule{
		Name:        name,
		Description: description,
		NewUserDays: newUserDays,
		MaxUseTimes: maxUseTimes,
		StartTime:   startTime,
		EndTime:     endTime,
		Status:      RuleStatusDraft,
	}
}

// CheckNewUser 检查是否为新用户
func (r *NewUserPriceRule) CheckNewUser(userRegisterTime time.Time, usedTimes int32) bool {
	if !r.IsActive() {
		return false
	}

	// 检查注册时间
	daysSinceRegister := int(time.Since(userRegisterTime).Hours() / 24)
	if daysSinceRegister > r.NewUserDays {
		return false
	}

	// 检查使用次数
	if r.MaxUseTimes > 0 && usedTimes >= r.MaxUseTimes {
		return false
	}

	return true
}

// GetNewUserPrice 获取新人价
func (r *NewUserPriceRule) GetNewUserPrice(originalPrice int64) int64 {
	if r.PriceType == "FIXED" {
		return r.Price
	}
	return originalPrice * int64(r.DiscountRate) / 100
}

// IsActive 检查规则是否生效
func (r *NewUserPriceRule) IsActive() bool {
	now := time.Now()
	return r.Status == RuleStatusActive && now.After(r.StartTime) && now.Before(r.EndTime)
}

// --- 促销计算器 ---

// OrderItemForPromotion 订单商品（用于促销计算）
type OrderItemForPromotion struct {
	SkuID      uint64 `json:"sku_id"`
	ProductID  uint64 `json:"product_id"`
	CategoryID uint64 `json:"category_id"`
	MerchantID uint64 `json:"merchant_id"`
	Quantity   int32  `json:"quantity"`
	UnitPrice  int64  `json:"unit_price"` // 单价（分）
	Amount     int64  `json:"amount"`     // 金额（分）
}

// PromotionResult 促销计算结果
type PromotionResult struct {
	TotalDiscount int64          `json:"total_discount"` // 总优惠金额
	AppliedRules  []*AppliedRule `json:"applied_rules"`  // 应用的规则
	GiftItems     []*GiftItem    `json:"gift_items"`     // 赠品
	Errors        []string       `json:"errors"`         // 错误信息
}

// AppliedRule 应用的规则
type AppliedRule struct {
	RuleID      uint64            `json:"rule_id"`
	RuleType    PromotionRuleType `json:"rule_type"`
	RuleName    string            `json:"rule_name"`
	Discount    int64             `json:"discount"` // 优惠金额
	Description string            `json:"description"`
}

// PromotionCalculator 促销计算器
type PromotionCalculator struct {
	TieredRules    []*TieredDiscountRule
	BuyGetRules    []*BuyGetRule
	PurchaseLimits []*PurchaseLimitRule
	MemberPrices   []*MemberPriceRule
	NewUserPrices  []*NewUserPriceRule
}

// Calculate 计算促销
func (c *PromotionCalculator) Calculate(ctx context.Context, items []*OrderItemForPromotion, userInfo *UserPromotionInfo) (*PromotionResult, error) {
	result := &PromotionResult{
		AppliedRules: make([]*AppliedRule, 0),
		GiftItems:    make([]*GiftItem, 0),
		Errors:       make([]string, 0),
	}

	// 计算订单总金额
	var totalAmount int64
	for _, item := range items {
		totalAmount += item.Amount
	}

	// 1. 检查限购
	for _, rule := range c.PurchaseLimits {
		for _, item := range items {
			if err := rule.CheckLimit(ctx, userInfo.UserID, item.SkuID, item.Quantity, userInfo.PurchaseHistory); err != nil {
				result.Errors = append(result.Errors, err.Error())
			}
		}
	}

	// 2. 应用会员价
	if userInfo.MemberLevel != "" {
		for _, rule := range c.MemberPrices {
			for _, item := range items {
				memberPrice := rule.GetMemberPrice(item.UnitPrice, userInfo.MemberLevel)
				if memberPrice < item.UnitPrice {
					discount := (item.UnitPrice - memberPrice) * int64(item.Quantity)
					result.TotalDiscount += discount
					result.AppliedRules = append(result.AppliedRules, &AppliedRule{
						RuleID:      rule.ID,
						RuleType:    RuleMemberPrice,
						RuleName:    rule.Name,
						Discount:    discount,
						Description: fmt.Sprintf("会员价优惠 %d 分", discount),
					})
				}
			}
		}
	}

	// 3. 应用新人价
	for _, rule := range c.NewUserPrices {
		if rule.CheckNewUser(userInfo.RegisterTime, userInfo.NewUserPromoUsed) {
			for _, item := range items {
				newUserPrice := rule.GetNewUserPrice(item.UnitPrice)
				if newUserPrice < item.UnitPrice {
					discount := (item.UnitPrice - newUserPrice) * int64(item.Quantity)
					result.TotalDiscount += discount
					result.AppliedRules = append(result.AppliedRules, &AppliedRule{
						RuleID:      rule.ID,
						RuleType:    RuleNewUserPrice,
						RuleName:    rule.Name,
						Discount:    discount,
						Description: fmt.Sprintf("新人专享价优惠 %d 分", discount),
					})
				}
			}
		}
	}

	// 4. 应用阶梯满减（互斥，取最优）
	var bestTieredDiscount int64
	var bestTieredRule *TieredDiscountRule
	for _, rule := range c.TieredRules {
		discount := rule.CalculateDiscount(totalAmount - result.TotalDiscount)
		if discount > bestTieredDiscount {
			bestTieredDiscount = discount
			bestTieredRule = rule
		}
	}
	if bestTieredRule != nil && bestTieredDiscount > 0 {
		result.TotalDiscount += bestTieredDiscount
		result.AppliedRules = append(result.AppliedRules, &AppliedRule{
			RuleID:      bestTieredRule.ID,
			RuleType:    RuleTieredDiscount,
			RuleName:    bestTieredRule.Name,
			Discount:    bestTieredDiscount,
			Description: fmt.Sprintf("满减优惠 %d 分", bestTieredDiscount),
		})
	}

	// 5. 应用买赠
	for _, rule := range c.BuyGetRules {
		if rule.CheckEligibility(items) {
			gifts := rule.GetAvailableGifts()
			result.GiftItems = append(result.GiftItems, gifts...)
			result.AppliedRules = append(result.AppliedRules, &AppliedRule{
				RuleID:      rule.ID,
				RuleType:    RuleBuyGetFree,
				RuleName:    rule.Name,
				Description: fmt.Sprintf("赠送 %d 件赠品", len(gifts)),
			})
		}
	}

	return result, nil
}

// UserPromotionInfo 用户促销信息
type UserPromotionInfo struct {
	UserID           uint64                 `json:"user_id"`
	MemberLevel      string                 `json:"member_level"`
	RegisterTime     time.Time              `json:"register_time"`
	NewUserPromoUsed int32                  `json:"new_user_promo_used"`
	PurchaseHistory  PurchaseHistoryService `json:"-"`
}

// --- 促销规则仓储接口 ---

// PromotionRuleRepository 促销规则仓储接口
type PromotionRuleRepository interface {
	SaveTieredDiscount(ctx context.Context, rule *TieredDiscountRule) error
	SaveBuyGet(ctx context.Context, rule *BuyGetRule) error
	SavePurchaseLimit(ctx context.Context, rule *PurchaseLimitRule) error
	SaveMemberPrice(ctx context.Context, rule *MemberPriceRule) error
	SaveNewUserPrice(ctx context.Context, rule *NewUserPriceRule) error

	FindActiveTieredDiscounts(ctx context.Context) ([]*TieredDiscountRule, error)
	FindActiveBuyGetRules(ctx context.Context) ([]*BuyGetRule, error)
	FindActivePurchaseLimits(ctx context.Context) ([]*PurchaseLimitRule, error)
	FindActiveMemberPrices(ctx context.Context) ([]*MemberPriceRule, error)
	FindActiveNewUserPrices(ctx context.Context) ([]*NewUserPriceRule, error)

	FindByScope(ctx context.Context, ruleType PromotionRuleType, scopeType string, scopeIDs []uint64) ([]interface{}, error)
}
