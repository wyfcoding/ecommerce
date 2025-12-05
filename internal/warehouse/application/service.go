package application

import (
	"context"
	"errors" // 导入标准错误处理库。
	"fmt"    // 导入格式化库。
	"time"   // 导入时间库。

	"github.com/wyfcoding/ecommerce/internal/warehouse/domain/entity"     // 导入仓库领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/warehouse/domain/repository" // 导入仓库领域的仓储接口。

	"log/slog" // 导入结构化日志库。
)

// WarehouseService 结构体定义了仓库管理相关的应用服务。
// 它协调领域层和基础设施层，处理仓库的创建、库存管理、库存调拨以及分布式事务中的库存操作。
type WarehouseService struct {
	repo   repository.WarehouseRepository // 依赖WarehouseRepository接口，用于数据持久化操作。
	logger *slog.Logger                   // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewWarehouseService 创建并返回一个新的 WarehouseService 实例。
func NewWarehouseService(repo repository.WarehouseRepository, logger *slog.Logger) *WarehouseService {
	return &WarehouseService{
		repo:   repo,
		logger: logger,
	}
}

// CreateWarehouse 创建一个新仓库。
// ctx: 上下文。
// code: 仓库代码，唯一标识。
// name: 仓库名称。
// 返回创建成功的Warehouse实体和可能发生的错误。
func (s *WarehouseService) CreateWarehouse(ctx context.Context, code, name string) (*entity.Warehouse, error) {
	warehouse := &entity.Warehouse{
		Code:   code,
		Name:   name,
		Status: entity.WarehouseStatusInactive, // 新仓库默认为非活跃状态。
	}
	// 通过仓储接口保存仓库。
	if err := s.repo.SaveWarehouse(ctx, warehouse); err != nil {
		s.logger.ErrorContext(ctx, "failed to create warehouse", "code", code, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "warehouse created successfully", "warehouse_id", warehouse.ID, "code", code)
	return warehouse, nil
}

// GetWarehouse 根据ID获取仓库详情。
func (s *WarehouseService) GetWarehouse(ctx context.Context, id uint64) (*entity.Warehouse, error) {
	return s.repo.GetWarehouse(ctx, id)
}

// UpdateStock 更新指定仓库和SKU的库存数量。
// quantity为正表示增加库存，为负表示减少库存。
// ctx: 上下文。
// warehouseID: 仓库ID。
// skuID: SKU ID。
// quantity: 待更新的数量。
// 返回可能发生的错误。
func (s *WarehouseService) UpdateStock(ctx context.Context, warehouseID, skuID uint64, quantity int32) error {
	stock, err := s.repo.GetStock(ctx, warehouseID, skuID)
	if err != nil {
		return err
	}
	if stock == nil {
		// 如果库存记录不存在，且是减少库存操作，则视为库存不足。
		if quantity < 0 {
			return errors.New("insufficient stock: SKU not found in warehouse")
		}
		// 如果是增加库存操作，则创建新的库存记录。
		stock = &entity.WarehouseStock{
			WarehouseID: warehouseID,
			SkuID:       skuID,
			Stock:       0, // 初始库存为0。
		}
	}

	// 检查更新后的库存是否为负数（防止库存透支）。
	if stock.Stock+quantity < 0 {
		return errors.New("insufficient stock: cannot reduce stock below zero")
	}

	stock.Stock += quantity // 更新库存数量。
	// 通过仓储接口保存更新后的库存。
	if err := s.repo.SaveStock(ctx, stock); err != nil {
		s.logger.ErrorContext(ctx, "failed to update stock", "warehouse_id", warehouseID, "sku_id", skuID, "change_quantity", quantity, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "stock updated successfully", "warehouse_id", warehouseID, "sku_id", skuID, "change_quantity", quantity, "current_stock", stock.Stock)
	return nil
}

// CreateTransfer 创建一个库存调拨单。
// ctx: 上下文。
// fromID: 源仓库ID。
// toID: 目标仓库ID。
// skuID: 调拨的SKU ID。
// quantity: 调拨数量。
// createdBy: 创建调拨单的用户ID。
// 返回创建成功的StockTransfer实体和可能发生的错误。
func (s *WarehouseService) CreateTransfer(ctx context.Context, fromID, toID, skuID uint64, quantity int32, createdBy uint64) (*entity.StockTransfer, error) {
	// 1. 检查源仓库的库存是否充足。
	stock, err := s.repo.GetStock(ctx, fromID, skuID)
	if err != nil {
		return nil, err
	}
	// AvailableStock = Stock - LockedStock
	if stock == nil || stock.AvailableStock() < quantity {
		return nil, errors.New("insufficient stock in source warehouse for transfer")
	}

	// 2. 锁定源仓库的库存（防止被其他操作占用）。
	stock.LockedStock += quantity
	if err := s.repo.SaveStock(ctx, stock); err != nil {
		s.logger.ErrorContext(ctx, "failed to lock stock for transfer", "warehouse_id", fromID, "sku_id", skuID, "quantity", quantity, "error", err)
		return nil, err
	}

	// 3. 创建调拨单实体。
	transfer := &entity.StockTransfer{
		TransferNo:      fmt.Sprintf("T%d%d", fromID, time.Now().UnixNano()), // 生成唯一的调拨单号。
		FromWarehouseID: fromID,
		ToWarehouseID:   toID,
		SkuID:           skuID,
		Quantity:        quantity,
		Status:          entity.StockTransferStatusPending, // 初始状态为待处理。
		CreatedBy:       createdBy,
	}

	// 4. 通过仓储接口保存调拨单。
	if err := s.repo.SaveTransfer(ctx, transfer); err != nil {
		s.logger.ErrorContext(ctx, "failed to save transfer", "from_id", fromID, "to_id", toID, "error", err)
		// TODO: 如果这里失败，需要解锁之前锁定的库存。
		return nil, err
	}
	s.logger.InfoContext(ctx, "stock transfer created successfully", "transfer_id", transfer.ID, "transfer_no", transfer.TransferNo)

	return transfer, nil
}

// CompleteTransfer 完成一个库存调拨单。
// ctx: 上下文。
// transferID: 调拨单ID。
// 返回可能发生的错误。
func (s *WarehouseService) CompleteTransfer(ctx context.Context, transferID uint64) error {
	transfer, err := s.repo.GetTransfer(ctx, transferID)
	if err != nil {
		return err
	}
	if transfer == nil {
		return errors.New("transfer not found")
	}
	// 如果调拨单已完成，则无需重复处理。
	if transfer.Status == entity.StockTransferStatusCompleted {
		return nil
	}
	// TODO: 考虑在事务中执行以下操作，以确保原子性。

	// 1. 从源仓库实际扣除库存和解锁库存。
	sourceStock, err := s.repo.GetStock(ctx, transfer.FromWarehouseID, transfer.SkuID)
	if err != nil {
		return err
	}
	if sourceStock == nil {
		return errors.New("source stock not found for transfer completion")
	}
	sourceStock.LockedStock -= transfer.Quantity // 解锁库存。
	sourceStock.Stock -= transfer.Quantity       // 实际扣减库存。
	if err := s.repo.SaveStock(ctx, sourceStock); err != nil {
		s.logger.ErrorContext(ctx, "failed to deduct source stock for transfer", "transfer_id", transferID, "error", err)
		return err
	}

	// 2. 增加目标仓库的库存。
	destStock, err := s.repo.GetStock(ctx, transfer.ToWarehouseID, transfer.SkuID)
	if err != nil {
		return err
	}
	if destStock == nil {
		// 如果目标仓库的库存记录不存在，则创建新的。
		destStock = &entity.WarehouseStock{
			WarehouseID: transfer.ToWarehouseID,
			SkuID:       transfer.SkuID,
			Stock:       0,
		}
	}
	destStock.Stock += transfer.Quantity // 增加目标库存。
	if err := s.repo.SaveStock(ctx, destStock); err != nil {
		s.logger.ErrorContext(ctx, "failed to add dest stock for transfer", "transfer_id", transferID, "error", err)
		// TODO: 如果这里失败，需要回滚源仓库的库存扣减操作。
		return err
	}

	// 3. 更新调拨单状态为已完成。
	transfer.Status = entity.StockTransferStatusCompleted
	now := time.Now()
	transfer.CompletedAt = &now // 记录完成时间。
	if err := s.repo.SaveTransfer(ctx, transfer); err != nil {
		s.logger.ErrorContext(ctx, "failed to complete transfer", "transfer_id", transferID, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "stock transfer completed successfully", "transfer_id", transferID)
	return nil
}

// GetTransfer 根据ID获取调拨单详情。
func (s *WarehouseService) GetTransfer(ctx context.Context, id uint64) (*entity.StockTransfer, error) {
	return s.repo.GetTransfer(ctx, id)
}

// ListTransfers 获取调拨单列表。
func (s *WarehouseService) ListTransfers(ctx context.Context, fromID, toID uint64, status *entity.StockTransferStatus, page, pageSize int) ([]*entity.StockTransfer, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListTransfers(ctx, fromID, toID, status, offset, pageSize)
}

// ListWarehouses 获取仓库列表。
// ctx: 上下文。
// page, pageSize: 分页参数。
// 返回Warehouse实体列表、总数和可能发生的错误。
func (s *WarehouseService) ListWarehouses(ctx context.Context, page, pageSize int) ([]*entity.Warehouse, int64, error) {
	offset := (page - 1) * pageSize
	// nil表示不按状态过滤，获取所有状态的仓库。
	return s.repo.ListWarehouses(ctx, nil, offset, pageSize)
}

// GetStock 获取指定仓库和SKU的库存信息。
// ctx: 上下文。
// warehouseID: 仓库ID。
// skuID: SKU ID。
// 返回WarehouseStock实体和可能发生的错误。
func (s *WarehouseService) GetStock(ctx context.Context, warehouseID, skuID uint64) (*entity.WarehouseStock, error) {
	return s.repo.GetStock(ctx, warehouseID, skuID)
}

// DeductStock 扣减库存（用于Saga分布式事务的正向操作）。
// ctx: 上下文。
// warehouseID: 仓库ID。
// skuID: SKU ID。
// quantity: 扣减数量。
// 返回可能发生的错误。
func (s *WarehouseService) DeductStock(ctx context.Context, warehouseID, skuID uint64, quantity int32) error {
	stock, err := s.repo.GetStock(ctx, warehouseID, skuID)
	if err != nil {
		return err
	}
	if stock == nil {
		return errors.New("stock not found for deduction")
	}
	if stock.Stock < quantity {
		return errors.New("insufficient stock for deduction")
	}

	stock.Stock -= quantity // 扣减库存。
	if err := s.repo.SaveStock(ctx, stock); err != nil {
		s.logger.ErrorContext(ctx, "failed to deduct stock for saga action", "warehouse_id", warehouseID, "sku_id", skuID, "quantity", quantity, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "stock deducted successfully for saga action", "warehouse_id", warehouseID, "sku_id", skuID, "quantity", quantity, "current_stock", stock.Stock)
	return nil
}

// RevertStock 回滚库存（用于Saga分布式事务的补偿操作）。
// ctx: 上下文。
// warehouseID: 仓库ID。
// skuID: SKU ID。
// quantity: 回滚数量。
// 返回可能发生的错误。
func (s *WarehouseService) RevertStock(ctx context.Context, warehouseID, skuID uint64, quantity int32) error {
	stock, err := s.repo.GetStock(ctx, warehouseID, skuID)
	if err != nil {
		return err
	}
	if stock == nil {
		// 备注：如果在回滚时库存记录不存在，这可能表示数据一致性问题，或者原始DeductStock操作从未成功。
		// 在补偿逻辑中，通常需要记录此异常或创建缺失的库存记录（如果业务允许）。
		s.logger.WarnContext(ctx, "stock record not found during revert, creating new stock record", "warehouse_id", warehouseID, "sku_id", skuID)
		stock = &entity.WarehouseStock{
			WarehouseID: warehouseID,
			SkuID:       skuID,
			Stock:       0,
		}
	}

	stock.Stock += quantity // 增加库存（回滚操作）。
	if err := s.repo.SaveStock(ctx, stock); err != nil {
		s.logger.ErrorContext(ctx, "failed to revert stock for saga compensation", "warehouse_id", warehouseID, "sku_id", skuID, "quantity", quantity, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "stock reverted successfully for saga compensation", "warehouse_id", warehouseID, "sku_id", skuID, "quantity", quantity, "current_stock", stock.Stock)
	return nil
}
