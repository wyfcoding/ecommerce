package repository

import (
	"context"

	"ecommerce/internal/pricing/model"
)

// PricingRepo defines the interface for pricing data access.
type PricingRepo interface {
	GetPriceRulesForSKUs(ctx context.Context, skuIDs []uint64) ([]*model.PriceRule, error)
	SaveDiscount(ctx context.Context, discount *model.Discount) (*model.Discount, error)
}