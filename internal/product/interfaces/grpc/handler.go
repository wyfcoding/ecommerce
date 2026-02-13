package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/product/v1"
	"github.com/wyfcoding/ecommerce/internal/product/application"
	"github.com/wyfcoding/ecommerce/internal/product/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedProductServiceServer
	cmdService   *application.ProductCommandService
	queryService *application.ProductQueryService
}

func NewServer(cmd *application.ProductCommandService, query *application.ProductQueryService) *Server {
	return &Server{cmdService: cmd, queryService: query}
}

// --- Product ---

func (s *Server) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.ProductInfo, error) {
	start := time.Now()
	slog.Info("gRPC CreateProduct received", "name", req.Name, "category_id", req.CategoryId)

	var pType domain.ProductType
	switch req.Type {
	case pb.ProductType_PHYSICAL:
		pType = domain.ProductTypePhysical
	case pb.ProductType_DIGITAL:
		pType = domain.ProductTypeDigital
	case pb.ProductType_ASSET:
		pType = domain.ProductTypeAsset
	default:
		pType = domain.ProductTypePhysical // Default
	}

	createReq := &application.CreateProductCommand{
		Name:        req.Name,
		Description: req.Description,
		CategoryID:  uint(req.CategoryId),
		BrandID:     uint(req.BrandId),
		Type:        pType,
		Price:       req.Price,
		Stock:       req.Stock,
	}

	product, err := s.cmdService.CreateProduct(ctx, createReq)
	if err != nil {
		slog.Error("gRPC CreateProduct failed", "name", req.Name, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create product: %v", err))
	}

	slog.Info("gRPC CreateProduct successful", "product_id", product.ID, "duration", time.Since(start))
	return convertProductToProto(product), nil
}

func (s *Server) GetProductByID(ctx context.Context, req *pb.GetProductByIDRequest) (*pb.ProductInfo, error) {
	product, err := s.queryService.GetProductByID(ctx, req.Id)
	if err != nil {
		slog.Error("gRPC GetProductByID failed", "id", req.Id, "error", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get product: %v", err))
	}
	if product == nil {
		return nil, status.Error(codes.NotFound, "product not found")
	}
	return convertProductToProto(product), nil
}

func (s *Server) UpdateProductInfo(ctx context.Context, req *pb.UpdateProductInfoRequest) (*pb.ProductInfo, error) {
	start := time.Now()
	slog.Info("gRPC UpdateProductInfo received", "id", req.Id)

	var name *string
	if req.Name != nil {
		v := req.Name.Value
		name = &v
	}
	var desc *string
	if req.Description != nil {
		v := req.Description.Value
		desc = &v
	}
	var categoryID *uint
	if req.CategoryId != nil {
		v := uint(req.CategoryId.Value)
		categoryID = &v
	}
	var brandID *uint
	if req.BrandId != nil {
		v := uint(req.BrandId.Value)
		brandID = &v
	}
	var statusVal *domain.ProductStatus
	if req.Status != pb.ProductStatus_PRODUCT_STATUS_UNSPECIFIED {
		st := domain.ProductStatus(req.Status)
		statusVal = &st
	}

	updateReq := &application.UpdateProductCommand{
		ID:          uint(req.Id),
		Name:        name,
		Description: desc,
		CategoryID:  categoryID,
		BrandID:     brandID,
		Status:      statusVal,
	}

	product, err := s.cmdService.UpdateProduct(ctx, updateReq)
	if err != nil {
		slog.Error("gRPC UpdateProductInfo failed", "id", req.Id, "error", err, "duration", time.Since(start))
		return nil, err
	}
	slog.Info("gRPC UpdateProductInfo successful", "id", req.Id, "duration", time.Since(start))
	return convertProductToProto(product), nil
}

func (s *Server) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*emptypb.Empty, error) {
	if err := s.cmdService.DeleteProduct(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete product: %v", err))
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	products, total, err := s.queryService.ListProducts(ctx, int(req.Page), int(req.PageSize), req.CategoryId, req.BrandId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list products: %v", err))
	}

	pbProducts := make([]*pb.ProductInfo, len(products))
	for i, p := range products {
		pbProducts[i] = convertProductToProto(p)
	}

	return &pb.ListProductsResponse{
		Products: pbProducts,
		Total:    int32(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// --- SKU ---

func (s *Server) AddSKUsToProduct(ctx context.Context, req *pb.AddSKUsToProductRequest) (*pb.AddSKUsToProductResponse, error) {
	var createdSKUs []*pb.SKU
	for _, skuReq := range req.Skus {
		specs := make(map[string]string)
		for _, sv := range skuReq.SpecValues {
			specs[sv.Key] = sv.Value
		}

		addReq := &application.AddSKUCommand{
			ProductID: uint(req.ProductId),
			Name:      skuReq.Name,
			Price:     skuReq.Price,
			Stock:     skuReq.StockQuantity,
			Image:     skuReq.ImageUrl,
			Specs:     specs,
		}

		sku, err := s.cmdService.AddSKU(ctx, addReq)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add SKU: %v", err))
		}
		createdSKUs = append(createdSKUs, convertSKUToProto(sku))
	}
	return &pb.AddSKUsToProductResponse{CreatedSkus: createdSKUs}, nil
}

func (s *Server) UpdateSKU(ctx context.Context, req *pb.UpdateSKURequest) (*pb.SKU, error) {
	var price *int64
	if req.Price != nil {
		v := req.Price.Value
		price = &v
	}
	var stock *int32
	if req.StockQuantity != nil {
		v := req.StockQuantity.Value
		stock = &v
	}
	var image *string
	if req.ImageUrl != nil {
		v := req.ImageUrl.Value
		image = &v
	}

	updateReq := &application.UpdateSKUCommand{
		ID:    uint(req.Id),
		Price: price,
		Stock: stock,
		Image: image,
	}

	sku, err := s.cmdService.UpdateSKU(ctx, updateReq)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update SKU: %v", err))
	}
	return convertSKUToProto(sku), nil
}

func (s *Server) DeleteSKU(ctx context.Context, req *pb.DeleteSKURequest) (*emptypb.Empty, error) {
	for _, id := range req.SkuIds {
		if err := s.cmdService.DeleteSKU(ctx, id); err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete SKU %d: %v", id, err))
		}
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) GetSKUByID(ctx context.Context, req *pb.GetSKUByIDRequest) (*pb.SKU, error) {
	sku, err := s.queryService.GetSKUByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if sku == nil {
		return nil, status.Error(codes.NotFound, "SKU not found")
	}
	return convertSKUToProto(sku), nil
}

// --- Category ---

func (s *Server) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.Category, error) {
	cmd := &application.CreateCategoryCommand{
		Name:     req.Name,
		ParentID: uint(req.ParentId),
	}
	category, err := s.cmdService.CreateCategory(ctx, cmd)
	if err != nil {
		return nil, err
	}
	return convertCategoryToProto(category), nil
}

func (s *Server) GetCategoryByID(ctx context.Context, req *pb.GetCategoryByIDRequest) (*pb.Category, error) {
	category, err := s.queryService.GetCategoryByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, status.Error(codes.NotFound, "category not found")
	}
	return convertCategoryToProto(category), nil
}

func (s *Server) ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	categories, err := s.queryService.ListCategories(ctx, req.ParentId)
	if err != nil {
		return nil, err
	}
	var pbCategories []*pb.Category
	for _, c := range categories {
		pbCategories = append(pbCategories, convertCategoryToProto(c))
	}
	return &pb.ListCategoriesResponse{Categories: pbCategories}, nil
}

