package repository

import (
	"context"

	"ecommerce/internal/flashsale/model"
)

// FlashSaleRepo defines the data storage interface for flash sale data.
// The business layer depends on this interface, not on a concrete data implementation.
type FlashSaleRepo interface {
	CreateFlashSaleEvent(ctx context.Context, event *model.FlashSaleEvent) (*model.FlashSaleEvent, error)
	GetFlashSaleEvent(ctx context.Context, id uint) (*model.FlashSaleEvent, error)
	ListActiveFlashSaleEvents(ctx context.Context) ([]*model.FlashSaleEvent, int32, error)
	GetFlashSaleProduct(ctx context.Context, eventID, productID string) (*model.FlashSaleProduct, error)
	UpdateFlashSaleProductStock(ctx context.Context, product *model.FlashSaleProduct, quantity int32) error
	CompensateFlashSaleProductStock(ctx context.Context, productID uint, quantity int32) error // For Saga compensation
	// TODO: Add methods for user purchase history in flash sale to check max_per_user
}