package mysql

import (
	"time"

	"github.com/wyfcoding/ecommerce/internal/pricing/domain"
	"gorm.io/gorm"
)

// PricingRuleModel 定价规则写模型。
type PricingRuleModel struct {
	gorm.Model
	Name       string                 `gorm:"type:varchar(255);not null;comment:规则名称"`
	ProductID  uint64                 `gorm:"not null;index;comment:商品ID"`
	SkuID      uint64                 `gorm:"not null;index;comment:SKU ID"`
	Strategy   domain.PricingStrategy `gorm:"type:varchar(32);not null;comment:策略"`
	BasePrice  uint64                 `gorm:"not null;comment:基础价格(分)"`
	MinPrice   uint64                 `gorm:"not null;comment:最低价格(分)"`
	MaxPrice   uint64                 `gorm:"not null;comment:最高价格(分)"`
	AdjustRate float64                `gorm:"type:decimal(10,2);not null;default:0;comment:调整率(%)"`
	Enabled    bool                   `gorm:"default:true;comment:是否启用"`
	StartTime  time.Time              `gorm:"comment:开始时间"`
	EndTime    time.Time              `gorm:"comment:结束时间"`
}

// PriceHistoryModel 价格历史写模型。
type PriceHistoryModel struct {
	gorm.Model
	ProductID  uint64  `gorm:"not null;index;comment:商品ID"`
	SkuID      uint64  `gorm:"not null;index;comment:SKU ID"`
	Price      uint64  `gorm:"not null;comment:价格(分)"`
	OldPrice   uint64  `gorm:"not null;comment:原价格(分)"`
	ChangeRate float64 `gorm:"type:decimal(10,2);comment:变动率(%)"`
	Reason     string  `gorm:"type:varchar(255);comment:变动原因"`
}

func (PricingRuleModel) TableName() string {
	return "pricing_rules"
}

func (PriceHistoryModel) TableName() string {
	return "price_histories"
}

func toPricingRuleModel(rule *domain.PricingRule) *PricingRuleModel {
	if rule == nil {
		return nil
	}
	return &PricingRuleModel{
		Model: gorm.Model{
			ID:        uint(rule.ID),
			CreatedAt: rule.CreatedAt,
			UpdatedAt: rule.UpdatedAt,
		},
		Name:       rule.Name,
		ProductID:  rule.ProductID,
		SkuID:      rule.SkuID,
		Strategy:   rule.Strategy,
		BasePrice:  rule.BasePrice,
		MinPrice:   rule.MinPrice,
		MaxPrice:   rule.MaxPrice,
		AdjustRate: rule.AdjustRate,
		Enabled:    rule.Enabled,
		StartTime:  rule.StartTime,
		EndTime:    rule.EndTime,
	}
}

func toPricingRule(model *PricingRuleModel) *domain.PricingRule {
	if model == nil {
		return nil
	}
	return &domain.PricingRule{
		ID:         uint64(model.ID),
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
		Name:       model.Name,
		ProductID:  model.ProductID,
		SkuID:      model.SkuID,
		Strategy:   model.Strategy,
		BasePrice:  model.BasePrice,
		MinPrice:   model.MinPrice,
		MaxPrice:   model.MaxPrice,
		AdjustRate: model.AdjustRate,
		Enabled:    model.Enabled,
		StartTime:  model.StartTime,
		EndTime:    model.EndTime,
	}
}

func toPriceHistoryModel(history *domain.PriceHistory) *PriceHistoryModel {
	if history == nil {
		return nil
	}
	return &PriceHistoryModel{
		Model: gorm.Model{
			ID:        uint(history.ID),
			CreatedAt: history.CreatedAt,
			UpdatedAt: history.UpdatedAt,
		},
		ProductID:  history.ProductID,
		SkuID:      history.SkuID,
		Price:      history.Price,
		OldPrice:   history.OldPrice,
		ChangeRate: history.ChangeRate,
		Reason:     history.Reason,
	}
}

func toPriceHistory(model *PriceHistoryModel) *domain.PriceHistory {
	if model == nil {
		return nil
	}
	return &domain.PriceHistory{
		ID:         uint64(model.ID),
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
		ProductID:  model.ProductID,
		SkuID:      model.SkuID,
		Price:      model.Price,
		OldPrice:   model.OldPrice,
		ChangeRate: model.ChangeRate,
		Reason:     model.Reason,
	}
}

// DynamicPriceModel 动态价格记录模型。
type DynamicPriceModel struct {
	gorm.Model
	SKUID            uint64  `gorm:"not null;index;comment:SKU ID"`
	BasePrice        int64   `gorm:"not null;comment:基础价格"`
	FinalPrice       int64   `gorm:"not null;comment:最终价格"`
	PriceAdjustment  float64 `gorm:"type:decimal(10,2);comment:价格调整幅度"`
	InventoryFactor  float64 `gorm:"type:decimal(10,2);comment:库存因子"`
	DemandFactor     float64 `gorm:"type:decimal(10,2);comment:需求因子"`
	CompetitorFactor float64 `gorm:"type:decimal(10,2);comment:竞品因子"`
	TimeFactor       float64 `gorm:"type:decimal(10,2);comment:时间因子"`
	UserFactor       float64 `gorm:"type:decimal(10,2);comment:用户因子"`
	EffectiveTime    time.Time
	ExpiryTime       time.Time
}

func (DynamicPriceModel) TableName() string {
	return "dynamic_prices"
}

