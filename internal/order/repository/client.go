package data

import (
	"context"

	cartV1 "ecommerce/api/cart/v1"
	productV1 "ecommerce/api/product/v1"
	"ecommerce/internal/order/biz"

	"google.golang.org/grpc"
)

// productClient 实现了 biz.ProductClient 接口。
type productClient struct {
	client productV1.ProductServiceClient
}

func NewProductClient(conn *grpc.ClientConn) biz.ProductClient {
	return &productClient{client: productV1.NewProductServiceClient(conn)}
}

func (pc *productClient) GetSkuInfos(ctx context.Context, skuIDs []uint64) (map[uint64]*biz.SkuInfo, error) {
	res, err := pc.client.GetSkuInfos(ctx, &productV1.GetSkuInfosRequest{SkuIds: skuIDs})
	if err != nil {
		return nil, err
	}

	resultMap := make(map[uint64]*biz.SkuInfo, len(res.Skus))
	for _, sku := range res.Skus {
		resultMap[sku.SkuId] = &biz.SkuInfo{
			SkuID: sku.SkuId,
			SpuID: sku.SpuId,
			Price: sku.Price,
			Stock: sku.Stock,
			Title: sku.Title,
			Image: sku.Image,
		}
	}
	return resultMap, nil
}

func (pc *productClient) LockStock(ctx context.Context, items map[uint64]uint32) error {
	reqItems := make([]*productV1.LockStockItem, 0, len(items))
	for id, qty := range items {
		reqItems = append(reqItems, &productV1.LockStockItem{SkuId: id, Quantity: qty})
	}
	_, err := pc.client.LockStock(ctx, &productV1.LockStockRequest{Items: reqItems})
	return err
}

func (pc *productClient) UnlockStock(ctx context.Context, items map[uint64]uint32) error {
	reqItems := make([]*productV1.LockStockItem, 0, len(items))
	for id, qty := range items {
		reqItems = append(reqItems, &productV1.LockStockItem{SkuId: id, Quantity: qty})
	}
	_, err := pc.client.UnlockStock(ctx, &productV1.LockStockRequest{Items: reqItems})
	return err
}

// cartClient 实现了 biz.CartClient 接口。
type cartClient struct {
	client cartV1.CartClient
}

func NewCartClient(conn *grpc.ClientConn) biz.CartClient {
	return &cartClient{client: cartV1.NewCartClient(conn)}
}

func (cc *cartClient) ClearCheckedItems(ctx context.Context, userID uint64) error {
	// 购物车服务需要提供一个清空已勾选商品的 gRPC 接口
	_, err := cc.client.ClearCheckedItems(ctx, &cartV1.ClearCheckedItemsRequest{UserId: userID})
	return err
}
