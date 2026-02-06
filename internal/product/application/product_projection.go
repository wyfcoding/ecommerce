package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/product/domain"
)

// ProductProjectionService 负责将事件转换为读模型更新。
type ProductProjectionService struct {
	repo             domain.ProductRepository
	skuRepo          domain.SKURepository
	brandRepo        domain.BrandRepository
	categoryRepo     domain.CategoryRepository
	readRepo         domain.ProductReadRepository
	skuReadRepo      domain.SKUReadRepository
	brandReadRepo    domain.BrandReadRepository
	categoryReadRepo domain.CategoryReadRepository
	searchRepo       domain.ProductSearchRepository
	logger           *slog.Logger
}

// NewProductProjectionService 创建商品投影服务。
func NewProductProjectionService(
	repo domain.ProductRepository,
	skuRepo domain.SKURepository,
	brandRepo domain.BrandRepository,
	categoryRepo domain.CategoryRepository,
	readRepo domain.ProductReadRepository,
	skuReadRepo domain.SKUReadRepository,
	brandReadRepo domain.BrandReadRepository,
	categoryReadRepo domain.CategoryReadRepository,
	searchRepo domain.ProductSearchRepository,
	logger *slog.Logger,
) *ProductProjectionService {
	return &ProductProjectionService{
		repo:             repo,
		skuRepo:          skuRepo,
		brandRepo:        brandRepo,
		categoryRepo:     categoryRepo,
		readRepo:         readRepo,
		skuReadRepo:      skuReadRepo,
		brandReadRepo:    brandReadRepo,
		categoryReadRepo: categoryReadRepo,
		searchRepo:       searchRepo,
		logger:           logger,
	}
}

func (s *ProductProjectionService) OnProductCreated(ctx context.Context, event *domain.ProductCreatedEvent) error {
	return s.refreshProduct(ctx, uint64(event.ID))
}

func (s *ProductProjectionService) OnProductUpdated(ctx context.Context, event *domain.ProductUpdatedEvent) error {
	return s.refreshProduct(ctx, uint64(event.ID))
}

func (s *ProductProjectionService) OnProductDeleted(ctx context.Context, event *domain.ProductDeletedEvent) error {
	if s.readRepo != nil {
		_ = s.readRepo.Delete(ctx, uint64(event.ID))
	}
	if s.searchRepo != nil {
		_ = s.searchRepo.Delete(ctx, uint64(event.ID))
	}
	return nil
}

func (s *ProductProjectionService) OnSKUAdded(ctx context.Context, event *domain.SKUAddedEvent) error {
	if err := s.refreshSKU(ctx, uint64(event.SKUID)); err != nil {
		return err
	}
	return s.refreshProduct(ctx, uint64(event.ProductID))
}

func (s *ProductProjectionService) OnSKUUpdated(ctx context.Context, event *domain.SKUUpdatedEvent) error {
	if err := s.refreshSKU(ctx, uint64(event.SKUID)); err != nil {
		return err
	}
	return s.refreshProduct(ctx, uint64(event.ProductID))
}

func (s *ProductProjectionService) OnSKUDeleted(ctx context.Context, event *domain.SKUDeletedEvent) error {
	if s.skuReadRepo != nil {
		_ = s.skuReadRepo.Delete(ctx, uint64(event.SKUID))
	}
	return s.refreshProduct(ctx, uint64(event.ProductID))
}

func (s *ProductProjectionService) OnBrandCreated(ctx context.Context, event *domain.BrandCreatedEvent) error {
	return s.refreshBrand(ctx, uint64(event.ID))
}

func (s *ProductProjectionService) OnBrandUpdated(ctx context.Context, event *domain.BrandUpdatedEvent) error {
	return s.refreshBrand(ctx, uint64(event.ID))
}

func (s *ProductProjectionService) OnBrandDeleted(ctx context.Context, event *domain.BrandDeletedEvent) error {
	if s.brandReadRepo != nil {
		_ = s.brandReadRepo.Delete(ctx, uint64(event.ID))
	}
	return nil
}

