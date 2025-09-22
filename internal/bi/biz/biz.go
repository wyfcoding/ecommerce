package biz

import (
	"context"
	"time"
)

// SalesOverview represents aggregated sales data for BI reports in the business logic layer.
type SalesOverview struct {
	TotalSalesAmount uint64
	TotalOrders      uint32
	TotalUsers       uint32
	ConversionRate   float64
}

// ProductSalesData represents sales data for a specific product in the business logic layer.
type ProductSalesData struct {
	ProductID    uint64
	ProductName  string
	SalesQuantity uint32
	SalesAmount  uint64
}

// BiRepo defines the interface for BI data access.
type BiRepo interface {
	GetSalesOverview(ctx context.Context, startDate, endDate *time.Time) (*SalesOverview, error)
	GetTopSellingProducts(ctx context.Context, limit uint32, startDate, endDate *time.Time) ([]*ProductSalesData, error)
}

// BiUsecase is the business logic for BI reports.
type BiUsecase struct {
	repo BiRepo
}

// NewBiUsecase creates a new BiUsecase.
func NewBiUsecase(repo BiRepo) *BiUsecase {
	return &BiUsecase{repo: repo}
}

// GetSalesOverview retrieves sales overview data.
func (uc *BiUsecase) GetSalesOverview(ctx context.Context, startDate, endDate *time.Time) (*SalesOverview, error) {
	// Add any business logic here, e.g., date validation
	return uc.repo.GetSalesOverview(ctx, startDate, endDate)
}

// GetTopSellingProducts retrieves top selling products data.
func (uc *BiUsecase) GetTopSellingProducts(ctx context.Context, limit uint32, startDate, endDate *time.Time) ([]*ProductSalesData, error) {
	// Add any business logic here
	return uc.repo.GetTopSellingProducts(ctx, limit, startDate, endDate)
}