// CompetitorPriceInfoModel 竞品汇总模型。
type CompetitorPriceInfoModel struct {
	gorm.Model
	SKUID        uint64 `gorm:"not null;index"`
	OurPrice     int64
	AveragePrice int64
	LowestPrice  int64
	HighestPrice int64
	PriceRank    int32
	LastUpdated  time.Time
}

func (CompetitorPriceInfoModel) TableName() string {
	return "competitor_price_infos"
}

// CompetitorPriceModel 具体竞品数据模型。
type CompetitorPriceModel struct {
	gorm.Model
	InfoID         uint64 `gorm:"not null;index"`
	CompetitorName string `gorm:"type:varchar(255);not null"`
	Price          int64  `gorm:"not null"`
	URL            string `gorm:"type:varchar(512)"`
	LastUpdated    time.Time
}

func (CompetitorPriceModel) TableName() string {
	return "competitor_prices"
}

// PriceElasticityModel 价格弹性模型。
type PriceElasticityModel struct {
	gorm.Model
	SKUID      uint64  `gorm:"not null;index"`
	Elasticity float64 `gorm:"type:decimal(10,4)"`
	Type       string  `gorm:"type:varchar(32)"`
	DataPoints int
	AnalyzedAt time.Time
}

func (PriceElasticityModel) TableName() string {
	return "price_elasticities"
}

func toDynamicPriceModel(p *domain.DynamicPrice) *DynamicPriceModel {
	if p == nil {
		return nil
	}
	return &DynamicPriceModel{
		Model: gorm.Model{
			ID:        p.ID,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		},
		SKUID:            p.SKUID,
		BasePrice:        p.BasePrice,
		FinalPrice:       p.FinalPrice,
		PriceAdjustment:  p.PriceAdjustment,
		InventoryFactor:  p.InventoryFactor,
		DemandFactor:     p.DemandFactor,
		CompetitorFactor: p.CompetitorFactor,
		TimeFactor:       p.TimeFactor,
		UserFactor:       p.UserFactor,
		EffectiveTime:    p.EffectiveTime,
		ExpiryTime:       p.ExpiryTime,
	}
}

func toDynamicPrice(m *DynamicPriceModel) *domain.DynamicPrice {
	if m == nil {
		return nil
	}
	return &domain.DynamicPrice{
		ID:               m.ID,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
		SKUID:            m.SKUID,
		BasePrice:        m.BasePrice,
		FinalPrice:       m.FinalPrice,
		PriceAdjustment:  m.PriceAdjustment,
		InventoryFactor:  m.InventoryFactor,
		DemandFactor:     m.DemandFactor,
		CompetitorFactor: m.CompetitorFactor,
		TimeFactor:       m.TimeFactor,
		UserFactor:       m.UserFactor,
		EffectiveTime:    m.EffectiveTime,
		ExpiryTime:       m.ExpiryTime,
	}
}

func toCompetitorPriceInfoModel(p *domain.CompetitorPriceInfo) *CompetitorPriceInfoModel {
	if p == nil {
		return nil
	}
	return &CompetitorPriceInfoModel{
		Model: gorm.Model{
			ID:        p.ID,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		},
		SKUID:        p.SKUID,
		OurPrice:     p.OurPrice,
		AveragePrice: p.AveragePrice,
		LowestPrice:  p.LowestPrice,
		HighestPrice: p.HighestPrice,
		PriceRank:    p.PriceRank,
		LastUpdated:  p.LastUpdated,
	}
}

func toCompetitorPriceInfo(m *CompetitorPriceInfoModel) *domain.CompetitorPriceInfo {
	if m == nil {
		return nil
	}
	return &domain.CompetitorPriceInfo{
		ID:           m.ID,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		SKUID:        m.SKUID,
		OurPrice:     m.OurPrice,
		AveragePrice: m.AveragePrice,
		LowestPrice:  m.LowestPrice,
		HighestPrice: m.HighestPrice,
		PriceRank:    m.PriceRank,
		LastUpdated:  m.LastUpdated,
	}
}

func toCompetitorPriceModel(p *domain.CompetitorPrice) *CompetitorPriceModel {
	if p == nil {
		return nil
	}
	return &CompetitorPriceModel{
		Model: gorm.Model{
			ID:        p.ID,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		},
		InfoID:         p.InfoID,
		CompetitorName: p.CompetitorName,
		Price:          p.Price,
		URL:            p.URL,
		LastUpdated:    p.LastUpdated,
	}
}

func toCompetitorPrice(m *CompetitorPriceModel) *domain.CompetitorPrice {
	if m == nil {
		return nil
	}
	return &domain.CompetitorPrice{
		ID:             m.ID,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
		InfoID:         m.InfoID,
		CompetitorName: m.CompetitorName,
		Price:          m.Price,
		URL:            m.URL,
		LastUpdated:    m.LastUpdated,
	}
}

func toPriceElasticityModel(p *domain.PriceElasticity) *PriceElasticityModel {
	if p == nil {
		return nil
	}
	return &PriceElasticityModel{
		Model: gorm.Model{
			ID:        p.ID,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		},
		SKUID:      p.SKUID,
		Elasticity: p.Elasticity,
		Type:       p.Type,
		DataPoints: p.DataPoints,
		AnalyzedAt: p.AnalyzedAt,
	}
}

func toPriceElasticity(m *PriceElasticityModel) *domain.PriceElasticity {
	if m == nil {
		return nil
	}
	return &domain.PriceElasticity{
		ID:         m.ID,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
		SKUID:      m.SKUID,
		Elasticity: m.Elasticity,
		Type:       m.Type,
		DataPoints: m.DataPoints,
		AnalyzedAt: m.AnalyzedAt,
	}
}
