package algorithm

import (
	"math"
	"time"
)

// PricingEngine 动态定价引擎
type PricingEngine struct {
	basePrice  int64   // 基础价格（分）
	minPrice   int64   // 最低价格
	maxPrice   int64   // 最高价格
	elasticity float64 // 需求价格弹性系数
}

// NewPricingEngine 创建定价引擎
func NewPricingEngine(basePrice, minPrice, maxPrice int64, elasticity float64) *PricingEngine {
	return &PricingEngine{
		basePrice:  basePrice,
		minPrice:   minPrice,
		maxPrice:   maxPrice,
		elasticity: elasticity,
	}
}

// PricingFactors 定价因素
type PricingFactors struct {
	Stock           int32   // 当前库存
	TotalStock      int32   // 总库存
	DemandLevel     float64 // 需求水平（0-1）
	CompetitorPrice int64   // 竞品价格
	TimeOfDay       int     // 时间（0-23）
	DayOfWeek       int     // 星期（0-6）
	IsHoliday       bool    // 是否节假日
	UserLevel       int     // 用户等级（1-10）
	SeasonFactor    float64 // 季节因素（0-1）
}

// CalculatePrice 计算动态价格
func (pe *PricingEngine) CalculatePrice(factors PricingFactors) int64 {
	price := float64(pe.basePrice)

	// 1. 库存因素（库存越少价格越高）
	stockRatio := float64(factors.Stock) / float64(factors.TotalStock)
	if stockRatio < 0.1 {
		price *= 1.2 // 库存不足10%，涨价20%
	} else if stockRatio < 0.3 {
		price *= 1.1 // 库存不足30%，涨价10%
	} else if stockRatio > 0.8 {
		price *= 0.9 // 库存超过80%，降价10%
	}

	// 2. 需求因素（需求越高价格越高）
	demandMultiplier := 1.0 + (factors.DemandLevel-0.5)*0.4 // 需求在0.5时价格不变
	price *= demandMultiplier

	// 3. 竞品价格因素
	if factors.CompetitorPrice > 0 {
		competitorRatio := float64(pe.basePrice) / float64(factors.CompetitorPrice)
		if competitorRatio > 1.1 {
			// 我们的价格比竞品高10%以上，适当降价
			price *= 0.95
		} else if competitorRatio < 0.9 {
			// 我们的价格比竞品低10%以上，可以适当涨价
			price *= 1.05
		}
	}

	// 4. 时间因素（高峰时段涨价）
	if factors.TimeOfDay >= 10 && factors.TimeOfDay <= 22 {
		price *= 1.05 // 白天涨价5%
	}

	// 5. 星期因素（周末涨价）
	if factors.DayOfWeek == 0 || factors.DayOfWeek == 6 {
		price *= 1.08 // 周末涨价8%
	}

	// 6. 节假日因素
	if factors.IsHoliday {
		price *= 1.15 // 节假日涨价15%
	}

	// 7. 用户等级因素（VIP用户享受折扣）
	if factors.UserLevel >= 8 {
		price *= 0.9 // VIP用户9折
	} else if factors.UserLevel >= 5 {
		price *= 0.95 // 高级用户95折
	}

	// 8. 季节因素
	price *= (1.0 + (factors.SeasonFactor-0.5)*0.2)

	// 限制价格范围
	finalPrice := int64(price)
	if finalPrice < pe.minPrice {
		finalPrice = pe.minPrice
	}
	if finalPrice > pe.maxPrice {
		finalPrice = pe.maxPrice
	}

	return finalPrice
}

// CalculateDemandElasticity 计算需求价格弹性
// 返回：价格变化1%时，需求量变化的百分比
func (pe *PricingEngine) CalculateDemandElasticity(currentPrice, currentDemand int64, newPrice, newDemand int64) float64 {
	priceChange := (float64(newPrice) - float64(currentPrice)) / float64(currentPrice)
	demandChange := (float64(newDemand) - float64(currentDemand)) / float64(currentDemand)

	if priceChange == 0 {
		return 0
	}

	return demandChange / priceChange
}

