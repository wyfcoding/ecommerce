package biz

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrCartItemNotFound = errors.New("购物车商品不存在")
	ErrInvalidQuantity  = errors.New("无效的商品数量")
)

// CartUsecase 封装了购物车相关的业务逻辑。
type CartUsecase struct {
	repo CartRepo
	productClient ProductClient // 用于获取商品信息，如价格、库存等
}

// NewCartUsecase 是 CartUsecase 的构造函数。
func NewCartUsecase(repo CartRepo, productClient ProductClient) *CartUsecase {
	return &CartUsecase{
		repo: repo,
		productClient: productClient,
	}
}

// AddItem 添加商品到购物车。
func (uc *CartUsecase) AddItem(ctx context.Context, item *UsecaseCartItem) error {
	if item.Quantity <= 0 {
		return ErrInvalidQuantity
	}

	// 检查商品是否存在、库存是否充足、价格等
	skuInfos, err := uc.productClient.GetSkuInfos(ctx, []uint64{item.SkuID})
	if err != nil {
		return fmt.Errorf("获取商品信息失败: %w", err)
	}
	if len(skuInfos) == 0 {
		return errors.New("商品不存在")
	}
	targetSku := skuInfos[0]

	if targetSku.Status != 1 { // 假设 1 为正常状态
		return errors.New("商品已下架或状态异常")
	}

	if targetSku.Stock < item.Quantity {
		return errors.New("库存不足")
	}

	return uc.repo.AddItem(ctx, item.UserID, item.SkuID, item.Quantity)
}

// UpdateItem 更新购物车中商品的数量。
func (uc *CartUsecase) UpdateItem(ctx context.Context, item *UsecaseCartItem) error {
	if item.Quantity <= 0 {
		return ErrInvalidQuantity
	}

	// 检查商品是否存在、库存是否充足、价格等
	skuInfos, err := uc.productClient.GetSkuInfos(ctx, []uint64{item.SkuID})
	if err != nil {
		return fmt.Errorf("获取商品信息失败: %w", err)
	}
	if len(skuInfos) == 0 {
		return errors.New("商品不存在")
	}
	targetSku := skuInfos[0]

	if targetSku.Status != 1 { // 假设 1 为正常状态
		return errors.New("商品已下架或状态异常")
	}

	if targetSku.Stock < item.Quantity {
		return errors.New("库存不足")
	}

	return uc.repo.UpdateItem(ctx, item.UserID, item.SkuID, item.Quantity)
}

// DeleteItem 从购物车中删除商品。
func (uc *CartUsecase) DeleteItem(ctx context.Context, userID, productID, skuID uint64) error {
	return uc.repo.RemoveItem(ctx, userID, skuID)
}

// GetCart 获取用户购物车中的所有商品。
func (uc *CartUsecase) GetCart(ctx context.Context, userID uint64) ([]*UsecaseCartItem, error) {
	repoCartItems, err := uc.repo.GetCart(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("从仓库获取购物车商品失败: %w", err)
	}

	if len(repoCartItems) == 0 {
		return []*UsecaseCartItem{}, nil
	}

	skuIDs := make([]uint64, 0, len(repoCartItems))
	for _, item := range repoCartItems {
		skuIDs = append(skuIDs, item.SkuID)
	}

	skuInfos, err := uc.productClient.GetSkuInfos(ctx, skuIDs)
	if err != nil {
		return nil, fmt.Errorf("从商品服务获取SKU信息失败: %w", err)
	}

	skuInfoMap := make(map[uint64]*SkuInfo)
	for _, info := range skuInfos {
		skuInfoMap[info.SkuID] = info
	}

	usecaseCartItems := make([]*UsecaseCartItem, 0, len(repoCartItems))
	for _, item := range repoCartItems {
		skuInfo, ok := skuInfoMap[item.SkuID]
		if !ok {
			// 如果SKU信息不存在，可能商品已下架或删除，跳过此商品
			continue
		}
		usecaseCartItems = append(usecaseCartItems, &UsecaseCartItem{
			SkuID:    item.SkuID,
			Quantity: item.Quantity,
			Checked:  item.Checked,
			SkuInfo:  skuInfo,
		})
	}

	return usecaseCartItems, nil
}

// ClearCart 清空用户购物车。
func (uc *CartUsecase) ClearCart(ctx context.Context, userID uint64) error {
	return uc.repo.ClearCart(ctx, userID)
}