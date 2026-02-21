// 变更说明：
// 从 pkg/algos/finance/pricing_engine.go 迁移。
// 实现了动态定价、需求预测、利润最优化、个性化定价等功能。
package domain

import (
	"math"
)

type PricingEngine struct {
	basePrice  int64
	minPrice   int64
	maxPrice   int64
	elasticity float64
}

func NewPricingEngine(basePrice, minPrice, maxPrice int64, elasticity float64) *PricingEngine {
	return &PricingEngine{
		basePrice:  basePrice,
		minPrice:   minPrice,
		maxPrice:   maxPrice,
		elasticity: elasticity,
	}
}

type PricingFactors struct {
	DemandLevel     float64
	SeasonFactor    float64
	CompetitorPrice int64
	Stock           int32
	TotalStock      int32
	TimeOfDay       int
	DayOfWeek       int
	UserLevel       int
	IsHoliday       bool
}

type PricingResult struct {
	InventoryFactor  float64
	DemandFactor     float64
	CompetitorFactor float64
	TimeFactor       float64
	SeasonFactor     float64
	UserFactor       float64
	FinalPrice       int64
	BasePrice        int64
}

func (pe *PricingEngine) CalculatePrice(factors PricingFactors) PricingResult {
	price := float64(pe.basePrice)
	result := PricingResult{
		BasePrice:        pe.basePrice,
		InventoryFactor:  1.0,
		DemandFactor:     1.0,
		CompetitorFactor: 1.0,
		TimeFactor:       1.0,
		SeasonFactor:     1.0,
		UserFactor:       1.0,
	}

	stockRatio := 0.0
	if factors.TotalStock > 0 {
		stockRatio = float64(factors.Stock) / float64(factors.TotalStock)
	}

	switch {
	case stockRatio < 0.1:
		result.InventoryFactor = 1.2
	case stockRatio < 0.3:
		result.InventoryFactor = 1.1
	case stockRatio > 0.8:
		result.InventoryFactor = 0.9
	}
	price *= result.InventoryFactor

	result.DemandFactor = 1.0 + (factors.DemandLevel-0.5)*0.4
	price *= result.DemandFactor

	if factors.CompetitorPrice > 0 {
		competitorRatio := float64(pe.basePrice) / float64(factors.CompetitorPrice)
		if competitorRatio > 1.1 {
			result.CompetitorFactor = 0.95
		} else if competitorRatio < 0.9 {
			result.CompetitorFactor = 1.05
		}
	}
	price *= result.CompetitorFactor

	if factors.TimeOfDay >= 10 && factors.TimeOfDay <= 22 {
		result.TimeFactor *= 1.05
	}
	if factors.DayOfWeek == 0 || factors.DayOfWeek == 6 {
		result.TimeFactor *= 1.08
	}
	if factors.IsHoliday {
		result.TimeFactor *= 1.15
	}
	price *= result.TimeFactor

	if factors.UserLevel >= 8 {
		result.UserFactor = 0.9
	} else if factors.UserLevel >= 5 {
		result.UserFactor = 0.95
	}
	price *= result.UserFactor

	result.SeasonFactor = 1.0 + (factors.SeasonFactor-0.5)*0.2
	price *= result.SeasonFactor

	result.FinalPrice = min(max(int64(price), pe.minPrice), pe.maxPrice)
	return result
}

func (pe *PricingEngine) PredictDemand(price int64, historicalData []DemandData) int64 {
	if len(historicalData) < 2 {
		return 0
	}
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
	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		return int64(sumY / n)
	}
	b := (n*sumXY - sumX*sumY) / denominator
	a := (sumY - b*sumX) / n
	predictedDemand := a + b*float64(price)
	if predictedDemand < 0 {
		predictedDemand = 0
	}
	return int64(predictedDemand)
}

type DemandData struct {
	Price  int64
	Demand int64
}

func (pe *PricingEngine) PersonalizedPrice(userProfile UserProfile) int64 {
	price := float64(pe.basePrice)
	if userProfile.PurchasePower > 8 {
		price *= 1.1
	} else if userProfile.PurchasePower < 3 {
		price *= 0.9
	}
	if userProfile.PriceSensitivity > 7 {
		price *= 0.95
	}
	if userProfile.Loyalty > 8 {
		price *= 0.92
	}
	return min(max(int64(price), pe.minPrice), pe.maxPrice)
}

type UserProfile struct {
	AvgOrderValue     int64
	PurchasePower     int
	PriceSensitivity  int
	Loyalty           int
	PurchaseFrequency int
}

func (pe *PricingEngine) OptimalPriceForProfit(cost int64, demandFunc func(int64) int64) int64 {
	goldenRatio := (math.Sqrt(5) - 1) / 2
	left := float64(pe.minPrice)
	right := float64(pe.maxPrice)
	for right-left > 1 {
		mid1 := left + (right-left)*(1-goldenRatio)
		mid2 := left + (right-left)*goldenRatio
		p1 := (int64(mid1) - cost) * demandFunc(int64(mid1))
		p2 := (int64(mid2) - cost) * demandFunc(int64(mid2))
		if p1 > p2 {
			right = mid2
		} else {
			left = mid1
		}
	}
	return int64((left + right) / 2)
}