// OptimalPrice 计算最优价格（利润最大化）
// 假设：利润 = (价格 - 成本) * 销量
// 销量 = 基础销量 * (1 + 弹性系数 * 价格变化率)
func (pe *PricingEngine) OptimalPrice(cost int64, baseDemand int64) int64 {
	// 使用导数求最优价格
	// 利润函数：P(x) = (x - c) * d * (1 + e * (x - b) / b)
	// 其中：x=价格, c=成本, d=基础需求, e=弹性系数, b=基础价格

	// 求导并令导数为0
	// 最优价格 = (c * e * b - b * d) / (2 * e * b - d)

	numerator := float64(cost)*pe.elasticity*float64(pe.basePrice) - float64(pe.basePrice)*float64(baseDemand)
	denominator := 2*pe.elasticity*float64(pe.basePrice) - float64(baseDemand)

	if denominator == 0 {
		return pe.basePrice
	}

	optimalPrice := int64(numerator / denominator)

	// 限制价格范围
	if optimalPrice < pe.minPrice {
		optimalPrice = pe.minPrice
	}
	if optimalPrice > pe.maxPrice {
		optimalPrice = pe.maxPrice
	}

	return optimalPrice
}

// SurgePrice 高峰定价（类似Uber动态定价）
func (pe *PricingEngine) SurgePrice(demandSupplyRatio float64) int64 {
	// demandSupplyRatio: 需求/供给比率
	// 比率越高，价格越高

	var multiplier float64
	if demandSupplyRatio < 0.5 {
		multiplier = 0.8 // 供大于求，降价
	} else if demandSupplyRatio < 1.0 {
		multiplier = 1.0 // 供需平衡
	} else if demandSupplyRatio < 2.0 {
		multiplier = 1.0 + (demandSupplyRatio-1.0)*0.5 // 需求略高
	} else if demandSupplyRatio < 5.0 {
		multiplier = 1.5 + (demandSupplyRatio-2.0)*0.3 // 需求较高
	} else {
		multiplier = 2.4 // 需求极高，最多涨价140%
	}

	price := int64(float64(pe.basePrice) * multiplier)

	if price < pe.minPrice {
		price = pe.minPrice
	}
	if price > pe.maxPrice {
		price = pe.maxPrice
	}

	return price
}

// TimeBasedPrice 基于时间的定价（早鸟价、尾货价）
func (pe *PricingEngine) TimeBasedPrice(startTime, endTime, currentTime time.Time) int64 {
	totalDuration := endTime.Sub(startTime).Seconds()
	elapsed := currentTime.Sub(startTime).Seconds()

	if elapsed < 0 {
		elapsed = 0
	}
	if elapsed > totalDuration {
		elapsed = totalDuration
	}

	progress := elapsed / totalDuration

	var multiplier float64
	if progress < 0.2 {
		// 前20%时间，早鸟价，8折
		multiplier = 0.8
	} else if progress < 0.5 {
		// 20%-50%时间，正常价
		multiplier = 1.0
	} else if progress < 0.8 {
		// 50%-80%时间，略微涨价
		multiplier = 1.1
	} else {
		// 最后20%时间，尾货价，7折清仓
		multiplier = 0.7
	}

	price := int64(float64(pe.basePrice) * multiplier)

	if price < pe.minPrice {
		price = pe.minPrice
	}
	if price > pe.maxPrice {
		price = pe.maxPrice
	}

	return price
}

// PersonalizedPrice 个性化定价（基于用户画像）
func (pe *PricingEngine) PersonalizedPrice(userProfile UserProfile) int64 {
	price := float64(pe.basePrice)

	// 1. 购买力因素
	if userProfile.PurchasePower > 8 {
		price *= 1.1 // 高购买力用户，价格略高
	} else if userProfile.PurchasePower < 3 {
		price *= 0.9 // 低购买力用户，价格略低
	}

	// 2. 价格敏感度
	if userProfile.PriceSensitivity > 7 {
		price *= 0.95 // 价格敏感用户，给予折扣
	}

	// 3. 忠诚度
	if userProfile.Loyalty > 8 {
		price *= 0.92 // 高忠诚度用户，VIP折扣
	}

	// 4. 购买频率
	if userProfile.PurchaseFrequency > 10 {
		price *= 0.95 // 高频用户，给予折扣
	}

	// 5. 客单价
	if userProfile.AvgOrderValue > 50000 {
		price *= 0.93 // 高客单价用户，给予折扣
	}

	finalPrice := int64(price)
	if finalPrice < pe.minPrice {
		finalPrice = pe.minPrice
	}
	if finalPrice > pe.maxPrice {
		finalPrice = pe.maxPrice
	}

	return finalPrice
}

