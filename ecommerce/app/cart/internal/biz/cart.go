package biz

import (
	"context"
	"errors"
	"time"
)

// --- Domain Models ---

// CartItem 是购物车的领域模型
type CartItem struct {
	SkuID    uint64
	Quantity uint32
	Checked  bool
	AddedAt  time.Time
	// 冗余的商品信息
	SpuTitle string
	SkuTitle string
	SkuImage string
	Price    uint64 // 商品当前价格
}

// ProductInfo 是商品信息的领域模型（用于服务间通信）
type ProductInfo struct {
	SkuID    uint64
	SpuID    uint64
	SpuTitle string
	SkuTitle string
	Image    string
	Price    uint64
	Stock    uint
}

// --- Repo & Greeter Interfaces ---

// CartRepo 定义了购物车数据仓库的接口
type CartRepo interface {
	GetItem(ctx context.Context, userID, skuID uint64) (*CartItem, error)
	SaveItem(ctx context.Context, userID uint64, item *CartItem) error
}

// ProductGreeter 定义了与商品服务通信的接口
type ProductGreeter interface {
	GetProductInfo(ctx context.Context, skuID uint64) (*ProductInfo, error)
}

// --- Usecase ---

// CartUsecase 是购物车的业务逻辑容器
type CartUsecase struct {
	repo    CartRepo
	product ProductGreeter
}

// NewCartUsecase 创建一个新的 CartUsecase
func NewCartUsecase(repo CartRepo, product ProductGreeter) *CartUsecase {
	return &CartUsecase{repo: repo, product: product}
}

// AddItem 实现了添加商品到购物车的业务逻辑
func (uc *CartUsecase) AddItem(ctx context.Context, userID, skuID uint64, quantity uint32) error {
	// 1. [跨服务调用] 从商品服务获取商品信息
	productInfo, err := uc.product.GetProductInfo(ctx, skuID)
	if err != nil {
		return err // 如果商品不存在或商品服务出错，直接返回错误
	}

	// 2. 检查库存 (简单检查)
	if uint32(productInfo.Stock) < quantity {
		// 实际项目中可能有更复杂的库存检查逻辑
		return errors.New("insufficient stock")
	}

	// 3. 从 repo 获取当前购物车中是否已有该商品
	item, err := uc.repo.GetItem(ctx, userID, skuID)
	if err != nil && err.Error() != "item not found" {
		return err // 其他类型的查询错误
	}

	if item != nil {
		// 购物车中已存在该商品，累加数量
		item.Quantity += quantity
	} else {
		// 购物车中不存在该商品，创建新条目
		item = &CartItem{
			SkuID:    skuID,
			Quantity: quantity,
			Checked:  true, // 默认添加到购物车即为选中状态
			AddedAt:  time.Now(),
			SpuTitle: productInfo.SpuTitle,
			SkuTitle: productInfo.SkuTitle,
			SkuImage: productInfo.Image,
			Price:    productInfo.Price,
		}
	}

	// 4. 将更新后的商品信息保存回 repo
	return uc.repo.SaveItem(ctx, userID, item)
}
