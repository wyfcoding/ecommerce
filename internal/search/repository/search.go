package repository

import (
	"context"

	"ecommerce/internal/search/model"
)

// SearchRepo defines the interface for search data access.
type SearchRepo interface {
	SearchProducts(ctx context.Context, query string, pageSize, pageToken int32) ([]*model.Product, int32, error)
}