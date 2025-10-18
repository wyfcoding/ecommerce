package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"ecommerce/internal/product/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ProductRepo 定义了商品 (SPU) 数据的存储接口。
type ProductRepo interface {
	// CreateProduct 创建一个新的商品。
	CreateProduct(ctx context.Context, product *model.Product) (*model.Product, error)
	// GetProductByID 根据ID获取商品详情，包含关联的SKU、分类和品牌信息。
	GetProductByID(ctx context.Context, id uint64) (*model.Product, error)
	// UpdateProduct 更新商品信息。
	UpdateProduct(ctx context.Context, product *model.Product) (*model.Product, error)
	// DeleteProduct 逻辑删除商品。
	DeleteProduct(ctx context.Context, id uint64) error
	// ListProducts 根据条件分页查询商品列表。
	ListProducts(ctx context.Context, query *ProductListQuery) ([]*model.Product, int64, error)
	// GetProductBySpuNo 根据SPU编码获取商品。
	GetProductBySpuNo(ctx context.Context, spuNo string) (*model.Product, error)
}

// SKURepo 定义了SKU数据的存储接口。
type SKURepo interface {
	// CreateSKUs 批量创建SKU。
	CreateSKUs(ctx context.Context, skus []*model.SKU) ([]*model.SKU, error)
	// UpdateSKU 更新单个SKU信息。
	UpdateSKU(ctx context.Context, sku *model.SKU) (*model.SKU, error)
	// DeleteSKUs 批量删除SKU。
	DeleteSKUs(ctx context.Context, productID uint64, skuIDs []uint64) error
	// GetSKUByID 根据ID获取SKU详情。
	GetSKUByID(ctx context.Context, id uint64) (*model.SKU, error)
	// ListSKUsByProductID 根据商品ID获取所有SKU。
	ListSKUsByProductID(ctx context.Context, productID uint64) ([]*model.SKU, error)
	// GetSKUBySkuNo 根据SKU编码获取SKU。
	GetSKUBySkuNo(ctx context.Context, skuNo string) (*model.SKU, error)
}

// CategoryRepo 定义了商品分类数据的存储接口。
type CategoryRepo interface {
	// CreateCategory 创建一个新的分类。
	CreateCategory(ctx context.Context, category *model.Category) (*model.Category, error)
	// GetCategoryByID 根据ID获取分类详情。
	GetCategoryByID(ctx context.Context, id uint64) (*model.Category, error)
	// UpdateCategory 更新分类信息。
	UpdateCategory(ctx context.Context, category *model.Category) (*model.Category, error)
	// DeleteCategory 逻辑删除分类。
	DeleteCategory(ctx context.Context, id uint64) error
	// ListCategories 根据父分类ID获取子分类列表。
	ListCategories(ctx context.Context, parentID uint64) ([]*model.Category, error)
	// GetCategoryByName 根据名称获取分类。
	GetCategoryByName(ctx context.Context, name string) (*model.Category, error)
}

// BrandRepo 定义了商品品牌数据的存储接口。
type BrandRepo interface {
	// CreateBrand 创建一个新的品牌。
	CreateBrand(ctx context.Context, brand *model.Brand) (*model.Brand, error)
	// GetBrandByID 根据ID获取品牌详情。
	GetBrandByID(ctx context.Context, id uint64) (*model.Brand, error)
	// UpdateBrand 更新品牌信息。
	UpdateBrand(ctx context.Context, brand *model.Brand) (*model.Brand, error)
	// DeleteBrand 逻辑删除品牌。
	DeleteBrand(ctx context.Context, id uint64) error
	// ListBrands 分页查询品牌列表。
	ListBrands(ctx context.Context, page, pageSize int32) ([]*model.Brand, int64, error)
	// GetBrandByName 根据名称获取品牌。
	GetBrandByName(ctx context.Context, name string) (*model.Brand, error)
}

