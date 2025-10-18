package service

import (
	"context"
	"fmt"
	"time"

	"ecommerce/internal/smart_product_selection/client"
	"ecommerce/internal/smart_product_selection/model"
	"ecommerce/internal/smart_product_selection/repository"
)

// SmartProductSelectionService is the business logic for smart product selection.
type SmartProductSelectionService struct {
	repo          repository.SmartProductSelectionRepo
	aiModelClient client.AIModelClient
}

// NewSmartProductSelectionService creates a new SmartProductSelectionService.
func NewSmartProductSelectionService(repo repository.SmartProductSelectionRepo, aiModelClient client.AIModelClient) *SmartProductSelectionService {
	return &SmartProductSelectionService{repo: repo, aiModelClient: aiModelClient}
}

// GetSmartProductRecommendations gets smart product recommendations for a merchant.
func (s *SmartProductSelectionService) GetSmartProductRecommendations(ctx context.Context, merchantID string, contextFeatures map[string]string) ([]*model.ProductRecommendation, error) {
	// 1. Call AI Model Service to get recommendations (simulated)
	// In a real system, this would involve:
	// - Fetching candidate products from product service.
	// - Preparing features for the AI model.
	// - Calling a specific AI model for product selection/ranking.

	// For now, use the simulated method from repo.
	recommendations, err := s.repo.SimulateGetSmartProductRecommendations(ctx, merchantID, contextFeatures)
	if err != nil {
		return nil, fmt.Errorf("failed to get smart product recommendations from AI model: %w", err)
	}

	// 2. Save recommendations (optional, for audit/feedback loop)
	for _, rec := range recommendations {
		rec.MerchantID = merchantID // Ensure merchant ID is set
		rec.ContextFeatures = contextFeatures
		_, err := s.repo.SaveProductRecommendation(ctx, rec)
		if err != nil {
			fmt.Printf("failed to save product recommendation: %v\n", err)
		}
	}

	return recommendations, nil
}
