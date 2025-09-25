package biz

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrCartItemNotFound = errors.New("cart item not found")
	ErrInvalidQuantity  = errors.New("invalid quantity")
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
// 检查商品是否存在、库存是否充足、价格等
	skuInfos, err := uc.productClient.GetSkuInfos(ctx, []uint64{item.SkuID})
	if err != nil {
		return fmt.Errorf("failed to get product information: %w", err)
	}
	if len(skuInfos) == 0 {
		return errors.New("product not found")
	}
	targetSku := skuInfos[0]

	if targetSku.Status != 1 { // Assuming 1 is the normal status.
		return errors.New("product is not available")
	}

	if targetSku.Stock < item.Quantity {
		return errors.New("insufficient stock")
	}

	return uc.repo.AddItem(ctx, item.UserID, item.SkuID, item.Quantity)
}

// UpdateItem updates the quantity of an item in the shopping cart.
func (uc *CartUsecase) UpdateItem(ctx context.Context, item *UsecaseCartItem) error {
	if item.Quantity <= 0 {
		return ErrInvalidQuantity
	}

	// Check if the product exists, has enough stock, and get its price.
	skuInfos, err := uc.productClient.GetSkuInfos(ctx, []uint64{item.SkuID})
	if err != nil {
		return fmt.Errorf("failed to get product information: %w", err)
	}
	if len(skuInfos) == 0 {
		return errors.New("product not found")
	}
	targetSku := skuInfos[0]

	if targetSku.Status != 1 { // Assuming 1 is the normal status.
		return errors.New("product is not available")
	}

	if targetSku.Stock < item.Quantity {
		return errors.New("insufficient stock")
	}

	return uc.repo.UpdateItem(ctx, item.UserID, item.SkuID, item.Quantity)
}

// DeleteItem removes an item from the shopping cart.
func (uc *CartUsecase) DeleteItem(ctx context.Context, userID, productID, skuID uint64) error {
	return uc.repo.RemoveItem(ctx, userID, skuID)
}

// GetCart gets all items in a user's shopping cart.
func (uc *CartUsecase) GetCart(ctx context.Context, userID uint64) ([]*UsecaseCartItem, error) {
	repoCartItems, err := uc.repo.GetCart(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart items from repository: %w", err)
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
		return nil, fmt.Errorf("failed to get SKU info from product service: %w", err)
	}

	skuInfoMap := make(map[uint64]*SkuInfo)
	for _, info := range skuInfos {
		skuInfoMap[info.SkuID] = info
	}

	usecaseCartItems := make([]*UsecaseCartItem, 0, len(repoCartItems))
	for _, item := range repoCartItems {
		skuInfo, ok := skuInfoMap[item.SkuID]
		if !ok {
			// If SKU info does not exist, the product may have been removed or is unavailable. Skip this item.
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

// ClearCart clears a user's shopping cart.
func (uc *CartUsecase) ClearCart(ctx context.Context, userID uint64) error {
	return uc.repo.ClearCart(ctx, userID)
}