// ProductListQuery 定义商品列表查询的参数。
type ProductListQuery struct {
	Page       int32
	PageSize   int32
	CategoryID uint64
	BrandID    uint64
	Status     model.ProductStatus
	SortBy     string // 例如: "price_asc", "created_at_desc"
}

// productRepoImpl 是 ProductRepo 接口的 GORM 实现。
type productRepoImpl struct {
	db *gorm.DB
}

// NewProductRepo 创建一个新的 ProductRepo 实例。
func NewProductRepo(db *gorm.DB) ProductRepo {
	return &productRepoImpl{db: db}
}

// CreateProduct 实现 ProductRepo.CreateProduct 方法。
func (r *productRepoImpl) CreateProduct(ctx context.Context, product *model.Product) (*model.Product, error) {
	// 确保关联数据也一并保存
	if err := r.db.WithContext(ctx).Create(product).Error; err != nil {
		zap.S().Errorf("failed to create product: %v", err)
		return nil, fmt.Errorf("failed to create product: %w", err)
	}
	return product, nil
}

// GetProductByID 实现 ProductRepo.GetProductByID 方法。
func (r *productRepoImpl) GetProductByID(ctx context.Context, id uint64) (*model.Product, error) {
	var product model.Product
	// 预加载关联的SKU、分类和品牌信息
	if err := r.db.WithContext(ctx).Preload("SKUs").Preload("Category").Preload("Brand").Preload("Attributes").First(&product, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 未找到记录
		}
		zap.S().Errorf("failed to get product by id %d: %v", id, err)
		return nil, fmt.Errorf("failed to get product by id: %w", err)
	}
	return &product, nil
}

// UpdateProduct 实现 ProductRepo.UpdateProduct 方法。
func (r *productRepoImpl) UpdateProduct(ctx context.Context, product *model.Product) (*model.Product, error) {
	// 使用 Select 指定更新字段，避免更新零值字段
	// 复杂的更新逻辑可能需要先查询再更新，或者使用 map[string]interface{} 进行部分更新
	if err := r.db.WithContext(ctx).Save(product).Error; err != nil {
		zap.S().Errorf("failed to update product %d: %v", product.ID, err)
		return nil, fmt.Errorf("failed to update product: %w", err)
	}
	return product, nil
}

// DeleteProduct 实现 ProductRepo.DeleteProduct 方法 (逻辑删除)。
func (r *productRepoImpl) DeleteProduct(ctx context.Context, id uint64) error {
	// GORM 的 Delete 方法默认执行软删除 (如果模型包含 gorm.DeletedAt 字段)
	if err := r.db.WithContext(ctx).Delete(&model.Product{}, id).Error; err != nil {
		zap.S().Errorf("failed to delete product %d: %v", id, err)
		return fmt.Errorf("failed to delete product: %w", err)
	}
	return nil
}

// ListProducts 实现 ProductRepo.ListProducts 方法。
func (r *productRepoImpl) ListProducts(ctx context.Context, query *ProductListQuery) ([]*model.Product, int64, error) {
	var products []*model.Product
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Product{}).Preload("Category").Preload("Brand")

	// 应用筛选条件
	if query.CategoryID != 0 {
		db = db.Where("category_id = ?", query.CategoryID)
	}
	if query.BrandID != 0 {
		db = db.Where("brand_id = ?", query.BrandID)
	}
	// 状态筛选，排除未指定状态
	if query.Status != model.ProductStatusUnspecified {
		db = db.Where("status = ?", query.Status)
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		zap.S().Errorf("failed to count products: %v", err)
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// 应用排序
	if query.SortBy != "" {
		sortOrder := ""
		switch query.SortBy {
		case "price_asc":
			sortOrder = "price ASC" // 注意: SPU没有直接价格，这里可能需要关联SKU或取最低价
		case "price_desc":
			sortOrder = "price DESC"
		case "created_at_asc":
			sortOrder = "created_at ASC"
		case "created_at_desc":
			sortOrder = "created_at DESC"
		default:
			zap.S().Warnf("unsupported sort_by parameter: %s", query.SortBy)
		}
		if sortOrder != "" {
			db = db.Order(sortOrder)
		}
	}

	// 应用分页
	if query.PageSize > 0 && query.Page > 0 {
		offset := (query.Page - 1) * query.PageSize
		db = db.Limit(int(query.PageSize)).Offset(int(offset))
	}

	// 查询数据
	if err := db.Find(&products).Error; err != nil {
		zap.S().Errorf("failed to list products: %v", err)
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}

	return products, total, nil
}

