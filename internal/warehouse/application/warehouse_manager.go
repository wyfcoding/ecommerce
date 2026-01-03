package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/warehouse/domain"
)

// WarehouseManager 处理仓库模块的写操作和业务逻辑。
type WarehouseManager struct {
	Repo   domain.WarehouseRepository
	logger *slog.Logger
}

// NewWarehouseManager 创建并返回一个新的 WarehouseManager 实例。
func NewWarehouseManager(repo domain.WarehouseRepository, logger *slog.Logger) *WarehouseManager {
	return &WarehouseManager{
		Repo:   repo,
		logger: logger,
	}
}

// CreateWarehouse 创建仓库。
func (m *WarehouseManager) CreateWarehouse(ctx context.Context, warehouse *domain.Warehouse) error {
	if err := m.Repo.SaveWarehouse(ctx, warehouse); err != nil {
		m.logger.Error("failed to create warehouse", "error", err, "code", warehouse.Code)
		return err
	}
	return nil
}

// UpdateWarehouse 更新仓库。
func (m *WarehouseManager) UpdateWarehouse(ctx context.Context, warehouse *domain.Warehouse) error {
	if err := m.Repo.SaveWarehouse(ctx, warehouse); err != nil {
		m.logger.Error("failed to update warehouse", "error", err, "id", warehouse.ID)
		return err
	}
	return nil
}

// AdjustStock 调整库存。
func (m *WarehouseManager) AdjustStock(ctx context.Context, stock *domain.WarehouseStock) error {
	if err := m.Repo.SaveStock(ctx, stock); err != nil {
		m.logger.Error("failed to adjust stock", "error", err, "warehouse_id", stock.WarehouseID, "sku_id", stock.SkuID)
		return err
	}
	return nil
}

// CreateTransfer 创建调拨单。
func (m *WarehouseManager) CreateTransfer(ctx context.Context, transfer *domain.StockTransfer) error {
	if err := m.Repo.SaveTransfer(ctx, transfer); err != nil {
		m.logger.Error("failed to create transfer", "error", err, "transfer_no", transfer.TransferNo)
		return err
	}
	return nil
}

// UpdateTransferStatus 更新调拨单状态。
func (m *WarehouseManager) UpdateTransferStatus(ctx context.Context, transfer *domain.StockTransfer) error {
	if err := m.Repo.SaveTransfer(ctx, transfer); err != nil {
		m.logger.Error("failed to update transfer status", "error", err, "id", transfer.ID)
		return err
	}
	return nil
}
