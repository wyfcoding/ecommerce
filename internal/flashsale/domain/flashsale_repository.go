package domain

import "context"

// FlashSaleRepository 是秒杀模块的写模型仓储接口。
// 它定义了对秒杀活动和秒杀订单实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type FlashSaleRepository interface {
	// --- tx helpers ---
	BeginTx(ctx context.Context) any
	CommitTx(tx any) error
	RollbackTx(tx any) error
	WithTx(ctx context.Context, fn func(tx any) error) error

	// --- 活动管理 (Flashsale methods) ---
	SaveFlashsale(ctx context.Context, flashsale *Flashsale) error
	SaveFlashsaleInTx(ctx context.Context, tx any, flashsale *Flashsale) error
	GetFlashsale(ctx context.Context, id uint64) (*Flashsale, error)
	ListFlashsales(ctx context.Context, query *FlashsaleQuery) ([]*Flashsale, int64, error)
	UpdateStock(ctx context.Context, id uint64, quantity int32) error

	// --- 订单管理 (FlashsaleOrder methods) ---
	SaveOrder(ctx context.Context, order *FlashsaleOrder) error
	SaveOrderInTx(ctx context.Context, tx any, order *FlashsaleOrder) error
	GetOrder(ctx context.Context, id uint64) (*FlashsaleOrder, error)
	GetUserOrders(ctx context.Context, userID, flashsaleID uint64) ([]*FlashsaleOrder, error)
	ListOrders(ctx context.Context, query *FlashsaleOrderQuery) ([]*FlashsaleOrder, int64, error)
	CountUserBought(ctx context.Context, userID, flashsaleID uint64) (int32, error)
}

// FlashsaleQuery 秒杀活动查询条件。
type FlashsaleQuery struct {
	Status    *FlashsaleStatus
	ProductID uint64
	Page      int
	PageSize  int
}

// FlashsaleOrderQuery 秒杀订单查询条件。
type FlashsaleOrderQuery struct {
	FlashsaleID uint64
	UserID      uint64
	Status      *FlashsaleOrderStatus
	Page        int
	PageSize    int
}

// FlashSaleCache 是秒杀模块的缓存接口。
// 它定义了在缓存中（通常是Redis）管理秒杀库存和用户限购的契约，以应对高并发场景。
type FlashSaleCache interface {
	// SetStock 在缓存中设置指定秒杀活动的库存。
	SetStock(ctx context.Context, flashsaleID uint64, stock int32) error
	// DeductStock 在缓存中原子性地扣减指定秒杀活动的库存，并检查用户限购。
	DeductStock(ctx context.Context, flashsaleID, userID uint64, quantity, limitPerUser int32) (bool, error)
	// RevertStock 在缓存中原子性地回滚已扣减的库存。
	RevertStock(ctx context.Context, flashsaleID, userID uint64, quantity int32) error
}