// GetProductBySpuNo 实现 ProductRepo.GetProductBySpuNo 方法。
func (r *productRepoImpl) GetProductBySpuNo(ctx context.Context, spuNo string) (*model.Product, error) {
	var product model.Product
	if err := r.db.WithContext(ctx).Where("spu_no = ?", spuNo).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		zap.S().Errorf("failed to get product by spu_no %s: %v", spuNo, err)
		return nil, fmt.Errorf("failed to get product by spu_no: %w", err)
	}
	return &product, nil
}

// skuRepoImpl 是 SKURepo 接口的 GORM 实现。
type skuRepoImpl struct {
	db *gorm.DB
}

// NewSKURepo 创建一个新的 SKURepo 实例。
func NewSKURepo(db *gorm.DB) SKURepo {
	return &skuRepoImpl{db: db}
}

// CreateSKUs 实现 SKURepo.CreateSKUs 方法。
func (r *skuRepoImpl) CreateSKUs(ctx context.Context, skus []*model.SKU) ([]*model.SKU, error) {
	if err := r.db.WithContext(ctx).Create(&skus).Error; err != nil {
		zap.S().Errorf("failed to create skus: %v", err)
		return nil, fmt.Errorf("failed to create skus: %w", err)
	}
	return skus, nil
}

// UpdateSKU 实现 SKURepo.UpdateSKU 方法。
func (r *skuRepoImpl) UpdateSKU(ctx context.Context, sku *model.SKU) (*model.SKU, error) {
	// 仅更新非零值字段，或者使用 Select 指定更新字段
	if err := r.db.WithContext(ctx).Save(sku).Error; err != nil {
		zap.S().Errorf("failed to update sku %d: %v", sku.ID, err)
		return nil, fmt.Errorf("failed to update sku: %w", err)
	}
	return sku, nil
}

// DeleteSKUs 实现 SKURepo.DeleteSKUs 方法 (逻辑删除)。
func (r *skuRepoImpl) DeleteSKUs(ctx context.Context, productID uint64, skuIDs []uint64) error {
	// 确保删除的SKU属于指定的商品
	if err := r.db.WithContext(ctx).Where("product_id = ? AND id IN (?) ", productID, skuIDs).Delete(&model.SKU{}).Error; err != nil {
		zap.S().Errorf("failed to delete skus for product %d, sku_ids %v: %v", productID, skuIDs, err)
		return fmt.Errorf("failed to delete skus: %w", err)
	}
	return nil
}

// GetSKUByID 实现 SKURepo.GetSKUByID 方法。
func (r *skuRepoImpl) GetSKUByID(ctx context.Context, id uint64) (*model.SKU, error) {
	var sku model.SKU
	if err := r.db.WithContext(ctx).First(&sku, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		zap.S().Errorf("failed to get sku by id %d: %v", id, err)
		return nil, fmt.Errorf("failed to get sku by id: %w", err)
	}
	return &sku, nil
}

// ListSKUsByProductID 实现 SKURepo.ListSKUsByProductID 方法。
func (r *skuRepoImpl) ListSKUsByProductID(ctx context.Context, productID uint64) ([]*model.SKU, error) {
	var skus []*model.SKU
	if err := r.db.WithContext(ctx).Where("product_id = ?", productID).Find(&skus).Error; err != nil {
		zap.S().Errorf("failed to list skus for product %d: %v", productID, err)
		return nil, fmt.Errorf("failed to list skus by product id: %w", err)
	}
	return skus, nil
}

