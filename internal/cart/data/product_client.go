package data

import (
	"context"
	"ecommerce/internal/cart/biz"
	productV1 "ecommerce/api/product/v1" // Import product service proto
	"google.golang.org/grpc"
)

// productClient 实现了 biz.ProductClient 接口。
type productClient struct {
	client productV1.ProductServiceClient
}

// NewProductClient 是 productClient 的构造函数。
func NewProductClient(conn *grpc.ClientConn) biz.ProductClient {
	return &productClient{
		client: productV1.NewProductServiceClient(conn),
	}
}

// GetSkuInfos 通过 gRPC 调用商品服务，批量获取 SKU 的详细信息。
func (pc *productClient) GetSkuInfos(ctx context.Context, skuIDs []uint64) ([]*biz.SkuInfo, error) {
	// 1. 调用 gRPC 接口
	res, err := pc.client.GetSkuInfos(ctx, &productV1.GetSkuInfosRequest{SkuIds: skuIDs})
	if err != nil {
		return nil, err
	}

	// 2. 将商品服务的 API 模型转换为购物车服务的 biz 模型
	bizSkus := make([]*biz.SkuInfo, 0, len(res.Skus))
	for _, s := range res.Skus {
		bizSkus = append(bizSkus, &biz.SkuInfo{
			SkuID:   s.SkuId,
			SpuID:   s.SpuId,
			Title:   s.Title,
			Price:   s.Price,
			Image:   s.Image,
			Specs:   s.Specs,
			Status:  s.Status,
		})
	}
	return bizSkus, nil
}
