package data

import (
	"context"
	"ecommerce/internal/inventory/biz"
	"ecommerce/internal/inventory/data/model"

	"gorm.io/gorm"
)

type inventoryRepo struct {
	data *Data
}

// NewInventoryRepo creates a new InventoryRepo.
func NewInventoryRepo(data *Data) biz.InventoryRepo {
	return &inventoryRepo{data: data}
}

// GetStock retrieves the current stock quantity for a SKU.
func (r *inventoryRepo) GetStock(ctx context.Context, skuID uint64) (*biz.Stock, error) {
	var stock model.Stock
	// Simplified: assuming one warehouse for now, or aggregate across warehouses
	if err := r.data.db.WithContext(ctx).Where("sku_id = ?", skuID).First(&stock).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Stock not found
		}
		return nil, err
	}
	return &biz.Stock{
		SKUID:          stock.SKUID,
		Quantity:       stock.Quantity,
		LockedQuantity: stock.LockedQuantity,
	}, nil
}

// DeductStock deducts stock quantity for a SKU.
func (r *inventoryRepo) DeductStock(ctx context.Context, skuID uint64, quantity uint32) error {
	// This is a simplified deduction. In a real system, this would involve:
	// - Transaction management
	// - Handling locked quantities (pre-allocated stock)
	// - Multi-warehouse logic
	// - Optimistic locking or other concurrency control
	return r.data.db.WithContext(ctx).Model(&model.Stock{}).Where("sku_id = ? AND quantity >= ?", skuID, quantity).
		UpdateColumn("quantity", gorm.Expr("quantity - ?", quantity)).Error
}

// ReleaseStock releases previously deducted or locked stock quantity for a SKU.
func (r *inventoryRepo) ReleaseStock(ctx context.Context, skuID uint64, quantity uint32) error {
	// This is a simplified release.
	return r.data.db.WithContext(ctx).Model(&model.Stock{}).Where("sku_id = ?", skuID).
		UpdateColumn("quantity", gorm.Expr("quantity + ?", quantity)).Error
}
