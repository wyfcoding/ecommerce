package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/product/domain"
	"github.com/wyfcoding/pkg/cache"
)

type ProductCommandService struct {
	repo         domain.ProductRepository
	skuRepo      domain.SKURepository
	brandRepo    domain.BrandRepository
	categoryRepo domain.CategoryRepository
	cache        cache.Cache
	publisher    domain.EventPublisher
	topic        string
	logger       *slog.Logger
}

func NewProductCommandService(
	repo domain.ProductRepository,
	skuRepo domain.SKURepository,
	brandRepo domain.BrandRepository,
	categoryRepo domain.CategoryRepository,
	cache cache.Cache,
	publisher domain.EventPublisher,
	topic string,
	logger *slog.Logger,
) *ProductCommandService {
	return &ProductCommandService{
		repo:         repo,
		skuRepo:      skuRepo,
		brandRepo:    brandRepo,
		categoryRepo: categoryRepo,
		cache:        cache,
		publisher:    publisher,
		topic:        topic,
		logger:       logger,
	}
}

// ---------------- Product ----------------

func (s *ProductCommandService) CreateProduct(ctx context.Context, cmd *CreateProductCommand) (*domain.Product, error) {
	product, err := domain.NewProduct(cmd.Name, cmd.Description, cmd.CategoryID, cmd.BrandID, cmd.Type, cmd.Price, cmd.Stock)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, product); err != nil {
		return nil, err
	}

	// 发布领域事件
	event := &domain.ProductCreatedEvent{
		ID:        product.ID,
		Name:      product.Name,
		Price:     product.Price,
		Stock:     product.Stock,
		Timestamp: time.Now(),
	}
	_ = s.publisher.Publish(ctx, s.topic, event)

	return product, nil
}

func (s *ProductCommandService) UpdateProduct(ctx context.Context, cmd *UpdateProductCommand) (*domain.Product, error) {
	product, err := s.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}

	if cmd.Name != nil {
		product.Name = *cmd.Name
	}
	if cmd.Description != nil {
		product.Description = *cmd.Description
	}
	if cmd.CategoryID != nil {
		product.CategoryID = *cmd.CategoryID
	}
	if cmd.BrandID != nil {
		product.BrandID = *cmd.BrandID
	}
	if cmd.Status != nil {
		product.Status = domain.ProductStatus(*cmd.Status)
	}

	if err := s.repo.Update(ctx, product); err != nil {
		return nil, err
	}

	_ = s.cache.Delete(ctx, fmt.Sprintf("product:%d", cmd.ID))

	// 发布领域事件
	event := &domain.ProductUpdatedEvent{
		ID:        product.ID,
		Status:    int(product.Status),
		Timestamp: time.Now(),
	}
	_ = s.publisher.Publish(ctx, s.topic, event)

	return product, nil
}

func (s *ProductCommandService) DeleteProduct(ctx context.Context, id uint64) error {
	if err := s.repo.Delete(ctx, uint(id)); err != nil {
		return err
	}

	_ = s.cache.Delete(ctx, fmt.Sprintf("product:%d", id))

	// 发布领域事件
	event := &domain.ProductDeletedEvent{
		ID:        uint(id),
		Timestamp: time.Now(),
	}
	_ = s.publisher.Publish(ctx, s.topic, event)

	return nil
}

// ---------------- SKU ----------------

func (s *ProductCommandService) AddSKU(ctx context.Context, cmd *AddSKUCommand) (*domain.SKU, error) {
	product, err := s.repo.FindByID(ctx, cmd.ProductID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}

	sku, err := domain.NewSKU(cmd.ProductID, cmd.Name, cmd.Price, cmd.Stock, cmd.Image, cmd.Specs)
	if err != nil {
		return nil, err
	}

	if err := s.skuRepo.Save(ctx, sku); err != nil {
		return nil, err
	}

	_ = s.cache.Delete(ctx, fmt.Sprintf("product:%d", cmd.ProductID))

	// 发布领域事件
	event := &domain.SKUAddedEvent{
		ProductID: cmd.ProductID,
		SKUID:     sku.ID,
		Timestamp: time.Now(),
	}
	_ = s.publisher.Publish(ctx, s.topic, event)

	return sku, nil
}

