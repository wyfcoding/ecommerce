package domain

import "time"

// DynamicPrice 实体记录了每次动态价格计算的结果。
type DynamicPrice struct {
	ID               uint      `json:"id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
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

// CompetitorPriceInfo 实体记录了SKU的竞品价格汇总信息。
type CompetitorPriceInfo struct {
	ID           uint               `json:"id"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	SKUID        uint64             `gorm:"not null;index;comment:SKU ID" json:"sku_id"`
	OurPrice     int64              `gorm:"comment:我方价格" json:"our_price"`
	Competitors  []*CompetitorPrice `gorm:"foreignKey:InfoID;comment:竞品列表" json:"competitors"`
	AveragePrice int64              `gorm:"comment:平均价格" json:"average_price"`
	LowestPrice  int64              `gorm:"comment:最低价格" json:"lowest_price"`
	HighestPrice int64              `gorm:"comment:最高价格" json:"highest_price"`
	PriceRank    int32              `gorm:"comment:价格排名(1:最低,2:中等,3:最高)" json:"price_rank"`
	LastUpdated  time.Time          `gorm:"comment:最后更新时间" json:"last_updated"`
}

// CompetitorPrice 实体记录了一个具体的竞争对手商品的价格信息。
type CompetitorPrice struct {
	ID             uint      `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	InfoID         uint64    `gorm:"not null;index;comment:关联Info ID" json:"info_id"`
	CompetitorName string    `gorm:"type:varchar(255);not null;comment:竞争对手名称" json:"competitor_name"`
	Price          int64     `gorm:"not null;comment:价格" json:"price"`
	URL            string    `gorm:"type:varchar(512);comment:商品链接" json:"url"`
	LastUpdated    time.Time `gorm:"comment:最后更新时间" json:"last_updated"`
}

// PriceElasticity 实体记录了商品的价格弹性数据。
type PriceElasticity struct {
	ID         uint      `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	SKUID      uint64    `gorm:"not null;index;comment:SKU ID" json:"sku_id"`
	Elasticity float64   `gorm:"type:decimal(10,4);comment:弹性系数" json:"elasticity"`
	Type       string    `gorm:"type:varchar(32);comment:类型(elastic/inelastic/normal)" json:"type"`
	DataPoints int       `gorm:"comment:数据点数量" json:"data_points"`
	AnalyzedAt time.Time `gorm:"comment:分析时间" json:"analyzed_at"`
}

// OptimizedPrice 实体记录了通过优化算法计算出的最优价格建议。
type OptimizedPrice struct {
	ID                     uint      `json:"id"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
	SKUID                  uint64    `gorm:"not null;index;comment:SKU ID" json:"sku_id"`
	BasePrice              int64     `gorm:"comment:基础价格" json:"base_price"`
	OptimizedPrice         int64     `gorm:"comment:优化后价格" json:"optimized_price"`
	Reason                 string    `gorm:"type:text;comment:优化原因" json:"reason"`
	EstimatedRevenueImpact float64   `gorm:"type:decimal(10,2);comment:预计营收影响" json:"estimated_revenue_impact"`
}

// PricingRequest 结构体定义了动态价格计算的输入请求参数。
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

// UserProfile 结构体定义了用于个性化定价的用户画像信息。
type UserProfile struct {
	PurchasePower     int   // 购买力
	PriceSensitivity  int   // 价格敏感度
	Loyalty           int   // 忠诚度
	PurchaseFrequency int   // 购买频率
	AvgOrderValue     int64 // 平均客单价
}
