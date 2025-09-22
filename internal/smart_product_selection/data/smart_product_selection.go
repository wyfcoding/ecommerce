package data

import (
	"context"
	"ecommerce/internal/smart_product_selection/biz"
	"ecommerce/internal/smart_product_selection/data/model"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type smartProductSelectionRepo struct {
	data *Data
	// TODO: Add client for AI Model Service to get actual recommendations
}

// NewSmartProductSelectionRepo creates a new SmartProductSelectionRepo.
func NewSmartProductSelectionRepo(data *Data) biz.SmartProductSelectionRepo {
	return &smartProductSelectionRepo{data: data}
}

// SaveProductRecommendation saves a product recommendation record to the database.
func (r *smartProductSelectionRepo) SaveProductRecommendation(ctx context.Context, rec *biz.ProductRecommendation) (*biz.ProductRecommendation, error) {
	contextFeaturesBytes, err := json.Marshal(rec.ContextFeatures)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal context features: %w", err)
	}

	po := &model.ProductRecommendation{
		MerchantID:      rec.MerchantID,
		ProductID:       rec.ProductID,
		ProductName:     rec.ProductName,
		Score:           rec.Score,
		Reason:          rec.Reason,
		ContextFeatures: string(contextFeaturesBytes),
		RecommendedAt:   time.Now(),
	}
	if err := r.data.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	rec.ID = po.ID
	return rec, nil
}

// SimulateGetSmartProductRecommendations simulates getting smart product recommendations.
func (r *smartProductSelectionRepo) SimulateGetSmartProductRecommendations(ctx context.Context, merchantID string, contextFeatures map[string]string) ([]*biz.ProductRecommendation, error) {
	// In a real system, this would call an external AI model or a complex algorithm.
	// For now, return dummy recommendations.
	recommendations := []*biz.ProductRecommendation{
		{
			ProductID:   1001,
			ProductName: "Smart Product A",
			Score:       0.95,
			Reason:      "High demand in current season",
			MerchantID:  merchantID,
		},
		{
			ProductID:   1002,
			ProductName: "Smart Product B",
			Score:       0.90,
			Reason:      "Trending product",
			MerchantID:  merchantID,
		},
	}
	return recommendations, nil
}
