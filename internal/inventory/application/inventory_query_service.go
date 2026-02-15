package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/inventory/domain"
)

// InventoryQueryService 处理库存的读操作。
type InventoryQueryService struct {
	repo       domain.InventoryRepository
	readRepo   domain.InventoryReadRepository
	searchRepo domain.InventorySearchRepository
	eventStore domain.EventStore
	logger     *slog.Logger
}

// NewInventoryQueryService 负责处理 NewInventory 相关的读操作和查询逻辑。
func NewInventoryQueryService(repo domain.InventoryRepository, readRepo domain.InventoryReadRepository, searchRepo domain.InventorySearchRepository, eventStore domain.EventStore, logger *slog.Logger) *InventoryQueryService {
	return &InventoryQueryService{
		repo:       repo,
		readRepo:   readRepo,
		searchRepo: searchRepo,
		eventStore: eventStore,
		logger:     logger,
	}
}

// GetInventory 获取指定SKU的库存记录。
func (q *InventoryQueryService) GetInventory(ctx context.Context, skuID uint64) (*domain.Inventory, error) {
	if q.readRepo != nil {
		if inv, err := q.readRepo.GetBySkuID(ctx, skuID); err == nil && inv != nil {
			return inv, nil
		}
	}
	if q.searchRepo != nil {
		if inv, err := q.searchRepo.FindBySkuID(ctx, skuID); err == nil && inv != nil {
			return inv, nil
		}
	}

	inv, err := q.repo.GetBySkuID(ctx, skuID)
	if err != nil || inv != nil {
		return inv, err
	}

	if q.eventStore != nil {
		aggregateID := fmt.Sprintf("%d", skuID)
		events, err := q.eventStore.Load(ctx, aggregateID)
		if err != nil {
			q.logger.WarnContext(ctx, "event store load failed", "sku_id", skuID, "error", err)
			return nil, nil
		}
		if len(events) > 0 {
			inventory, rebuildErr := domain.RebuildInventoryFromEvents(events)
			if rebuildErr != nil {
				q.logger.WarnContext(ctx, "inventory rebuild failed", "sku_id", skuID, "error", rebuildErr)
				return nil, nil
			}
			return inventory, nil
		}
	}

	return nil, nil
}

// ListInventories 获取库存列表。
func (q *InventoryQueryService) ListInventories(ctx context.Context, page, pageSize int) ([]*domain.Inventory, int64, error) {
	offset := (page - 1) * pageSize

	if q.searchRepo != nil {
		list, total, err := q.searchRepo.Search(ctx, nil, nil, nil, nil, offset, pageSize, nil, nil, "created_at")
		if err == nil {
			return list, total, nil
		}
		q.logger.WarnContext(ctx, "inventory search fallback to mysql", "error", err)
	}

	return q.repo.List(ctx, offset, pageSize)
}

// GetInventoryLogs 获取指定库存的日志列表。
func (q *InventoryQueryService) GetInventoryLogs(ctx context.Context, skuID uint64, inventoryID uint64, page, pageSize int) ([]*domain.InventoryLog, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.GetLogs(ctx, skuID, inventoryID, offset, pageSize)
}

// ListInventoriesByCondition 支持更复杂的查询条件（可选扩展）。
func (q *InventoryQueryService) ListInventoriesByCondition(ctx context.Context, skuID, productID, warehouseID *uint64, status *domain.InventoryStatus, page, pageSize int, startTime, endTime *time.Time, sortBy string) ([]*domain.Inventory, int64, error) {
	offset := (page - 1) * pageSize
	if q.searchRepo != nil {
		return q.searchRepo.Search(ctx, skuID, productID, warehouseID, status, offset, pageSize, startTime, endTime, sortBy)
	}
	return q.repo.List(ctx, offset, pageSize)
}
