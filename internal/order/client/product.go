package client

import (
	"context"
	"fmt"

	productv1 "ecommerce/api/product/v1"
	"ecommerce/internal/order/model"
	"google.golang.org/grpc"
)

// ProductClient 定义了与商品服务交互的接口
type ProductClient interface {
	GetProductSKU(ctx context.Context, skuID uint64) (*productv1.SkuInfo, error)
	LockStock(ctx context.Context, skuID uint64, quantity uint32) error
	UnlockStock(ctx context.Context, skuID uint64, quantity uint32) error
}

type productClient struct {
	client productv1.ProductClient
}

func NewProductClient(conn *grpc.ClientConn) ProductClient {
	return &productClient{
		client: productv1.NewProductClient(conn),
	}
}

func (c *productClient) GetProductSKU(ctx context.Context, skuID uint64) (*productv1.SkuInfo, error) {
	req := &productv1.GetSkuInfoRequest{SkuId: skuID}
	res, err := c.client.GetSkuInfo(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get product SKU: %w", err)
	}
	return res.GetSkuInfo(), nil
}

func (c *productClient) LockStock(ctx context.Context, skuID uint64, quantity uint32) error {
	req := &productv1.LockStockRequest{SkuId: skuID, Quantity: quantity}
	_, err := c.client.LockStock(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to lock stock: %w", err)
	}
	return nil
}

func (c *productClient) UnlockStock(ctx context.Context, skuID uint64, quantity uint32) error {
	req := &productv1.UnlockStockRequest{SkuId: skuID, Quantity: quantity}
	_, err := c.client.UnlockStock(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to unlock stock: %w", err)
	}
	return nil
}
