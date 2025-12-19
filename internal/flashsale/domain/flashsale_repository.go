package domain

import (
	"context"
)

// FlashSaleRepository 是秒杀模块的仓储接口。
// 它定义了对秒杀活动和秒杀订单实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type FlashSaleRepository interface {
	// --- 活动管理 (Flashsale methods) ---

	// SaveFlashsale 将秒杀活动实体保存到数据存储中。
	// 如果活动已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// flashsale: 待保存的秒杀活动实体。
	SaveFlashsale(ctx context.Context, flashsale *Flashsale) error
	// GetFlashsale 根据ID获取秒杀活动实体。
	GetFlashsale(ctx context.Context, id uint64) (*Flashsale, error)
	// ListFlashsales 列出所有秒杀活动实体，支持通过状态过滤和分页。
	ListFlashsales(ctx context.Context, status *FlashsaleStatus, offset, limit int) ([]*Flashsale, int64, error)
	// UpdateStock 更新秒杀活动的商品库存。
	UpdateStock(ctx context.Context, id uint64, quantity int32) error

	// --- 订单管理 (FlashsaleOrder methods) ---

	// SaveOrder 将秒杀订单实体保存到数据存储中。
	SaveOrder(ctx context.Context, order *FlashsaleOrder) error
	// GetOrder 根据ID获取秒杀订单实体。
	GetOrder(ctx context.Context, id uint64) (*FlashsaleOrder, error)
	// GetUserOrders 获取指定用户和秒杀活动ID的所有订单实体。
	GetUserOrders(ctx context.Context, userID, flashsaleID uint64) ([]*FlashsaleOrder, error)
	// CountUserBought 统计指定用户在某个秒杀活动中已购买的数量。
	CountUserBought(ctx context.Context, userID, flashsaleID uint64) (int32, error)
}

// FlashSaleCache 是秒杀模块的缓存接口。
// 它定义了在缓存中（通常是Redis）管理秒杀库存和用户限购的契约，以应对高并发场景。
type FlashSaleCache interface {
	// SetStock 在缓存中设置指定秒杀活动的库存。
	// ctx: 上下文。
	// flashsaleID: 秒杀活动ID。
	// stock: 要设置的库存数量。
	SetStock(ctx context.Context, flashsaleID uint64, stock int32) error
	// DeductStock 在缓存中原子性地扣减指定秒杀活动的库存，并检查用户限购。
	// ctx: 上下文。
	// flashsaleID: 秒杀活动ID。
	// userID: 用户ID。
	// quantity: 待扣减的数量。
	// limitPerUser: 每用户限购数量。
	// 返回一个布尔值，表示扣减是否成功（例如，库存充足且未超限购），以及可能发生的错误。
	DeductStock(ctx context.Context, flashsaleID, userID uint64, quantity, limitPerUser int32) (bool, error)
	// RevertStock 在缓存中原子性地回滚已扣减的库存。
	// ctx: 上下文。
	// flashsaleID: 秒杀活动ID。
	// userID: 用户ID。
	// quantity: 待回滚的数量。
	RevertStock(ctx context.Context, flashsaleID, userID uint64, quantity int32) error
}
