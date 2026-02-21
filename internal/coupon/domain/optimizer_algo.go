// 变更说明：
// 从 pkg/algos/optimization/coupon_optimizer.go 迁移。
// 实现了优惠券最优组合算法（暴力方案与贪心方案）。
package domain

import (
	"slices"
)

type CouponType int

const (
	CouponTypeDiscount CouponType = iota + 1
	CouponTypeReduction
	CouponTypeCash
)

type Coupon struct {
	ID              uint64
	DiscountRate    float64
	Threshold       int64
	ReductionAmount int64
	MaxDiscount     int64
	Type            CouponType
	Priority        int
	CanStack        bool
}

type CouponOptimizer struct{}

func NewCouponOptimizer() *CouponOptimizer {
	return &CouponOptimizer{}
}

func (co *CouponOptimizer) OptimalCombination(
	originalPrice int64,
	coupons []Coupon,
) (bestCombinations []uint64, finalPrice, totalDiscount int64) {
	if len(coupons) == 0 {
		return nil, originalPrice, 0
	}
	if len(coupons) > 20 {
		return co.GreedyOptimization(originalPrice, coupons)
	}

	available := make([]Coupon, 0)
	for _, c := range coupons {
		if originalPrice >= c.Threshold {
			available = append(available, c)
		}
	}

	if len(available) == 0 {
		return nil, originalPrice, 0
	}

	slices.SortFunc(available, func(a, b Coupon) int {
		if a.Type != b.Type {
			if a.Type < b.Type {
				return -1
			}
			return 1
		}
		if a.Priority > b.Priority {
			return -1
		}
		return 1
	})

	bestCombination := make([]uint64, 0)
	bestPrice := originalPrice
	maxDiscount := int64(0)

	n := len(available)
	for mask := 1; mask < (1 << n); mask++ {
		combination := make([]Coupon, 0, n)
		for i := 0; i < n; i++ {
			if mask&(1<<i) != 0 {
				combination = append(combination, available[i])
			}
		}

		if !co.isValidCombination(combination) {
			continue
		}

		fPrice := co.calculatePriceFast(originalPrice, combination)
		discount := originalPrice - fPrice

		if fPrice < bestPrice {
			bestPrice = fPrice
			maxDiscount = discount
			bestCombination = make([]uint64, len(combination))
			for i, c := range combination {
				bestCombination[i] = c.ID
			}
		}
	}

	return bestCombination, bestPrice, maxDiscount
}

func (co *CouponOptimizer) calculatePriceFast(originalPrice int64, sortedCoupons []Coupon) int64 {
	currentPrice := originalPrice
	for _, c := range sortedCoupons {
		switch c.Type {
		case CouponTypeDiscount:
			discount := int64(float64(currentPrice) * (1 - c.DiscountRate))
			if c.MaxDiscount > 0 && discount > c.MaxDiscount {
				discount = c.MaxDiscount
			}
			currentPrice -= discount
		case CouponTypeReduction:
			if currentPrice >= c.Threshold {
				currentPrice -= c.ReductionAmount
			}
		case CouponTypeCash:
			currentPrice -= c.ReductionAmount
		}
		if currentPrice < 0 {
			currentPrice = 0
		}
	}
	return currentPrice
}

func (co *CouponOptimizer) GreedyOptimization(
	originalPrice int64,
	coupons []Coupon,
) (ids []uint64, finalPrice, totalDiscount int64) {
	available := make([]Coupon, 0)
	for _, c := range coupons {
		if originalPrice >= c.Threshold {
			available = append(available, c)
		}
	}
	if len(available) == 0 {
		return nil, originalPrice, 0
	}

	type cdItem struct {
		coupon   Coupon
		discount int64
	}
	discounts := make([]cdItem, 0)
	for _, c := range available {
		discounts = append(discounts, cdItem{c, co.calculateSingleDiscount(originalPrice, c)})
	}
	slices.SortFunc(discounts, func(a, b cdItem) int {
		if a.discount > b.discount {
			return -1
		}
		return 1
	})

	selected := make([]Coupon, 0)
	currentPrice := originalPrice
	for _, cd := range discounts {
		test := append(slices.Clone(selected), cd.coupon)
		if co.isValidCombination(test) {
			newPrice := co.calculatePrice(originalPrice, test)
			if newPrice < currentPrice {
				selected = test
				currentPrice = newPrice
			}
		}
	}

	resIDs := make([]uint64, len(selected))
	for i, c := range selected {
		resIDs[i] = c.ID
	}
	return resIDs, currentPrice, originalPrice - currentPrice
}

func (co *CouponOptimizer) isValidCombination(coupons []Coupon) bool {
	if len(coupons) <= 1 {
		return len(coupons) == 1
	}
	for _, c := range coupons {
		if !c.CanStack {
			return false
		}
	}
	return true
}

func (co *CouponOptimizer) calculateSingleDiscount(originalPrice int64, coupon Coupon) int64 {
	switch coupon.Type {
	case CouponTypeDiscount:
		d := int64(float64(originalPrice) * (1 - coupon.DiscountRate))
		if coupon.MaxDiscount > 0 && d > coupon.MaxDiscount {
			d = coupon.MaxDiscount
		}
		return d
	case CouponTypeReduction:
		if originalPrice >= coupon.Threshold {
			return coupon.ReductionAmount
		}
		return 0
	case CouponTypeCash:
		return coupon.ReductionAmount
	}
	return 0
}

func (co *CouponOptimizer) calculatePrice(originalPrice int64, coupons []Coupon) int64 {
	sorted := slices.Clone(coupons)
	slices.SortFunc(sorted, func(a, b Coupon) int {
		if a.Type != b.Type {
			if a.Type < b.Type {
				return -1
			}
			return 1
		}
		if a.Priority > b.Priority {
			return -1
		}
		return 1
	})
	return co.calculatePriceFast(originalPrice, sorted)
}
