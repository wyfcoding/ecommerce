package biz

import (
	"context"
	// 导入 product.v1 的 proto 包，以便复用其消息体
	productV1 "ecommerce/ecommerce/api/product/v1"
)

// ProductGreeter 定义了与 product-service 通信的接口
type ProductGreeter interface {
	// 这里的请求和响应可以直接使用 productV1 的消息体，简化 BFF 的模型转换
	CreateProduct(ctx context.Context, req *productV1.CreateProductRequest) (*productV1.CreateProductResponse, error)
	ListProducts(ctx context.Context, req *productV1.ListProductsRequest) (*productV1.ListProductsResponse, error)
}

// ProductUsecase 封装了商品管理的业务逻辑
type ProductUsecase struct {
	greeter ProductGreeter
}

// NewProductUsecase 创建一个新的 ProductUsecase
func NewProductUsecase(greeter ProductGreeter) *ProductUsecase {
	return &ProductUsecase{greeter: greeter}
}

// CreateProduct 调用下游服务创建商品
func (uc *ProductUsecase) CreateProduct(ctx context.Context, req *productV1.CreateProductRequest) (*productV1.CreateProductResponse, error) {
	// BFF 层的业务逻辑可以很简单，直接透传
	// 也可以很复杂，比如在这里聚合其他服务的数据
	return uc.greeter.CreateProduct(ctx, req)
}

// ListProducts 调用下游服务获取商品列表
func (uc *ProductUsecase) ListProducts(ctx context.Context, req *productV1.ListProductsRequest) (*productV1.ListProductsResponse, error) {
	return uc.greeter.ListProducts(ctx, req)
}
