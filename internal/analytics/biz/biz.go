package biz

import (
	"context"
	"time"
)

// ProductViewEvent represents a product view event in the business logic layer.
type ProductViewEvent struct {
	UserID    uint64
	ProductID uint64
	ViewTime  time.Time
}

// AnalyticsRepo defines the interface for analytics data access.
type AnalyticsRepo interface {
	RecordProductView(ctx context.Context, event *ProductViewEvent) error
}

// AnalyticsUsecase is the business logic for analytics.
type AnalyticsUsecase struct {
	repo AnalyticsRepo
}

// NewAnalyticsUsecase creates a new AnalyticsUsecase.
func NewAnalyticsUsecase(repo AnalyticsRepo) *AnalyticsUsecase {
	return &AnalyticsUsecase{repo: repo}
}

// RecordProductView records a product view event.
func (uc *AnalyticsUsecase) RecordProductView(ctx context.Context, userID, productID uint64) error {
	event := &ProductViewEvent{
		UserID:    userID,
		ProductID: productID,
		ViewTime:  time.Now(),
	}
	return uc.repo.RecordProductView(ctx, event)
}
