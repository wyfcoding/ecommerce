package service

import (
	"context"
	"errors"

	"ecommerce/internal/marketing/model"
	"ecommerce/internal/marketing/repository"

	"go.uber.org/zap"
)

type PromotionService struct {
	repo repository.PromotionRepo
	Log  *zap.SugaredLogger
}

// NewPromotionService creates a new PromotionService instance
func NewPromotionService(repo repository.PromotionRepo, logger *zap.SugaredLogger) *PromotionService {
	return &PromotionService{repo: repo, Log: logger}
}

// CreatePromotion implements the business logic for creating a promotion
func (s *PromotionService) CreatePromotion(ctx context.Context, promotion *model.Promotion) (*model.Promotion, error) {
	// Business rule validation
	if promotion.Name == "" {
		return nil, errors.New("promotion name cannot be empty")
	}
	if promotion.StartTime == nil || promotion.EndTime == nil || promotion.EndTime.Before(*promotion.StartTime) {
		return nil, errors.New("invalid promotion time range")
	}
	// ... other rule validation

	return s.repo.CreatePromotion(ctx, promotion)
}

// UpdatePromotion implements the business logic for updating a promotion
func (s *PromotionService) UpdatePromotion(ctx context.Context, promotion *model.Promotion) (*model.Promotion, error) {
	// Business rule validation
	if promotion.ID == 0 {
		return nil, errors.New("ID is required to update a promotion")
	}
	// ... other rule validation

	return s.repo.UpdatePromotion(ctx, promotion)
}

// DeletePromotion implements the business logic for deleting a promotion
func (s *PromotionService) DeletePromotion(ctx context.Context, id uint64) error {
	return s.repo.DeletePromotion(ctx, id)
}

// GetPromotion implements the business logic for getting promotion details
func (s *PromotionService) GetPromotion(ctx context.Context, id uint64) (*model.Promotion, error) {
	return s.repo.GetPromotion(ctx, id)
}

// ListPromotions implements the business logic for getting a list of promotions
func (s *PromotionService) ListPromotions(ctx context.Context, pageSize, pageNum uint32, name *string, promoType *uint32, status *uint32) ([]*model.Promotion, uint64, error) {
	return s.repo.ListPromotions(ctx, pageSize, pageNum, name, promoType, status)
}
