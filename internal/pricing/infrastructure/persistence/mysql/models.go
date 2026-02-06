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
