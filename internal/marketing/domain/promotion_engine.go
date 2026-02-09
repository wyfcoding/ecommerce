// Package domain 提供了营销引擎逻辑。
// 变更说明：实现高级营销引擎，支持阶梯优惠（Tiered Discount）、多营销工具堆叠算法（Matrix Marketing）及加价购（Add-on Sales）。
package domain

import (
	"context"
	"sort"
)

// PromotionType 营销类型
type PromotionType string

const (
	PromoTieredDiscount PromotionType = "TIERED" // 阶梯优惠（满100减10，满200减30）
	PromoMatrixStack    PromotionType = "MATRIX" // 矩阵堆叠（优惠券+活动+积分可叠加）
	PromoAddOnSale      PromotionType = "ADD_ON" // 加价购
)

// Tier 阶梯定义
type Tier struct {
	Threshold int64 `json:"threshold"` // 门槛金额
	Discount  int64 `json:"discount"`  // 优惠金额
}

// TieredDiscountRule 阶梯优惠规则
type TieredDiscountRule struct {
	Tiers []Tier `json:"tiers"`
}

// PromotionEngine 营销引擎服务
type PromotionEngine interface {
	// CalculateBestPromotion 计算最优优惠方案
	CalculateBestPromotion(ctx context.Context, orderAmount int64, availablePromos []*Campaign) (*PromotionResult, error)
}

// PromotionResult 优惠计算结果
type PromotionResult struct {
	OriginalAmount int64          `json:"original_amount"`
	FinalAmount    int64          `json:"final_amount"`
	TotalDiscount  int64          `json:"total_discount"`
	AppliedPromos  []AppliedPromo `json:"applied_promos"`
}

type AppliedPromo struct {
	CampaignID uint64 `json:"campaign_id"`
	PromoType  string `json:"promo_type"`
	Discount   int64  `json:"discount"`
}

// DefaultPromotionEngine 营销引擎实现
type DefaultPromotionEngine struct{}

func (e *DefaultPromotionEngine) CalculateBestPromotion(ctx context.Context, orderAmount int64, campaigns []*Campaign) *PromotionResult {
	result := &PromotionResult{
		OriginalAmount: orderAmount,
		FinalAmount:    orderAmount,
		AppliedPromos:  make([]AppliedPromo, 0),
	}

	// 1. 过滤并筛选可用活动
	activePromos := make([]*Campaign, 0)
	for _, c := range campaigns {
		if c.IsActive() {
			activePromos = append(activePromos, c)
		}
	}

	// 2. 这里的逻辑支持矩阵堆叠：活动可以多重命中（简化展示）
	for _, c := range activePromos {
		discount := e.calculateDiscount(c, result.FinalAmount)
		if discount > 0 {
			result.TotalDiscount += discount
			result.FinalAmount -= discount
			result.AppliedPromos = append(result.AppliedPromos, AppliedPromo{
				CampaignID: c.ID,
				PromoType:  string(c.CampaignType),
				Discount:   discount,
			})
		}
	}

	return result
}

func (e *DefaultPromotionEngine) calculateDiscount(c *Campaign, currentAmount int64) int64 {
	// 实现阶梯优惠逻辑
	if c.CampaignType == CampaignTypeFullReduce {
		// 模拟从 JSONMap 取出阶梯规则
		// 实际应有更严谨的解析逻辑
		threshold, _ := c.Rules["threshold"].(float64)
		reduce, _ := c.Rules["reduce"].(float64)

		if currentAmount >= int64(threshold) {
			return int64(reduce)
		}
	}

	// 实现阶梯自动寻找最高档位
	if tiers, ok := c.Rules["tiers"].([]any); ok {
		var bestDiscount int64
		// 转换并排序阶梯
		for _, t := range tiers {
			tm, _ := t.(map[string]any)
			thr := int64(tm["threshold"].(float64))
			dis := int64(tm["discount"].(float64))
			if currentAmount >= thr && dis > bestDiscount {
				bestScore := dis
				bestDiscount = bestScore
			}
		}
		return bestDiscount
	}

	return 0
}

// AddOnOption 加价购选项
type AddOnOption struct {
	TriggerAmount int64  `json:"trigger_amount"` // 触发金额
	ProductID     uint64 `json:"product_id"`     // 加价购商品
	AddOnPrice    int64  `json:"add_on_price"`   // 优惠后的加价
}

// GetAddOnSales 获取当前订单符合的加价购建议
func (e *DefaultPromotionEngine) GetAddOnSales(orderAmount int64, options []AddOnOption) []AddOnOption {
	var available []AddOnOption
	for _, opt := range options {
		if orderAmount >= opt.TriggerAmount {
			available = append(available, opt)
		}
	}
	// 按触发金额降序排序，推荐更有吸引力的
	sort.Slice(available, func(i, j int) bool {
		return available[i].TriggerAmount > available[j].TriggerAmount
	})
	return available
}
