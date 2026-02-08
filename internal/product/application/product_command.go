package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/product/domain"
	"github.com/wyfcoding/pkg/messagequeue"
)

type ProductCommandService struct {
	repo         domain.ProductRepository
	skuRepo      domain.SKURepository
	brandRepo    domain.BrandRepository
	categoryRepo domain.CategoryRepository
	publisher    messagequeue.EventPublisher
	logger       *slog.Logger
}

func NewProductCommandService(
	repo domain.ProductRepository,
	skuRepo domain.SKURepository,
	brandRepo domain.BrandRepository,
	categoryRepo domain.CategoryRepository,
	publisher messagequeue.EventPublisher,
	logger *slog.Logger,
) *ProductCommandService {
	return &ProductCommandService{
		repo:         repo,
		skuRepo:      skuRepo,
		brandRepo:    brandRepo,
		categoryRepo: categoryRepo,
		publisher:    publisher,
		logger:       logger,
	}
}

// ---------------- Product ----------------

func (s *ProductCommandService) CreateProduct(ctx context.Context, cmd *CreateProductCommand) (*domain.Product, error) {
	product, err := domain.NewProduct(cmd.Name, cmd.Description, cmd.CategoryID, cmd.BrandID, cmd.Type, cmd.Price, cmd.Stock)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Transaction(ctx, func(tx any) error {
		repo := s.repo.WithTx(tx)
		if err := repo.Save(ctx, product); err != nil {
			return err
		}
		event := &domain.ProductCreatedEvent{
			ID:        product.ID,
			Name:      product.Name,
			Price:     product.Price,
			Stock:     product.Stock,
			Timestamp: time.Now(),
		}
		return s.publisher.PublishInTx(ctx, tx, domain.ProductCreatedEventType, fmt.Sprintf("%d", product.ID), event)
	}); err != nil {
		return nil, err
	}

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

	if err := s.repo.Transaction(ctx, func(tx any) error {
		repo := s.repo.WithTx(tx)
		if err := repo.Update(ctx, product); err != nil {
			return err
		}
		event := &domain.ProductUpdatedEvent{
			ID:        product.ID,
			Status:    int(product.Status),
			Timestamp: time.Now(),
		}
		return s.publisher.PublishInTx(ctx, tx, domain.ProductUpdatedEventType, fmt.Sprintf("%d", product.ID), event)
	}); err != nil {
		return nil, err
	}

	return product, nil
}

func (s *ProductCommandService) DeleteProduct(ctx context.Context, id uint64) error {
	return s.repo.Transaction(ctx, func(tx any) error {
		repo := s.repo.WithTx(tx)
		if err := repo.Delete(ctx, uint(id)); err != nil {
			return err
		}
		event := &domain.ProductDeletedEvent{
			ID:        uint(id),
			Timestamp: time.Now(),
		}
		return s.publisher.PublishInTx(ctx, tx, domain.ProductDeletedEventType, fmt.Sprintf("%d", id), event)
	})
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

	if err := s.skuRepo.Transaction(ctx, func(tx any) error {
		repo := s.skuRepo.WithTx(tx)
		if err := repo.Save(ctx, sku); err != nil {
			return err
		}
		event := &domain.SKUAddedEvent{
			ProductID: cmd.ProductID,
			SKUID:     sku.ID,
			Timestamp: time.Now(),
		}
		return s.publisher.PublishInTx(ctx, tx, domain.SKUAddedEventType, fmt.Sprintf("%d", sku.ID), event)
	}); err != nil {
		return nil, err
	}

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

	if err := s.skuRepo.Transaction(ctx, func(tx any) error {
		repo := s.skuRepo.WithTx(tx)
		if err := repo.Update(ctx, sku); err != nil {
			return err
		}
		event := &domain.SKUUpdatedEvent{
			ProductID: sku.ProductID,
			SKUID:     sku.ID,
			Timestamp: time.Now(),
		}
		return s.publisher.PublishInTx(ctx, tx, domain.SKUUpdatedEventType, fmt.Sprintf("%d", sku.ID), event)
	}); err != nil {
		return nil, err
	}

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

	return s.skuRepo.Transaction(ctx, func(tx any) error {
		repo := s.skuRepo.WithTx(tx)
		if err := repo.Delete(ctx, uint(id)); err != nil {
			return err
		}
		event := &domain.SKUDeletedEvent{
			ProductID: sku.ProductID,
			SKUID:     sku.ID,
			Timestamp: time.Now(),
		}
		return s.publisher.PublishInTx(ctx, tx, domain.SKUDeletedEventType, fmt.Sprintf("%d", sku.ID), event)
	})
}

