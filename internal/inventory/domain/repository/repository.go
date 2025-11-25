package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/inventory/domain/entity"
)

// InventoryRepository 库存仓储接口
type InventoryRepository interface {
	Save(ctx context.Context, inventory *entity.Inventory) error
	GetBySkuID(ctx context.Context, skuID uint64) (*entity.Inventory, error)
	GetBySkuIDs(ctx context.Context, skuIDs []uint64) ([]*entity.Inventory, error)
	List(ctx context.Context, offset, limit int) ([]*entity.Inventory, int64, error)
	GetLogs(ctx context.Context, inventoryID uint64, offset, limit int) ([]*entity.InventoryLog, int64, error)
}

// WarehouseRepository 仓库仓储接口
type WarehouseRepository interface {
	Save(ctx context.Context, warehouse *entity.Warehouse) error
	GetByID(ctx context.Context, id uint64) (*entity.Warehouse, error)
	ListAll(ctx context.Context) ([]*entity.Warehouse, error)
}