// GetSKUBySkuNo 实现 SKURepo.GetSKUBySkuNo 方法。
func (r *skuRepoImpl) GetSKUBySkuNo(ctx context.Context, skuNo string) (*model.SKU, error) {
	var sku model.SKU
	if err := r.db.WithContext(ctx).Where("sku_no = ?", skuNo).First(&sku).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		zap.S().Errorf("failed to get sku by sku_no %s: %v", skuNo, err)
		return nil, fmt.Errorf("failed to get sku by sku_no: %w", err)
	}
	return &sku, nil
}

// categoryRepoImpl 是 CategoryRepo 接口的 GORM 实现。
type categoryRepoImpl struct {
	db *gorm.DB
}

// NewCategoryRepo 创建一个新的 CategoryRepo 实例。
func NewCategoryRepo(db *gorm.DB) CategoryRepo {
	return &categoryRepoImpl{db: db}
}

// CreateCategory 实现 CategoryRepo.CreateCategory 方法。
func (r *categoryRepoImpl) CreateCategory(ctx context.Context, category *model.Category) (*model.Category, error) {
	// 自动计算层级
	if category.ParentID != 0 {
		parent, err := r.GetCategoryByID(ctx, category.ParentID)
		if err != nil || parent == nil {
			zap.S().Errorf("parent category %d not found or error: %v", category.ParentID, err)
			return nil, fmt.Errorf("parent category not found")
		}
		category.Level = parent.Level + 1
	} else {
		category.Level = 1
	}

	if err := r.db.WithContext(ctx).Create(category).Error; err != nil {
		zap.S().Errorf("failed to create category: %v", err)
		return nil, fmt.Errorf("failed to create category: %w", err)
	}
	return category, nil
}

// GetCategoryByID 实现 CategoryRepo.GetCategoryByID 方法。
func (r *categoryRepoImpl) GetCategoryByID(ctx context.Context, id uint64) (*model.Category, error) {
	var category model.Category
	if err := r.db.WithContext(ctx).First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		zap.S().Errorf("failed to get category by id %d: %v", id, err)
		return nil, fmt.Errorf("failed to get category by id: %w", err)
	}
	return &category, nil
}

// UpdateCategory 实现 CategoryRepo.UpdateCategory 方法。
func (r *categoryRepoImpl) UpdateCategory(ctx context.Context, category *model.Category) (*model.Category, error) {
	// 自动更新层级，如果 ParentID 改变
	if category.ParentID != 0 {
		parent, err := r.GetCategoryByID(ctx, category.ParentID)
		if err != nil || parent == nil {
			zap.S().Errorf("parent category %d not found or error: %v", category.ParentID, err)
			return nil, fmt.Errorf("parent category not found")
		}
		category.Level = parent.Level + 1
	} else {
		category.Level = 1
	}

	if err := r.db.WithContext(ctx).Save(category).Error; err != nil {
		zap.S().Errorf("failed to update category %d: %v", category.ID, err)
		return nil, fmt.Errorf("failed to update category: %w", err)
	}
	return category, nil
}

// DeleteCategory 实现 CategoryRepo.DeleteCategory 方法 (逻辑删除)。
func (r *categoryRepoImpl) DeleteCategory(ctx context.Context, id uint64) error {
	// 检查是否有子分类或关联商品，实际业务中可能需要更复杂的检查
	if err := r.db.WithContext(ctx).Delete(&model.Category{}, id).Error; err != nil {
		zap.S().Errorf("failed to delete category %d: %v", id, err)
		return fmt.Errorf("failed to delete category: %w", err)
	}
	return nil
}

