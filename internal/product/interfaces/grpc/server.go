package grpc

import (
	"context" // 导入上下文。
	"fmt"     // 导入格式化库。

	pb "github.com/wyfcoding/ecommerce/go-api/product/v1"         // 导入产品模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/product/application" // 导入产品模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/product/domain"      // 导入产品模块的领域实体。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/emptypb"     // 导入空消息类型。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
	// "google.golang.org/protobuf/types/known/wrapperspb" // 导入包装类型，用于可选字段，此处代码已内联处理。
)

// Server 结构体实现了 ProductService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedProductServer                                        // 嵌入生成的UnimplementedProductServer，确保前向兼容性。
	app                           *application.ProductApplicationService // 依赖Product应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 Product gRPC 服务端实例。
func NewServer(app *application.ProductApplicationService) *Server {
	return &Server{app: app}
}

// --- 商品 (Product) 相关接口实现 ---

// CreateProduct 处理创建商品的gRPC请求。
// req: 包含商品名称、描述、分类ID、品牌ID的请求体。
// 返回创建成功的商品信息响应和可能发生的gRPC错误。
func (s *Server) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.ProductInfo, error) {
	// 应用服务层的 CreateProduct 方法期望 price 和 stock 参数，但在 gRPC 请求中这些字段可能缺失或默认。
	// 当前实现暂时传递0值。这需要根据业务需求决定是在接口层、应用服务层补充或验证这些字段。
	product, err := s.app.CreateProduct(
		ctx,
		req.Name,
		req.Description,
		req.CategoryId,
		req.BrandId,
		0, // price: 暂时传递0，Proto请求中没有此字段。
		0, // stock: 暂时传递0，Proto请求中没有此字段。
	)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create product: %v", err))
	}

	// 备注：如果商品主图（MainImage）和图片列表（Images）需要在此处创建时一并处理，
	// 需要修改应用服务层的 CreateProduct 签名或提供额外的更新方法。
	// 目前将遵循应用服务层当前支持的字段。

	return convertProductToProto(product), nil
}

// GetProductByID 处理根据ID获取商品信息的gRPC请求。
// req: 包含商品ID的请求体。
// 返回商品信息响应和可能发生的gRPC错误。
func (s *Server) GetProductByID(ctx context.Context, req *pb.GetProductByIDRequest) (*pb.ProductInfo, error) {
	product, err := s.app.GetProductByID(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get product: %v", err))
	}
	if product == nil {
		return nil, status.Error(codes.NotFound, "product not found") // 如果商品未找到，返回NotFound错误。
	}
	return convertProductToProto(product), nil
}

// UpdateProductInfo 处理更新商品信息的gRPC请求。
// req: 包含商品ID和待更新字段的请求体。
// 返回更新后的商品信息响应和可能发生的gRPC错误。
func (s *Server) UpdateProductInfo(ctx context.Context, req *pb.UpdateProductInfoRequest) (*pb.ProductInfo, error) {
	// 将protobuf的包装类型（Wrapper Types）转换为Go的指针类型，以便应用服务层进行选择性更新。
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

	// 转换商品状态。
	var status *domain.ProductStatus
	if req.Status != pb.ProductStatus_PRODUCT_STATUS_UNSPECIFIED { // 检查Proto状态是否已指定。
		s := domain.ProductStatus(req.Status)
		status = &s
	}

	product, err := s.app.UpdateProductInfo(ctx, req.Id, name, desc, categoryID, brandID, status)
	if err != nil {
		return nil, err
	}
	return convertProductToProto(product), nil
}

