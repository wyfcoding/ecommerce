package repository

import (
	"context"
	"ecommerce/internal/product/model"
	"gorm.io/gorm"
)

type ProductRepository interface {
	CreateProduct(ctx context.Context, product *model.Product) error
	GetProductByID(ctx context.Context, id uint64) (*model.Product, error)
	UpdateProduct(ctx context.Context, product *model.Product) error
	DeleteProduct(ctx context.Context, id uint64) error
	ListProducts(ctx context.Context, offset, limit int) ([]*model.Product, int64, error)
	
	CreateSKU(ctx context.Context, sku *model.SKU) error
	GetSKUByID(ctx context.Context, id uint64) (*model.SKU, error)
	UpdateSKU(ctx context.Context, sku *model.SKU) error
	ListSKUsByProductID(ctx context.Context, productID uint64) ([]*model.SKU, error)
}

type CategoryRepository interface {
	CreateCategory(ctx context.Context, category *model.Category) error
	GetCategoryByID(ctx context.Context, id uint64) (*model.Category, error)
	UpdateCategory(ctx context.Context, category *model.Category) error
	DeleteCategory(ctx context.Context, id uint64) error
	ListCategories(ctx context.Context) ([]*model.Category, error)
}

type BrandRepository interface {
	CreateBrand(ctx context.Context, brand *model.Brand) error
	GetBrandByID(ctx context.Context, id uint64) (*model.Brand, error)
	UpdateBrand(ctx context.Context, brand *model.Brand) error
	DeleteBrand(ctx context.Context, id uint64) error
	ListBrands(ctx context.Context) ([]*model.Brand, error)
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) CreateProduct(ctx context.Context, product *model.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

func (r *productRepository) GetProductByID(ctx context.Context, id uint64) (*model.Product, error) {
	var product model.Product
	err := r.db.WithContext(ctx).First(&product, id).Error
	return &product, err
}

func (r *productRepository) UpdateProduct(ctx context.Context, product *model.Product) error {
	return r.db.WithContext(ctx).Save(product).Error
}

func (r *productRepository) DeleteProduct(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Product{}, id).Error
}

func (r *productRepository) ListProducts(ctx context.Context, offset, limit int) ([]*model.Product, int64, error) {
	var products []*model.Product
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.Product{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&products).Error
	return products, total, err
}

func (r *productRepository) CreateSKU(ctx context.Context, sku *model.SKU) error {
	return r.db.WithContext(ctx).Create(sku).Error
}

func (r *productRepository) GetSKUByID(ctx context.Context, id uint64) (*model.SKU, error) {
	var sku model.SKU
	err := r.db.WithContext(ctx).First(&sku, id).Error
	return &sku, err
}

func (r *productRepository) UpdateSKU(ctx context.Context, sku *model.SKU) error {
	return r.db.WithContext(ctx).Save(sku).Error
}

func (r *productRepository) ListSKUsByProductID(ctx context.Context, productID uint64) ([]*model.SKU, error) {
	var skus []*model.SKU
	err := r.db.WithContext(ctx).Where("product_id = ?", productID).Find(&skus).Error
	return skus, err
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) CreateCategory(ctx context.Context, category *model.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *categoryRepository) GetCategoryByID(ctx context.Context, id uint64) (*model.Category, error) {
	var category model.Category
	err := r.db.WithContext(ctx).First(&category, id).Error
	return &category, err
}

func (r *categoryRepository) UpdateCategory(ctx context.Context, category *model.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

func (r *categoryRepository) DeleteCategory(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Category{}, id).Error
}

func (r *categoryRepository) ListCategories(ctx context.Context) ([]*model.Category, error) {
	var categories []*model.Category
	err := r.db.WithContext(ctx).Order("sort ASC").Find(&categories).Error
	return categories, err
}

type brandRepository struct {
	db *gorm.DB
}

func NewBrandRepository(db *gorm.DB) BrandRepository {
	return &brandRepository{db: db}
}

func (r *brandRepository) CreateBrand(ctx context.Context, brand *model.Brand) error {
	return r.db.WithContext(ctx).Create(brand).Error
}

func (r *brandRepository) GetBrandByID(ctx context.Context, id uint64) (*model.Brand, error) {
	var brand model.Brand
	err := r.db.WithContext(ctx).First(&brand, id).Error
	return &brand, err
}

func (r *brandRepository) UpdateBrand(ctx context.Context, brand *model.Brand) error {
	return r.db.WithContext(ctx).Save(brand).Error
}

func (r *brandRepository) DeleteBrand(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Brand{}, id).Error
}

func (r *brandRepository) ListBrands(ctx context.Context) ([]*model.Brand, error) {
	var brands []*model.Brand
	err := r.db.WithContext(ctx).Find(&brands).Error
	return brands, err
}
