package repository

import (
	"context"

	"ecommerce/internal/recommendation/model"
)

// RecommendationRepo defines the interface for recommendation data access.
type RecommendationRepo interface {
	GetRecommendedProducts(ctx context.Context, userID string, count int32) ([]*model.Product, error)
	SaveProductRelationshipToNeo4j(ctx context.Context, rel *model.ProductRelationship) error
	GetRelatedProductsFromNeo4j(ctx context.Context, productID string, count int32) ([]*model.Product, error)
}