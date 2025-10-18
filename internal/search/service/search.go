package service

import (
	"context"

	"ecommerce/internal/search/model"
	"ecommerce/internal/search/repository"
)

// SearchService is the business logic for search.
type SearchService struct {
	repo repository.SearchRepo
}

// NewSearchService creates a new SearchService.
func NewSearchService(repo repository.SearchRepo) *SearchService {
	return &SearchService{repo: repo}
}

// SearchProducts searches for products.
func (s *SearchService) SearchProducts(ctx context.Context, query string, pageSize, pageToken int32) ([]*model.Product, int32, error) {
	// Add any business logic here, e.g., query validation, result re-ranking
	return s.repo.SearchProducts(ctx, query, pageSize, pageToken)
}