// ---------------- Brand ----------------

func (s *ProductCommandService) CreateBrand(ctx context.Context, cmd *CreateBrandCommand) (*domain.Brand, error) {
	brand, err := domain.NewBrand(cmd.Name, cmd.Logo)
	if err != nil {
		return nil, err
	}
	if err := s.brandRepo.Transaction(ctx, func(tx any) error {
		repo := s.brandRepo.WithTx(tx)
		if err := repo.Save(ctx, brand); err != nil {
			return err
		}
		event := &domain.BrandCreatedEvent{
			ID:        brand.ID,
			Name:      brand.Name,
			Logo:      brand.Logo,
			Status:    brand.Status,
			Timestamp: time.Now(),
		}
		return s.publisher.PublishInTx(ctx, tx, domain.BrandCreatedEventType, fmt.Sprintf("%d", brand.ID), event)
	}); err != nil {
		return nil, err
	}
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
	if err := s.brandRepo.Transaction(ctx, func(tx any) error {
		repo := s.brandRepo.WithTx(tx)
		if err := repo.Update(ctx, brand); err != nil {
			return err
		}
		event := &domain.BrandUpdatedEvent{
			ID:        brand.ID,
			Name:      brand.Name,
			Logo:      brand.Logo,
			Status:    brand.Status,
			Timestamp: time.Now(),
		}
		return s.publisher.PublishInTx(ctx, tx, domain.BrandUpdatedEventType, fmt.Sprintf("%d", brand.ID), event)
	}); err != nil {
		return nil, err
	}
	return brand, nil
}

func (s *ProductCommandService) DeleteBrand(ctx context.Context, id uint64) error {
	return s.brandRepo.Transaction(ctx, func(tx any) error {
		repo := s.brandRepo.WithTx(tx)
		if err := repo.Delete(ctx, uint(id)); err != nil {
			return err
		}
		event := &domain.BrandDeletedEvent{
			ID:        uint(id),
			Timestamp: time.Now(),
		}
		return s.publisher.PublishInTx(ctx, tx, domain.BrandDeletedEventType, fmt.Sprintf("%d", id), event)
	})
}

// ---------------- Category ----------------

func (s *ProductCommandService) CreateCategory(ctx context.Context, cmd *CreateCategoryCommand) (*domain.Category, error) {
	category, err := domain.NewCategory(cmd.Name, cmd.ParentID)
	if err != nil {
		return nil, err
	}
	if err := s.categoryRepo.Transaction(ctx, func(tx any) error {
		repo := s.categoryRepo.WithTx(tx)
		if err := repo.Save(ctx, category); err != nil {
			return err
		}
		event := &domain.CategoryCreatedEvent{
			ID:        category.ID,
			Name:      category.Name,
			ParentID:  category.ParentID,
			Sort:      category.Sort,
			Status:    category.Status,
			Timestamp: time.Now(),
		}
		return s.publisher.PublishInTx(ctx, tx, domain.CategoryCreatedEventType, fmt.Sprintf("%d", category.ID), event)
	}); err != nil {
		return nil, err
	}
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
	if err := s.categoryRepo.Transaction(ctx, func(tx any) error {
		repo := s.categoryRepo.WithTx(tx)
		if err := repo.Update(ctx, category); err != nil {
			return err
		}
		event := &domain.CategoryUpdatedEvent{
			ID:        category.ID,
			Name:      category.Name,
			ParentID:  category.ParentID,
			Sort:      category.Sort,
			Status:    category.Status,
			Timestamp: time.Now(),
		}
		return s.publisher.PublishInTx(ctx, tx, domain.CategoryUpdatedEventType, fmt.Sprintf("%d", category.ID), event)
	}); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *ProductCommandService) DeleteCategory(ctx context.Context, id uint64) error {
	return s.categoryRepo.Transaction(ctx, func(tx any) error {
		repo := s.categoryRepo.WithTx(tx)
		if err := repo.Delete(ctx, uint(id)); err != nil {
			return err
		}
		event := &domain.CategoryDeletedEvent{
			ID:        uint(id),
			Timestamp: time.Now(),
		}
		return s.publisher.PublishInTx(ctx, tx, domain.CategoryDeletedEventType, fmt.Sprintf("%d", id), event)
	})
}
