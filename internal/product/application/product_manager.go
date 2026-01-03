package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/product/domain"
	"github.com/wyfcoding/pkg/cache"
	"github.com/wyfcoding/pkg/messagequeue/outbox"
	"gorm.io/gorm"
)

type ProductManager struct {
	repo         domain.ProductRepository
	skuRepo      domain.SKURepository
	brandRepo    domain.BrandRepository
	categoryRepo domain.CategoryRepository
	cache        cache.Cache
	outbox       *outbox.Manager
	db           *gorm.DB
	logger       *slog.Logger
}

func NewProductManager(
	repo domain.ProductRepository,
	skuRepo domain.SKURepository,
	brandRepo domain.BrandRepository,
	categoryRepo domain.CategoryRepository,
	cache cache.Cache,
	outboxMgr *outbox.Manager,
	db *gorm.DB,
	logger *slog.Logger,
) *ProductManager {
	return &ProductManager{
		repo:         repo,
		skuRepo:      skuRepo,
		brandRepo:    brandRepo,
		categoryRepo: categoryRepo,
		cache:        cache,
		outbox:       outboxMgr,
		db:           db,
		logger:       logger,
	}
}

// ---------------- Product ----------------

func (m *ProductManager) CreateProduct(ctx context.Context, req *CreateProductRequest) (*domain.Product, error) {
	product, err := domain.NewProduct(req.Name, req.Description, uint(req.CategoryID), uint(req.BrandID), req.Price, req.Stock)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to create new product entity", "error", err)
		return nil, err
	}

	err = m.repo.Transaction(ctx, func(tx any) error {
		txRepo := m.repo.WithTx(tx)
		if err := txRepo.Save(ctx, product); err != nil {
			return err
		}

		// 发布“商品创建”事件用于同步搜索索引
		event := map[string]any{
			"action":     "create",
			"product_id": product.ID,
			"name":       product.Name,
			"price":      product.Price,
			"stock":      product.Stock,
		}
		gormTx := tx.(*gorm.DB)
		return m.outbox.PublishInTx(ctx, gormTx, "product.index.sync", fmt.Sprintf("%d", product.ID), event)
	})

	if err != nil {
		return nil, err
	}

	m.logger.InfoContext(ctx, "product created successfully", "product_id", product.ID)
	return product, nil
}

func (m *ProductManager) UpdateProduct(ctx context.Context, id uint64, req *UpdateProductRequest) (*domain.Product, error) {
	product, err := m.repo.FindByID(ctx, uint(id))
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}

	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.CategoryID != nil {
		product.CategoryID = uint(*req.CategoryID)
	}
	if req.BrandID != nil {
		product.BrandID = uint(*req.BrandID)
	}
	if req.Status != nil {
		product.Status = *req.Status
	}

	err = m.repo.Transaction(ctx, func(tx any) error {
		txRepo := m.repo.WithTx(tx)
		if err := txRepo.Update(ctx, product); err != nil {
			return err
		}

		// 发布“商品更新”事件
		event := map[string]any{
			"action":     "update",
			"product_id": id,
			"name":       product.Name,
			"price":      product.Price,
			"status":     product.Status,
		}
		gormTx := tx.(*gorm.DB)
		return m.outbox.PublishInTx(ctx, gormTx, "product.index.sync", fmt.Sprintf("%d", id), event)
	})

	if err != nil {
		return nil, err
	}

	m.logger.InfoContext(ctx, "product updated and sync event recorded", "product_id", id)

	// 异步清理缓存
	_ = m.cache.Delete(ctx, fmt.Sprintf("product:%d", id))

	return product, nil
}

func (m *ProductManager) DeleteProduct(ctx context.Context, id uint64) error {
	err := m.repo.Transaction(ctx, func(tx any) error {
		txRepo := m.repo.WithTx(tx)
		if err := txRepo.Delete(ctx, uint(id)); err != nil {
			return err
		}

		// 发布“商品删除”事件
		event := map[string]any{
			"action":     "delete",
			"product_id": id,
		}
		gormTx := tx.(*gorm.DB)
		return m.outbox.PublishInTx(ctx, gormTx, "product.index.sync", fmt.Sprintf("%d", id), event)
	})

	if err != nil {
		return err
	}

	m.logger.InfoContext(ctx, "product deleted successfully", "product_id", id)
	return m.cache.Delete(ctx, fmt.Sprintf("product:%d", id))
}

// ---------------- SKU ----------------

