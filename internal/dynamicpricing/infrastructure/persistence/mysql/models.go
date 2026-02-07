package mysql

import (
	"time"

	"github.com/wyfcoding/ecommerce/internal/dynamicpricing/domain"
	"gorm.io/gorm"
)

// DynamicPriceModel 动态价格写模型。
type DynamicPriceModel struct {
	gorm.Model
	SKUID            uint64    `gorm:"column:sku_id;not null;index;comment:SKU ID"`
	BasePrice        int64     `gorm:"column:base_price;not null;comment:基础价格"`
	FinalPrice       int64     `gorm:"column:final_price;not null;comment:最终价格"`
	PriceAdjustment  float64   `gorm:"column:price_adjustment;type:decimal(10,2);comment:价格调整幅度"`
	InventoryFactor  float64   `gorm:"column:inventory_factor;type:decimal(10,2);comment:库存因子"`
	DemandFactor     float64   `gorm:"column:demand_factor;type:decimal(10,2);comment:需求因子"`
	CompetitorFactor float64   `gorm:"column:competitor_factor;type:decimal(10,2);comment:竞品因子"`
	TimeFactor       float64   `gorm:"column:time_factor;type:decimal(10,2);comment:时间因子"`
	UserFactor       float64   `gorm:"column:user_factor;type:decimal(10,2);comment:用户因子"`
	EffectiveTime    time.Time `gorm:"column:effective_time;comment:生效时间"`
	ExpiryTime       time.Time `gorm:"column:expiry_time;comment:过期时间"`
}

// PricingStrategyModel 定价策略写模型。
type PricingStrategyModel struct {
	gorm.Model
	SKUID                 uint64 `gorm:"column:sku_id;not null;index;comment:SKU ID"`
	StrategyType          string `gorm:"column:strategy_type;type:varchar(32);not null;comment:策略类型"`
	MinPrice              int64  `gorm:"column:min_price;comment:最低价格"`
	MaxPrice              int64  `gorm:"column:max_price;comment:最高价格"`
	InventoryThreshold    int32  `gorm:"column:inventory_threshold;comment:库存阈值"`
	DemandThreshold       int32  `gorm:"column:demand_threshold;comment:需求阈值"`
	CompetitorPriceOffset int64  `gorm:"column:competitor_price_offset;comment:竞品价格偏移量"`
	Enabled               bool   `gorm:"column:enabled;default:true;comment:是否启用"`
}

// CompetitorPriceInfoModel 竞品价格汇总写模型。
type CompetitorPriceInfoModel struct {
	gorm.Model
	SKUID        uint64                 `gorm:"column:sku_id;not null;index;comment:SKU ID"`
	OurPrice     int64                  `gorm:"column:our_price;comment:我方价格"`
	AveragePrice int64                  `gorm:"column:average_price;comment:平均价格"`
	LowestPrice  int64                  `gorm:"column:lowest_price;comment:最低价格"`
	HighestPrice int64                  `gorm:"column:highest_price;comment:最高价格"`
	PriceRank    int32                  `gorm:"column:price_rank;comment:价格排名"`
	LastUpdated  time.Time              `gorm:"column:last_updated;comment:最后更新时间"`
	Competitors  []CompetitorPriceModel `gorm:"foreignKey:InfoID"`
}

// CompetitorPriceModel 竞品价格明细写模型。
type CompetitorPriceModel struct {
	gorm.Model
	InfoID         uint64    `gorm:"column:info_id;not null;index;comment:关联Info ID"`
	CompetitorName string    `gorm:"column:competitor_name;type:varchar(255);not null;comment:竞争对手名称"`
	Price          int64     `gorm:"column:price;not null;comment:价格"`
	URL            string    `gorm:"column:url;type:varchar(512);comment:商品链接"`
	LastUpdated    time.Time `gorm:"column:last_updated;comment:最后更新时间"`
}

