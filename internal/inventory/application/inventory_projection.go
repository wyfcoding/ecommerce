package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/inventory/domain"
)

// InventoryProjectionService 负责将事件转换为读模型更新。
type InventoryProjectionService struct {
	repo       domain.InventoryRepository
	readRepo   domain.InventoryReadRepository
	searchRepo domain.InventorySearchRepository
	logger     *slog.Logger
}

// NewInventoryProjectionService 创建库存投影服务。
func NewInventoryProjectionService(repo domain.InventoryRepository, readRepo domain.InventoryReadRepository, searchRepo domain.InventorySearchRepository, logger *slog.Logger) *InventoryProjectionService {
	return &InventoryProjectionService{
		repo:       repo,
		readRepo:   readRepo,
		searchRepo: searchRepo,
		logger:     logger,
	}
}

// OnStockLocked 处理库存锁定事件。
func (s *InventoryProjectionService) OnStockLocked(ctx context.Context, event *domain.StockLockedEvent) error {
	return s.refreshReadModel(ctx, event.SkuID)
}

// OnStockUnlocked 处理库存解锁事件。
func (s *InventoryProjectionService) OnStockUnlocked(ctx context.Context, event *domain.StockUnlockedEvent) error {
	return s.refreshReadModel(ctx, event.SkuID)
}

// OnStockDeducted 处理库存扣减事件。
func (s *InventoryProjectionService) OnStockDeducted(ctx context.Context, event *domain.StockDeductedEvent) error {
	return s.refreshReadModel(ctx, event.SkuID)
}

// OnStockAdded 处理库存增加事件。
func (s *InventoryProjectionService) OnStockAdded(ctx context.Context, event *domain.StockAddedEvent) error {
	return s.refreshReadModel(ctx, event.SkuID)
}

// OnStockWarning 处理库存预警事件。
func (s *InventoryProjectionService) OnStockWarning(ctx context.Context, event *domain.StockWarningEvent) error {
	return s.refreshReadModel(ctx, event.SkuID)
}

func (s *InventoryProjectionService) refreshReadModel(ctx context.Context, skuID uint64) error {
	inventory, err := s.repo.GetBySkuID(ctx, skuID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load inventory for projection", "sku_id", skuID, "error", err)
		return err
	}

	if inventory == nil {
		if s.readRepo != nil {
			_ = s.readRepo.Delete(ctx, skuID)
		}
		if s.searchRepo != nil {
			_ = s.searchRepo.Delete(ctx, skuID)
		}
		return nil
	}

	if s.readRepo != nil {
		if err := s.readRepo.Save(ctx, inventory); err != nil {
			s.logger.ErrorContext(ctx, "failed to save inventory read model", "sku_id", skuID, "error", err)
			return err
		}
	}
	if s.searchRepo != nil {
		if err := s.searchRepo.Index(ctx, inventory); err != nil {
			s.logger.ErrorContext(ctx, "failed to index inventory search model", "sku_id", skuID, "error", err)
			return err
		}
	}

	return nil
}
