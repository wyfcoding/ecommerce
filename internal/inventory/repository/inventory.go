package repository

import (
	"context"

	"ecommerce/internal/inventory/model"
)

// InventoryRepo defines the interface for inventory data access.
type InventoryRepo interface {
	GetStock(ctx context.Context, skuID uint64) (*model.Stock, error)
	DeductStock(ctx context.Context, skuID uint64, quantity uint32) error
	ReleaseStock(ctx context.Context, skuID uint64, quantity uint32) error
}