package biz

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrStockNotFound     = errors.New("stock not found")
)

// Stock represents the stock quantity of a SKU in the business logic layer.
type Stock struct {
	SKUID          uint64
	Quantity       uint32
	LockedQuantity uint32
}

// InventoryRepo defines the interface for inventory data access.
type InventoryRepo interface {
	GetStock(ctx context.Context, skuID uint64) (*Stock, error)
	DeductStock(ctx context.Context, skuID uint64, quantity uint32) error
	ReleaseStock(ctx context.Context, skuID uint64, quantity uint32) error
}

// InventoryUsecase is the business logic for inventory management.
type InventoryUsecase struct {
	repo InventoryRepo
}

// NewInventoryUsecase creates a new InventoryUsecase.
func NewInventoryUsecase(repo InventoryRepo) *InventoryUsecase {
	return &InventoryUsecase{repo: repo}
}

// DeductStock deducts stock for a SKU.
func (uc *InventoryUsecase) DeductStock(ctx context.Context, skuID uint64, quantity uint32) error {
	stock, err := uc.repo.GetStock(ctx, skuID)
	if err != nil {
		return err
	}
	if stock == nil {
		return ErrStockNotFound
	}
	if stock.Quantity < quantity {
		return fmt.Errorf("%w: SKU %d, available: %d, requested: %d", ErrInsufficientStock, skuID, stock.Quantity, quantity)
	}
	return uc.repo.DeductStock(ctx, skuID, quantity)
}

// ReleaseStock releases stock for a SKU.
func (uc *InventoryUsecase) ReleaseStock(ctx context.Context, skuID uint64, quantity uint32) error {
	return uc.repo.ReleaseStock(ctx, skuID, quantity)
}

// GetStock gets the current stock quantity for a SKU.
func (uc *InventoryUsecase) GetStock(ctx context.Context, skuID uint64) (*Stock, error) {
	stock, err := uc.repo.GetStock(ctx, skuID)
	if err != nil {
		return nil, err
	}
	if stock == nil {
		return nil, ErrStockNotFound
	}
	return stock, nil
}
