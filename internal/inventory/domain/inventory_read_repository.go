package domain

import "context"

// InventoryReadRepository 定义库存读模型的高性能访问接口。
type InventoryReadRepository interface {
	// Save 保存或更新库存读模型。
	Save(ctx context.Context, inventory *Inventory) error
	// GetBySkuID 根据 SKU ID 获取读模型。
	GetBySkuID(ctx context.Context, skuID uint64) (*Inventory, error)
	// Delete 删除读模型数据。
	Delete(ctx context.Context, skuID uint64) error
}
