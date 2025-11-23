package entity

import (
	"time"

	"gorm.io/gorm"
)

// PricingStrategy 定价策略
type PricingStrategy string

const (
	PricingStrategyFixed       PricingStrategy = "FIXED"       // 固定价格
	PricingStrategyDynamic     PricingStrategy = "DYNAMIC"     // 动态定价
	PricingStrategyCompetitive PricingStrategy = "COMPETITIVE" // 竞争定价
	PricingStrategyPromotion   PricingStrategy = "PROMOTION"   // 促销定价
)

// PricingRule 定价规则实体
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

// IsActive 是否激活
func (r *PricingRule) IsActive() bool {
	now := time.Now()
	return r.Enabled && now.After(r.StartTime) && now.Before(r.EndTime)
}

// CalculatePrice 计算价格
func (r *PricingRule) CalculatePrice(demand float64, competition float64) uint64 {
	var price uint64

	switch r.Strategy {
	case PricingStrategyFixed:
		price = r.BasePrice

	case PricingStrategyDynamic:
		// 根据需求动态调整价格
		adjustment := float64(r.BasePrice) * (r.AdjustRate * demand) / 100
		price = r.BasePrice + uint64(adjustment)

	case PricingStrategyCompetitive:
		// 根据竞争对手价格调整
		competitorPrice := uint64(float64(r.BasePrice) * competition)
		adjustment := float64(competitorPrice) * r.AdjustRate / 100
		price = competitorPrice - uint64(adjustment)

	case PricingStrategyPromotion:
		// 促销价格
		price = r.BasePrice - uint64(float64(r.BasePrice)*r.AdjustRate/100)

	default:
		price = r.BasePrice
	}

	// 确保价格在范围内
	if price < r.MinPrice {
		price = r.MinPrice
	}
	if price > r.MaxPrice {
		price = r.MaxPrice
	}

	return price
}

// PriceHistory 价格历史实体
type PriceHistory struct {
	gorm.Model
	ProductID  uint64  `gorm:"not null;index;comment:商品ID" json:"product_id"`
	SkuID      uint64  `gorm:"not null;index;comment:SKU ID" json:"sku_id"`
	Price      uint64  `gorm:"not null;comment:价格(分)" json:"price"`
	OldPrice   uint64  `gorm:"not null;comment:原价格(分)" json:"old_price"`
	ChangeRate float64 `gorm:"type:decimal(10,2);comment:变动率(%)" json:"change_rate"`
	Reason     string  `gorm:"type:varchar(255);comment:变动原因" json:"reason"`
}
