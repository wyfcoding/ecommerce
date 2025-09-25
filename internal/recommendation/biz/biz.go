package biz

import (
	"context"
	"fmt"
)

// Product represents a product in the business logic layer.
type Product struct {
	ID          string
	Name        string
	Description string
	Price       float64
	ImageURL    string
}

// ProductRelationship represents a relationship between two products in Neo4j.
type ProductRelationship struct {
	ProductID1 string
	ProductID2 string
	Type       string
	Weight     float64
}

// RecommendationRepo defines the interface for recommendation data access.
type RecommendationRepo interface {
	GetRecommendedProducts(ctx context.Context, userID string, count int32) ([]*Product, error)
	SaveProductRelationshipToNeo4j(ctx context.Context, rel *ProductRelationship) error
	GetRelatedProductsFromNeo4j(ctx context.Context, productID string, count int32) ([]*Product, error)
}

// AIModelClient defines the interface to interact with the AI Model Service.
type AIModelClient interface {
	RankRecommendations(ctx context.Context, userID string, itemIDs []string, userFeatures, itemFeatures map[string]string) ([]string, map[string]float64, error)
}

// RecommendationUsecase is the business logic for recommendation.
type RecommendationUsecase struct {
	repo RecommendationRepo
	aiModelClient AIModelClient // Added AI Model client
}

// NewRecommendationUsecase creates a new RecommendationUsecase.
func NewRecommendationUsecase(repo RecommendationRepo, aiModelClient AIModelClient) *RecommendationUsecase {
	return &RecommendationUsecase{repo: repo, aiModelClient: aiModelClient}
}

// GetRecommendedProducts gets recommended products for a user.
func (uc *RecommendationUsecase) GetRecommendedProducts(ctx context.Context, userID string, count int32) ([]*Product, error) {
	// Add any business logic here, e.g., user profiling, filtering
	return uc.repo.GetRecommendedProducts(ctx, userID, count)
}

// IndexProductRelationship indexes a product relationship in Neo4j.
func (uc *RecommendationUsecase) IndexProductRelationship(ctx context.Context, rel *ProductRelationship) error {
	return uc.repo.SaveProductRelationshipToNeo4j(ctx, rel)
}

// GetGraphRecommendedProducts gets recommended products based on graph data.
func (uc *RecommendationUsecase) GetGraphRecommendedProducts(ctx context.Context, productID string, count int32) ([]*Product, error) {
	return uc.repo.GetRelatedProductsFromNeo4j(ctx, productID, count)
}

// GetAdvancedRecommendedProducts gets advanced personalized recommendations using an AI model.
func (uc *RecommendationUsecase) GetAdvancedRecommendedProducts(ctx context.Context, userID string, count int32, contextFeatures map[string]string) ([]*Product, string, error) {
	// 1. Get candidate items (e.g., from popular products, user's history, etc.)
	// For simplicity, let's use dummy product IDs for now.
	candidateItemIDs := []string{"prod_101", "prod_102", "prod_103", "prod_104", "prod_105"}

	// 2. Prepare user and item features (simplified)
	userFeatures := map[string]string{"user_segment": "premium", "device": "mobile"}
	itemFeatures := map[string]string{"category": "electronics", "brand": "xyz"}

	// 3. Call AI Model Service for ranking
	rankedItemIDs, scores, err := uc.aiModelClient.RankRecommendations(ctx, userID, candidateItemIDs, userFeatures, itemFeatures)
	if err != nil {
		return nil, "", fmt.Errorf("failed to rank recommendations with AI model: %w", err)
	}

	// 4. Fetch product details for ranked items (simplified)
	// In a real system, you'd call the Product Service to get full product details.
	products := make([]*Product, 0, len(rankedItemIDs))
	for _, itemID := range rankedItemIDs {
		// Simulate fetching product details
		products = append(products, &Product{
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
