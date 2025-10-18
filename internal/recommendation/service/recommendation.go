package service

import (
	"context"
	"fmt"

	"ecommerce/internal/recommendation/client"
	"ecommerce/internal/recommendation/model"
	"ecommerce/internal/recommendation/repository"
)

// RecommendationService is the business logic for recommendation.
type RecommendationService struct {
	repo          repository.RecommendationRepo
	aiModelClient client.AIModelClient // Added AI Model client
}

// NewRecommendationService creates a new RecommendationService.
func NewRecommendationService(repo repository.RecommendationRepo, aiModelClient client.AIModelClient) *RecommendationService {
	return &RecommendationService{repo: repo, aiModelClient: aiModelClient}
}

// GetRecommendedProducts gets recommended products for a user.
func (s *RecommendationService) GetRecommendedProducts(ctx context.Context, userID string, count int32) ([]*model.Product, error) {
	// Add any business logic here, e.g., user profiling, filtering
	return s.repo.GetRecommendedProducts(ctx, userID, count)
}

// IndexProductRelationship indexes a product relationship in Neo4j.
func (s *RecommendationService) IndexProductRelationship(ctx context.Context, rel *model.ProductRelationship) error {
	return s.repo.SaveProductRelationshipToNeo4j(ctx, rel)
}

// GetGraphRecommendedProducts gets recommended products based on graph data.
func (s *RecommendationService) GetGraphRecommendedProducts(ctx context.Context, productID string, count int32) ([]*model.Product, error) {
	return s.repo.GetRelatedProductsFromNeo4j(ctx, productID, count)
}

// GetAdvancedRecommendedProducts gets advanced personalized recommendations using an AI model.
func (s *RecommendationService) GetAdvancedRecommendedProducts(ctx context.Context, userID string, count int32, contextFeatures map[string]string) ([]*model.Product, string, error) {
	// 1. Get candidate items (e.g., from popular products, user's history, etc.)
	// For simplicity, let's use dummy product IDs for now.
	candidateItemIDs := []string{"prod_101", "prod_102", "prod_103", "prod_104", "prod_105"}

	// 2. Prepare user and item features (simplified)
	userFeatures := map[string]string{"user_segment": "premium", "device": "mobile"}
	itemFeatures := map[string]string{"category": "electronics", "brand": "xyz"}

	// 3. Call AI Model Service for ranking
	rankedItemIDs, scores, err := s.aiModelClient.RankRecommendations(ctx, userID, candidateItemIDs, userFeatures, itemFeatures)
	if err != nil {
		return nil, "", fmt.Errorf("failed to rank recommendations with AI model: %w", err)
	}

	// 4. Fetch product details for ranked items (simplified)
	// In a real system, you'd call the Product Service to get full product details.
	products := make([]*model.Product, 0, len(rankedItemIDs))
	for _, itemID := range rankedItemIDs {
		// Simulate fetching product details
		products = append(products, &model.Product{
			ID:          itemID,
			Name:        fmt.Sprintf("AI Recommended Product %s", itemID),
			Description: fmt.Sprintf("Highly personalized for user %s", userID),
			Price:       scores[itemID] * 100, // Dummy price based on score
			ImageURL:    fmt.Sprintf("http://example.com/ai_rec_%s.jpg", itemID),
		})
	}

	// 5. Generate explanation (simplified)
	explanation := "Based on your recent activity and similar users."

	return products, explanation, nil
}