func (m *ProductManager) AddSKU(ctx context.Context, productID uint64, req *AddSKURequest) (*domain.SKU, error) {
	product, err := m.repo.FindByID(ctx, uint(productID))
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}

	sku, err := domain.NewSKU(uint(productID), req.Name, req.Price, req.Stock, req.Image, req.Specs)
	if err != nil {
		return nil, err
	}

	if err := m.skuRepo.Save(ctx, sku); err != nil {
		m.logger.ErrorContext(ctx, "failed to save SKU", "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "SKU added successfully", "sku_id", sku.ID, "product_id", productID)

	if err := m.cache.Delete(ctx, fmt.Sprintf("product:%d", productID)); err != nil {
		m.logger.ErrorContext(ctx, "failed to delete product cache after adding SKU", "product_id", productID, "error", err)
	}

	return sku, nil
}

func (m *ProductManager) UpdateSKU(ctx context.Context, id uint64, req *UpdateSKURequest) (*domain.SKU, error) {
	sku, err := m.skuRepo.FindByID(ctx, uint(id))
	if err != nil {
		return nil, err
	}
	if sku == nil {
		return nil, errors.New("SKU not found")
	}

	if req.Price != nil {
		sku.Price = *req.Price
	}
	if req.Stock != nil {
		sku.Stock = *req.Stock
	}
	if req.Image != nil {
		sku.Image = *req.Image
	}

	if err := m.skuRepo.Update(ctx, sku); err != nil {
		m.logger.ErrorContext(ctx, "failed to update SKU", "sku_id", id, "error", err)
		return nil, err
	}

	if err := m.cache.Delete(ctx, fmt.Sprintf("product:%d", sku.ProductID)); err != nil {
		m.logger.ErrorContext(ctx, "failed to delete product cache after updating SKU", "sku_id", id, "product_id", sku.ProductID, "error", err)
	}

	return sku, nil
}

func (m *ProductManager) DeleteSKU(ctx context.Context, id uint64) error {
	sku, err := m.skuRepo.FindByID(ctx, uint(id))
	if err != nil {
		return err
	}
	if sku != nil {
		if err := m.skuRepo.Delete(ctx, uint(id)); err != nil {
			return err
		}
		if err := m.cache.Delete(ctx, fmt.Sprintf("product:%d", sku.ProductID)); err != nil {
			m.logger.ErrorContext(ctx, "failed to delete product cache after deleting SKU", "sku_id", id, "product_id", sku.ProductID, "error", err)
		}
	}
	return nil
}

// ---------------- Brand ----------------

func (m *ProductManager) CreateBrand(ctx context.Context, req *CreateBrandRequest) (*domain.Brand, error) {
	brand, err := domain.NewBrand(req.Name, req.Logo)
	if err != nil {
		return nil, err
	}

	if err := m.brandRepo.Save(ctx, brand); err != nil {
		return nil, err
	}
	return brand, nil
}

func (m *ProductManager) UpdateBrand(ctx context.Context, id uint64, req *UpdateBrandRequest) (*domain.Brand, error) {
	brand, err := m.brandRepo.FindByID(ctx, uint(id))
	if err != nil {
		return nil, err
	}
	if brand == nil {
		return nil, errors.New("brand not found")
	}

	if req.Name != nil {
		brand.Name = *req.Name
	}
	if req.Logo != nil {
		brand.Logo = *req.Logo
	}

	if err := m.brandRepo.Update(ctx, brand); err != nil {
		return nil, err
	}
	return brand, nil
}

func (m *ProductManager) DeleteBrand(ctx context.Context, id uint64) error {
	return m.brandRepo.Delete(ctx, uint(id))
}

// ---------------- Category ----------------

func (m *ProductManager) CreateCategory(ctx context.Context, req *CreateCategoryRequest) (*domain.Category, error) {
	category, err := domain.NewCategory(req.Name, uint(req.ParentID))
	if err != nil {
		return nil, err
	}

	if err := m.categoryRepo.Save(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (m *ProductManager) UpdateCategory(ctx context.Context, id uint64, req *UpdateCategoryRequest) (*domain.Category, error) {
	category, err := m.categoryRepo.FindByID(ctx, uint(id))
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, errors.New("category not found")
	}

	if req.Name != nil {
		category.Name = *req.Name
	}
	if req.ParentID != nil {
		category.ParentID = uint(*req.ParentID)
	}
	if req.Sort != nil {
		category.Sort = *req.Sort
	}

	if err := m.categoryRepo.Update(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (m *ProductManager) DeleteCategory(ctx context.Context, id uint64) error {
	return m.categoryRepo.Delete(ctx, uint(id))
}
