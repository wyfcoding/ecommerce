package biz

import (
	"context"
	"fmt"
	"time"
)

// ProductRecommendation represents a smart product selection recommendation in the business logic layer.
type ProductRecommendation struct {
	ID            uint
	MerchantID    string
	ProductID     uint64
	ProductName   string
	Score         float64
	Reason        string
	ContextFeatures map[string]string
	RecommendedAt time.Time
}

// SmartProductSelectionRepo defines the interface for smart product selection data access.
type SmartProductSelectionRepo interface {
	SaveProductRecommendation(ctx context.Context, rec *ProductRecommendation) (*ProductRecommendation, error)
	SimulateGetSmartProductRecommendations(ctx context.Context, merchantID string, contextFeatures map[string]string) ([]*ProductRecommendation, error)
}

// AIModelClient defines the interface to interact with the AI Model Service.
type AIModelClient interface {
	// Assuming a generic recommendation/ranking model in AI Model Service
	RankRecommendations(ctx context.Context, userID string, itemIDs []string, userFeatures, itemFeatures map[string]string) ([]string, map[string]float64, error)
	// Or a more specific product selection model
	// PredictProductSelection(ctx context.Context, merchantID string, productIDs []uint64, contextFeatures map[string]string) ([]*ProductSelectionScore, error)
}

// SmartProductSelectionUsecase is the business logic for smart product selection.
type SmartProductSelectionUsecase struct {
	repo        SmartProductSelectionRepo
	aiModelClient AIModelClient
}

// NewSmartProductSelectionUsecase creates a new SmartProductSelectionUsecase.
func NewSmartProductSelectionUsecase(repo SmartProductSelectionRepo, aiModelClient AIModelClient) *SmartProductSelectionUsecase {
	return &SmartProductSelectionUsecase{repo: repo, aiModelClient: aiModelClient}
}

// GetSmartProductRecommendations gets smart product recommendations for a merchant.
func (uc *SmartProductSelectionUsecase) GetSmartProductRecommendations(ctx context.Context, merchantID string, contextFeatures map[string]string) ([]*ProductRecommendation, error) {
	// 1. Call AI Model Service to get recommendations (simulated)
	// In a real system, this would involve:
	// - Fetching candidate products from product service.
	// - Preparing features for the AI model.
	// - Calling a specific AI model for product selection/ranking.

	// For now, use the simulated method from repo.
	recommendations, err := uc.repo.SimulateGetSmartProductRecommendations(ctx, merchantID, contextFeatures)
	if err != nil {
		return nil, fmt.Errorf("failed to get smart product recommendations from AI model: %w", err)
	}

	// 2. Save recommendations (optional, for audit/feedback loop)
	for _, rec := range recommendations {
		rec.MerchantID = merchantID // Ensure merchant ID is set
		rec.ContextFeatures = contextFeatures
		_, err := uc.repo.SaveProductRecommendation(ctx, rec)
		if err != nil {
			fmt.Printf("failed to save product recommendation: %v\n", err)
		}
	}

	return recommendations, nil
}
