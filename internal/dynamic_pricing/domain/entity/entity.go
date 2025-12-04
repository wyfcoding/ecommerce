package entity

import (
	"time"

	"gorm.io/gorm" // 导入GORM库。
)

// PricingRequest 结构体定义了动态价格计算的输入请求参数。
// 它包含了计算价格所需的所有实时和历史数据。
type PricingRequest struct {
	SKUID              uint64 `json:"sku_id"`               // 待计算价格的SKU ID。
	BasePrice          int64  `json:"base_price"`           // 商品的基础价格（单位：分）。
	CurrentStock       int32  `json:"current_stock"`        // 当前库存量。
	TotalStock         int32  `json:"total_stock"`          // 总库存量。
	DailyDemand        int32  `json:"daily_demand"`         // 当天的实际需求量。
	AverageDailyDemand int32  `json:"average_daily_demand"` // 历史平均每日需求量。
	CompetitorPrice    int64  `json:"competitor_price"`     // 竞争对手的同类商品价格（单位：分）。
	UserLevel          string `json:"user_level"`           // 用户等级或类型，用于个性化定价。
}

// DynamicPrice 实体记录了每次动态价格计算的结果。
// 它包含了SKU的最终价格、调整幅度以及影响价格的各个因子。
type DynamicPrice struct {
	gorm.Model                 // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	SKUID            uint64    `gorm:"not null;index;comment:SKU ID" json:"sku_id"`               // 关联的SKU ID，索引字段。
	BasePrice        int64     `gorm:"not null;comment:基础价格" json:"base_price"`                   // 计算时的商品基础价格。
	FinalPrice       int64     `gorm:"not null;comment:最终价格" json:"final_price"`                  // 动态计算出的最终价格。
	PriceAdjustment  float64   `gorm:"type:decimal(10,2);comment:价格调整幅度" json:"price_adjustment"` // 价格调整的乘数或百分比。
	InventoryFactor  float64   `gorm:"type:decimal(10,2);comment:库存因子" json:"inventory_factor"`   // 库存对价格的影响因子。
	DemandFactor     float64   `gorm:"type:decimal(10,2);comment:需求因子" json:"demand_factor"`      // 需求对价格的影响因子。
	CompetitorFactor float64   `gorm:"type:decimal(10,2);comment:竞品因子" json:"competitor_factor"`  // 竞品价格对价格的影响因子。
	TimeFactor       float64   `gorm:"type:decimal(10,2);comment:时间因子" json:"time_factor"`        // 时间（如时段、季节）对价格的影响因子。
	UserFactor       float64   `gorm:"type:decimal(10,2);comment:用户因子" json:"user_factor"`        // 用户属性对价格的影响因子。
	EffectiveTime    time.Time `gorm:"comment:生效时间" json:"effective_time"`                        // 当前价格生效的时间。
	ExpiryTime       time.Time `gorm:"comment:过期时间" json:"expiry_time"`                           // 当前价格失效的时间。
}

// CompetitorPriceInfo 实体记录了SKU的竞品价格汇总信息。
// 它包含了我方价格、竞品列表、平均价格、最低价格等。
type CompetitorPriceInfo struct {
	gorm.Model                      // 嵌入gorm.Model。
	SKUID        uint64             `gorm:"not null;index;comment:SKU ID" json:"sku_id"`       // 关联的SKU ID，索引字段。
	OurPrice     int64              `gorm:"comment:我方价格" json:"our_price"`                     // 我方当前商品价格。
	Competitors  []*CompetitorPrice `gorm:"foreignKey:InfoID;comment:竞品列表" json:"competitors"` // 关联的竞品价格列表，一对多关系。
	AveragePrice int64              `gorm:"comment:平均价格" json:"average_price"`                 // 所有竞品价格的平均值。
	LowestPrice  int64              `gorm:"comment:最低价格" json:"lowest_price"`                  // 所有竞品价格中的最低值。
	HighestPrice int64              `gorm:"comment:最高价格" json:"highest_price"`                 // 所有竞品价格中的最高值。
	PriceRank    int32              `gorm:"comment:价格排名(1:最低,2:中等,3:最高)" json:"price_rank"`    // 我方价格在竞品中的排名（例如，1表示我方价格最低）。
	LastUpdated  time.Time          `gorm:"comment:最后更新时间" json:"last_updated"`                // 竞品价格信息最后更新时间。
}

