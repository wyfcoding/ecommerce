package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	v1 "ecommerce/api/product/v1"
	"ecommerce/internal/product/model"
	"ecommerce/internal/product/repository"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ProductService 封装了商品、SKU、分类和品牌相关的业务逻辑，实现了 product.proto 中定义的 ProductServer 接口。
type ProductService struct {
	// 嵌入 v1.UnimplementedProductServer 以确保向前兼容性
	v1.UnimplementedProductServer
	productRepo  repository.ProductRepo
	skuRepo      repository.SKURepo
	categoryRepo repository.CategoryRepo
	brandRepo    repository.BrandRepo
	validator    *validator.Validate
	defaultPageSize int32
	maxPageSize     int32
}

// NewProductService 是 ProductService 的构造函数。
func NewProductService(
	productRepo repository.ProductRepo,
	skuRepo repository.SKURepo,
	categoryRepo repository.CategoryRepo,
	brandRepo repository.BrandRepo,
	defaultPageSize, maxPageSize int32,
) *ProductService {
	return &ProductService{
		productRepo:  productRepo,
		skuRepo:      skuRepo,
		categoryRepo: categoryRepo,
		brandRepo:    brandRepo,
		validator:    validator.New(),
		defaultPageSize: defaultPageSize,
		maxPageSize:     maxPageSize,
	}
}

// --- 商品 (SPU) 核心接口实现 ---

// CreateProduct 实现了创建商品的 RPC 方法。
func (s *ProductService) CreateProduct(ctx context.Context, req *v1.CreateProductRequest) (*v1.ProductInfo, error) {
	zap.S().Infof("CreateProduct request received: %v", req.Name)

	// 1. 参数校验
	if err := s.validator.Struct(req); err != nil {
		zap.S().Warnf("CreateProduct request validation failed: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid argument: %v", err)
	}

	// 2. 检查分类和品牌是否存在
	category, err := s.categoryRepo.GetCategoryByID(ctx, req.CategoryId)
	if err != nil {
		zap.S().Errorf("failed to get category %d: %v", req.CategoryId, err)
		return nil, status.Errorf(codes.Internal, "failed to check category")
	}
	if category == nil {
		zap.S().Warnf("category %d not found for product creation", req.CategoryId)
		return nil, status.Errorf(codes.NotFound, "category not found")
	}

	brand, err := s.brandRepo.GetBrandByID(ctx, req.BrandId)
	if err != nil {
		zap.S().Errorf("failed to get brand %d: %v", req.BrandId, err)
		return nil, status.Errorf(codes.Internal, "failed to check brand")
	}
	if brand == nil {
		zap.S().Warnf("brand %d not found for product creation", req.BrandId)
		return nil, status.Errorf(codes.NotFound, "brand not found")
	}

	// 3. 构建商品模型
	product := &model.Product{
		Name:            req.Name,
		Description:     req.Description,
		CategoryID:      req.CategoryId,
		BrandID:         req.BrandId,
		Status:          model.ProductStatus(req.Status),
		MainImageURL:    req.MainImageURL,
		GalleryImageURLs: strings.Join(req.GalleryImageUrls, ","), // 存储为逗号分隔字符串或JSON
		Weight:          req.Weight.GetValue(),
	}

	// 自动生成SPU编码 (示例: P + 时间戳 + 随机数)
	product.SpuNo = fmt.Sprintf("P%d%s", time.Now().UnixNano()/int64(time.Millisecond), strconv.Itoa(int(time.Now().Nanosecond()%1000)))

	// 处理SEO信息
	if req.SeoInfo != nil {
		product.MetaTitle = req.SeoInfo.MetaTitle
		product.MetaDescription = req.SeoInfo.MetaDescription
		product.MetaKeywords = req.SeoInfo.MetaKeywords
		product.URLSlug = req.SeoInfo.UrlSlug
	}

	// 处理商品属性
	for _, attr := range req.Attributes {
		product.Attributes = append(product.Attributes, model.ProductAttribute{Key: attr.Key, Value: attr.Value})
	}

	// 4. 调用仓库层创建商品
	createdProduct, err := s.productRepo.CreateProduct(ctx, product)
	if err != nil {
		zap.S().Errorf("failed to create product %s: %v", req.Name, err)
		return nil, status.Errorf(codes.Internal, "failed to create product")
	}

	zap.S().Infof("Product %d created successfully", createdProduct.ID)
	return s.bizProductToProto(ctx, createdProduct) // 转换并返回Protobuf类型
}

