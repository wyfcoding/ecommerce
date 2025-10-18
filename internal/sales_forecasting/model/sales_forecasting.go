package model

import "time"

// ForecastResult represents a sales forecast result in the business logic layer.
type ForecastResult struct {
	ID                      uint
	ProductID               uint64
	ForecastDate            time.Time
	PredictedSalesQuantity  float64
	ConfidenceIntervalLower float64
	ConfidenceIntervalUpper float64
	ModelID                 string
}
