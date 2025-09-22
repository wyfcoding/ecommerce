package service

import (
	"context"
	v1 "ecommerce/api/product/v1"
	"ecommerce/internal/product/biz"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// --- SPU Converters ---

func protoToBizSpu(spu *v1.SpuInfo) *biz.Spu {
	if spu == nil {
		return nil
	}
	return &biz.Spu{
		ID:            spu.SpuId, // For update operations, SpuId from proto is the ID
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

func bizSpuToProto(spu *biz.Spu) *v1.SpuInfo {
	if spu == nil {
		return nil
	}
	res := &v1.SpuInfo{
		SpuId:         spu.ID, // Map internal ID to external SpuId
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

// --- SKU Converters ---

func protoToBizSku(sku *v1.SkuInfo) *biz.Sku {
	if sku == nil {
		return nil
	}
	return &biz.Sku{
		ID:            sku.SkuId, // For update operations, SkuId from proto is the ID
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

func bizSkuToProto(sku *biz.Sku) *v1.SkuInfo {
	if sku == nil {
		return nil
	}
	res := &v1.SkuInfo{
		SkuId:         sku.ID, // Map internal ID to external SkuId
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

// --- RPC Implementations ---

// CreateProduct 实现了创建商品的 RPC。
func (s *service) CreateProduct(ctx context.Context, req *v1.CreateProductRequest) (*v1.CreateProductResponse, error) {
	if req.Spu == nil {
		return nil, status.Error(codes.InvalidArgument, "SPU information is required")
	}

	bizSpu := protoToBizSpu(req.Spu)
	var bizSkus []*biz.Sku
	for _, skuProto := range req.Skus {
		bizSkus = append(bizSkus, protoToBizSku(skuProto))
	}

	createdSpu, createdSkus, err := s.productUsecase.CreateProduct(ctx, bizSpu, bizSkus)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create product: %v", err)
	}

	var skuIDs []uint64
	for _, sku := range createdSkus {
		skuIDs = append(skuIDs, sku.ID) // Use internal ID for response
	}

	return &v1.CreateProductResponse{
		SpuId:  createdSpu.ID, // Use internal ID for response
		SkuIds: skuIDs,
	}, nil
}

// UpdateProduct 实现了更新商品的 RPC。
func (s *service) UpdateProduct(ctx context.Context, req *v1.UpdateProductRequest) (*v1.GetSpuDetailResponse, error) {
	if req.Spu == nil || req.Spu.SpuId == 0 {
		return nil, status.Error(codes.InvalidArgument, "SPU with spu_id is required")
	}

	bizSpu := protoToBizSpu(req.Spu)
	var bizSkus []*biz.Sku
	for _, skuProto := range req.Skus {
		bizSkus = append(bizSkus, protoToBizSku(skuProto))
	}

	updatedSpu, updatedSkus, err := s.productUsecase.UpdateProduct(ctx, bizSpu, bizSkus)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update product: %v", err)
	}

	var skuProtos []*v1.SkuInfo
	for _, sku := range updatedSkus {
		skuProtos = append(skuProtos, bizSkuToProto(sku))
	}

	return &v1.GetSpuDetailResponse{
		Spu:  bizSpuToProto(updatedSpu),
		Skus: skuProtos,
	}, nil
}

// DeleteProduct 实现了删除商品的 RPC。
func (s *service) DeleteProduct(ctx context.Context, req *v1.DeleteProductRequest) (*emptypb.Empty, error) {
	if req.SpuId == 0 {
		return nil, status.Error(codes.InvalidArgument, "spu_id is required")
	}
	if err := s.productUsecase.DeleteProduct(ctx, req.SpuId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete product: %v", err)
	}
	return &emptypb.Empty{}, nil
}


// GetSpuDetail 实现了获取商品详情的 RPC。
func (s *service) GetSpuDetail(ctx context.Context, req *v1.GetSpuDetailRequest) (*v1.GetSpuDetailResponse, error) {
	spu, skus, err := s.productUsecase.GetProductDetails(ctx, req.SpuId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get product detail: %v", err)
	}

	var skuProtos []*v1.SkuInfo
	for _, sku := range skus {
		skuProtos = append(skuProtos, bizSkuToProto(sku))
	}

	return &v1.GetSpuDetailResponse{
		Spu:  bizSpuToProto(spu),
		Skus: skuProtos,
	}, nil
}

// ListProducts 实现了获取产品列表的 RPC。
func (s *service) ListProducts(ctx context.Context, req *v1.ListProductsRequest) (*v1.ListProductsResponse, error) {
	// 调用 biz 层获取产品列表
	products, total, err := s.productUsecase.ListProducts(
		ctx,
		req.PageSize,
		req.PageNum,
		req.GetCategoryId(),
		req.GetStatus(),
		req.GetBrandId(),
		req.GetMinPrice(),
		req.GetMaxPrice(),
		req.GetQuery(),
		req.GetSortBy(),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list products: %v", err)
	}

	var productInfos []*v1.SpuInfo
	for _, p := range products {
		productInfos = append(productInfos, bizSpuToProto(p))
	}

	return &v1.ListProductsResponse{Products: productInfos, TotalCount: total}, nil
}

// IndexProduct 实现了产品索引的 RPC。
func (s *service) IndexProduct(ctx context.Context, req *v1.IndexProductRequest) (*v1.IndexProductResponse, error) {
	if req.SpuId == 0 {
		return nil, status.Error(codes.InvalidArgument, "spu_id is required")
	}

	err := s.productUsecase.IndexProduct(ctx, req.SpuId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to index product: %v", err)
	}

	return &v1.IndexProductResponse{}, nil
}
// CreateBrand 实现创建品牌接口
func (s *service) CreateBrand(ctx context.Context, req *v1.CreateBrandRequest) (*v1.BrandInfo, error) {
	// 调用 biz 层创建品牌
	brand, err := s.brandUsecase.CreateBrand(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create brand: %v", err)
	}
	return brand, nil
}

// UpdateBrand 实现更新品牌接口
func (s *service) UpdateBrand(ctx context.Context, req *v1.UpdateBrandRequest) (*v1.BrandInfo, error) {
	// 调用 biz 层更新品牌
	brand, err := s.brandUsecase.UpdateBrand(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update brand: %v", err)
	}
	return brand, nil
}

// DeleteBrand 实现删除品牌接口
func (s *service) DeleteBrand(ctx context.Context, req *v1.DeleteBrandRequest) (*emptypb.Empty, error) {
	// 调用 biz 层删除品牌
	err := s.brandUsecase.DeleteBrand(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete brand: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// ListBrands 实现获取品牌列表接口
func (s *service) ListBrands(ctx context.Context, req *v1.ListBrandsRequest) (*v1.ListBrandsResponse, error) {
	// 调用 biz 层获取品牌列表
	brands, total, err := s.brandUsecase.ListBrands(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list brands: %v", err)
	}
	return &v1.ListBrandsResponse{Brands: brands, TotalCount: total}, nil
}

// CreateReview 实现创建评论接口
func (s *service) CreateReview(ctx context.Context, req *v1.CreateReviewRequest) (*v1.ReviewInfo, error) {
	// 调用 biz 层创建评论
	review, err := s.reviewUsecase.CreateReview(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create review: %v", err)
	}
	return review, nil
}

// ListReviews 实现获取评论列表接口
func (s *service) ListReviews(ctx context.Context, req *v1.ListReviewsRequest) (*v1.ListReviewsResponse, error) {
	// 调用 biz 层获取评论列表
	reviews, total, err := s.reviewUsecase.ListReviews(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list reviews: %v", err)
	}
	return &v1.ListReviewsResponse{Reviews: reviews, TotalCount: total}, nil
}

// DeleteReview 实现删除评论接口
func (s *service) DeleteReview(ctx context.Context, req *v1.DeleteReviewRequest) (*emptypb.Empty, error) {
	// 调用 biz 层删除评论
	err := s.reviewUsecase.DeleteReview(ctx, req.Id, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete review: %v", err)
	}
	return &emptypb.Empty{}, nil
}
