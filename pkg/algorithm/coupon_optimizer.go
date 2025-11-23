package algorithm

import (
	"sort"
)

// CouponType 优惠券类型
type CouponType int

const (
	CouponTypeDiscount  CouponType = iota + 1 // 折扣券
	CouponTypeReduction                       // 满减券
	CouponTypeCash                            // 立减券
)

// Coupon 优惠券
type Coupon struct {
	ID              uint64
	Type            CouponType
	Threshold       int64   // 门槛金额（分）
	DiscountRate    float64 // 折扣率（0.8表示8折）
	ReductionAmount int64   // 减免金额（分）
	MaxDiscount     int64   // 最大优惠金额（分）
	CanStack        bool    // 是否可叠加
	Priority        int     // 优先级
}

// CouponOptimizer 优惠券优化器
type CouponOptimizer struct{}

// NewCouponOptimizer 创建优惠券优化器
func NewCouponOptimizer() *CouponOptimizer {
	return &CouponOptimizer{}
}

// OptimalCombination 计算最优优惠组合
// 返回：最优优惠券组合、最终价格、优惠金额
func (co *CouponOptimizer) OptimalCombination(
	originalPrice int64,
	coupons []Coupon,
) ([]uint64, int64, int64) {

	if len(coupons) == 0 {
		return nil, originalPrice, 0
	}

	// 过滤可用优惠券
	available := make([]Coupon, 0)
	for _, c := range coupons {
		if originalPrice >= c.Threshold {
			available = append(available, c)
		}
	}

	if len(available) == 0 {
		return nil, originalPrice, 0
	}

	// 按优先级排序
	sort.Slice(available, func(i, j int) bool {
		return available[i].Priority > available[j].Priority
	})

	// 尝试所有可能的组合（使用位运算枚举）
	n := len(available)
	bestCombination := make([]uint64, 0)
	bestPrice := originalPrice
	maxDiscount := int64(0)

	// 枚举所有子集
	for mask := 1; mask < (1 << n); mask++ {
		combination := make([]Coupon, 0)

		for i := 0; i < n; i++ {
			if mask&(1<<i) != 0 {
				combination = append(combination, available[i])
			}
		}

		// 检查组合是否合法（是否都可叠加）
		if !co.isValidCombination(combination) {
			continue
		}

		// 计算该组合的最终价格
		finalPrice := co.calculatePrice(originalPrice, combination)
		discount := originalPrice - finalPrice

		if finalPrice < bestPrice {
			bestPrice = finalPrice
			maxDiscount = discount
			bestCombination = make([]uint64, len(combination))
			for i, c := range combination {
				bestCombination[i] = c.ID
			}
		}
	}

	return bestCombination, bestPrice, maxDiscount
}

// GreedyOptimization 贪心算法（快速但不一定最优）
func (co *CouponOptimizer) GreedyOptimization(
	originalPrice int64,
	coupons []Coupon,
) ([]uint64, int64, int64) {

	// 过滤可用优惠券
	available := make([]Coupon, 0)
	for _, c := range coupons {
		if originalPrice >= c.Threshold {
			available = append(available, c)
		}
	}

	if len(available) == 0 {
		return nil, originalPrice, 0
	}

	// 计算每个优惠券的优惠金额，按优惠金额降序排序
	type couponDiscount struct {
		coupon   Coupon
		discount int64
	}

	discounts := make([]couponDiscount, 0)
	for _, c := range available {
		discount := co.calculateSingleDiscount(originalPrice, c)
		discounts = append(discounts, couponDiscount{c, discount})
	}

	sort.Slice(discounts, func(i, j int) bool {
		return discounts[i].discount > discounts[j].discount
	})

	// 贪心选择
	selected := make([]Coupon, 0)
	currentPrice := originalPrice

	for _, cd := range discounts {
		// 尝试添加这个优惠券
		testCombination := append(selected, cd.coupon)

		if co.isValidCombination(testCombination) {
			newPrice := co.calculatePrice(originalPrice, testCombination)
			if newPrice < currentPrice {
				selected = testCombination
				currentPrice = newPrice
			}
		}
	}

	result := make([]uint64, len(selected))
	for i, c := range selected {
		result[i] = c.ID
	}

	discount := originalPrice - currentPrice
	return result, currentPrice, discount
}

