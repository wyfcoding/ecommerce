package data

import (
	"context"
	"ecommerce/internal/sales_forecasting/biz"
	"ecommerce/internal/sales_forecasting/data/model"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type salesForecastingRepo struct {
	data *Data
	// TODO: Add client for external ML platform (e.g., MLflow, SageMaker)
}

// NewSalesForecastingRepo creates a new SalesForecastingRepo.
func NewSalesForecastingRepo(data *Data) biz.SalesForecastingRepo {
	return &salesForecastingRepo{data: data}
}

// GetProductSalesForecast retrieves sales forecasts for a product from the database.
func (r *salesForecastingRepo) GetProductSalesForecast(ctx context.Context, productID uint64, forecastDate time.Time) (*biz.ForecastResult, error) {
	var po model.ForecastResult
	if err := r.data.db.WithContext(ctx).Where("product_id = ? AND forecast_date = ?", productID, forecastDate).First(&po).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Forecast not found
		}
		return nil, err
	}
	return &biz.ForecastResult{
		ID:                      po.ID,
		ProductID:               po.ProductID,
		ForecastDate:            po.ForecastDate,
		PredictedSalesQuantity:  po.PredictedSalesQuantity,
		ConfidenceIntervalLower: po.ConfidenceIntervalLower,
		ConfidenceIntervalUpper: po.ConfidenceIntervalUpper,
		ModelID:                 po.ModelID,
	}, nil
}

// SaveProductSalesForecast saves a sales forecast result to the database.
func (r *salesForecastingRepo) SaveProductSalesForecast(ctx context.Context, result *biz.ForecastResult) (*biz.ForecastResult, error) {
	po := &model.ForecastResult{
		ProductID:               result.ProductID,
		ForecastDate:            result.ForecastDate,
		PredictedSalesQuantity:  result.PredictedSalesQuantity,
		ConfidenceIntervalLower: result.ConfidenceIntervalLower,
		ConfidenceIntervalUpper: result.ConfidenceIntervalUpper,
		ModelID:                 result.ModelID,
	}
	if err := r.data.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	result.ID = po.ID
	return result, nil
}

// SimulateTrainSalesForecastModel simulates training a sales forecast model.
func (r *salesForecastingRepo) SimulateTrainSalesForecastModel(ctx context.Context, modelName, dataSource string, parameters map[string]string) (string, string, error) {
	// In a real system, this would trigger a job on an ML platform (e.g., Kubeflow, SageMaker).
	// For now, simulate a successful training.
	modelID := fmt.Sprintf("model_%s_%d", modelName, time.Now().UnixNano())
	return modelID, "COMPLETED", nil
}