func (s *ProductProjectionService) OnCategoryCreated(ctx context.Context, event *domain.CategoryCreatedEvent) error {
	return s.refreshCategory(ctx, uint64(event.ID))
}

func (s *ProductProjectionService) OnCategoryUpdated(ctx context.Context, event *domain.CategoryUpdatedEvent) error {
	return s.refreshCategory(ctx, uint64(event.ID))
}

func (s *ProductProjectionService) OnCategoryDeleted(ctx context.Context, event *domain.CategoryDeletedEvent) error {
	if s.categoryReadRepo != nil {
		_ = s.categoryReadRepo.Delete(ctx, uint64(event.ID))
	}
	return nil
}

func (s *ProductProjectionService) refreshProduct(ctx context.Context, id uint64) error {
	product, err := s.repo.FindByID(ctx, uint(id))
	if err != nil {
		if s.logger != nil {
			s.logger.ErrorContext(ctx, "failed to load product for projection", "product_id", id, "error", err)
		}
		return err
	}

	if product == nil {
		if s.readRepo != nil {
			_ = s.readRepo.Delete(ctx, id)
		}
		if s.searchRepo != nil {
			_ = s.searchRepo.Delete(ctx, id)
		}
		return nil
	}

	if s.readRepo != nil {
		if err := s.readRepo.Save(ctx, product); err != nil && s.logger != nil {
			s.logger.ErrorContext(ctx, "failed to save product read model", "product_id", id, "error", err)
			return err
		}
	}
	if s.searchRepo != nil {
		if err := s.searchRepo.Index(ctx, product); err != nil && s.logger != nil {
			s.logger.ErrorContext(ctx, "failed to index product search model", "product_id", id, "error", err)
			return err
		}
	}
	return nil
}

func (s *ProductProjectionService) refreshSKU(ctx context.Context, id uint64) error {
	sku, err := s.skuRepo.FindByID(ctx, uint(id))
	if err != nil {
		if s.logger != nil {
			s.logger.ErrorContext(ctx, "failed to load sku for projection", "sku_id", id, "error", err)
		}
		return err
	}
	if sku == nil {
		if s.skuReadRepo != nil {
			_ = s.skuReadRepo.Delete(ctx, id)
		}
		return nil
	}
	if s.skuReadRepo != nil {
		if err := s.skuReadRepo.Save(ctx, sku); err != nil && s.logger != nil {
			s.logger.ErrorContext(ctx, "failed to save sku read model", "sku_id", id, "error", err)
			return err
		}
	}
	return nil
}

func (s *ProductProjectionService) refreshBrand(ctx context.Context, id uint64) error {
	brand, err := s.brandRepo.FindByID(ctx, uint(id))
	if err != nil {
		if s.logger != nil {
			s.logger.ErrorContext(ctx, "failed to load brand for projection", "brand_id", id, "error", err)
		}
		return err
	}
	if brand == nil {
		if s.brandReadRepo != nil {
			_ = s.brandReadRepo.Delete(ctx, id)
		}
		return nil
	}
	if s.brandReadRepo != nil {
		if err := s.brandReadRepo.Save(ctx, brand); err != nil && s.logger != nil {
			s.logger.ErrorContext(ctx, "failed to save brand read model", "brand_id", id, "error", err)
			return err
		}
	}
	return nil
}

func (s *ProductProjectionService) refreshCategory(ctx context.Context, id uint64) error {
	category, err := s.categoryRepo.FindByID(ctx, uint(id))
	if err != nil {
		if s.logger != nil {
			s.logger.ErrorContext(ctx, "failed to load category for projection", "category_id", id, "error", err)
		}
		return err
	}
	if category == nil {
		if s.categoryReadRepo != nil {
			_ = s.categoryReadRepo.Delete(ctx, id)
		}
		return nil
	}
	if s.categoryReadRepo != nil {
		if err := s.categoryReadRepo.Save(ctx, category); err != nil && s.logger != nil {
			s.logger.ErrorContext(ctx, "failed to save category read model", "category_id", id, "error", err)
			return err
		}
	}
	return nil
}
