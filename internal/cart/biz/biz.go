package biz

import (
	"context"
	"fmt"
)

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

// CartUsecase 是购物车的业务用例。
type CartUsecase struct {
	repo          CartRepo
	productClient ProductClient
}

// NewCartUsecase 创建一个新的 CartUsecase。
func NewCartUsecase(repo CartRepo, productClient ProductClient) *CartUsecase {
	return &CartUsecase{
		repo:          repo,
		productClient: productClient,
	}
}

// AddItem 向购物车中添加商品。
func (uc *CartUsecase) AddItem(ctx context.Context, userID, skuID uint64, quantity uint32, checked bool) error {
	// 1. 验证 SKU 是否存在 (可以批量验证，这里简化为单个)
	skuInfos, err := uc.productClient.GetSkuInfos(ctx, []uint64{skuID})
	if err != nil {
		return err
	}
	if len(skuInfos) == 0 || skuInfos[0].SkuID != skuID {
		return fmt.Errorf("SKU %d not found", skuID)
	}

	// 2. 添加商品到购物车
	return uc.repo.AddItem(ctx, userID, skuID, quantity)
}