// PriceHistoryDataModel 价格历史数据写模型。
type PriceHistoryDataModel struct {
	gorm.Model
	SKUID    uint64    `gorm:"column:sku_id;not null;index;comment:SKU ID"`
	Date     time.Time `gorm:"column:date;comment:日期"`
	Price    int64     `gorm:"column:price;comment:价格"`
	Quantity int32     `gorm:"column:quantity;comment:销量"`
}

// PriceElasticityModel 价格弹性写模型。
type PriceElasticityModel struct {
	gorm.Model
	SKUID      uint64    `gorm:"column:sku_id;not null;index;comment:SKU ID"`
	Elasticity float64   `gorm:"column:elasticity;type:decimal(10,4);comment:弹性系数"`
	Type       string    `gorm:"column:type;type:varchar(32);comment:类型"`
	DataPoints int       `gorm:"column:data_points;comment:数据点数量"`
	AnalyzedAt time.Time `gorm:"column:analyzed_at;comment:分析时间"`
}

func (DynamicPriceModel) TableName() string        { return "dynamic_prices" }
func (PricingStrategyModel) TableName() string     { return "pricing_strategies" }
func (CompetitorPriceInfoModel) TableName() string { return "competitor_price_infos" }
func (CompetitorPriceModel) TableName() string     { return "competitor_prices" }
func (PriceHistoryDataModel) TableName() string    { return "price_history" }
func (PriceElasticityModel) TableName() string     { return "price_elasticities" }

func toDynamicPriceModel(price *domain.DynamicPrice) *DynamicPriceModel {
	if price == nil {
		return nil
	}
	return &DynamicPriceModel{
		Model: gorm.Model{
			ID:        price.ID,
			CreatedAt: price.CreatedAt,
			UpdatedAt: price.UpdatedAt,
		},
		SKUID:            price.SKUID,
		BasePrice:        price.BasePrice,
		FinalPrice:       price.FinalPrice,
		PriceAdjustment:  price.PriceAdjustment,
		InventoryFactor:  price.InventoryFactor,
		DemandFactor:     price.DemandFactor,
		CompetitorFactor: price.CompetitorFactor,
		TimeFactor:       price.TimeFactor,
		UserFactor:       price.UserFactor,
		EffectiveTime:    price.EffectiveTime,
		ExpiryTime:       price.ExpiryTime,
	}
}

func toDynamicPrice(model *DynamicPriceModel) *domain.DynamicPrice {
	if model == nil {
		return nil
	}
	return &domain.DynamicPrice{
		ID:               model.ID,
		CreatedAt:        model.CreatedAt,
		UpdatedAt:        model.UpdatedAt,
		SKUID:            model.SKUID,
		BasePrice:        model.BasePrice,
		FinalPrice:       model.FinalPrice,
		PriceAdjustment:  model.PriceAdjustment,
		InventoryFactor:  model.InventoryFactor,
		DemandFactor:     model.DemandFactor,
		CompetitorFactor: model.CompetitorFactor,
		TimeFactor:       model.TimeFactor,
		UserFactor:       model.UserFactor,
		EffectiveTime:    model.EffectiveTime,
		ExpiryTime:       model.ExpiryTime,
	}
}

func toPricingStrategyModel(strategy *domain.PricingStrategy) *PricingStrategyModel {
	if strategy == nil {
		return nil
	}
	return &PricingStrategyModel{
		Model: gorm.Model{
			ID:        strategy.ID,
			CreatedAt: strategy.CreatedAt,
			UpdatedAt: strategy.UpdatedAt,
		},
		SKUID:                 strategy.SKUID,
		StrategyType:          strategy.StrategyType,
		MinPrice:              strategy.MinPrice,
		MaxPrice:              strategy.MaxPrice,
		InventoryThreshold:    strategy.InventoryThreshold,
		DemandThreshold:       strategy.DemandThreshold,
		CompetitorPriceOffset: strategy.CompetitorPriceOffset,
		Enabled:               strategy.Enabled,
	}
}

