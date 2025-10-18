package repository

import (
	"context"

	"ecommerce/internal/marketing/model"
)

type PromotionRepo interface {
	CreatePromotion(ctx context.Context, promotion *model.Promotion) (*model.Promotion, error)
	UpdatePromotion(ctx context.Context, promotion *model.Promotion) (*model.Promotion, error)
	DeletePromotion(ctx context.Context, id uint64) error
	GetPromotion(ctx context.Context, id uint64) (*model.Promotion, error)
	ListPromotions(ctx context.Context, pageSize, pageNum uint32, name *string, promoType *uint32, status *uint32) ([]*model.Promotion, uint64, error)
}
