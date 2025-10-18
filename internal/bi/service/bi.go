package service

import (
	"context"
	"time"

	"ecommerce/internal/bi/model"
	"ecommerce/internal/bi/repository"
)

// BiService is the business logic for BI reports.
type BiService struct {
	repo repository.BiRepo
}

// NewBiService creates a new BiService.
func NewBiService(repo repository.BiRepo) *BiService {
	return &BiService{repo: repo}
}

// GetSalesOverview retrieves sales overview data.
func (s *BiService) GetSalesOverview(ctx context.Context, startDate, endDate *time.Time) (*model.SalesOverview, error) {
	// Add any business logic here, e.g., date validation
	return s.repo.GetSalesOverview(ctx, startDate, endDate)
}

// GetTopSellingProducts retrieves top selling products data.
func (s *BiService) GetTopSellingProducts(ctx context.Context, limit uint32, startDate, endDate *time.Time) ([]*model.ProductSalesData, error) {
	// Add any business logic here
	return s.repo.GetTopSellingProducts(ctx, limit, startDate, endDate)
}
