package repository

import (
	"context"

	"ecommerce/internal/smart_product_selection/model"
)

// SmartProductSelectionRepo defines the interface for smart product selection data access.
type SmartProductSelectionRepo interface {
	SaveProductRecommendation(ctx context.Context, rec *model.ProductRecommendation) (*model.ProductRecommendation, error)
	SimulateGetSmartProductRecommendations(ctx context.Context, merchantID string, contextFeatures map[string]string) ([]*model.ProductRecommendation, error)
}