func (s *ProductCommandService) UpdateSKU(ctx context.Context, cmd *UpdateSKUCommand) (*domain.SKU, error) {
	sku, err := s.skuRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}
	if sku == nil {
		return nil, errors.New("SKU not found")
	}

	if cmd.Price != nil {
		sku.Price = *cmd.Price
	}
	if cmd.Stock != nil {
		sku.Stock = *cmd.Stock
	}
	if cmd.Image != nil {
		sku.Image = *cmd.Image
	}

	if err := s.skuRepo.Update(ctx, sku); err != nil {
		return nil, err
	}

	_ = s.cache.Delete(ctx, fmt.Sprintf("product:%d", sku.ProductID))
	_ = s.cache.Delete(ctx, fmt.Sprintf("sku:%d", sku.ID))

	// 发布领域事件
	event := &domain.SKUUpdatedEvent{
		ProductID: sku.ProductID,
		SKUID:     sku.ID,
		Timestamp: time.Now(),
	}
	_ = s.publisher.Publish(ctx, s.topic, event)

	return sku, nil
}

func (s *ProductCommandService) DeleteSKU(ctx context.Context, id uint64) error {
	sku, err := s.skuRepo.FindByID(ctx, uint(id))
	if err != nil {
		return err
	}
	if sku == nil {
		return nil // Already deleted
	}

	if err := s.skuRepo.Delete(ctx, uint(id)); err != nil {
		return err
	}

	_ = s.cache.Delete(ctx, fmt.Sprintf("product:%d", sku.ProductID))
	_ = s.cache.Delete(ctx, fmt.Sprintf("sku:%d", sku.ID))

	// 发布领域事件
	event := &domain.SKUDeletedEvent{
		ProductID: sku.ProductID,
		SKUID:     sku.ID,
		Timestamp: time.Now(),
	}
	_ = s.publisher.Publish(ctx, s.topic, event)

	return nil
}

// ---------------- Brand ----------------

func (s *ProductCommandService) CreateBrand(ctx context.Context, cmd *CreateBrandCommand) (*domain.Brand, error) {
	brand, err := domain.NewBrand(cmd.Name, cmd.Logo)
	if err != nil {
		return nil, err
	}
	if err := s.brandRepo.Save(ctx, brand); err != nil {
		return nil, err
	}
	_ = s.cache.Delete(ctx, "brand:list")
	return brand, nil
}

func (s *ProductCommandService) UpdateBrand(ctx context.Context, cmd *UpdateBrandCommand) (*domain.Brand, error) {
	brand, err := s.brandRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}
	if brand == nil {
		return nil, errors.New("brand not found")
	}
	if cmd.Name != nil {
		brand.Name = *cmd.Name
	}
	if cmd.Logo != nil {
		brand.Logo = *cmd.Logo
	}
	if err := s.brandRepo.Update(ctx, brand); err != nil {
		return nil, err
	}
	_ = s.cache.Delete(ctx, fmt.Sprintf("brand:%d", cmd.ID))
	_ = s.cache.Delete(ctx, "brand:list")
	return brand, nil
}

func (s *ProductCommandService) DeleteBrand(ctx context.Context, id uint64) error {
	if err := s.brandRepo.Delete(ctx, uint(id)); err != nil {
		return err
	}
	_ = s.cache.Delete(ctx, fmt.Sprintf("brand:%d", id))
	_ = s.cache.Delete(ctx, "brand:list")
	return nil
}

// ---------------- Category ----------------

func (s *ProductCommandService) CreateCategory(ctx context.Context, cmd *CreateCategoryCommand) (*domain.Category, error) {
	category, err := domain.NewCategory(cmd.Name, cmd.ParentID)
	if err != nil {
		return nil, err
	}
	if err := s.categoryRepo.Save(ctx, category); err != nil {
		return nil, err
	}
	_ = s.cache.Delete(ctx, "category:list")
	return category, nil
}

func (s *ProductCommandService) UpdateCategory(ctx context.Context, cmd *UpdateCategoryCommand) (*domain.Category, error) {
	category, err := s.categoryRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, errors.New("category not found")
	}
	if cmd.Name != nil {
		category.Name = *cmd.Name
	}
	if cmd.ParentID != nil {
		category.ParentID = *cmd.ParentID
	}
	if cmd.Sort != nil {
		category.Sort = *cmd.Sort
	}
	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return nil, err
	}
	_ = s.cache.Delete(ctx, fmt.Sprintf("category:%d", cmd.ID))
	_ = s.cache.Delete(ctx, "category:list")
	return category, nil
}

func (s *ProductCommandService) DeleteCategory(ctx context.Context, id uint64) error {
	if err := s.categoryRepo.Delete(ctx, uint(id)); err != nil {
		return err
	}
	_ = s.cache.Delete(ctx, fmt.Sprintf("category:%d", id))
	_ = s.cache.Delete(ctx, "category:list")
	return nil
}
