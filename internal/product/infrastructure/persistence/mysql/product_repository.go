package mysql

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/product/domain"
	"gorm.io/gorm"
)

// ProductRepository 结构体是 ProductRepository 接口的 MySQL 实现。
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
	if product == nil {
		return nil
	}
	model := toProductModel(product)
	db := r.db.WithContext(ctx)
	if model.ID == 0 {
		if err := db.Create(model).Error; err != nil {
			return err
		}
	} else {
		if err := db.Save(model).Error; err != nil {
			return err
		}
	}

	// 保存 SKU
	for _, sku := range product.SKUs {
		if sku == nil {
			continue
		}
		sku.ProductID = model.ID
		skuModel := toSKUModel(sku)
		if skuModel.ID == 0 {
			if err := db.Create(skuModel).Error; err != nil {
				return err
			}
		} else {
			if err := db.Save(skuModel).Error; err != nil {
				return err
			}
		}
		if synced := toDomainSKU(skuModel); synced != nil {
			*sku = *synced
		}
	}

	if synced := toDomainProduct(model); synced != nil {
		*product = *synced
	}
	return nil
}

// FindByID 根据ID从数据库获取商品记录，并预加载其关联的SKU列表。
func (r *ProductRepository) FindByID(ctx context.Context, id uint) (*domain.Product, error) {
	var product ProductModel
	if err := r.db.WithContext(ctx).Preload("SKUs").First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainProduct(&product), nil
}

// FindByName 根据名称从数据库获取商品记录，并预加载其关联的SKU列表。
func (r *ProductRepository) FindByName(ctx context.Context, name string) (*domain.Product, error) {
	var product ProductModel
	if err := r.db.WithContext(ctx).Preload("SKUs").Where("name = ?", name).First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainProduct(&product), nil
}

// Update 更新商品实体。
func (r *ProductRepository) Update(ctx context.Context, product *domain.Product) error {
	if product == nil {
		return nil
	}
	model := toProductModel(product)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	if synced := toDomainProduct(model); synced != nil {
		*product = *synced
	}
	return nil
}

// Delete 根据ID从数据库删除商品记录。
func (r *ProductRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&ProductModel{}, id).Error
}

