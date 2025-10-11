package data

import (
	"context"
	"ecommerce/internal/pricing/biz"
	"ecommerce/internal/pricing/data/model"
	"time"
)

type pricingRepo struct {
	data *Data
}

// NewPricingRepo creates a new PricingRepo.
func NewPricingRepo(data *Data) biz.PricingRepo {
	return &pricingRepo{data: data}
}

// GetPriceRulesForSKUs retrieves pricing rules for given SKUs.
func (r *pricingRepo) GetPriceRulesForSKUs(ctx context.Context, skuIDs []uint64) ([]*biz.PriceRule, error) {
	var priceRules []*model.PriceRule
	if err := r.data.db.WithContext(ctx).Where("sku_id IN (?) AND valid_from <= ? AND valid_to >= ?", skuIDs, time.Now(), time.Now()).Find(&priceRules).Error; err != nil {
		return nil, err
	}

	bizPriceRules := make([]*biz.PriceRule, len(priceRules))
	for i, pr := range priceRules {
		bizPriceRules[i] = &biz.PriceRule{
			ID:        pr.ID,
			ProductID: pr.ProductID,
			SKUID:     pr.SKUID,
			RuleType:  pr.RuleType,
			Price:     pr.Price,
			ValidFrom: pr.ValidFrom,
			ValidTo:   pr.ValidTo,
			Priority:  pr.Priority,
		}
	}
	return bizPriceRules, nil
}

// SaveDiscount saves a discount record.
func (r *pricingRepo) SaveDiscount(ctx context.Context, discount *biz.Discount) (*biz.Discount, error) {
	po := &model.Discount{
		DiscountType:  discount.DiscountType,
		DiscountID:    discount.DiscountID,
		Amount:        discount.Amount,
		AppliedTo:     discount.AppliedTo,
		AppliedItemID: discount.AppliedItemID,
	}
	if err := r.data.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	discount.ID = po.ID
	return discount, nil
}