// Note: Update/Delete Category/Brand implementations omitted for brevity but should map to cmdService methods.
// I will implement Brand Create/Get to ensure main paths work.

func (s *Server) CreateBrand(ctx context.Context, req *pb.CreateBrandRequest) (*pb.Brand, error) {
	cmd := &application.CreateBrandCommand{Name: req.Name, Logo: req.LogoUrl}
	brand, err := s.cmdService.CreateBrand(ctx, cmd)
	if err != nil {
		return nil, err
	}
	return convertBrandToProto(brand), nil
}

func (s *Server) GetBrandByID(ctx context.Context, req *pb.GetBrandByIDRequest) (*pb.Brand, error) {
	brand, err := s.queryService.GetBrandByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if brand == nil {
		return nil, status.Error(codes.NotFound, "brand not found")
	}
	return convertBrandToProto(brand), nil
}

func (s *Server) ListBrands(ctx context.Context, req *pb.ListBrandsRequest) (*pb.ListBrandsResponse, error) {
	brands, err := s.queryService.ListBrands(ctx)
	if err != nil {
		return nil, err
	}
	var pbBrands []*pb.Brand
	for _, b := range brands {
		pbBrands = append(pbBrands, convertBrandToProto(b))
	}
	return &pb.ListBrandsResponse{Brands: pbBrands, Total: int32(len(brands))}, nil
}

// --- Helpers ---
// (Same as before)
func convertProductToProto(p *domain.Product) *pb.ProductInfo {
	if p == nil {
		return nil
	}
	pbSKUs := make([]*pb.SKU, len(p.SKUs))
	for i, sku := range p.SKUs {
		pbSKUs[i] = convertSKUToProto(sku)
	}

	return &pb.ProductInfo{
		Id:               uint64(p.ID),
		Name:             p.Name,
		Description:      p.Description,
		Category:         &pb.Category{Id: uint64(p.CategoryID)},
		Brand:            &pb.Brand{Id: uint64(p.BrandID)},
		Status:           pb.ProductStatus(p.Status),
		Skus:             pbSKUs,
		MainImageUrl:     p.MainImage,
		GalleryImageUrls: p.Images,
		CreatedAt:        timestamppb.New(p.CreatedAt),
		UpdatedAt:        timestamppb.New(p.UpdatedAt),
	}
}

