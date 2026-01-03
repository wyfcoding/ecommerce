package mysql

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/product/domain"
	"gorm.io/gorm"
)

// ProductRepository 结构体是 ProductRepository 接口的MySQL实现。
type ProductRepository struct {
	db *gorm.DB
}

// NewProductRepository 创建并返回一个新的 ProductRepository 实例。
func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Transaction 实现事务包装
func (r *ProductRepository) Transaction(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// WithTx 返回带事务的副本
func (r *ProductRepository) WithTx(tx any) domain.ProductRepository {
	if tx == nil {
		return r
	}
	return &ProductRepository{db: tx.(*gorm.DB)}
}

// Save 将商品实体保存到数据库。
func (r *ProductRepository) Save(ctx context.Context, product *domain.Product) error {
	db := r.db.WithContext(ctx)
	if err := db.Create(product).Error; err != nil {
		return err
	}
	for _, sku := range product.SKUs {
		sku.ProductID = product.ID
		if err := db.Create(sku).Error; err != nil {
			return err
		}
	}
	return nil
}

// FindByID 根据ID从数据库获取商品记录，并预加载其关联的SKU列表。
func (r *ProductRepository) FindByID(ctx context.Context, id uint) (*domain.Product, error) {
	var product domain.Product
	if err := r.db.WithContext(ctx).Preload("SKUs").First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

// FindByName 根据名称从数据库获取商品记录，并预加载其关联的SKU列表。
func (r *ProductRepository) FindByName(ctx context.Context, name string) (*domain.Product, error) {
	var product domain.Product
	if err := r.db.WithContext(ctx).Preload("SKUs").Where("name = ?", name).First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

// Update 更新商品实体。
func (r *ProductRepository) Update(ctx context.Context, product *domain.Product) error {
	return r.db.WithContext(ctx).Save(product).Error
}

// Delete 根据ID从数据库删除商品记录。
func (r *ProductRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Product{}, id).Error
}

// List 从数据库列出所有商品记录，支持分页。
func (r *ProductRepository) List(ctx context.Context, offset, limit int) ([]*domain.Product, int64, error) {
	var products []*domain.Product
	var total int64

	if err := r.db.WithContext(ctx).Model(&domain.Product{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Preload("SKUs").Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// ListByCategory 从数据库列出指定分类ID下的商品记录。
func (r *ProductRepository) ListByCategory(ctx context.Context, categoryID uint, offset, limit int) ([]*domain.Product, int64, error) {
	var products []*domain.Product
	var total int64

	if err := r.db.WithContext(ctx).Model(&domain.Product{}).Where("category_id = ?", categoryID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Preload("SKUs").Where("category_id = ?", categoryID).Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// ListByBrand 从数据库列出指定品牌ID下的商品记录。
func (r *ProductRepository) ListByBrand(ctx context.Context, brandID uint, offset, limit int) ([]*domain.Product, int64, error) {
	var products []*domain.Product
	var total int64

	if err := r.db.WithContext(ctx).Model(&domain.Product{}).Where("brand_id = ?", brandID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Preload("SKUs").Where("brand_id = ?", brandID).Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// SKURepository 结构体是 SKURepository 接口的MySQL实现。
type SKURepository struct {
	db *gorm.DB
}

func NewSKURepository(db *gorm.DB) *SKURepository {
	return &SKURepository{db: db}
}

func (r *SKURepository) Save(ctx context.Context, sku *domain.SKU) error {
	return r.db.WithContext(ctx).Create(sku).Error
}

func (r *SKURepository) FindByID(ctx context.Context, id uint) (*domain.SKU, error) {
	var sku domain.SKU
	if err := r.db.WithContext(ctx).First(&sku, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sku, nil
}

func (r *SKURepository) FindByProductID(ctx context.Context, productID uint) ([]*domain.SKU, error) {
	var skus []*domain.SKU
	if err := r.db.WithContext(ctx).Where("product_id = ?", productID).Find(&skus).Error; err != nil {
		return nil, err
	}
	return skus, nil
}

func (r *SKURepository) Update(ctx context.Context, sku *domain.SKU) error {
	return r.db.WithContext(ctx).Save(sku).Error
}

func (r *SKURepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.SKU{}, id).Error
}

// CategoryRepository 结构体是 CategoryRepository 接口的MySQL实现。
type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Save(ctx context.Context, category *domain.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *CategoryRepository) FindByID(ctx context.Context, id uint) (*domain.Category, error) {
	var category domain.Category
	if err := r.db.WithContext(ctx).First(&category, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

func (r *CategoryRepository) FindByName(ctx context.Context, name string) (*domain.Category, error) {
	var category domain.Category
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

func (r *CategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

func (r *CategoryRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Category{}, id).Error
}

func (r *CategoryRepository) List(ctx context.Context) ([]*domain.Category, error) {
	var categories []*domain.Category
	if err := r.db.WithContext(ctx).Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *CategoryRepository) FindByParentID(ctx context.Context, parentID uint) ([]*domain.Category, error) {
	var categories []*domain.Category
	if err := r.db.WithContext(ctx).Where("parent_id = ?", parentID).Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// BrandRepository 结构体是 BrandRepository 接口的MySQL实现。
type BrandRepository struct {
	db *gorm.DB
}

func NewBrandRepository(db *gorm.DB) *BrandRepository {
	return &BrandRepository{db: db}
}

func (r *BrandRepository) Save(ctx context.Context, brand *domain.Brand) error {
	return r.db.WithContext(ctx).Create(brand).Error
}

func (r *BrandRepository) FindByID(ctx context.Context, id uint) (*domain.Brand, error) {
	var brand domain.Brand
	if err := r.db.WithContext(ctx).First(&brand, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &brand, nil
}

func (r *BrandRepository) FindByName(ctx context.Context, name string) (*domain.Brand, error) {
	var brand domain.Brand
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&brand).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &brand, nil
}

func (r *BrandRepository) Update(ctx context.Context, brand *domain.Brand) error {
	return r.db.WithContext(ctx).Save(brand).Error
}

func (r *BrandRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Brand{}, id).Error
}

func (r *BrandRepository) List(ctx context.Context) ([]*domain.Brand, error) {
	var brands []*domain.Brand
	if err := r.db.WithContext(ctx).Find(&brands).Error; err != nil {
		return nil, err
	}
	return brands, nil
}
