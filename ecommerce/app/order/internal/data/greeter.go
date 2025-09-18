package data

import (
	"context"

	cartV1 "ecommerce/ecommerce/api/cart/v1"
	productV1 "ecommerce/ecommerce/api/product/v1"
	"ecommerce/ecommerce/app/order/internal/biz"
)

// productGreeter and cartGreeter implementations...

func (g *productGreeter) GetSkuInfos(ctx context.Context, skuIDs []uint64) (map[uint64]*biz.SkuInfo, error) {
	// 真实调用 product-service
	res, err := g.client.GetSkuInfos(ctx, &productV1.GetSkuInfosRequest{SkuIds: skuIDs})
	if err != nil {
		return nil, err
	}

	resultMap := make(map[uint64]*biz.SkuInfo)
	for _, sku := range res.Skus {
		resultMap[sku.SkuId] = &biz.SkuInfo{
			// ... 模型转换 ...
		}
	}
	return resultMap, nil
}

func (g *productGreeter) LockStock(ctx context.Context, items []*biz.OrderItem) error {
	// 真实调用 product-service
	reqItems := make([]*productV1.LockStockItem, len(items))
	for i, item := range items {
		reqItems[i] = &productV1.LockStockItem{SkuId: item.SkuID, Quantity: item.Quantity}
	}
	_, err := g.client.LockStock(ctx, &productV1.LockStockRequest{Items: reqItems})
	return err
}

func (c *cartGreeter) ClearCartItems(ctx context.Context, userID uint64, skuIDs []uint64) error {
	// 真实调用 cart-service
	_, err := c.client.RemoveItem(ctx, &cartV1.RemoveItemRequest{UserId: userID, SkuIds: skuIDs})
	return err
}
