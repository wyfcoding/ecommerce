package domain

import "time"

// PricingStrategy 定义了定价规则所采用的策略类型。
type PricingStrategy string

const (
	PricingStrategyFixed              PricingStrategy = "FIXED"
	PricingStrategyDynamic            PricingStrategy = "DYNAMIC"
	PricingStrategyCompetitive        PricingStrategy = "COMPETITIVE"
	PricingStrategyPromotion          PricingStrategy = "PROMOTION"
	PricingStrategyProfitMaximization PricingStrategy = "PROFIT_MAXIMIZATION"
	PricingStrategyDemandBased        PricingStrategy = "DEMAND_BASED"
	PricingStrategyInventoryBased     PricingStrategy = "INVENTORY_BASED"
)

// PricingRule 实体代表一个商品的定价规则。
type PricingRule struct {
	ID         uint64          `json:"id"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
	Name       string          `json:"name"`
	ProductID  uint64          `json:"product_id"`
	SkuID      uint64          `json:"sku_id"`
	Strategy   PricingStrategy `json:"strategy"`
	BasePrice  uint64          `json:"base_price"`
	MinPrice   uint64          `json:"min_price"`
	MaxPrice   uint64          `json:"max_price"`
	AdjustRate float64         `json:"adjust_rate"`
	Enabled    bool            `json:"enabled"`
	StartTime  time.Time       `json:"start_time"`
	EndTime    time.Time       `json:"end_time"`
}

// IsActive 检查定价规则当前是否处于活跃状态。
func (r *PricingRule) IsActive() bool {
	now := time.Now()
	return r.Enabled &&
		now.After(r.StartTime) &&
		now.Before(r.EndTime)
}

// CalculatePrice 根据定价规则的策略和外部因素计算商品的最终价格。
func (r *PricingRule) CalculatePrice(demand float64, competition float64) uint64 {
	var price uint64

	switch r.Strategy {
	case PricingStrategyFixed:
		price = r.BasePrice

	case PricingStrategyDynamic:
		adjustment := float64(r.BasePrice) * (r.AdjustRate * demand) / 100
		price = r.BasePrice + uint64(adjustment)

	case PricingStrategyCompetitive:
		competitorPrice := uint64(float64(r.BasePrice) * competition)
		adjustment := float64(competitorPrice) * r.AdjustRate / 100
		price = competitorPrice - uint64(adjustment)

	case PricingStrategyPromotion:
		price = r.BasePrice - uint64(float64(r.BasePrice)*r.AdjustRate/100)

	default:
		price = r.BasePrice
	}

	if price < r.MinPrice {
		price = r.MinPrice
	}
	if price > r.MaxPrice {
		price = r.MaxPrice
	}

	return price
}

// PriceHistory 实体代表商品价格的变动历史记录。
type PriceHistory struct {
	ID         uint64    `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	ProductID  uint64    `json:"product_id"`
	SkuID      uint64    `json:"sku_id"`
	Price      uint64    `json:"price"`
	OldPrice   uint64    `json:"old_price"`
	ChangeRate float64   `json:"change_rate"`
	Reason     string    `json:"reason"`
}
