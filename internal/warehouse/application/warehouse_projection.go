package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/warehouse/domain"
)

// WarehouseProjectionService 负责将事件转换为读模型更新。
type WarehouseProjectionService struct {
	repo       domain.WarehouseRepository
	readRepo   domain.WarehouseReadRepository
	searchRepo domain.WarehouseSearchRepository
	logger     *slog.Logger
}

// NewWarehouseProjectionService 创建仓库投影服务。
func NewWarehouseProjectionService(repo domain.WarehouseRepository, readRepo domain.WarehouseReadRepository, searchRepo domain.WarehouseSearchRepository, logger *slog.Logger) *WarehouseProjectionService {
	return &WarehouseProjectionService{
		repo:       repo,
		readRepo:   readRepo,
		searchRepo: searchRepo,
		logger:     logger,
	}
}

// OnWarehouseCreated 处理仓库创建事件。
func (s *WarehouseProjectionService) OnWarehouseCreated(ctx context.Context, event *domain.WarehouseCreatedEvent) error {
	return s.refreshWarehouse(ctx, event.WarehouseID)
}

// OnStockAdjusted 处理库存调整事件。
func (s *WarehouseProjectionService) OnStockAdjusted(ctx context.Context, event *domain.StockAdjustedEvent) error {
	return s.refreshStock(ctx, event.WarehouseID, event.SkuID)
}

// OnStockDeducted 处理库存扣减事件。
func (s *WarehouseProjectionService) OnStockDeducted(ctx context.Context, event *domain.StockDeductedEvent) error {
	return s.refreshStock(ctx, event.WarehouseID, event.SkuID)
}

// OnStockReverted 处理库存回滚事件。
func (s *WarehouseProjectionService) OnStockReverted(ctx context.Context, event *domain.StockRevertedEvent) error {
	return s.refreshStock(ctx, event.WarehouseID, event.SkuID)
}

// OnTransferCreated 处理调拨单创建事件。
func (s *WarehouseProjectionService) OnTransferCreated(ctx context.Context, event *domain.StockTransferCreatedEvent) error {
	return s.refreshTransfer(ctx, event.TransferID)
}

// OnTransferCompleted 处理调拨单完成事件。
func (s *WarehouseProjectionService) OnTransferCompleted(ctx context.Context, event *domain.StockTransferCompletedEvent) error {
	return s.refreshTransfer(ctx, event.TransferID)
}

func (s *WarehouseProjectionService) refreshWarehouse(ctx context.Context, id uint64) error {
	warehouse, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load warehouse for projection", "warehouse_id", id, "error", err)
		return err
	}

	if warehouse == nil {
		if s.readRepo != nil {
			_ = s.readRepo.DeleteWarehouse(ctx, id)
		}
		if s.searchRepo != nil {
			_ = s.searchRepo.DeleteWarehouse(ctx, id)
		}
		return nil
	}

	if s.readRepo != nil {
		if err := s.readRepo.SaveWarehouse(ctx, warehouse); err != nil {
			s.logger.ErrorContext(ctx, "failed to save warehouse read model", "warehouse_id", id, "error", err)
			return err
		}
	}
	if s.searchRepo != nil {
		if err := s.searchRepo.IndexWarehouse(ctx, warehouse); err != nil {
			s.logger.ErrorContext(ctx, "failed to index warehouse search model", "warehouse_id", id, "error", err)
			return err
		}
	}
	return nil
}

func (s *WarehouseProjectionService) refreshStock(ctx context.Context, warehouseID, skuID uint64) error {
	stock, err := s.repo.GetStock(ctx, warehouseID, skuID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load stock for projection", "warehouse_id", warehouseID, "sku_id", skuID, "error", err)
		return err
	}

	if stock == nil {
		if s.readRepo != nil {
			_ = s.readRepo.DeleteStock(ctx, warehouseID, skuID)
		}
		return nil
	}

	if s.readRepo != nil {
		if err := s.readRepo.SaveStock(ctx, stock); err != nil {
			s.logger.ErrorContext(ctx, "failed to save stock read model", "warehouse_id", warehouseID, "sku_id", skuID, "error", err)
			return err
		}
	}
	return nil
}

func (s *WarehouseProjectionService) refreshTransfer(ctx context.Context, id uint64) error {
	transfer, err := s.repo.FindTransferByID(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load transfer for projection", "transfer_id", id, "error", err)
		return err
	}

	if transfer == nil {
		if s.readRepo != nil {
			_ = s.readRepo.DeleteTransfer(ctx, id)
		}
		if s.searchRepo != nil {
			_ = s.searchRepo.DeleteTransfer(ctx, id)
		}
		return nil
	}

	if s.readRepo != nil {
		if err := s.readRepo.SaveTransfer(ctx, transfer); err != nil {
			s.logger.ErrorContext(ctx, "failed to save transfer read model", "transfer_id", id, "error", err)
			return err
		}
	}
	if s.searchRepo != nil {
		if err := s.searchRepo.IndexTransfer(ctx, transfer); err != nil {
			s.logger.ErrorContext(ctx, "failed to index transfer search model", "transfer_id", id, "error", err)
			return err
		}
	}
	return nil
}
