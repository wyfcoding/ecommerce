package domain

import (
	"context"
	"time"
)

// InventorySearchRepository 定义库存搜索仓储接口。
type InventorySearchRepository interface {
	// Index 将库存写入搜索索引。
	Index(ctx context.Context, inventory *Inventory) error
	// Delete 从索引中删除库存。
	Delete(ctx context.Context, skuID uint64) error
	// Search 按条件检索库存（支持 SKU / 商品 / 仓库 / 状态过滤、分页）。
	Search(ctx context.Context, skuID, productID, warehouseID *uint64, status *InventoryStatus, offset, limit int, startTime, endTime *time.Time, sortBy string) ([]*Inventory, int64, error)
	// FindBySkuID 通过 SKU ID 检索库存。
	FindBySkuID(ctx context.Context, skuID uint64) (*Inventory, error)
}