// DeleteProduct 处理删除商品的gRPC请求。
// req: 包含商品ID的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*emptypb.Empty, error) {
	if err := s.app.DeleteProduct(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete product: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListProducts 处理列出商品列表的gRPC请求。
// req: 包含分页参数、分类ID和品牌ID过滤的请求体。
// 返回商品列表响应和可能发生的gRPC错误。
func (s *Server) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	products, total, err := s.app.ListProducts(ctx, int(req.Page), int(req.PageSize), req.CategoryId, req.BrandId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list products: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbProducts := make([]*pb.ProductInfo, len(products))
	for i, p := range products {
		pbProducts[i] = convertProductToProto(p)
	}

	return &pb.ListProductsResponse{
		Products: pbProducts,
		Total:    int32(total), // 总记录数。
		Page:     req.Page,     // 当前页码。
		PageSize: req.PageSize, // 每页大小。
	}, nil
}

// --- SKU 相关接口实现 ---

// AddSKUsToProduct 处理为商品添加SKU的gRPC请求。
// req: 包含商品ID和SKU列表的请求体。
// 返回创建成功的SKU列表响应和可能发生的gRPC错误。
func (s *Server) AddSKUsToProduct(ctx context.Context, req *pb.AddSKUsToProductRequest) (*pb.AddSKUsToProductResponse, error) {
	var createdSKUs []*pb.SKU
	for _, skuReq := range req.Skus {
		// 将Proto的SpecValues（键值对列表）映射到Go的map[string]string。
		specs := make(map[string]string)
		for _, sv := range skuReq.SpecValues {
			specs[sv.Key] = sv.Value
		}

		sku, err := s.app.AddSKU(ctx, req.ProductId, skuReq.Name, skuReq.Price, skuReq.StockQuantity, skuReq.ImageUrl, specs)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add SKU to product %d: %v", req.ProductId, err))
		}
		createdSKUs = append(createdSKUs, convertSKUToProto(sku)) // 将创建的SKU转换为Proto格式。
	}
	return &pb.AddSKUsToProductResponse{CreatedSkus: createdSKUs}, nil
}

// UpdateSKU 更新单个SKU的信息。
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

	sku, err := s.app.UpdateSKU(ctx, req.Id, price, stock, image)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update SKU: %v", err))
	}
	return convertSKUToProto(sku), nil
}

// DeleteSKU 删除商品下的一个或多个SKU。
func (s *Server) DeleteSKU(ctx context.Context, req *pb.DeleteSKURequest) (*emptypb.Empty, error) {
	// 目前应用服务只支持单个删除，这里循环删除
	for _, id := range req.SkuIds {
		if err := s.app.DeleteSKU(ctx, id); err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete SKU %d: %v", id, err))
		}
	}
	return &emptypb.Empty{}, nil
}

// GetSKUByID 根据SKU ID获取其详细信息。
func (s *Server) GetSKUByID(ctx context.Context, req *pb.GetSKUByIDRequest) (*pb.SKU, error) {
	sku, err := s.app.GetSKUByID(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get SKU: %v", err))
	}
	if sku == nil {
		return nil, status.Error(codes.NotFound, "SKU not found")
	}
	return convertSKUToProto(sku), nil
}

// --- 分类 (Category) 相关接口实现 ---

// CreateCategory 创建一个新的商品分类。
func (s *Server) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.Category, error) {
	category, err := s.app.CreateCategory(ctx, req.Name, req.ParentId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create category: %v", err))
	}
	return convertCategoryToProto(category), nil
}

// GetCategoryByID 根据ID获取分类信息。
func (s *Server) GetCategoryByID(ctx context.Context, req *pb.GetCategoryByIDRequest) (*pb.Category, error) {
	category, err := s.app.GetCategoryByID(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get category: %v", err))
	}
	if category == nil {
		return nil, status.Error(codes.NotFound, "category not found")
	}
	return convertCategoryToProto(category), nil
}

// UpdateCategory 更新分类信息。
func (s *Server) UpdateCategory(ctx context.Context, req *pb.UpdateCategoryRequest) (*pb.Category, error) {
	var name *string
	if req.Name != nil {
		v := req.Name.Value
		name = &v
	}
	var parentID *uint64
	if req.ParentId != nil {
		v := req.ParentId.Value
		parentID = &v
	}
	var sort *int
	if req.SortOrder != nil {
		v := int(req.SortOrder.Value)
		sort = &v
	}

	category, err := s.app.UpdateCategory(ctx, req.Id, name, parentID, sort)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update category: %v", err))
	}
	return convertCategoryToProto(category), nil
}

// DeleteCategory 删除一个分类。
func (s *Server) DeleteCategory(ctx context.Context, req *pb.DeleteCategoryRequest) (*emptypb.Empty, error) {
	if err := s.app.DeleteCategory(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete category: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListCategories 获取分类列表。
func (s *Server) ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	categories, err := s.app.ListCategories(ctx, req.ParentId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list categories: %v", err))
	}

	var pbCategories []*pb.Category
	for _, c := range categories {
		pbCategories = append(pbCategories, convertCategoryToProto(c))
	}
	return &pb.ListCategoriesResponse{Categories: pbCategories}, nil
}

// --- 品牌 (Brand) 相关接口实现 ---

// CreateBrand 创建一个新的品牌。
func (s *Server) CreateBrand(ctx context.Context, req *pb.CreateBrandRequest) (*pb.Brand, error) {
	brand, err := s.app.CreateBrand(ctx, req.Name, req.LogoUrl)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create brand: %v", err))
	}
	return convertBrandToProto(brand), nil
}