func toPricingStrategy(model *PricingStrategyModel) *domain.PricingStrategy {
	if model == nil {
		return nil
	}
	return &domain.PricingStrategy{
		ID:                    model.ID,
		CreatedAt:             model.CreatedAt,
		UpdatedAt:             model.UpdatedAt,
		SKUID:                 model.SKUID,
		StrategyType:          model.StrategyType,
		MinPrice:              model.MinPrice,
		MaxPrice:              model.MaxPrice,
		InventoryThreshold:    model.InventoryThreshold,
		DemandThreshold:       model.DemandThreshold,
		CompetitorPriceOffset: model.CompetitorPriceOffset,
		Enabled:               model.Enabled,
	}
}

func toCompetitorPriceInfo(model *CompetitorPriceInfoModel) *domain.CompetitorPriceInfo {
	if model == nil {
		return nil
	}
	competitors := make([]*domain.CompetitorPrice, len(model.Competitors))
	for i := range model.Competitors {
		competitors[i] = toCompetitorPrice(&model.Competitors[i])
	}
	return &domain.CompetitorPriceInfo{
		ID:           model.ID,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
		SKUID:        model.SKUID,
		OurPrice:     model.OurPrice,
		Competitors:  competitors,
		AveragePrice: model.AveragePrice,
		LowestPrice:  model.LowestPrice,
		HighestPrice: model.HighestPrice,
		PriceRank:    model.PriceRank,
		LastUpdated:  model.LastUpdated,
	}
}

func toCompetitorPriceInfoModel(info *domain.CompetitorPriceInfo) *CompetitorPriceInfoModel {
	if info == nil {
		return nil
	}
	competitors := make([]CompetitorPriceModel, len(info.Competitors))
	for i, c := range info.Competitors {
		competitors[i] = *toCompetitorPriceModel(c)
	}
	return &CompetitorPriceInfoModel{
		Model: gorm.Model{
			ID:        info.ID,
			CreatedAt: info.CreatedAt,
			UpdatedAt: info.UpdatedAt,
		},
		SKUID:        info.SKUID,
		OurPrice:     info.OurPrice,
		AveragePrice: info.AveragePrice,
		LowestPrice:  info.LowestPrice,
		HighestPrice: info.HighestPrice,
		PriceRank:    info.PriceRank,
		LastUpdated:  info.LastUpdated,
		Competitors:  competitors,
	}
}

func toCompetitorPrice(model *CompetitorPriceModel) *domain.CompetitorPrice {
	if model == nil {
		return nil
	}
	return &domain.CompetitorPrice{
		ID:             model.ID,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
		InfoID:         model.InfoID,
		CompetitorName: model.CompetitorName,
		Price:          model.Price,
		URL:            model.URL,
		LastUpdated:    model.LastUpdated,
	}
}

func toCompetitorPriceModel(price *domain.CompetitorPrice) *CompetitorPriceModel {
	if price == nil {
		return nil
	}
	return &CompetitorPriceModel{
		Model: gorm.Model{
			ID:        price.ID,
			CreatedAt: price.CreatedAt,
			UpdatedAt: price.UpdatedAt,
		},
		InfoID:         price.InfoID,
		CompetitorName: price.CompetitorName,
		Price:          price.Price,
		URL:            price.URL,
		LastUpdated:    price.LastUpdated,
	}
}

func toPriceHistoryData(list []PriceHistoryDataModel) []domain.PriceHistoryData {
	items := make([]domain.PriceHistoryData, len(list))
	for i := range list {
		items[i] = domain.PriceHistoryData{
			ID:        list[i].ID,
			CreatedAt: list[i].CreatedAt,
			UpdatedAt: list[i].UpdatedAt,
			SKUID:     list[i].SKUID,
			Date:      list[i].Date,
			Price:     list[i].Price,
			Quantity:  list[i].Quantity,
		}
	}
	return items
}

func toPriceElasticity(model *PriceElasticityModel) *domain.PriceElasticity {
	if model == nil {
		return nil
	}
	return &domain.PriceElasticity{
		ID:         model.ID,
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
		SKUID:      model.SKUID,
		Elasticity: model.Elasticity,
		Type:       model.Type,
		DataPoints: model.DataPoints,
		AnalyzedAt: model.AnalyzedAt,
	}
}