// ListCategories 实现 CategoryRepo.ListCategories 方法。
func (r *categoryRepoImpl) ListCategories(ctx context.Context, parentID uint64) ([]*model.Category, error) {
	var categories []*model.Category
	db := r.db.WithContext(ctx).Order("sort_order ASC, id ASC")

	if parentID == 0 {
		// 获取所有顶级分类
		db = db.Where("parent_id = ?", 0)
	} else {
		// 获取指定父分类下的子分类
		db = db.Where("parent_id = ?", parentID)
	}

	if err := db.Find(&categories).Error; err != nil {
		zap.S().Errorf("failed to list categories for parent_id %d: %v", parentID, err)
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	return categories, nil
}

// GetCategoryByName 实现 CategoryRepo.GetCategoryByName 方法。
func (r *categoryRepoImpl) GetCategoryByName(ctx context.Context, name string) (*model.Category, error) {
	var category model.Category
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		zap.S().Errorf("failed to get category by name %s: %v", name, err)
		return nil, fmt.Errorf("failed to get category by name: %w", err)
	}
	return &category, nil
}

// brandRepoImpl 是 BrandRepo 接口的 GORM 实现。
type brandRepoImpl struct {
	db *gorm.DB
}

// NewBrandRepo 创建一个新的 BrandRepo 实例。
func NewBrandRepo(db *gorm.DB) BrandRepo {
	return &brandRepoImpl{db: db}
}

// CreateBrand 实现 BrandRepo.CreateBrand 方法。
func (r *brandRepoImpl) CreateBrand(ctx context.Context, brand *model.Brand) (*model.Brand, error) {
	if err := r.db.WithContext(ctx).Create(brand).Error; err != nil {
		zap.S().Errorf("failed to create brand: %v", err)
		return nil, fmt.Errorf("failed to create brand: %w", err)
	}
	return brand, nil
}

// GetBrandByID 实现 BrandRepo.GetBrandByID 方法。
func (r *brandRepoImpl) GetBrandByID(ctx context.Context, id uint64) (*model.Brand, error) {
	var brand model.Brand
	if err := r.db.WithContext(ctx).First(&brand, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		zap.S().Errorf("failed to get brand by id %d: %v", id, err)
		return nil, fmt.Errorf("failed to get brand by id: %w", err)
	}
	return &brand, nil
}

// UpdateBrand 实现 BrandRepo.UpdateBrand 方法。
func (r *brandRepoImpl) UpdateBrand(ctx context.Context, brand *model.Brand) (*model.Brand, error) {
	if err := r.db.WithContext(ctx).Save(brand).Error; err != nil {
		zap.S().Errorf("failed to update brand %d: %v", brand.ID, err)
		return nil, fmt.Errorf("failed to update brand: %w", err)
	}
	return brand, nil
}

// DeleteBrand 实现 BrandRepo.DeleteBrand 方法 (逻辑删除)。
func (r *brandRepoImpl) DeleteBrand(ctx context.Context, id uint64) error {
	// 实际业务中可能需要检查品牌下是否有商品
	if err := r.db.WithContext(ctx).Delete(&model.Brand{}, id).Error; err != nil {
		zap.S().Errorf("failed to delete brand %d: %v", id, err)
		return fmt.Errorf("failed to delete brand: %w", err)
	}
	return nil
}

// ListBrands 实现 BrandRepo.ListBrands 方法。
func (r *brandRepoImpl) ListBrands(ctx context.Context, page, pageSize int32) ([]*model.Brand, int64, error) {
	var brands []*model.Brand
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Brand{})

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		zap.S().Errorf("failed to count brands: %v", err)
		return nil, 0, fmt.Errorf("failed to count brands: %w", err)
	}

	// 应用分页
	if pageSize > 0 && page > 0 {
		offset := (page - 1) * pageSize
		db = db.Limit(int(pageSize)).Offset(int(offset))
	}

	if err := db.Find(&brands).Error; err != nil {
		zap.S().Errorf("failed to list brands: %v", err)
		return nil, 0, fmt.Errorf("failed to list brands: %w", err)
	}

	return brands, total, nil
}
