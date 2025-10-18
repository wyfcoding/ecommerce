package model

import "time"

// ProductRecommendation represents a smart product selection recommendation in the business logic layer.
type ProductRecommendation struct {
	ID              uint
	MerchantID      string
	ProductID       uint64
	ProductName     string
	Score           float64
	Reason          string
	ContextFeatures map[string]string
	RecommendedAt   time.Time
}