// UserProfile 用户画像
type UserProfile struct {
	PurchasePower     int   // 购买力（1-10）
	PriceSensitivity  int   // 价格敏感度（1-10）
	Loyalty           int   // 忠诚度（1-10）
	PurchaseFrequency int   // 购买频率（次/月）
	AvgOrderValue     int64 // 平均客单价（分）
}

// BundlePrice 捆绑定价
func (pe *PricingEngine) BundlePrice(items []int64, bundleDiscount float64) int64 {
	totalPrice := int64(0)
	for _, price := range items {
		totalPrice += price
	}

	// 应用捆绑折扣
	bundlePrice := int64(float64(totalPrice) * (1.0 - bundleDiscount))

	return bundlePrice
}

// CompetitivePricing 竞争定价策略
func (pe *PricingEngine) CompetitivePricing(competitorPrices []int64, strategy string) int64 {
	if len(competitorPrices) == 0 {
		return pe.basePrice
	}

	var price int64

	switch strategy {
	case "lowest":
		// 最低价策略
		price = competitorPrices[0]
		for _, p := range competitorPrices {
			if p < price {
				price = p
			}
		}
		price = int64(float64(price) * 0.95) // 比最低价再低5%

	case "average":
		// 平均价策略
		sum := int64(0)
		for _, p := range competitorPrices {
			sum += p
		}
		price = sum / int64(len(competitorPrices))

	case "premium":
		// 溢价策略
		sum := int64(0)
		for _, p := range competitorPrices {
			sum += p
		}
		avgPrice := sum / int64(len(competitorPrices))
		price = int64(float64(avgPrice) * 1.1) // 比平均价高10%

	default:
		price = pe.basePrice
	}

	if price < pe.minPrice {
		price = pe.minPrice
	}
	if price > pe.maxPrice {
		price = pe.maxPrice
	}

	return price
}

// PredictDemand 预测需求（简化版）
func (pe *PricingEngine) PredictDemand(price int64, historicalData []DemandData) int64 {
	if len(historicalData) == 0 {
		return 0
	}

	// 使用线性回归预测
	// y = a + b*x
	// 其中 y=需求量, x=价格

	n := float64(len(historicalData))
	var sumX, sumY, sumXY, sumX2 float64

	for _, data := range historicalData {
		x := float64(data.Price)
		y := float64(data.Demand)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// 计算回归系数
	b := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	a := (sumY - b*sumX) / n

	// 预测需求
	predictedDemand := a + b*float64(price)

	if predictedDemand < 0 {
		predictedDemand = 0
	}

	return int64(predictedDemand)
}

// DemandData 需求数据
type DemandData struct {
	Price  int64
	Demand int64
}

// CalculateRevenue 计算收入
func (pe *PricingEngine) CalculateRevenue(price, demand int64) int64 {
	return price * demand
}

// CalculateProfit 计算利润
func (pe *PricingEngine) CalculateProfit(price, cost, demand int64) int64 {
	return (price - cost) * demand
}

// OptimalPriceForProfit 计算利润最大化的最优价格
func (pe *PricingEngine) OptimalPriceForProfit(cost int64, demandFunc func(int64) int64) int64 {
	// 使用黄金分割法搜索最优价格
	goldenRatio := (math.Sqrt(5) - 1) / 2

	left := float64(pe.minPrice)
	right := float64(pe.maxPrice)

	for right-left > 1 {
		mid1 := left + (right-left)*(1-goldenRatio)
		mid2 := left + (right-left)*goldenRatio

		profit1 := pe.CalculateProfit(int64(mid1), cost, demandFunc(int64(mid1)))
		profit2 := pe.CalculateProfit(int64(mid2), cost, demandFunc(int64(mid2)))

		if profit1 > profit2 {
			right = mid2
		} else {
			left = mid1
		}
	}

	return int64((left + right) / 2)
}
