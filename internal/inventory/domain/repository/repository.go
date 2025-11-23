package repository

import (
	"context"
	"ecommerce/internal/inventory/domain/entity"
)

// InventoryRepository 库存仓储接口
type InventoryRepository interface {
	Save(ctx context.Context, inventory *entity.Inventory) error
	GetBySkuID(ctx context.Context, skuID uint64) (*entity.Inventory, error)
	List(ctx context.Context, offset, limit int) ([]*entity.Inventory, int64, error)
	GetLogs(ctx context.Context, inventoryID uint64, offset, limit int) ([]*entity.InventoryLog, int64, error)
}
