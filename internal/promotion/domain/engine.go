// 变更说明：
// 新增促销计算引擎，负责根据购物车商品列表匹配并计算所有适用的促销优惠。
// 核心算法：
// 1. 收集所有适用促销 → 2. 按优先级排序 → 3. 处理互斥/叠加 → 4. 逐一计算优惠 → 5. 汇总结果。
// 时间复杂度：O(P * I)，P 为促销数量，I 为商品数量。
package domain

import (
	"sort"

	"github.com/shopspring/decimal"
)

// PromotionEngine 促销计算引擎。
// 负责根据购物车商品列表，匹配并计算所有适用的促销优惠。
type PromotionEngine struct{}

// NewPromotionEngine 创建促销计算引擎实例。
func NewPromotionEngine() *PromotionEngine {
	return &PromotionEngine{}
}

// Calculate 执行促销计算。
// items: 购物车商品列表。
// promotions: 当前可用的促销活动列表。
// 返回促销计算结果，包含总优惠金额和命中的促销详情。
func (e *PromotionEngine) Calculate(items []*CartItem, promotions []*Promotion) *PromotionCalculation {
	result := &PromotionCalculation{
		OriginalAmount: decimal.Zero,
		TotalDiscount:  decimal.Zero,
		Promotions:     make([]*PromotionResult, 0),
	}

	// 1. 计算原始总金额。
	for _, item := range items {
		result.OriginalAmount = result.OriginalAmount.Add(item.TotalAmount())
	}

	if len(promotions) == 0 || len(items) == 0 {
		result.FinalAmount = result.OriginalAmount
		return result
	}

	// 2. 按优先级从高到低排序。
	sorted := make([]*Promotion, len(promotions))
	copy(sorted, promotions)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority > sorted[j].Priority
	})

	// 3. 记录已被互斥促销占用的商品 SKU。
	exclusiveUsed := make(map[uint64]bool)
	hasExclusive := false

	// 4. 逐一匹配并计算。
	for _, promo := range sorted {
		// 跳过已过期或无额度的促销。
		if !promo.HasUsageQuota() {
			continue
		}

		// 如果已有互斥促销命中，跳过后续互斥促销。
		if hasExclusive && promo.StackMode == StackModeExclusive {
			continue
		}

		// 筛选适用于该促销的商品。
		matchedItems := e.filterItems(items, promo, exclusiveUsed)
		if len(matchedItems) == 0 {
			continue
		}

		// 计算匹配商品的总金额和总数量。
		matchedAmount := decimal.Zero
		matchedQty := int32(0)
		appliedSKUs := make([]uint64, 0, len(matchedItems))
		for _, item := range matchedItems {
			matchedAmount = matchedAmount.Add(item.TotalAmount())
			matchedQty += item.Quantity
			appliedSKUs = append(appliedSKUs, item.SkuID)
		}

		// 计算优惠金额。
		discount := promo.CalculateDiscount(matchedAmount, matchedQty)
		if discount.IsZero() && promo.Type != PromotionTypeFreeShipping && promo.Type != PromotionTypeGift {
			continue
		}

		// 处理赠品。
		var giftSKU uint64
		var giftQty int32
		if promo.Type == PromotionTypeBuyNGetM || promo.Type == PromotionTypeGift {
			rule := promo.MatchRule(matchedAmount, matchedQty)
			if rule != nil {
				giftSKU = rule.GiftSKUID
				giftQty = rule.GiftQty
			}
		}

		// 处理包邮。
		if promo.Type == PromotionTypeFreeShipping {
			result.FreeShipping = true
		}

		// 记录结果。
		promoResult := &PromotionResult{
			PromotionID:   promo.ID,
			PromotionName: promo.Name,
			PromotionType: promo.Type,
			Label:         promo.Label,
			Discount:      discount,
			GiftSKUID:     giftSKU,
			GiftQty:       giftQty,
			AppliedItems:  appliedSKUs,
		}
		result.Promotions = append(result.Promotions, promoResult)
		result.TotalDiscount = result.TotalDiscount.Add(discount)

		// 如果是互斥促销，标记已占用的商品。
		if promo.StackMode == StackModeExclusive {
			hasExclusive = true
			for _, sku := range appliedSKUs {
				exclusiveUsed[sku] = true
			}
		}
	}

	// 5. 计算最终金额（不能为负数）。
	result.FinalAmount = result.OriginalAmount.Sub(result.TotalDiscount)
	if result.FinalAmount.IsNegative() {
		result.FinalAmount = decimal.Zero
	}

	return result
}

// SelectBest 最优促销选择（当 StackMode 为 BEST 时使用）。
// 从多个互斥促销中选择优惠最大的一个。
func (e *PromotionEngine) SelectBest(items []*CartItem, promotions []*Promotion) *PromotionResult {
	var bestResult *PromotionResult
	bestDiscount := decimal.Zero

	for _, promo := range promotions {
		matchedItems := e.filterItems(items, promo, nil)
		if len(matchedItems) == 0 {
			continue
		}

		matchedAmount := decimal.Zero
		matchedQty := int32(0)
		for _, item := range matchedItems {
			matchedAmount = matchedAmount.Add(item.TotalAmount())
			matchedQty += item.Quantity
		}

		discount := promo.CalculateDiscount(matchedAmount, matchedQty)
		if discount.GreaterThan(bestDiscount) {
			bestDiscount = discount
			appliedSKUs := make([]uint64, 0, len(matchedItems))
			for _, item := range matchedItems {
				appliedSKUs = append(appliedSKUs, item.SkuID)
			}
			bestResult = &PromotionResult{
				PromotionID:   promo.ID,
				PromotionName: promo.Name,
				PromotionType: promo.Type,
				Label:         promo.Label,
				Discount:      discount,
				AppliedItems:  appliedSKUs,
			}
		}
	}

	return bestResult
}

// filterItems 筛选适用于指定促销的商品列表。
func (e *PromotionEngine) filterItems(items []*CartItem, promo *Promotion, exclusiveUsed map[uint64]bool) []*CartItem {
	matched := make([]*CartItem, 0)

	// 构建排除集合。
	excludeSet := make(map[uint64]bool, len(promo.ExcludeIDs))
	for _, id := range promo.ExcludeIDs {
		excludeSet[id] = true
	}

	// 构建适用范围集合。
	scopeSet := make(map[uint64]bool, len(promo.ScopeIDs))
	for _, id := range promo.ScopeIDs {
		scopeSet[id] = true
	}

	for _, item := range items {
		// 跳过已被互斥促销占用的商品。
		if exclusiveUsed != nil && exclusiveUsed[item.SkuID] {
			continue
		}

		// 检查是否在排除列表中。
		if excludeSet[item.ProductID] || excludeSet[item.SkuID] {
			continue
		}

		// 根据适用范围匹配。
		switch promo.Scope {
		case PromotionScopeAll:
			matched = append(matched, item)
		case PromotionScopeProduct:
			if scopeSet[item.ProductID] {
				matched = append(matched, item)
			}
		case PromotionScopeSKU:
			if scopeSet[item.SkuID] {
				matched = append(matched, item)
			}
		case PromotionScopeCategory:
			if scopeSet[item.CategoryID] {
				matched = append(matched, item)
			}
		case PromotionScopeBrand:
			if scopeSet[item.BrandID] {
				matched = append(matched, item)
			}
		case PromotionScopeMerchant:
			if scopeSet[item.MerchantID] {
				matched = append(matched, item)
			}
		}
	}

	return matched
}
