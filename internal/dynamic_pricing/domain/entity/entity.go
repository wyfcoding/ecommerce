package entity

import (
	"time"

	"gorm.io/gorm"
)

// PricingRequest 定价请求
type PricingRequest struct {
	SKUID              uint64 `json:"sku_id"`
	BasePrice          int64  `json:"base_price"`
	CurrentStock       int32  `json:"current_stock"`
	TotalStock         int32  `json:"total_stock"`
	DailyDemand        int32  `json:"daily_demand"`
	AverageDailyDemand int32  `json:"average_daily_demand"`
	CompetitorPrice    int64  `json:"competitor_price"`
	UserLevel          string `json:"user_level"`
}

// DynamicPrice 动态价格
type DynamicPrice struct {
	gorm.Model
	SKUID            uint64    `gorm:"not null;index;comment:SKU ID" json:"sku_id"`
	BasePrice        int64     `gorm:"not null;comment:基础价格" json:"base_price"`
	FinalPrice       int64     `gorm:"not null;comment:最终价格" json:"final_price"`
	PriceAdjustment  float64   `gorm:"type:decimal(10,2);comment:价格调整幅度" json:"price_adjustment"`
	InventoryFactor  float64   `gorm:"type:decimal(10,2);comment:库存因子" json:"inventory_factor"`
	DemandFactor     float64   `gorm:"type:decimal(10,2);comment:需求因子" json:"demand_factor"`
	CompetitorFactor float64   `gorm:"type:decimal(10,2);comment:竞品因子" json:"competitor_factor"`
	TimeFactor       float64   `gorm:"type:decimal(10,2);comment:时间因子" json:"time_factor"`
	UserFactor       float64   `gorm:"type:decimal(10,2);comment:用户因子" json:"user_factor"`
	EffectiveTime    time.Time `gorm:"comment:生效时间" json:"effective_time"`
	ExpiryTime       time.Time `gorm:"comment:过期时间" json:"expiry_time"`
}

// CompetitorPriceInfo 竞品价格信息
type CompetitorPriceInfo struct {
	gorm.Model
	SKUID        uint64             `gorm:"not null;index;comment:SKU ID" json:"sku_id"`
	OurPrice     int64              `gorm:"comment:我方价格" json:"our_price"`
	Competitors  []*CompetitorPrice `gorm:"foreignKey:InfoID;comment:竞品列表" json:"competitors"`
	AveragePrice int64              `gorm:"comment:平均价格" json:"average_price"`
	LowestPrice  int64              `gorm:"comment:最低价格" json:"lowest_price"`
	HighestPrice int64              `gorm:"comment:最高价格" json:"highest_price"`
	PriceRank    int32              `gorm:"comment:价格排名(1:最低,2:中等,3:最高)" json:"price_rank"`
	LastUpdated  time.Time          `gorm:"comment:最后更新时间" json:"last_updated"`
}

// CompetitorPrice 竞争对手价格
type CompetitorPrice struct {
	gorm.Model
	InfoID         uint64    `gorm:"not null;index;comment:关联Info ID" json:"info_id"`
	CompetitorName string    `gorm:"type:varchar(255);not null;comment:竞争对手名称" json:"competitor_name"`
	Price          int64     `gorm:"not null;comment:价格" json:"price"`
	URL            string    `gorm:"type:varchar(512);comment:商品链接" json:"url"`
	LastUpdated    time.Time `gorm:"comment:最后更新时间" json:"last_updated"`
}

// PriceHistoryData 价格历史数据
type PriceHistoryData struct {
	gorm.Model
	SKUID    uint64    `gorm:"not null;index;comment:SKU ID" json:"sku_id"`
	Date     time.Time `gorm:"comment:日期" json:"date"`
	Price    int64     `gorm:"comment:价格" json:"price"`
	Quantity int32     `gorm:"comment:销量" json:"quantity"`
}

// PriceElasticity 价格弹性
type PriceElasticity struct {
	gorm.Model
	SKUID      uint64    `gorm:"not null;index;comment:SKU ID" json:"sku_id"`
	Elasticity float64   `gorm:"type:decimal(10,4);comment:弹性系数" json:"elasticity"`
	Type       string    `gorm:"type:varchar(32);comment:类型(elastic/inelastic/normal)" json:"type"`
	DataPoints int       `gorm:"comment:数据点数量" json:"data_points"`
	AnalyzedAt time.Time `gorm:"comment:分析时间" json:"analyzed_at"`
}

// OptimizedPrice 优化价格
type OptimizedPrice struct {
	gorm.Model
	SKUID                  uint64  `gorm:"not null;index;comment:SKU ID" json:"sku_id"`
	BasePrice              int64   `gorm:"comment:基础价格" json:"base_price"`
	OptimizedPrice         int64   `gorm:"comment:优化后价格" json:"optimized_price"`
	Reason                 string  `gorm:"type:text;comment:优化原因" json:"reason"`
	EstimatedRevenueImpact float64 `gorm:"type:decimal(10,2);comment:预计营收影响" json:"estimated_revenue_impact"`
}

// PricingStrategy 定价策略
type PricingStrategy struct {
	gorm.Model
	SKUID                 uint64 `gorm:"not null;index;comment:SKU ID" json:"sku_id"`
	StrategyType          string `gorm:"type:varchar(32);not null;comment:策略类型(dynamic/fixed/tiered)" json:"strategy_type"`
	MinPrice              int64  `gorm:"comment:最低价格" json:"min_price"`
	MaxPrice              int64  `gorm:"comment:最高价格" json:"max_price"`
	InventoryThreshold    int32  `gorm:"comment:库存阈值" json:"inventory_threshold"`
	DemandThreshold       int32  `gorm:"comment:需求阈值" json:"demand_threshold"`
	CompetitorPriceOffset int64  `gorm:"comment:竞品价格偏移量" json:"competitor_price_offset"`
	Enabled               bool   `gorm:"default:true;comment:是否启用" json:"enabled"`
}
