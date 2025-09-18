package service

import (
	"context"

	v1 "ecommerce/ecommerce/api/product/v1"
	"ecommerce/ecommerce/app/product/internal/biz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProductService 是 gRPC 服务的实现
type ProductService struct {
	v1.UnimplementedProductServer // 必须嵌入

	uc *biz.ProductUsecase
}

// NewProductService 创建一个新的 ProductService
func NewProductService(uc *biz.ProductUsecase) *ProductService {
	return &ProductService{uc: uc}
}

// ListCategories 实现了获取商品分类列表的接口
func (s *ProductService) ListCategories(ctx context.Context, req *v1.ListCategoriesRequest) (*v1.ListCategoriesResponse, error) {
	// 1. 调用 biz 层
	categories, err := s.uc.ListCategories(ctx, req.ParentId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// 2. 将 biz.Category 转换为 v1.CategoryInfo
	res := make([]*v1.CategoryInfo, 0, len(categories))
	for _, c := range categories {
		res = append(res, &v1.CategoryInfo{
			Id:        c.ID,
			ParentId:  c.ParentID,
			Name:      c.Name,
			Level:     uint32(c.Level),
			Icon:      c.Icon,
			SortOrder: uint32(c.SortOrder),
			IsVisible: c.IsVisible,
		})
	}

	return &v1.ListCategoriesResponse{
		Categories: res,
	}, nil
}

// GetSpuDetail 实现了获取SPU商品详情的接口
func (s *ProductService) GetSpuDetail(ctx context.Context, req *v1.GetSpuDetailRequest) (*v1.GetSpuDetailResponse, error) {
	// 1. 调用 biz 层
	spu, skus, err := s.uc.GetSpuDetail(ctx, req.SpuId)
	if err != nil {
		// 这里可以根据 biz 层返回的具体错误类型，转换为不同的 gRPC 状态码
		// 例如，如果是 "spu not found"，可以返回 codes.NotFound
		return nil, status.Error(codes.Internal, err.Error())
	}

	// 2. 将 biz.Spu 转换为 v1.SpuInfo
	spuInfo := &v1.SpuInfo{
		SpuId:         spu.SpuID,
		CategoryId:    spu.CategoryID,
		BrandId:       spu.BrandID,
		Title:         spu.Title,
		SubTitle:      spu.SubTitle,
		MainImage:     spu.MainImage,
		GalleryImages: spu.GalleryImages,
		DetailHtml:    spu.DetailHTML,
		Status:        int32(spu.Status),
	}

	// 3. 将 biz.Sku 列表转换为 v1.SkuInfo 列表
	skuInfos := make([]*v1.SkuInfo, 0, len(skus))
	for _, s := range skus {
		skuInfos = append(skuInfos, &v1.SkuInfo{
			SkuId:         s.SkuID,
			SpuId:         s.SpuID,
			Title:         s.Title,
			Price:         s.Price,
			OriginalPrice: s.OriginalPrice,
			Stock:         uint32(s.Stock),
			Image:         s.Image,
			Specs:         s.Specs,
			Status:        int32(s.Status),
		})
	}

	return &v1.GetSpuDetailResponse{
		Spu:  spuInfo,
		Skus: skuInfos,
	}, nil
}
