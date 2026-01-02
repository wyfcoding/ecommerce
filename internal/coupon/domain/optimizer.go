package domain

import (
	"sort"

	"github.com/shopspring/decimal"
)

// OptimizationResult 最终方案
type OptimizationResult struct {
	BestCoupons     []*Coupon
	TotalDiscount   decimal.Decimal
	FinalPrice      decimal.Decimal
	CalculationPath string // 用于展示计算过程
}

// CartItem 购物车项
type CartItem struct {
	SKUID      uint64
	CategoryID uint64
	Price      decimal.Decimal
	Quantity   int32
}

// CouponOptimizer 优惠券组合优化器
type CouponOptimizer struct {
	// 深度限制，防止计算爆炸
	MaxRecursionDepth int
}

func NewCouponOptimizer() *CouponOptimizer {
	return &CouponOptimizer{MaxRecursionDepth: 10}
}

// FindBestCombination 使用回溯算法 + 剪枝 寻找最优优惠方案
func (o *CouponOptimizer) FindBestCombination(items []*CartItem, availableCoupons []*Coupon) *OptimizationResult {
	if len(availableCoupons) == 0 {
		return &OptimizationResult{FinalPrice: calculateTotalPrice(items)}
	}

	// 1. 预处理：按优惠券编号或默认顺序排序
	// 注意：在已有定义中没有 Priority 字段，我们可以根据优惠类型进行排序，优先处理折扣高的
	sort.Slice(availableCoupons, func(i, j int) bool {
		return availableCoupons[i].DiscountAmount > availableCoupons[j].DiscountAmount
	})

	var bestCoupons []*Coupon
	maxDiscount := decimal.Zero
	initialPrice := calculateTotalPrice(items)

	// 2. 递归回溯搜索
	var backtrack func(startIndex int, currentPrice decimal.Decimal, usedCoupons []*Coupon)
	backtrack = func(startIndex int, currentPrice decimal.Decimal, usedCoupons []*Coupon) {
		// 更新全局最优解
		discount := initialPrice.Sub(currentPrice)
		if discount.GreaterThan(maxDiscount) {
			maxDiscount = discount
			bestCoupons = make([]*Coupon, len(usedCoupons))
			copy(bestCoupons, usedCoupons)
		}

		if startIndex >= len(availableCoupons) || len(usedCoupons) >= o.MaxRecursionDepth {
			return
		}

		for i := startIndex; i < len(availableCoupons); i++ {
			coupon := availableCoupons[i]

			// 剪枝逻辑 1：叠加检查 (基于已有定义的 CanStack 字段)
			if !coupon.CanStack && len(usedCoupons) > 0 {
				continue
			}
			for _, used := range usedCoupons {
				if !used.CanStack {
					goto nextCoupon
				}
			}

			// 检查门槛 (MinOrderAmount)
			if o.canApply(coupon, items, currentPrice) {
				// 应用优惠
				newPrice := o.applyDiscount(coupon, items, currentPrice)

				// 剪枝逻辑 2：如果优惠后价格没变，跳过
				if newPrice.Equal(currentPrice) {
					goto nextCoupon
				}

				// 进入下一层
				backtrack(i+1, newPrice, append(usedCoupons, coupon))
			}

		nextCoupon:
		}
	}

	backtrack(0, initialPrice, []*Coupon{})

	return &OptimizationResult{
		BestCoupons:   bestCoupons,
		TotalDiscount: maxDiscount,
		FinalPrice:    initialPrice.Sub(maxDiscount),
	}
}

// canApply 检查优惠券是否适用于当前购物车
func (o *CouponOptimizer) canApply(c *Coupon, items []*CartItem, currentPrice decimal.Decimal) bool {
	threshold := decimal.NewFromInt(c.MinOrderAmount)

	// 1. 检查范围金额是否达标
	var scopeAmount decimal.Decimal
	if c.ApplicableScope == "全场通用" || len(c.ApplicableIDs) == 0 {
		scopeAmount = currentPrice
	} else {
		// 计算特定品类的金额
		for _, item := range items {
			if contains(c.ApplicableIDs, item.CategoryID) || contains(c.ApplicableIDs, item.SKUID) {
				scopeAmount = scopeAmount.Add(item.Price.Mul(decimal.NewFromInt32(item.Quantity)))
			}
		}
	}

	return scopeAmount.GreaterThanOrEqual(threshold)
}

// applyDiscount 计算应用优惠券后的价格
func (o *CouponOptimizer) applyDiscount(c *Coupon, items []*CartItem, currentPrice decimal.Decimal) decimal.Decimal {
	val := decimal.NewFromInt(c.DiscountAmount)

	// 1. 计算该券覆盖的商品总额基数
	var applicableBase decimal.Decimal
	if c.ApplicableScope == "全场通用" || len(c.ApplicableIDs) == 0 {
		// 全场券：基数即为当前价格（或者初始价格，取决于业务规则，这里假设按原价基数折算）
		applicableBase = calculateTotalPrice(items)
	} else {
		// 局部券：只计算范围内商品的金额
		for _, item := range items {
			if contains(c.ApplicableIDs, item.CategoryID) || contains(c.ApplicableIDs, item.SKUID) {
				applicableBase = applicableBase.Add(item.Price.Mul(decimal.NewFromInt32(item.Quantity)))
			}
		}
	}

	// 2. 根据类型计算扣减额
	var discountAmount decimal.Decimal
	switch c.Type {
	case CouponTypeDiscount:
		// 如果是折扣券，DiscountAmount 为 80 代表 8折，即扣除 20%
		rate := val.Div(decimal.NewFromInt(100))
		discountAmount = applicableBase.Mul(decimal.NewFromInt(1).Sub(rate))
	case CouponTypeCash:
		// 现金券直接扣减额
		discountAmount = val
	}

	// 3. 返回扣减后的价格，确保不为负数
	newPrice := currentPrice.Sub(discountAmount)
	if newPrice.IsNegative() {
		return decimal.Zero
	}
	return newPrice
}

func calculateTotalPrice(items []*CartItem) decimal.Decimal {
	total := decimal.Zero
	for _, item := range items {
		total = total.Add(item.Price.Mul(decimal.NewFromInt32(item.Quantity)))
	}
	return total
}

func contains(slice []uint64, val uint64) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}
