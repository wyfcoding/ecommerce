package repository

import (
	"context"
	"time"

	"ecommerce/internal/sales_forecasting/model"
)

// SalesForecastingRepo defines the interface for sales forecasting data access.
type SalesForecastingRepo interface {
	GetProductSalesForecast(ctx context.Context, productID uint64, forecastDate time.Time) (*model.ForecastResult, error)
	SaveProductSalesForecast(ctx context.Context, result *model.ForecastResult) (*model.ForecastResult, error)
	SimulateTrainSalesForecastModel(ctx context.Context, modelName, dataSource string, parameters map[string]string) (string, string, error)
}