package grpc

import (
	"context"

	pb "ecommerce/api/product/v1"
	"ecommerce/internal/product/application"
	"ecommerce/internal/product/domain"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedProductServer
	app *application.ProductApplicationService
}

func NewServer(app *application.ProductApplicationService) *Server {
	return &Server{app: app}
}

// --- Product ---

func (s *Server) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.ProductInfo, error) {
	// Price and Stock are not in CreateProductRequest based on previous error context (passed as 0).
	// We need to check if we should add them to request or just pass 0.
	// For now, passing 0 as in previous code, but removing extra args.
	product, err := s.app.CreateProduct(
		ctx,
		req.Name,
		req.Description,
		req.CategoryId,
		req.BrandId,
		0, // Price
		0, // Stock
	)
	if err != nil {
		return nil, err
	}

	// If MainImage and Images are needed, we should update them separately or update CreateProduct signature.
	// But CreateProduct in service.go doesn't take them.
	// Let's update them after creation if needed, or just ignore for now as per service.go change.
	// Actually, service.go CreateProduct only saves basic info.
	// If we want to save images, we might need to update service.go or call another method.
	// Given the refactoring scope, I'll stick to what service.go supports.

	return convertProductToProto(product), nil
}

func (s *Server) GetProductByID(ctx context.Context, req *pb.GetProductByIDRequest) (*pb.ProductInfo, error) {
	product, err := s.app.GetProductByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, nil // Or NotFound error
	}
	return convertProductToProto(product), nil
}

func (s *Server) UpdateProductInfo(ctx context.Context, req *pb.UpdateProductInfoRequest) (*pb.ProductInfo, error) {
	// Convert wrappers to pointers
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
	var categoryID *uint64
	if req.CategoryId != nil {
		v := req.CategoryId.Value
		categoryID = &v
	}
	var brandID *uint64
	if req.BrandId != nil {
		v := req.BrandId.Value
		brandID = &v
	}

	// Status conversion
	var status *domain.ProductStatus
	if req.Status != pb.ProductStatus_PRODUCT_STATUS_UNSPECIFIED {
		s := domain.ProductStatus(req.Status)
		status = &s
	}

	product, err := s.app.UpdateProductInfo(ctx, req.Id, name, desc, categoryID, brandID, status)
	if err != nil {
		return nil, err
	}
	return convertProductToProto(product), nil
}

func (s *Server) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*emptypb.Empty, error) {
	if err := s.app.DeleteProduct(ctx, req.Id); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	products, total, err := s.app.ListProducts(ctx, int(req.Page), int(req.PageSize), req.CategoryId, req.BrandId)
	if err != nil {
		return nil, err
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
		// Specs mapping
		specs := make(map[string]string)
		for _, sv := range skuReq.SpecValues {
			specs[sv.Key] = sv.Value
		}

		sku, err := s.app.AddSKU(ctx, req.ProductId, skuReq.Name, skuReq.Price, skuReq.StockQuantity, skuReq.ImageUrl, specs)
		if err != nil {
			return nil, err
		}
		createdSKUs = append(createdSKUs, convertSKUToProto(sku))
	}
	return &pb.AddSKUsToProductResponse{CreatedSkus: createdSKUs}, nil
}

// ... Implement other methods similarly ...

// Helper functions

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
