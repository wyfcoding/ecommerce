package data

import (
	"context"
	"errors"

	productV1 "ecommerce/ecommerce/api/product/v1"
	"ecommerce/ecommerce/app/cart/internal/biz"

	"google.golang.org/grpc"
)

type productGreeter struct {
	client productV1.ProductClient
}

// NewProductGreeter 创建一个商品服务的客户端
func NewProductGreeter(conn *grpc.ClientConn) biz.ProductGreeter {
	return &productGreeter{client: productV1.NewProductClient(conn)}
}

func (g *productGreeter) GetProductInfo(ctx context.Context, skuID uint64) (*biz.ProductInfo, error) {
	// 调用 product-service 的 GetSku 接口
	// 注意：这里我们假设 product.proto 已经增加了 GetSku 接口
	// req := &productV1.GetSkuRequest{SkuId: skuID}
	// res, err := g.client.GetSku(ctx, req)
	// if err != nil {
	// 	return nil, err
	// }
	//
	// return &biz.ProductInfo{
	// 	SkuID:    res.SkuId,
	// 	SpuID:    res.SpuId,
	// 	SpuTitle: "后端接口待实现", // 需从 Spu 信息中获取
	// 	SkuTitle: res.Title,
	// 	Image:    res.Image,
	// 	Price:    res.Price,
	// 	Stock:    uint(res.Stock),
	// }, nil

	// --- MOCK DATA START (因为后端 product-service 接口尚未实现) ---
	if skuID > 1000 && skuID < 2000 {
		return &biz.ProductInfo{
			SkuID:    skuID,
			SpuID:    101,
			SpuTitle: "超轻薄笔记本电脑",
			SkuTitle: "16GB+512GB 银色",
			Image:    "https://via.placeholder.com/100x100.png?text=Laptop",
			Price:    899900,
			Stock:    99,
		}, nil
	}
	return nil, errors.New("mock product not found")
	// --- MOCK DATA END ---
}