// GetBrandByID 根据ID获取品牌信息。
func (s *Server) GetBrandByID(ctx context.Context, req *pb.GetBrandByIDRequest) (*pb.Brand, error) {
	brand, err := s.app.GetBrandByID(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get brand: %v", err))
	}
	if brand == nil {
		return nil, status.Error(codes.NotFound, "brand not found")
	}
	return convertBrandToProto(brand), nil
}

// UpdateBrand 更新品牌信息。
func (s *Server) UpdateBrand(ctx context.Context, req *pb.UpdateBrandRequest) (*pb.Brand, error) {
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

	brand, err := s.app.UpdateBrand(ctx, req.Id, name, logo)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update brand: %v", err))
	}
	return convertBrandToProto(brand), nil
}

// DeleteBrand 删除一个品牌。
func (s *Server) DeleteBrand(ctx context.Context, req *pb.DeleteBrandRequest) (*emptypb.Empty, error) {
	if err := s.app.DeleteBrand(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete brand: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListBrands 分页列出所有品牌。
func (s *Server) ListBrands(ctx context.Context, req *pb.ListBrandsRequest) (*pb.ListBrandsResponse, error) {
	// 目前应用服务 ListBrands 不支持分页，这里暂时忽略分页参数
	brands, err := s.app.ListBrands(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list brands: %v", err))
	}

	var pbBrands []*pb.Brand
	for _, b := range brands {
		pbBrands = append(pbBrands, convertBrandToProto(b))
	}
	return &pb.ListBrandsResponse{Brands: pbBrands, Total: int32(len(brands))}, nil
}

// --- 辅助函数：领域实体到Proto消息的转换 ---

// convertProductToProto 是一个辅助函数，将领域层的 Product 实体转换为 protobuf 的 ProductInfo 消息。
func convertProductToProto(p *domain.Product) *pb.ProductInfo {
	if p == nil {
		return nil
	}
	// 转换关联的SKU列表。
	pbSKUs := make([]*pb.SKU, len(p.SKUs))
	for i, sku := range p.SKUs {
		pbSKUs[i] = convertSKUToProto(sku)
	}

	return &pb.ProductInfo{
		Id:               uint64(p.ID),                           // 商品ID。
		Name:             p.Name,                                 // 名称。
		Description:      p.Description,                          // 描述。
		Category:         &pb.Category{Id: uint64(p.CategoryID)}, // 分类信息（仅ID）。
		Brand:            &pb.Brand{Id: uint64(p.BrandID)},       // 品牌信息（仅ID）。
		Status:           pb.ProductStatus(p.Status),             // 状态。
		Skus:             pbSKUs,                                 // SKU列表。
		MainImageUrl:     p.MainImage,                            // 主图URL。
		GalleryImageUrls: p.Images,                               // 图片列表。
		CreatedAt:        timestamppb.New(p.CreatedAt),           // 创建时间。
		UpdatedAt:        timestamppb.New(p.UpdatedAt),           // 更新时间。
	}
}

// convertSKUToProto 是一个辅助函数，将领域层的 SKU 实体转换为 protobuf 的 SKU 消息。
func convertSKUToProto(s *domain.SKU) *pb.SKU {
	if s == nil {
		return nil
	}
	// 转换SKU的规格参数。
	var specValues []*pb.SpecValue
	for k, v := range s.Specs {
		specValues = append(specValues, &pb.SpecValue{Key: k, Value: v})
	}

	return &pb.SKU{
		Id:            uint64(s.ID),                 // SKU ID。
		ProductId:     uint64(s.ProductID),          // 商品ID。
		Name:          s.Name,                       // 名称。
		Price:         s.Price,                      // 价格。
		StockQuantity: s.Stock,                      // 库存数量。
		ImageUrl:      s.Image,                      // 图片URL。
		SpecValues:    specValues,                   // 规格参数。
		CreatedAt:     timestamppb.New(s.CreatedAt), // 创建时间。
		UpdatedAt:     timestamppb.New(s.UpdatedAt), // 更新时间。
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
		// IconUrl missing in domain entity
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
		// Description missing in domain entity
	}
}
