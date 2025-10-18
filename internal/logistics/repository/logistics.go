package repository

import (
	"context"

	"ecommerce/internal/logistics/model"
)

// LogisticsRepo defines the interface for logistics data access.
type LogisticsRepo interface {
	GetShippingRules(ctx context.Context, origin, destination string) ([]*model.ShippingRule, error)
}