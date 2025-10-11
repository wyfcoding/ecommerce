package data

import (
	"time"

	"gorm.io/gorm"
)

// ForecastResult represents a sales forecast result stored in the database.
type ForecastResult struct {
	gorm.Model
	ProductID               uint64    `gorm:"index;not null;comment:商品ID" json:"productId"`
	ForecastDate            time.Time `gorm:"index;not null;comment:预测日期" json:"forecastDate"`
	PredictedSalesQuantity  float64   `gorm:"not null;comment:预测销量" json:"predictedSalesQuantity"`
	ConfidenceIntervalLower float64   `gorm:"comment:置信区间下限" json:"confidenceIntervalLower"`
	ConfidenceIntervalUpper float64   `gorm:"comment:置信区间上限" json:"confidenceIntervalUpper"`
	ModelID                 string    `gorm:"size:255;comment:使用的模型ID" json:"modelId"`
	// Add other fields like region, forecast version, etc.
}

// TableName specifies the table name for ForecastResult.
func (ForecastResult) TableName() string {
	return "sales_forecast_results"
}