// List 从数据库列出所有商品记录，支持分页。
func (r *ProductRepository) List(ctx context.Context, offset, limit int) ([]*domain.Product, int64, error) {
	var products []*ProductModel
	var total int64

	if err := r.db.WithContext(ctx).Model(&ProductModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Preload("SKUs").Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	result := make([]*domain.Product, 0, len(products))
	for _, p := range products {
		result = append(result, toDomainProduct(p))
	}

	return result, total, nil
}

// ListByCategory 从数据库列出指定分类ID下的商品记录。
func (r *ProductRepository) ListByCategory(ctx context.Context, categoryID uint, offset, limit int) ([]*domain.Product, int64, error) {
	var products []*ProductModel
	var total int64

	if err := r.db.WithContext(ctx).Model(&ProductModel{}).Where("category_id = ?", categoryID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Preload("SKUs").Where("category_id = ?", categoryID).Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	result := make([]*domain.Product, 0, len(products))
	for _, p := range products {
		result = append(result, toDomainProduct(p))
	}

	return result, total, nil
}

// ListByBrand 从数据库列出指定品牌ID下的商品记录。
func (r *ProductRepository) ListByBrand(ctx context.Context, brandID uint, offset, limit int) ([]*domain.Product, int64, error) {
	var products []*ProductModel
	var total int64

	if err := r.db.WithContext(ctx).Model(&ProductModel{}).Where("brand_id = ?", brandID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Preload("SKUs").Where("brand_id = ?", brandID).Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	result := make([]*domain.Product, 0, len(products))
	for _, p := range products {
		result = append(result, toDomainProduct(p))
	}

	return result, total, nil
}

// SKURepository 结构体是 SKURepository 接口的 MySQL 实现。
type SKURepository struct {
	db *gorm.DB
}

func NewSKURepository(db *gorm.DB) *SKURepository {
	return &SKURepository{db: db}
}

func (r *SKURepository) Transaction(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

func (r *SKURepository) WithTx(tx any) domain.SKURepository {
	if tx == nil {
		return r
	}
	return &SKURepository{db: tx.(*gorm.DB)}
}

func (r *SKURepository) Save(ctx context.Context, sku *domain.SKU) error {
	if sku == nil {
		return nil
	}
	model := toSKUModel(sku)
	if model.ID == 0 {
		if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
			return err
		}
	} else {
		if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
			return err
		}
	}
	if synced := toDomainSKU(model); synced != nil {
		*sku = *synced
	}
	return nil
}

func (r *SKURepository) FindByID(ctx context.Context, id uint) (*domain.SKU, error) {
	var sku SKUModel
	if err := r.db.WithContext(ctx).First(&sku, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainSKU(&sku), nil
}

func (r *SKURepository) FindByProductID(ctx context.Context, productID uint) ([]*domain.SKU, error) {
	var skus []*SKUModel
	if err := r.db.WithContext(ctx).Where("product_id = ?", productID).Find(&skus).Error; err != nil {
		return nil, err
	}
	result := make([]*domain.SKU, 0, len(skus))
	for _, s := range skus {
		result = append(result, toDomainSKU(s))
	}
	return result, nil
}

func (r *SKURepository) Update(ctx context.Context, sku *domain.SKU) error {
	return r.Save(ctx, sku)
}

func (r *SKURepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&SKUModel{}, id).Error
}

// CategoryRepository 结构体是 CategoryRepository 接口的 MySQL 实现。
type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Transaction(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

func (r *CategoryRepository) WithTx(tx any) domain.CategoryRepository {
	if tx == nil {
		return r
	}
	return &CategoryRepository{db: tx.(*gorm.DB)}
}

func (r *CategoryRepository) Save(ctx context.Context, category *domain.Category) error {
	if category == nil {
		return nil
	}
	model := toCategoryModel(category)
	if model.ID == 0 {
		if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
			return err
		}
	} else {
		if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
			return err
		}
	}
	if synced := toDomainCategory(model); synced != nil {
		*category = *synced
	}
	return nil
}

func (r *CategoryRepository) FindByID(ctx context.Context, id uint) (*domain.Category, error) {
	var category CategoryModel
	if err := r.db.WithContext(ctx).First(&category, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainCategory(&category), nil
}

func (r *CategoryRepository) FindByName(ctx context.Context, name string) (*domain.Category, error) {
	var category CategoryModel
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainCategory(&category), nil
}

func (r *CategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	return r.Save(ctx, category)
}

func (r *CategoryRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&CategoryModel{}, id).Error
}

func (r *CategoryRepository) List(ctx context.Context) ([]*domain.Category, error) {
	var categories []*CategoryModel
	if err := r.db.WithContext(ctx).Find(&categories).Error; err != nil {
		return nil, err
	}
	result := make([]*domain.Category, 0, len(categories))
	for _, c := range categories {
		result = append(result, toDomainCategory(c))
	}
	return result, nil
}

func (r *CategoryRepository) FindByParentID(ctx context.Context, parentID uint) ([]*domain.Category, error) {
	var categories []*CategoryModel
	if err := r.db.WithContext(ctx).Where("parent_id = ?", parentID).Find(&categories).Error; err != nil {
		return nil, err
	}
	result := make([]*domain.Category, 0, len(categories))
	for _, c := range categories {
		result = append(result, toDomainCategory(c))
	}
	return result, nil
}

// BrandRepository 结构体是 BrandRepository 接口的 MySQL 实现。
type BrandRepository struct {
	db *gorm.DB
}

func NewBrandRepository(db *gorm.DB) *BrandRepository {
	return &BrandRepository{db: db}
}

func (r *BrandRepository) Transaction(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

func (r *BrandRepository) WithTx(tx any) domain.BrandRepository {
	if tx == nil {
		return r
	}
	return &BrandRepository{db: tx.(*gorm.DB)}
}

func (r *BrandRepository) Save(ctx context.Context, brand *domain.Brand) error {
	if brand == nil {
		return nil
	}
	model := toBrandModel(brand)
	if model.ID == 0 {
		if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
			return err
		}
	} else {
		if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
			return err
		}
	}
	if synced := toDomainBrand(model); synced != nil {
		*brand = *synced
	}
	return nil
}

func (r *BrandRepository) FindByID(ctx context.Context, id uint) (*domain.Brand, error) {
	var brand BrandModel
	if err := r.db.WithContext(ctx).First(&brand, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainBrand(&brand), nil
}

func (r *BrandRepository) FindByName(ctx context.Context, name string) (*domain.Brand, error) {
	var brand BrandModel
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&brand).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainBrand(&brand), nil
}

func (r *BrandRepository) Update(ctx context.Context, brand *domain.Brand) error {
	return r.Save(ctx, brand)
}

func (r *BrandRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&BrandModel{}, id).Error
}

func (r *BrandRepository) List(ctx context.Context) ([]*domain.Brand, error) {
	var brands []*BrandModel
	if err := r.db.WithContext(ctx).Find(&brands).Error; err != nil {
		return nil, err
	}
	result := make([]*domain.Brand, 0, len(brands))
	for _, b := range brands {
		result = append(result, toDomainBrand(b))
	}
	return result, nil
}