func convertSKUToProto(s *domain.SKU) *pb.SKU {
	if s == nil {
		return nil
	}
	var specValues []*pb.SpecValue
	for k, v := range s.Specs {
		specValues = append(specValues, &pb.SpecValue{Key: k, Value: v})
	}

	return &pb.SKU{
		Id:            uint64(s.ID),
		ProductId:     uint64(s.ProductID),
		Name:          s.Name,
		Price:         s.Price,
		StockQuantity: s.Stock,
		ImageUrl:      s.Image,
		SpecValues:    specValues,
		CreatedAt:     timestamppb.New(s.CreatedAt),
		UpdatedAt:     timestamppb.New(s.UpdatedAt),
	}
}

func convertCategoryToProto(c *domain.Category) *pb.Category {
	if c == nil {
		return nil
	}
	return &pb.Category{
		Id:        uint64(c.ID),
		Name:      c.Name,
		ParentId:  uint64(c.ParentID),
		SortOrder: int32(c.Sort),
	}
}

func convertBrandToProto(b *domain.Brand) *pb.Brand {
	if b == nil {
		return nil
	}
	return &pb.Brand{
		Id:      uint64(b.ID),
		Name:    b.Name,
		LogoUrl: b.Logo,
	}
}

// Stub missing Update/Delete Category/Brand to avoid compilation errors if Proto defines them
func (s *Server) UpdateCategory(ctx context.Context, req *pb.UpdateCategoryRequest) (*pb.Category, error) {
	start := time.Now()
	slog.Info("gRPC UpdateCategory received", "id", req.Id)

	var name *string
	if req.Name != nil {
		v := req.Name.Value
		name = &v
	}
	var parentID *uint
	if req.ParentId != nil {
		v := uint(req.ParentId.Value)
		parentID = &v
	}
	var sort *int
	if req.SortOrder != nil {
		v := int(req.SortOrder.Value)
		sort = &v
	}

	cmd := &application.UpdateCategoryCommand{
		ID:       uint(req.Id),
		Name:     name,
		ParentID: parentID,
		Sort:     sort,
	}

	category, err := s.cmdService.UpdateCategory(ctx, cmd)
	if err != nil {
		slog.Error("gRPC UpdateCategory failed", "id", req.Id, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update category: %v", err))
	}

	slog.Info("gRPC UpdateCategory successful", "id", req.Id, "duration", time.Since(start))
	return convertCategoryToProto(category), nil
}

func (s *Server) DeleteCategory(ctx context.Context, req *pb.DeleteCategoryRequest) (*emptypb.Empty, error) {
	start := time.Now()
	slog.Info("gRPC DeleteCategory received", "id", req.Id)

	if err := s.cmdService.DeleteCategory(ctx, req.Id); err != nil {
		slog.Error("gRPC DeleteCategory failed", "id", req.Id, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete category: %v", err))
	}

	slog.Info("gRPC DeleteCategory successful", "id", req.Id, "duration", time.Since(start))
	return &emptypb.Empty{}, nil
}

func (s *Server) UpdateBrand(ctx context.Context, req *pb.UpdateBrandRequest) (*pb.Brand, error) {
	start := time.Now()
	slog.Info("gRPC UpdateBrand received", "id", req.Id)

	var name *string
	if req.Name != nil {
		v := req.Name.Value
		name = &v
	}
	var logo *string
	if req.LogoUrl != nil {
		v := req.LogoUrl.Value
		logo = &v
	}

	cmd := &application.UpdateBrandCommand{
		ID:    uint(req.Id),
		Name:  name,
		Logo:  logo,
	}

	brand, err := s.cmdService.UpdateBrand(ctx, cmd)
	if err != nil {
		slog.Error("gRPC UpdateBrand failed", "id", req.Id, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update brand: %v", err))
	}

	slog.Info("gRPC UpdateBrand successful", "id", req.Id, "duration", time.Since(start))
	return convertBrandToProto(brand), nil
}

func (s *Server) DeleteBrand(ctx context.Context, req *pb.DeleteBrandRequest) (*emptypb.Empty, error) {
	start := time.Now()
	slog.Info("gRPC DeleteBrand received", "id", req.Id)

	if err := s.cmdService.DeleteBrand(ctx, req.Id); err != nil {
		slog.Error("gRPC DeleteBrand failed", "id", req.Id, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete brand: %v", err))
	}

	slog.Info("gRPC DeleteBrand successful", "id", req.Id, "duration", time.Since(start))
	return &emptypb.Empty{}, nil
}