// CompetitorPrice 实体记录了一个具体的竞争对手商品的价格信息。
type CompetitorPrice struct {
	gorm.Model               // 嵌入gorm.Model。
	InfoID         uint64    `gorm:"not null;index;comment:关联Info ID" json:"info_id"`                  // 关联的CompetitorPriceInfo ID，索引字段。
	CompetitorName string    `gorm:"type:varchar(255);not null;comment:竞争对手名称" json:"competitor_name"` // 竞争对手的名称。
	Price          int64     `gorm:"not null;comment:价格" json:"price"`                                 // 竞争对手的商品价格。
	URL            string    `gorm:"type:varchar(512);comment:商品链接" json:"url"`                        // 竞争对手的商品链接。
	LastUpdated    time.Time `gorm:"comment:最后更新时间" json:"last_updated"`                               // 该竞品价格信息最后更新时间。
}

// PriceHistoryData 实体记录了商品的每日价格和销量历史数据。
type PriceHistoryData struct {
	gorm.Model           // 嵌入gorm.Model。
	SKUID      uint64    `gorm:"not null;index;comment:SKU ID" json:"sku_id"` // 关联的SKU ID，索引字段。
	Date       time.Time `gorm:"comment:日期" json:"date"`                      // 数据记录的日期。
	Price      int64     `gorm:"comment:价格" json:"price"`                     // 当天的商品价格。
	Quantity   int32     `gorm:"comment:销量" json:"quantity"`                  // 当天的商品销量。
}

// PriceElasticity 实体记录了商品的价格弹性数据。
type PriceElasticity struct {
	gorm.Model           // 嵌入gorm.Model。
	SKUID      uint64    `gorm:"not null;index;comment:SKU ID" json:"sku_id"`                       // 关联的SKU ID，索引字段。
	Elasticity float64   `gorm:"type:decimal(10,4);comment:弹性系数" json:"elasticity"`                 // 价格弹性系数。
	Type       string    `gorm:"type:varchar(32);comment:类型(elastic/inelastic/normal)" json:"type"` // 弹性类型，例如“elastic”（弹性）、“inelastic”（非弹性）。
	DataPoints int       `gorm:"comment:数据点数量" json:"data_points"`                                  // 用于计算弹性的数据点数量。
	AnalyzedAt time.Time `gorm:"comment:分析时间" json:"analyzed_at"`                                   // 弹性数据分析时间。
}

// OptimizedPrice 实体记录了通过优化算法计算出的最优价格建议。
type OptimizedPrice struct {
	gorm.Model                     // 嵌入gorm.Model。
	SKUID                  uint64  `gorm:"not null;index;comment:SKU ID" json:"sku_id"`                       // 关联的SKU ID，索引字段。
	BasePrice              int64   `gorm:"comment:基础价格" json:"base_price"`                                    // 进行优化时的商品基础价格。
	OptimizedPrice         int64   `gorm:"comment:优化后价格" json:"optimized_price"`                              // 优化算法给出的建议价格。
	Reason                 string  `gorm:"type:text;comment:优化原因" json:"reason"`                              // 给出此优化价格的原因或依据。
	EstimatedRevenueImpact float64 `gorm:"type:decimal(10,2);comment:预计营收影响" json:"estimated_revenue_impact"` // 采纳此优化价格后，预计对营收的影响。
}

// PricingStrategy 实体定义了一个SKU的定价策略。
// 策略可以是固定价格、动态调整价格或阶梯价格等。
type PricingStrategy struct {
	gorm.Model                   // 嵌入gorm.Model。
	SKUID                 uint64 `gorm:"not null;index;comment:SKU ID" json:"sku_id"`                                       // 关联的SKU ID，索引字段。
	StrategyType          string `gorm:"type:varchar(32);not null;comment:策略类型(dynamic/fixed/tiered)" json:"strategy_type"` // 策略类型，例如“dynamic”（动态定价）、“fixed”（固定价格）。
	MinPrice              int64  `gorm:"comment:最低价格" json:"min_price"`                                                     // 策略允许的最低价格。
	MaxPrice              int64  `gorm:"comment:最高价格" json:"max_price"`                                                     // 策略允许的最高价格。
	InventoryThreshold    int32  `gorm:"comment:库存阈值" json:"inventory_threshold"`                                           // 触发价格调整的库存阈值。
	DemandThreshold       int32  `gorm:"comment:需求阈值" json:"demand_threshold"`                                              // 触发价格调整的需求阈值。
	CompetitorPriceOffset int64  `gorm:"comment:竞品价格偏移量" json:"competitor_price_offset"`                                    // 相对于竞品价格的偏移量。
	Enabled               bool   `gorm:"default:true;comment:是否启用" json:"enabled"`                                          // 策略是否启用，默认为启用。
}