// GetProductByID 实现了根据ID获取商品详情的 RPC 方法。
func (s *ProductService) GetProductByID(ctx context.Context, req *v1.GetProductByIDRequest) (*v1.ProductInfo, error) {
	zap.S().Infof("GetProductByID request received for product ID: %d", req.Id)

	// 1. 参数校验
	if req.Id == 0 {
		zap.S().Warn("GetProductByID request with zero product ID")
		return nil, status.Errorf(codes.InvalidArgument, "product ID cannot be zero")
	}

	// 2. 调用仓库层获取商品
	product, err := s.productRepo.GetProductByID(ctx, req.Id)
	if err != nil {
		zap.S().Errorf("failed to get product by ID %d: %v", req.Id, err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve product")
	}
	if product == nil {
		zap.S().Warnf("product with ID %d not found", req.Id)
		return nil, status.Errorf(codes.NotFound, "product not found")
	}

	zap.S().Infof("Product %d retrieved successfully", req.Id)
	return s.bizProductToProto(ctx, product) // 转换并返回Protobuf类型
}

// UpdateProductInfo 实现了更新商品核心信息的 RPC 方法。
func (s *ProductService) UpdateProductInfo(ctx context.Context, req *v1.UpdateProductInfoRequest) (*v1.ProductInfo, error) {
	zap.S().Infof("UpdateProductInfo request received for product ID: %d", req.Id)

	// 1. 参数校验
	if req.Id == 0 {
		zap.S().Warn("UpdateProductInfo request with zero product ID")
		return nil, status.Errorf(codes.InvalidArgument, "product ID cannot be zero")
	}

	// 2. 获取现有商品信息
	existingProduct, err := s.productRepo.GetProductByID(ctx, req.Id)
	if err != nil {
		zap.S().Errorf("failed to get existing product %d for update: %v", req.Id, err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve product for update")
	}
	if existingProduct == nil {
		zap.S().Warnf("product with ID %d not found for update", req.Id)
		return nil, status.Errorf(codes.NotFound, "product not found")
	}

	// 3. 根据请求更新字段
	if req.Name != nil {
		existingProduct.Name = req.Name.GetValue()
	}
	if req.Description != nil {
		existingProduct.Description = req.Description.GetValue()
	}
	if req.CategoryId != nil {
		// 检查新分类是否存在
		category, err := s.categoryRepo.GetCategoryByID(ctx, req.CategoryId.GetValue())
		if err != nil || category == nil {
			zap.S().Warnf("new category %d not found for product %d update", req.CategoryId.GetValue(), req.Id)
			return nil, status.Errorf(codes.NotFound, "new category not found")
		}
		existingProduct.CategoryID = req.CategoryId.GetValue()
	}
	if req.BrandId != nil {
		// 检查新品牌是否存在
		brand, err := s.brandRepo.GetBrandByID(ctx, req.BrandId.GetValue())
		if err != nil || brand == nil {
			zap.S().Warnf("new brand %d not found for product %d update", req.BrandId.GetValue(), req.Id)
			return nil, status.Errorf(codes.NotFound, "new brand not found")
		}
		existingProduct.BrandID = req.BrandId.GetValue()
	}
	// 状态更新
	if req.Status != v1.ProductStatus_PRODUCT_STATUS_UNSPECIFIED {
		existingProduct.Status = model.ProductStatus(req.Status)
	}
	if req.MainImageURL != "" {
		existingProduct.MainImageURL = req.MainImageURL
	}
	if len(req.GalleryImageUrls) > 0 {
		existingProduct.GalleryImageURLs = strings.Join(req.GalleryImageUrls, ",")
	}
	if req.SeoInfo != nil {
		existingProduct.MetaTitle = req.SeoInfo.MetaTitle
		existingProduct.MetaDescription = req.SeoInfo.MetaDescription
		existingProduct.MetaKeywords = req.SeoInfo.MetaKeywords
		existingProduct.URLSlug = req.SeoInfo.UrlSlug
	}
	if req.Weight != nil {
		existingProduct.Weight = req.Weight.GetValue()
	}

	// 4. 调用仓库层更新商品
	updatedProduct, err := s.productRepo.UpdateProduct(ctx, existingProduct)
	if err != nil {
		zap.S().Errorf("failed to update product %d: %v", req.Id, err)
		return nil, status.Errorf(codes.Internal, "failed to update product")
	}

	zap.S().Infof("Product %d updated successfully", req.Id)
	return s.bizProductToProto(ctx, updatedProduct) // 转换并返回Protobuf类型
}

// DeleteProduct 实现了删除商品的 RPC 方法 (逻辑删除)。
func (s *ProductService) DeleteProduct(ctx context.Context, req *v1.DeleteProductRequest) (*emptypb.Empty, error) {
	zap.S().Infof("DeleteProduct request received for product ID: %d", req.Id)

	// 1. 参数校验
	if req.Id == 0 {
		zap.S().Warn("DeleteProduct request with zero product ID")
		return nil, status.Errorf(codes.InvalidArgument, "product ID cannot be zero")
	}

	// 2. 检查商品是否存在
	existingProduct, err := s.productRepo.GetProductByID(ctx, req.Id)
	if err != nil {
		zap.S().Errorf("failed to get product %d for deletion: %v", req.Id, err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve product for deletion")
	}
	if existingProduct == nil {
		zap.S().Warnf("product with ID %d not found for deletion", req.Id)
		return nil, status.Errorf(codes.NotFound, "product not found")
	}

	// 3. 调用仓库层逻辑删除商品
	if err := s.productRepo.DeleteProduct(ctx, req.Id); err != nil {
		zap.S().Errorf("failed to delete product %d: %v", req.Id, err)
		return nil, status.Errorf(codes.Internal, "failed to delete product")
	}

	zap.S().Infof("Product %d deleted successfully", req.Id)
	return &emptypb.Empty{}, nil
}

// ListProducts 实现了分页列出商品的 RPC 方法。
func (s *ProductService) ListProducts(ctx context.Context, req *v1.ListProductsRequest) (*v1.ListProductsResponse, error) {
	zap.S().Infof("ListProducts request received: page=%d, page_size=%d, category_id=%d, brand_id=%d, status=%s",
		req.Page, req.PageSize, req.CategoryId, req.BrandId, req.Status.String())

	// 1. 参数校验与默认值设置
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = s.defaultPageSize
	}
	if pageSize > s.maxPageSize {
		pageSize = s.maxPageSize
	}
	page := req.Page
	if page == 0 {
		page = 1
	}

	query := &repository.ProductListQuery{
		Page:       page,
		PageSize:   pageSize,
		CategoryID: req.CategoryId,
		BrandID:    req.BrandId,
		Status:     model.ProductStatus(req.Status),
		SortBy:     req.SortBy,
	}

	// 2. 调用仓库层查询商品列表
	products, total, err := s.productRepo.ListProducts(ctx, query)
	if err != nil {
		zap.S().Errorf("failed to list products: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to list products")
	}

	// 3. 转换为Protobuf类型
	protoProducts := make([]*v1.ProductInfo, len(products))
	for i, p := range products {
		protoProduct, err := s.bizProductToProto(ctx, p)
		if err != nil {
			zap.S().Errorf("failed to convert product %d to proto: %v", p.ID, err)
			// 即使转换失败，也尝试返回部分成功的数据，或者直接返回错误
			return nil, status.Errorf(codes.Internal, "failed to process product data")
		}
		protoProducts[i] = protoProduct
	}

	zap.S().Infof("Listed %d products (total: %d)", len(products), total)
	return &v1.ListProductsResponse{
		Products: protoProducts,
		Total:    int32(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// --- SKU (Stock Keeping Unit) 接口实现 ---

// AddSKUsToProduct 实现了为指定商品批量添加SKU的 RPC 方法。
func (s *ProductService) AddSKUsToProduct(ctx context.Context, req *v1.AddSKUsToProductRequest) (*v1.AddSKUsToProductResponse, error) {
	zap.S().Infof("AddSKUsToProduct request received for product ID: %d, SKUs count: %d", req.ProductId, len(req.Skus))

	// 1. 参数校验
	if req.ProductId == 0 || len(req.Skus) == 0 {
		zap.S().Warn("AddSKUsToProduct request with invalid product ID or empty SKUs list")
		return nil, status.Errorf(codes.InvalidArgument, "product ID cannot be zero or SKUs list cannot be empty")
	}

	// 2. 检查商品是否存在
	product, err := s.productRepo.GetProductByID(ctx, req.ProductId)
	if err != nil {
		zap.S().Errorf("failed to get product %d for adding SKUs: %v", req.ProductId, err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve product")
	}
	if product == nil {
		zap.S().Warnf("product with ID %d not found for adding SKUs", req.ProductId)
		return nil, status.Errorf(codes.NotFound, "product not found")
	}

	// 3. 构建SKU模型列表
	var skusToCreate []*model.SKU
	for _, skuReq := range req.Skus {
		// 自动生成SKU编码 (示例: S + SPU_ID + 时间戳 + 随机数)
		skuNo := fmt.Sprintf("S%d%d%s", req.ProductId, time.Now().UnixNano()/int64(time.Millisecond), strconv.Itoa(int(time.Now().Nanosecond()%1000)))

		specValuesJSON, err := json.Marshal(skuReq.SpecValues)
		if err != nil {
			zap.S().Errorf("failed to marshal spec values for SKU %s: %v", skuReq.Name, err)
			return nil, status.Errorf(codes.Internal, "failed to process SKU spec values")
		}

		skusToCreate = append(skusToCreate, &model.SKU{
			ProductID:     req.ProductId,
			SkuNo:         skuNo,
			Name:          skuReq.Name,
			Price:         skuReq.Price,
			StockQuantity: skuReq.StockQuantity,
			ImageURL:      skuReq.ImageURL,
			SpecValues:    string(specValuesJSON),
		})
	}

	// 4. 调用仓库层批量创建SKU
	createdSKUs, err := s.skuRepo.CreateSKUs(ctx, skusToCreate)
	if err != nil {
		zap.S().Errorf("failed to create SKUs for product %d: %v", req.ProductId, err)
		return nil, status.Errorf(codes.Internal, "failed to create SKUs")
	}

	// 5. 转换为Protobuf类型
	protoSKUs := make([]*v1.SKU, len(createdSKUs))
	for i, sku := range createdSKUs {
		protoSKUs[i] = s.bizSKUToProto(sku)
	}

	zap.S().Infof("Added %d SKUs to product %d successfully", len(createdSKUs), req.ProductId)
	return &v1.AddSKUsToProductResponse{
		CreatedSkus: protoSKUs,
	}, nil
}

// UpdateSKU 实现了更新单个SKU信息的 RPC 方法。
func (s *ProductService) UpdateSKU(ctx context.Context, req *v1.UpdateSKURequest) (*v1.SKU, error) {
	zap.S().Infof("UpdateSKU request received for SKU ID: %d", req.Id)

	// 1. 参数校验
	if req.Id == 0 {
		zap.S().Warn("UpdateSKU request with zero SKU ID")
		return nil, status.Errorf(codes.InvalidArgument, "SKU ID cannot be zero")
	}

	// 2. 获取现有SKU信息
	existingSKU, err := s.skuRepo.GetSKUByID(ctx, req.Id)
	if err != nil {
		zap.S().Errorf("failed to get existing SKU %d for update: %v", req.Id, err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve SKU for update")
	}
	if existingSKU == nil {
		zap.S().Warnf("SKU with ID %d not found for update", req.Id)
		return nil, status.Errorf(codes.NotFound, "SKU not found")
	}

	// 3. 根据请求更新字段
	if req.Price != nil {
		existingSKU.Price = req.Price.GetValue()
	}
	if req.StockQuantity != nil {
		existingSKU.StockQuantity = req.StockQuantity.GetValue()
	}
	if req.ImageURL != nil {
		existingSKU.ImageURL = req.ImageURL.GetValue()
	}

	// 4. 调用仓库层更新SKU
	updatedSKU, err := s.skuRepo.UpdateSKU(ctx, existingSKU)
	if err != nil {
		zap.S().Errorf("failed to update SKU %d: %v", req.Id, err)
		return nil, status.Errorf(codes.Internal, "failed to update SKU")
	}

	zap.S().Infof("SKU %d updated successfully", req.Id)
	return s.bizSKUToProto(updatedSKU), nil
}

// DeleteSKU 实现了删除商品下的一个或多个SKU的 RPC 方法 (逻辑删除)。
func (s *ProductService) DeleteSKU(ctx context.Context, req *v1.DeleteSKURequest) (*emptypb.Empty, error) {
	zap.S().Infof("DeleteSKU request received for product ID: %d, SKU IDs: %v", req.ProductId, req.SkuIds)

	// 1. 参数校验
	if req.ProductId == 0 || len(req.SkuIds) == 0 {
		zap.S().Warn("DeleteSKU request with invalid product ID or empty SKU IDs list")
		return nil, status.Errorf(codes.InvalidArgument, "product ID cannot be zero or SKU IDs list cannot be empty")
	}

	// 2. 检查商品是否存在 (可选，但建议进行)
	product, err := s.productRepo.GetProductByID(ctx, req.ProductId)
	if err != nil {
		zap.S().Errorf("failed to get product %d for SKU deletion: %v", req.ProductId, err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve product")
	}
	if product == nil {
		zap.S().Warnf("product with ID %d not found for SKU deletion", req.ProductId)
		return nil, status.Errorf(codes.NotFound, "product not found")
	}

	// 3. 调用仓库层批量删除SKU
	if err := s.skuRepo.DeleteSKUs(ctx, req.ProductId, req.SkuIds); err != nil {
		zap.S().Errorf("failed to delete SKUs %v for product %d: %v", req.SkuIds, req.ProductId, err)
		return nil, status.Errorf(codes.Internal, "failed to delete SKUs")
	}

	zap.S().Infof("SKUs %v for product %d deleted successfully", req.SkuIds, req.ProductId)
	return &emptypb.Empty{}, nil
}

// GetSKUByID 实现了根据SKU ID获取其详细信息的 RPC 方法。
func (s *ProductService) GetSKUByID(ctx context.Context, req *v1.GetSKUByIDRequest) (*v1.SKU, error) {
	zap.S().Infof("GetSKUByID request received for SKU ID: %d", req.Id)

	// 1. 参数校验
	if req.Id == 0 {
		zap.S().Warn("GetSKUByID request with zero SKU ID")
		return nil, status.Errorf(codes.InvalidArgument, "SKU ID cannot be zero")
	}

	// 2. 调用仓库层获取SKU
	sku, err := s.skuRepo.GetSKUByID(ctx, req.Id)
	if err != nil {
		zap.S().Errorf("failed to get SKU by ID %d: %v", req.Id, err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve SKU")
	}
	if sku == nil {
		zap.S().Warnf("SKU with ID %d not found", req.Id)
		return nil, status.Errorf(codes.NotFound, "SKU not found")
	}

	zap.S().Infof("SKU %d retrieved successfully", req.Id)
	return s.bizSKUToProto(sku), nil
}

// --- 商品分类接口实现 ---

// CreateCategory 实现了创建商品分类的 RPC 方法。
func (s *ProductService) CreateCategory(ctx context.Context, req *v1.CreateCategoryRequest) (*v1.Category, error) {
	zap.S().Infof("CreateCategory request received: %v (ParentID: %d)", req.Name, req.ParentId)

	// 1. 参数校验
	if err := s.validator.Struct(req); err != nil {
		zap.S().Warnf("CreateCategory request validation failed: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid argument: %v", err)
	}

	// 2. 检查同级分类名称是否重复
	existingCategory, err := s.categoryRepo.GetCategoryByName(ctx, req.Name)
	if err != nil {
		zap.S().Errorf("failed to check existing category name %s: %v", req.Name, err)
		return nil, status.Errorf(codes.Internal, "failed to check category name")
	}
	if existingCategory != nil && existingCategory.ParentID == req.ParentId {
		zap.S().Warnf("category name '%s' already exists under parent %d", req.Name, req.ParentId)
		return nil, status.Errorf(codes.AlreadyExists, "category name '%s' already exists under this parent", req.Name)
	}

	// 3. 构建分类模型
	category := &model.Category{
		Name:      req.Name,
		ParentID:  req.ParentId,
		IconURL:   req.IconUrl,
		SortOrder: req.SortOrder,
	}

	// 4. 调用仓库层创建分类
	createdCategory, err := s.categoryRepo.CreateCategory(ctx, category)
	if err != nil {
		zap.S().Errorf("failed to create category %s: %v", req.Name, err)
		return nil, status.Errorf(codes.Internal, "failed to create category")
	}

	zap.S().Infof("Category %d created successfully", createdCategory.ID)
	return s.bizCategoryToProto(createdCategory), nil
}

// GetCategoryByID 实现了根据ID获取分类信息的 RPC 方法。
func (s *ProductService) GetCategoryByID(ctx context.Context, req *v1.GetCategoryByIDRequest) (*v1.Category, error) {
	zap.S().Infof("GetCategoryByID request received for category ID: %d", req.Id)

	// 1. 参数校验
	if req.Id == 0 {
		zap.S().Warn("GetCategoryByID request with zero category ID")
		return nil, status.Errorf(codes.InvalidArgument, "category ID cannot be zero")
	}

	// 2. 调用仓库层获取分类
	category, err := s.categoryRepo.GetCategoryByID(ctx, req.Id)
	if err != nil {
		zap.S().Errorf("failed to get category by ID %d: %v", req.Id, err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve category")
	}
	if category == nil {
		zap.S().Warnf("category with ID %d not found", req.Id)
		return nil, status.Errorf(codes.NotFound, "category not found")
	}

	zap.S().Infof("Category %d retrieved successfully", req.Id)
	return s.bizCategoryToProto(category), nil
}

// UpdateCategory 实现了更新分类信息的 RPC 方法。
func (s *ProductService) UpdateCategory(ctx context.Context, req *v1.UpdateCategoryRequest) (*v1.Category, error) {
	zap.S().Infof("UpdateCategory request received for category ID: %d", req.Id)

	// 1. 参数校验
	if req.Id == 0 {
		zap.S().Warn("UpdateCategory request with zero category ID")
		return nil, status.Errorf(codes.InvalidArgument, "category ID cannot be zero")
	}

	// 2. 获取现有分类信息
	existingCategory, err := s.categoryRepo.GetCategoryByID(ctx, req.Id)
	if err != nil {
		zap.S().Errorf("failed to get existing category %d for update: %v", req.Id, err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve category for update")
	}
	if existingCategory == nil {
		zap.S().Warnf("category with ID %d not found for update", req.Id)
		return nil, status.Errorf(codes.NotFound, "category not found")
	}

	// 3. 根据请求更新字段
	if req.Name != nil {
		// 检查新名称是否与同级其他分类重复
		if existingCategory.Name != req.Name.GetValue() {
			checkCategory, err := s.categoryRepo.GetCategoryByName(ctx, req.Name.GetValue())
			if err != nil {
				zap.S().Errorf("failed to check category name %s: %v", req.Name.GetValue(), err)
				return nil, status.Errorf(codes.Internal, "failed to check category name")
			}
			if checkCategory != nil && checkCategory.ParentID == existingCategory.ParentID {
				zap.S().Warnf("category name '%s' already exists under parent %d", req.Name.GetValue(), existingCategory.ParentID)
				return nil, status.Errorf(codes.AlreadyExists, "category name '%s' already exists under this parent", req.Name.GetValue())
			}
		}
		existingCategory.Name = req.Name.GetValue()
	}
	if req.ParentId != nil {
		// 检查新父分类是否存在
		if req.ParentId.GetValue() != 0 {
			parent, err := s.categoryRepo.GetCategoryByID(ctx, req.ParentId.GetValue())
			if err != nil || parent == nil {
				zap.S().Warnf("new parent category %d not found for category %d update", req.ParentId.GetValue(), req.Id)
				return nil, status.Errorf(codes.NotFound, "new parent category not found")
			}
		}
		existingCategory.ParentID = req.ParentId.GetValue()
	}
	if req.IconUrl != nil {
		existingCategory.IconURL = req.IconUrl.GetValue()
	}
	if req.SortOrder != nil {
		existingCategory.SortOrder = req.SortOrder.GetValue()
	}

	// 4. 调用仓库层更新分类
	updatedCategory, err := s.categoryRepo.UpdateCategory(ctx, existingCategory)
	if err != nil {
		zap.S().Errorf("failed to update category %d: %v", req.Id, err)
		return nil, status.Errorf(codes.Internal, "failed to update category")
	}

	zap.S().Infof("Category %d updated successfully", req.Id)
	return s.bizCategoryToProto(updatedCategory), nil
}

// DeleteCategory 实现了删除分类的 RPC 方法 (逻辑删除)。
func (s *ProductService) DeleteCategory(ctx context.Context, req *v1.DeleteCategoryRequest) (*emptypb.Empty, error) {
	zap.S().Infof("DeleteCategory request received for category ID: %d", req.Id)

	// 1. 参数校验
	if req.Id == 0 {
		zap.S().Warn("DeleteCategory request with zero category ID")
		return nil, status.Errorf(codes.InvalidArgument, "category ID cannot be zero")
	}

	// 2. 检查分类是否存在
	existingCategory, err := s.categoryRepo.GetCategoryByID(ctx, req.Id)
	if err != nil {
		zap.S().Errorf("failed to get category %d for deletion: %v", req.Id, err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve category for deletion")
	}
	if existingCategory == nil {
		zap.S().Warnf("category with ID %d not found for deletion", req.Id)
		return nil, status.Errorf(codes.NotFound, "category not found")
	}

	// 3. 检查分类下是否有子分类或关联商品 (此处仅为示例，实际业务中可能需要更复杂的检查)
	childCategories, err := s.categoryRepo.ListCategories(ctx, req.Id)
	if err != nil {
		zap.S().Errorf("failed to check child categories for category %d: %v", req.Id, err)
		return nil, status.Errorf(codes.Internal, "failed to check child categories")
	}
	if len(childCategories) > 0 {
		zap.S().Warnf("category %d has %d child categories, cannot delete", req.Id, len(childCategories))
		return nil, status.Errorf(codes.FailedPrecondition, "category has child categories, cannot delete")
	}

	// 4. 调用仓库层逻辑删除分类
	if err := s.categoryRepo.DeleteCategory(ctx, req.Id); err != nil {
		zap.S().Errorf("failed to delete category %d: %v", req.Id, err)
		return nil, status.Errorf(codes.Internal, "failed to delete category")
	}

	zap.S().Infof("Category %d deleted successfully", req.Id)
	return &emptypb.Empty{}, nil
}

// ListCategories 实现了获取分类列表的 RPC 方法。
func (s *ProductService) ListCategories(ctx context.Context, req *v1.ListCategoriesRequest) (*v1.ListCategoriesResponse, error) {
	zap.S().Infof("ListCategories request received for parent ID: %d", req.ParentId)

	// 1. 调用仓库层查询分类列表
	categories, err := s.categoryRepo.ListCategories(ctx, req.ParentId)
	if err != nil {
		zap.S().Errorf("failed to list categories for parent ID %d: %v", req.ParentId, err)
		return nil, status.Errorf(codes.Internal, "failed to list categories")
	}

	// 2. 转换为Protobuf类型
	protoCategories := make([]*v1.Category, len(categories))
	for i, c := range categories {
		protoCategories[i] = s.bizCategoryToProto(c)
	}

	zap.S().Infof("Listed %d categories for parent ID %d", len(categories), req.ParentId)
	return &v1.ListCategoriesResponse{
		Categories: protoCategories,
	}, nil
}

// --- 商品品牌接口实现 ---

// CreateBrand 实现了创建品牌的 RPC 方法。
func (s *ProductService) CreateBrand(ctx context.Context, req *v1.CreateBrandRequest) (*v1.Brand, error) {
	zap.S().Infof("CreateBrand request received: %v", req.Name)

	// 1. 参数校验
	if err := s.validator.Struct(req); err != nil {
		zap.S().Warnf("CreateBrand request validation failed: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid argument: %v", err)
	}

	// 2. 检查品牌名称是否重复
	existingBrand, err := s.brandRepo.GetBrandByName(ctx, req.Name)
	if err != nil {
		zap.S().Errorf("failed to check existing brand name %s: %v", req.Name, err)
		return nil, status.Errorf(codes.Internal, "failed to check brand name")
	}
	if existingBrand != nil {
		zap.S().Warnf("brand name '%s' already exists", req.Name)
		return nil, status.Errorf(codes.AlreadyExists, "brand name '%s' already exists", req.Name)
	}

	// 3. 构建品牌模型
	brand := &model.Brand{
		Name:        req.Name,
		LogoURL:     req.LogoUrl,
		Description: req.Description,
	}

	// 4. 调用仓库层创建品牌
	createdBrand, err := s.brandRepo.CreateBrand(ctx, brand)
	if err != nil {
		zap.S().Errorf("failed to create brand %s: %v", req.Name, err)
		return nil, status.Errorf(codes.Internal, "failed to create brand")
	}

	zap.S().Infof("Brand %d created successfully", createdBrand.ID)
	return s.bizBrandToProto(createdBrand), nil
}

// GetBrandByID 实现了根据ID获取品牌信息的 RPC 方法。
func (s *ProductService) GetBrandByID(ctx context.Context, req *v1.GetBrandByIDRequest) (*v1.Brand, error) {
	zap.S().Infof("GetBrandByID request received for brand ID: %d", req.Id)

	// 1. 参数校验
	if req.Id == 0 {
		zap.S().Warn("GetBrandByID request with zero brand ID")
		return nil, status.Errorf(codes.InvalidArgument, "brand ID cannot be zero")
	}

	// 2. 调用仓库层获取品牌
	brand, err := s.brandRepo.GetBrandByID(ctx, req.Id)
	if err != nil {
		zap.S().Errorf("failed to get brand by ID %d: %v", req.Id, err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve brand")
	}
	if brand == nil {
		zap.S().Warnf("brand with ID %d not found", req.Id)
		return nil, status.Errorf(codes.NotFound, "brand not found")
	}

	zap.S().Infof("Brand %d retrieved successfully", req.Id)
	return s.bizBrandToProto(brand), nil
}

// UpdateBrand 实现了更新品牌信息的 RPC 方法。
func (s *ProductService) UpdateBrand(ctx context.Context, req *v1.UpdateBrandRequest) (*v1.Brand, error) {
	zap.S().Infof("UpdateBrand request received for brand ID: %d", req.Id)

	// 1. 参数校验
	if req.Id == 0 {
		zap.S().Warn("UpdateBrand request with zero brand ID")
		return nil, status.Errorf(codes.InvalidArgument, "brand ID cannot be zero")
	}

	// 2. 获取现有品牌信息
	existingBrand, err := s.brandRepo.GetBrandByID(ctx, req.Id)
	if err != nil {
		zap.S().Errorf("failed to get existing brand %d for update: %v", req.Id, err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve brand for update")
	}
	if existingBrand == nil {
		zap.S().Warnf("brand with ID %d not found for update", req.Id)
		return nil, status.Errorf(codes.NotFound, "brand not found")
	}

	// 3. 根据请求更新字段
	if req.Name != nil {
		// 检查新名称是否重复
		if existingBrand.Name != req.Name.GetValue() {
			checkBrand, err := s.brandRepo.GetBrandByName(ctx, req.Name.GetValue())
			if err != nil {
				zap.S().Errorf("failed to check brand name %s: %v", req.Name.GetValue(), err)
				return nil, status.Errorf(codes.Internal, "failed to check brand name")
			}
			if checkBrand != nil {
				zap.S().Warnf("brand name '%s' already exists", req.Name.GetValue())
				return nil, status.Errorf(codes.AlreadyExists, "brand name '%s' already exists", req.Name.GetValue())
			}
		}
		existingBrand.Name = req.Name.GetValue()
	}
	if req.LogoUrl != nil {
		existingBrand.LogoURL = req.LogoUrl.GetValue()
	}
	if req.Description != nil {
		existingBrand.Description = req.Description.GetValue()
	}

	// 4. 调用仓库层更新品牌
	updatedBrand, err := s.brandRepo.UpdateBrand(ctx, existingBrand)
	if err != nil {
		zap.S().Errorf("failed to update brand %d: %v", req.Id, err)
		return nil, status.Errorf(codes.Internal, "failed to update brand")
	}

	zap.S().Infof("Brand %d updated successfully", req.Id)
	return s.bizBrandToProto(updatedBrand), nil
}

// DeleteBrand 实现了删除品牌的 RPC 方法 (逻辑删除)。
func (s *ProductService) DeleteBrand(ctx context.Context, req *v1.DeleteBrandRequest) (*emptypb.Empty, error) {
	zap.S().Infof("DeleteBrand request received for brand ID: %d", req.Id)

	// 1. 参数校验
	if req.Id == 0 {
		zap.S().Warn("DeleteBrand request with zero brand ID")
		return nil, status.Errorf(codes.InvalidArgument, "brand ID cannot be zero")
	}

	// 2. 检查品牌是否存在
	existingBrand, err := s.brandRepo.GetBrandByID(ctx, req.Id)
	if err != nil {
		zap.S().Errorf("failed to get brand %d for deletion: %v", req.Id, err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve brand for deletion")
	}
	if existingBrand == nil {
		zap.S().Warnf("brand with ID %d not found for deletion", req.Id)
		return nil, status.Errorf(codes.NotFound, "brand not found")
	}

	// 3. 检查品牌下是否有商品 (此处仅为示例，实际业务中可能需要更复杂的检查)
	// 假设有一个方法可以检查品牌下的商品数量
	// productCount, err := s.productRepo.CountProductsByBrandID(ctx, req.Id)
	// if err != nil || productCount > 0 {
	// 	return nil, status.Errorf(codes.FailedPrecondition, "brand has associated products, cannot delete")
	// }

	// 4. 调用仓库层逻辑删除品牌
	if err := s.brandRepo.DeleteBrand(ctx, req.Id); err != nil {
		zap.S().Errorf("failed to delete brand %d: %v", req.Id, err)
		return nil, status.Errorf(codes.Internal, "failed to delete brand")
	}

	zap.S().Infof("Brand %d deleted successfully", req.Id)
	return &emptypb.Empty{}, nil
}

// ListBrands 实现了分页列出所有品牌的 RPC 方法。
func (s *ProductService) ListBrands(ctx context.Context, req *v1.ListBrandsRequest) (*v1.ListBrandsResponse, error) {
	zap.S().Infof("ListBrands request received: page=%d, page_size=%d", req.Page, req.PageSize)

	// 1. 参数校验与默认值设置
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = s.defaultPageSize
	}
	if pageSize > s.maxPageSize {
		pageSize = s.maxPageSize
	}
	page := req.Page
	if page == 0 {
		page = 1
	}

	// 2. 调用仓库层查询品牌列表
	brands, total, err := s.brandRepo.ListBrands(ctx, page, pageSize)
	if err != nil {
		zap.S().Errorf("failed to list brands: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to list brands")
	}

	// 3. 转换为Protobuf类型
	protoBrands := make([]*v1.Brand, len(brands))
	for i, b := range brands {
		protoBrands[i] = s.bizBrandToProto(b)
	}

	zap.S().Infof("Listed %d brands (total: %d)", len(brands), total)
	return &v1.ListBrandsResponse{
		Brands: protoBrands,
		Total:  int32(total),
	},
	nil
}

// --- 辅助函数：模型转换 ---

// bizProductToProto 将 model.Product 领域模型转换为 v1.ProductInfo API 模型。
func (s *ProductService) bizProductToProto(ctx context.Context, p *model.Product) (*v1.ProductInfo, error) {
	if p == nil {
		return nil, nil
	}

	protoSKUs := make([]*v1.SKU, len(p.SKUs))
	for i, sku := range p.SKUs {
		protoSKUs[i] = s.bizSKUToProto(&sku)
	}

	protoAttributes := make([]*v1.ProductAttribute, len(p.Attributes))
	for i, attr := range p.Attributes {
		protoAttributes[i] = &v1.ProductAttribute{Key: attr.Key, Value: attr.Value}
	}

	galleryImageUrls := []string{}
	if p.GalleryImageURLs != "" {
		galleryImageUrls = strings.Split(p.GalleryImageURLs, ",")
	}

	return &v1.ProductInfo{
		Id:               p.ID,
		Name:             p.Name,
		SpuNo:            p.SpuNo,
		Description:      p.Description,
		Category:         s.bizCategoryToProto(&p.Category),
		Brand:            s.bizBrandToProto(&p.Brand),
		Status:           v1.ProductStatus(p.Status),
		Skus:             protoSKUs,
		Attributes:       protoAttributes,
		MainImageURL:     p.MainImageURL,
		GalleryImageUrls: galleryImageUrls,
		SeoInfo: &v1.SeoInfo{
			MetaTitle:       p.MetaTitle,
			MetaDescription: p.MetaDescription,
			MetaKeywords:    p.MetaKeywords,
			UrlSlug:         p.URLSlug,
		},
		Weight:           &v1.Google_Protobuf_DoubleValue{Value: p.Weight},
		CreatedAt:        timestamppb.New(p.CreatedAt),
		UpdatedAt:        timestamppb.New(p.UpdatedAt),
	}, nil
}

// bizSKUToProto 将 model.SKU 领域模型转换为 v1.SKU API 模型。
func (s *ProductService) bizSKUToProto(sku *model.SKU) *v1.SKU {
	if sku == nil {
		return nil
	}

	var specValues []*v1.SpecValue
	if sku.SpecValues != "" {
		err := json.Unmarshal([]byte(sku.SpecValues), &specValues)
		if err != nil {
			zap.S().Errorf("failed to unmarshal SKU %d spec values: %v", sku.ID, err)
			// 即使反序列化失败，也返回部分数据，或者根据业务需求决定是否返回错误
		}
	}

	return &v1.SKU{
		Id:            sku.ID,
		ProductId:     sku.ProductID,
		SkuNo:         sku.SkuNo,
		Name:          sku.Name,
		Price:         sku.Price,
		StockQuantity: sku.StockQuantity,
		ImageURL:      sku.ImageURL,
		SpecValues:    specValues,
		CreatedAt:     timestamppb.New(sku.CreatedAt),
		UpdatedAt:     timestamppb.New(sku.UpdatedAt),
	}
}

// bizCategoryToProto 将 model.Category 领域模型转换为 v1.Category API 模型。
func (s *ProductService) bizCategoryToProto(c *model.Category) *v1.Category {
	if c == nil {
		return nil
	}
	return &v1.Category{
		Id:        c.ID,
		Name:      c.Name,
		ParentId:  c.ParentID,
		Level:     c.Level,
		IconUrl:   c.IconURL,
		SortOrder: c.SortOrder,
	}
}

// bizBrandToProto 将 model.Brand 领域模型转换为 v1.Brand API 模型。
func (s *ProductService) bizBrandToProto(b *model.Brand) *v1.Brand {
	if b == nil {
		return nil
	}
	return &v1.Brand{
		Id:          b.ID,
		Name:        b.Name,
		LogoUrl:     b.LogoURL,
		Description: b.Description,
	}
}