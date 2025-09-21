package biz

import "context"

// SkuInfo 是购物车服务内部使用的商品SKU信息模型。
// 它是一个数据传输对象(DTO)，用于解耦购物车服务和商品服务。
// 即使商品服务的内部模型`product.Sku`发生变化，购物车服务也不受影响。
type SkuInfo struct {
	SkuID         uint64
	SpuID         uint64
	Title         string
	Price         uint64
	Image         string
	Specs         map[string]string
	Status        int32
}

type CartItem struct {
	SkuID    uint64
	Quantity uint32
	Checked  bool
}

// UsecaseCartItem 是购物车的业务领域模型。
type UsecaseCartItem struct {
	SkuID    uint64
	Quantity uint32
	Checked  bool // 增加勾选状态
	SkuInfo  *SkuInfo
}

// CartRepo 定义了购物车数据仓库需要实现的接口 (CURD)。
type CartRepo interface {
	AddItem(ctx context.Context, userID, skuID uint64, quantity uint32) error
	GetCart(ctx context.Context, userID uint64) ([]*CartItem, error)
	UpdateItem(ctx context.Context, userID, skuID uint64, quantity uint32) error
	RemoveItem(ctx context.Context, userID uint64, skuIDs ...uint64) error // 支持批量删除
	UpdateCheckStatus(ctx context.Context, userID uint64, skuIDs []uint64, checked bool) error
	GetCheckStatus(ctx context.Context, userID uint64) (map[uint64]bool, error)
	GetCartItemCount(ctx context.Context, userID uint64) (uint32, error)
	ClearCart(ctx context.Context, userID uint64) error
}

// ProductClient 定义了购物车服务依赖的商品服务客户端接口。
type ProductClient interface {
	// GetSkuInfos 批量获取 SKU 的详细信息。
	GetSkuInfos(ctx context.Context, skuIDs []uint64) ([]*SkuInfo, error)
}
