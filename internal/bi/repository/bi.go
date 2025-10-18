package repository

import (
	"context"
	"time"

	"ecommerce/internal/bi/model"
)

// BiRepo defines the interface for BI data access.
type BiRepo interface {
	GetSalesOverview(ctx context.Context, startDate, endDate *time.Time) (*model.SalesOverview, error)
	GetTopSellingProducts(ctx context.Context, limit uint32, startDate, endDate *time.Time) ([]*model.ProductSalesData, error)
}