package data

import (
	"context"

	productV1 "ecommerce/ecommerce/api/product/v1"
	"ecommerce/ecommerce/app/admin/internal/biz"

	"google.golang.org/grpc"
)

type productGreeter struct {
	client productV1.ProductClient
}

// NewProductGreeter 创建一个 product-service 的客户端
func NewProductGreeter(conn *grpc.ClientConn) biz.ProductGreeter {
	return &productGreeter{client: productV1.NewProductClient(conn)}
}

func (g *productGreeter) CreateProduct(ctx context.Context, req *productV1.CreateProductRequest) (*productV1.CreateProductResponse, error) {
	// 真实调用 product-service
	return g.client.CreateProduct(ctx, req)
}

func (g *productGreeter) ListProducts(ctx context.Context, req *productV1.ListProductsRequest) (*productV1.ListProductsResponse, error) {
	// 真实调用 product-service
	return g.client.ListProducts(ctx, req)
}
