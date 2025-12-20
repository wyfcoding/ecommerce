package domain

import (
	"time"

	"gorm.io/gorm"
)

// PricingStrategy 定义了定价规则所采用的策略类型。
type PricingStrategy string

const (
	PricingStrategyFixed       PricingStrategy = "FIXED"
	PricingStrategyDynamic     PricingStrategy = "DYNAMIC"
	PricingStrategyCompetitive PricingStrategy = "COMPETITIVE"
	PricingStrategyPromotion   PricingStrategy = "PROMOTION"
)

// PricingRule 实体代表一个商品的定价规则。
type PricingRule struct {
	gorm.Model
	Name       string          `gorm:"type:varchar(255);not null;comment:规则名称" json:"name"`
	ProductID  uint64          `gorm:"not null;index;comment:商品ID" json:"product_id"`
	SkuID      uint64          `gorm:"not null;index;comment:SKU ID" json:"sku_id"`
	Strategy   PricingStrategy `gorm:"type:varchar(32);not null;comment:策略" json:"strategy"`
	BasePrice  uint64          `gorm:"not null;comment:基础价格(分)" json:"base_price"`
	MinPrice   uint64          `gorm:"not null;comment:最低价格(分)" json:"min_price"`
	MaxPrice   uint64          `gorm:"not null;comment:最高价格(分)" json:"max_price"`
	AdjustRate float64         `gorm:"type:decimal(10,2);not null;default:0;comment:调整率(%)" json:"adjust_rate"`
	Enabled    bool            `gorm:"default:true;comment:是否启用" json:"enabled"`
	StartTime  time.Time       `gorm:"comment:开始时间" json:"start_time"`
	EndTime    time.Time       `gorm:"comment:结束时间" json:"end_time"`
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
	gorm.Model
	ProductID  uint64  `gorm:"not null;index;comment:商品ID" json:"product_id"`
	SkuID      uint64  `gorm:"not null;index;comment:SKU ID" json:"sku_id"`
	Price      uint64  `gorm:"not null;comment:价格(分)" json:"price"`
	OldPrice   uint64  `gorm:"not null;comment:原价格(分)" json:"old_price"`
	ChangeRate float64 `gorm:"type:decimal(10,2);comment:变动率(%)" json:"change_rate"`
	Reason     string  `gorm:"type:varchar(255);comment:变动原因" json:"reason"`
}
