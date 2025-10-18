package client

import (
	"context"

	productv1 "ecommerce/api/product/v1"
	"ecommerce/internal/cart/model"
	"google.golang.org/grpc"
)

// ProductClient defines the interface for the cart service to interact with the product service.
type ProductClient interface {
	GetSkuInfos(ctx context.Context, skuIDs []uint64) ([]*model.SkuInfo, error)
}

type productClient struct {
	client productv1.ProductClient
}

func NewProductClient(conn *grpc.ClientConn) ProductClient {
	return &productClient{
		client: productv1.NewProductClient(conn),
	}
}

func (pc *productClient) GetSkuInfos(ctx context.Context, skuIDs []uint64) ([]*model.SkuInfo, error) {
	req := &productv1.GetSkuInfosRequest{
		SkuIds: skuIDs,
	}
	res, err := pc.client.GetSkuInfos(ctx, req)
	if err != nil {
		return nil, err
	}

	skuInfos := make([]*model.SkuInfo, len(res.GetSkuInfos()))
	for i, sku := range res.GetSkuInfos() {
		skuInfos[i] = &model.SkuInfo{
			SkuID:  sku.GetSkuId(),
			SpuID:  sku.GetSpuId(),
			Title:  sku.GetTitle(),
			Price:  sku.GetPrice(),
			Image:  sku.GetImage(),
			Specs:  sku.GetSpecs(),
			Status: sku.GetStatus(),
		}
	}
	return skuInfos, nil
}
