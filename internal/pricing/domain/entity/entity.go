package entity

import (
	"time" // 导入时间库。

	"gorm.io/gorm" // 导入GORM库。
)

// PricingStrategy 定义了定价规则所采用的策略类型。
type PricingStrategy string

const (
	PricingStrategyFixed       PricingStrategy = "FIXED"       // 固定价格：商品价格为预设值。
	PricingStrategyDynamic     PricingStrategy = "DYNAMIC"     // 动态定价：根据市场需求等因素实时调整价格。
	PricingStrategyCompetitive PricingStrategy = "COMPETITIVE" // 竞争定价：根据竞争对手的价格调整。
	PricingStrategyPromotion   PricingStrategy = "PROMOTION"   // 促销定价：为促销活动设置的特殊价格。
)

// PricingRule 实体代表一个商品的定价规则。
// 它包含了规则的名称、关联的商品/SKU、定价策略、价格范围和启用状态等。
type PricingRule struct {
	gorm.Model                 // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	Name       string          `gorm:"type:varchar(255);not null;comment:规则名称" json:"name"`                     // 规则名称。
	ProductID  uint64          `gorm:"not null;index;comment:商品ID" json:"product_id"`                           // 关联的商品ID，索引字段。
	SkuID      uint64          `gorm:"not null;index;comment:SKU ID" json:"sku_id"`                             // 关联的SKU ID，索引字段。
	Strategy   PricingStrategy `gorm:"type:varchar(32);not null;comment:策略" json:"strategy"`                    // 定价策略类型。
	BasePrice  uint64          `gorm:"not null;comment:基础价格(分)" json:"base_price"`                              // 定价的基础价格（单位：分）。
	MinPrice   uint64          `gorm:"not null;comment:最低价格(分)" json:"min_price"`                               // 价格的下限（单位：分）。
	MaxPrice   uint64          `gorm:"not null;comment:最高价格(分)" json:"max_price"`                               // 价格的上限（单位：分）。
	AdjustRate float64         `gorm:"type:decimal(10,2);not null;default:0;comment:调整率(%)" json:"adjust_rate"` // 价格调整的百分比或系数。
	Enabled    bool            `gorm:"default:true;comment:是否启用" json:"enabled"`                                // 规则是否启用。
	StartTime  time.Time       `gorm:"comment:开始时间" json:"start_time"`                                          // 规则的生效开始时间。
	EndTime    time.Time       `gorm:"comment:结束时间" json:"end_time"`                                            // 规则的失效结束时间。
}

// IsActive 检查定价规则当前是否处于活跃状态。
func (r *PricingRule) IsActive() bool {
	now := time.Now()
	return r.Enabled && // 规则已启用。
		now.After(r.StartTime) && // 当前时间在开始时间之后。
		now.Before(r.EndTime) // 当前时间在结束时间之前。
}

// CalculatePrice 根据定价规则的策略和外部因素计算商品的最终价格。
// demand: 需求系数（例如，0.8表示需求较低，1.2表示需求较高）。
// competition: 竞争系数（例如，0.9表示竞争激烈，1.1表示竞争不激烈）。
// 返回计算后的价格（单位：分）。
func (r *PricingRule) CalculatePrice(demand float64, competition float64) uint64 {
	var price uint64

	switch r.Strategy {
	case PricingStrategyFixed:
		price = r.BasePrice // 固定价格，直接使用基础价格。

	case PricingStrategyDynamic:
		// 根据需求动态调整价格：基础价格 + 基础价格 * 调整率 * 需求系数。
		adjustment := float64(r.BasePrice) * (r.AdjustRate * demand) / 100
		price = r.BasePrice + uint64(adjustment)

	case PricingStrategyCompetitive:
		// 根据竞争对手价格调整：假设 competition 为竞争对手的价格系数。
		// competitorPrice 假设是根据 competition 调整后的价格。
		// 例如，如果 competition < 1，表示竞争对手价格更低，需要降低价格。
		competitorPrice := uint64(float64(r.BasePrice) * competition)
		adjustment := float64(competitorPrice) * r.AdjustRate / 100
		price = competitorPrice - uint64(adjustment) // 假设AdjustRate为折扣率。

	case PricingStrategyPromotion:
		// 促销价格：基础价格 - 基础价格 * 调整率（折扣百分比）。
		price = r.BasePrice - uint64(float64(r.BasePrice)*r.AdjustRate/100)

	default:
		price = r.BasePrice // 默认为基础价格。
	}

	// 确保计算出的价格在设定的最低价格和最高价格之间。
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
	gorm.Model         // 嵌入gorm.Model。
	ProductID  uint64  `gorm:"not null;index;comment:商品ID" json:"product_id"`        // 关联的商品ID，索引字段。
	SkuID      uint64  `gorm:"not null;index;comment:SKU ID" json:"sku_id"`          // 关联的SKU ID，索引字段。
	Price      uint64  `gorm:"not null;comment:价格(分)" json:"price"`                  // 变动后的新价格（单位：分）。
	OldPrice   uint64  `gorm:"not null;comment:原价格(分)" json:"old_price"`             // 变动前的原价格（单位：分）。
	ChangeRate float64 `gorm:"type:decimal(10,2);comment:变动率(%)" json:"change_rate"` // 价格变动百分比。
	Reason     string  `gorm:"type:varchar(255);comment:变动原因" json:"reason"`         // 价格变动的原因。
}
