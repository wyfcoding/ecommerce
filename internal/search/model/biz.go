package biz

import (
	"context"
)

// Product represents a product in the business logic layer.
type Product struct {
	ID          string
	Name        string
	Description string
	Price       float64
	ImageURL    string
}

// SearchRepo defines the interface for search data access.
type SearchRepo interface {
	SearchProducts(ctx context.Context, query string, pageSize, pageToken int32) ([]*Product, int32, error)
}

// SearchUsecase is the business logic for search.
type SearchUsecase struct {
	repo SearchRepo
}

// NewSearchUsecase creates a new SearchUsecase.
func NewSearchUsecase(repo SearchRepo) *SearchUsecase {
	return &SearchUsecase{repo: repo}
}

// SearchProducts searches for products.
func (uc *SearchUsecase) SearchProducts(ctx context.Context, query string, pageSize, pageToken int32) ([]*Product, int32, error) {
	// Add any business logic here, e.g., query validation, result re-ranking
	return uc.repo.SearchProducts(ctx, query, pageSize, pageToken)
}
