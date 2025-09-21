package data

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"ecommerce/internal/admin/biz"
	"ecommerce/internal/admin/data/model"
	productV1 "ecommerce/api/product/v1"

	"google.golang.org/grpc"
	"gorm.io/gorm"
)

// productClient 实现了 biz.ProductClient 接口。
type productClient struct {
	client productV1.ProductServiceClient
}

// NewProductClient 是 productClient 的构造函数。
func NewProductClient(conn *grpc.ClientConn) biz.ProductClient {
	return &productClient{client: productV1.NewProductServiceClient(conn)}
}

// protoToBizSpu 将 productV1.SpuInfo 转换为 biz.Spu。
func (pc *productClient) protoToBizSpu(spu *productV1.SpuInfo) *biz.Spu {
	if spu == nil {
		return nil
	}
	return &biz.Spu{
		ID:            spu.SpuId,
		CategoryID:    spu.CategoryId,
		BrandID:       spu.BrandId,
		Title:         spu.Title,
		SubTitle:      spu.SubTitle,
		MainImage:     spu.MainImage,
		GalleryImages: spu.GalleryImages,
		DetailHTML:    spu.DetailHtml,
		Status:        spu.Status,
	}
}

// bizSpuToProto 将 biz.Spu 转换为 productV1.SpuInfo。
func (pc *productClient) bizSpuToProto(spu *biz.Spu) *productV1.SpuInfo {
	if spu == nil {
		return nil
	}
	res := &productV1.SpuInfo{
		SpuId:         spu.ID,
		CategoryId:    spu.CategoryID,
		BrandId:       spu.BrandID,
		Title:         spu.Title,
		SubTitle:      spu.SubTitle,
		MainImage:     spu.MainImage,
		GalleryImages: spu.GalleryImages,
		DetailHtml:    spu.DetailHTML,
		Status:        spu.Status,
	}
	return res
}

// protoToBizSku 将 productV1.SkuInfo 转换为 biz.Sku。
func (pc *productClient) protoToBizSku(sku *productV1.SkuInfo) *biz.Sku {
	if sku == nil {
		return nil
	}
	return &biz.Sku{
		ID:            sku.SkuId,
		SpuID:         sku.SpuId,
		Title:         sku.Title,
		Price:         sku.Price,
		OriginalPrice: sku.OriginalPrice,
		Stock:         sku.Stock,
		Image:         sku.Image,
		Specs:         sku.Specs,
		Status:        sku.Status,
	}
}

// bizSkuToProto 将 biz.Sku 转换为 productV1.SkuInfo。
func (pc *productClient) bizSkuToProto(sku *biz.Sku) *productV1.SkuInfo {
	if sku == nil {
		return nil
	}
	res := &productV1.SkuInfo{
		SkuId:         sku.ID,
		SpuId:         sku.SpuID,
		Title:         sku.Title,
		Price:         sku.Price,
		OriginalPrice: sku.OriginalPrice,
		Stock:         sku.Stock,
		Image:         sku.Image,
		Specs:         sku.Specs,
		Status:        sku.Status,
	}
	return res
}

// protoToBizSkuList 将 productV1.SkuInfo 列表转换为 biz.Sku 列表。
func (pc *productClient) protoToBizSkuList(skus []*productV1.SkuInfo) []*biz.Sku {
	if skus == nil {
		return nil
	}
	bizSkus := make([]*biz.Sku, 0, len(skus))
	for _, skuProto := range skus {
		bizSkus = append(bizSkus, pc.protoToBizSku(skuProto))
	}
	return bizSkus
}

// CreateProduct 调用商品服务创建商品。
func (pc *productClient) CreateProduct(ctx context.Context, spu *biz.Spu, skus []*biz.Sku) (*biz.Spu, []*biz.Sku, error) {
	protoSpu := pc.bizSpuToProto(spu)
	protoSkus := make([]*productV1.SkuInfo, 0, len(skus))
	for _, sku := range skus {
		protoSkus = append(protoSkus, pc.bizSkuToProto(sku))
	}

	res, err := pc.client.CreateProduct(ctx, &productV1.CreateProductRequest{
		Spu:  protoSpu,
		Skus: protoSkus,
	})
	if err != nil {
		return nil, nil, err
	}

	return pc.protoToBizSpu(res.Spu), pc.protoToBizSkuList(res.Skus), nil
}

// UpdateProduct 调用商品服务更新商品。
func (pc *productClient) UpdateProduct(ctx context.Context, spu *biz.Spu, skus []*biz.Sku) (*biz.Spu, []*biz.Sku, error) {
	protoSpu := pc.bizSpuToProto(spu)
	protoSkus := make([]*productV1.SkuInfo, 0, len(skus))
	for _, sku := range skus {
		protoSkus = append(protoSkus, pc.bizSkuToProto(sku))
	}

	res, err := pc.client.UpdateProduct(ctx, &productV1.UpdateProductRequest{
		Spu:  protoSpu,
		Skus: protoSkus,
	})
	if err != nil {
		return nil, nil, err
	}

	// 假设 UpdateProduct 返回更新后的完整 SPU 和 SKU 信息
	updatedSpu := pc.protoToBizSpu(res.Spu)
	updatedSkus := make([]*biz.Sku, 0, len(res.Skus))
	for _, skuProto := range res.Skus {
		updatedSkus = append(updatedSkus, pc.protoToBizSku(skuProto))
	}

	return updatedSpu, updatedSkus, nil
}

// GetSpuDetail 调用商品服务获取商品详情。
func (pc *productClient) GetSpuDetail(ctx context.Context, spuID uint64) (*biz.Spu, []*biz.Sku, error) {
	res, err := pc.client.GetSpuDetail(ctx, &productV1.GetSpuDetailRequest{SpuId: spuID})
	if err != nil {
		return nil, nil, err
	}

	bizSpu := pc.protoToBizSpu(res.Spu)
	bizSkus := make([]*biz.Sku, 0, len(res.Skus))
	for _, skuProto := range res.Skus {
		bizSkus = append(bizSkus, pc.protoToBizSku(skuProto))
	}

	return bizSpu, bizSkus, nil
}