// isValidCombination 检查优惠券组合是否合法
func (co *CouponOptimizer) isValidCombination(coupons []Coupon) bool {
	if len(coupons) == 0 {
		return false
	}

	if len(coupons) == 1 {
		return true
	}

	// 检查是否都可叠加
	for _, c := range coupons {
		if !c.CanStack {
			return false
		}
	}

	return true
}

// calculatePrice 计算使用优惠券后的价格
func (co *CouponOptimizer) calculatePrice(originalPrice int64, coupons []Coupon) int64 {
	if len(coupons) == 0 {
		return originalPrice
	}

	// 按优先级排序（折扣券 -> 满减券 -> 立减券）
	sorted := make([]Coupon, len(coupons))
	copy(sorted, coupons)

	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Type != sorted[j].Type {
			return sorted[i].Type < sorted[j].Type
		}
		return sorted[i].Priority > sorted[j].Priority
	})

	currentPrice := originalPrice

	for _, c := range sorted {
		switch c.Type {
		case CouponTypeDiscount:
			// 折扣券
			discount := int64(float64(currentPrice) * (1 - c.DiscountRate))
			if c.MaxDiscount > 0 && discount > c.MaxDiscount {
				discount = c.MaxDiscount
			}
			currentPrice -= discount

		case CouponTypeReduction:
			// 满减券
			if currentPrice >= c.Threshold {
				currentPrice -= c.ReductionAmount
			}

		case CouponTypeCash:
			// 立减券
			currentPrice -= c.ReductionAmount
		}

		if currentPrice < 0 {
			currentPrice = 0
		}
	}

	return currentPrice
}

// calculateSingleDiscount 计算单个优惠券的优惠金额
func (co *CouponOptimizer) calculateSingleDiscount(originalPrice int64, coupon Coupon) int64 {
	switch coupon.Type {
	case CouponTypeDiscount:
		discount := int64(float64(originalPrice) * (1 - coupon.DiscountRate))
		if coupon.MaxDiscount > 0 && discount > coupon.MaxDiscount {
			discount = coupon.MaxDiscount
		}
		return discount

	case CouponTypeReduction:
		if originalPrice >= coupon.Threshold {
			return coupon.ReductionAmount
		}
		return 0

	case CouponTypeCash:
		return coupon.ReductionAmount

	default:
		return 0
	}
}

// DynamicProgramming 动态规划求解（适合优惠券数量较多的情况）
func (co *CouponOptimizer) DynamicProgramming(
	originalPrice int64,
	coupons []Coupon,
) ([]uint64, int64, int64) {

	// 过滤可用优惠券
	available := make([]Coupon, 0)
	for _, c := range coupons {
		if originalPrice >= c.Threshold {
			available = append(available, c)
		}
	}

	if len(available) == 0 {
		return nil, originalPrice, 0
	}

	n := len(available)

	// dp[i][j] 表示使用前i个优惠券，是否选择第i个（j=0不选，j=1选）的最低价格
	dp := make([][]int64, n+1)
	for i := range dp {
		dp[i] = make([]int64, 2)
		dp[i][0] = originalPrice
		dp[i][1] = originalPrice
	}

	// 记录选择路径
	choice := make([][]bool, n+1)
	for i := range choice {
		choice[i] = make([]bool, 2)
	}

	for i := 1; i <= n; i++ {
		coupon := available[i-1]

		// 不选第i个优惠券
		dp[i][0] = dp[i-1][0]

		// 选第i个优惠券
		if coupon.CanStack {
			// 可叠加，可以在之前的基础上继续优惠
			testPrice := co.calculatePrice(dp[i-1][1], []Coupon{coupon})
			if testPrice < dp[i-1][1] {
				dp[i][1] = testPrice
				choice[i][1] = true
			} else {
				dp[i][1] = dp[i-1][1]
			}
		} else {
			// 不可叠加，单独使用
			dp[i][1] = co.calculatePrice(originalPrice, []Coupon{coupon})
			choice[i][1] = true
		}

		// 更新不选的情况
		if dp[i][1] < dp[i][0] {
			dp[i][0] = dp[i][1]
		}
	}

	// 回溯找出选择的优惠券
	selected := make([]uint64, 0)
	minPrice := dp[n][0]
	if dp[n][1] < minPrice {
		minPrice = dp[n][1]
	}

	for i := n; i > 0; i-- {
		if choice[i][1] && dp[i][1] == minPrice {
			selected = append(selected, available[i-1].ID)
			minPrice = dp[i-1][1]
		}
	}

	discount := originalPrice - minPrice
	return selected, minPrice, discount
}
