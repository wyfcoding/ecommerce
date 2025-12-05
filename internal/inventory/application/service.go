package application

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/inventory/domain/entity"     // 导入库存领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/inventory/domain/repository" // 导入库存领域的仓储接口。
	"github.com/wyfcoding/ecommerce/pkg/algorithm"                        // 导入算法包，用于仓库分配。

	"log/slog" // 导入结构化日志库。
)

// InventoryService 结构体定义了库存管理相关的应用服务。
// 它协调领域层和基础设施层，处理库存的创建、增减、锁定、解锁以及分配等业务逻辑。
type InventoryService struct {
	repo          repository.InventoryRepository // 依赖InventoryRepository接口，用于库存数据的持久化操作。
	warehouseRepo repository.WarehouseRepository // 依赖WarehouseRepository接口，用于仓库数据的持久化操作。
	allocator     *algorithm.WarehouseAllocator  // 依赖仓库分配算法，用于优化订单的仓库分配。
	logger        *slog.Logger                   // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewInventoryService 创建并返回一个新的 InventoryService 实例。
func NewInventoryService(repo repository.InventoryRepository, warehouseRepo repository.WarehouseRepository, logger *slog.Logger) *InventoryService {
	return &InventoryService{
		repo:          repo,
		warehouseRepo: warehouseRepo,
		allocator:     algorithm.NewWarehouseAllocator(), // 初始化仓库分配器。
		logger:        logger,
	}
}

// CreateInventory 创建一个新的库存记录。
// ctx: 上下文。
// skuID: SKU ID。
// productID: 商品ID。
// warehouseID: 仓库ID。
// totalStock: 总库存量。
// warningThreshold: 警告阈值。
// 返回created successfully的Inventory实体和可能发生的错误。
func (s *InventoryService) CreateInventory(ctx context.Context, skuID, productID, warehouseID uint64, totalStock, warningThreshold int32) (*entity.Inventory, error) {
	// 检查指定SKU的库存记录是否已存在。
	// 注意：如果一个SKU可以存在于多个仓库，这个检查逻辑可能需要调整为检查特定仓库的SKU库存。
	existing, err := s.repo.GetBySkuID(ctx, skuID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("inventory already exists for this SKU") // 如果已存在，则返回错误。
	}

	inventory := entity.NewInventory(skuID, productID, warehouseID, totalStock, warningThreshold) // 创建Inventory实体。
	// 通过仓储接口保存库存。
	if err := s.repo.Save(ctx, inventory); err != nil {
		s.logger.ErrorContext(ctx, "failed to save inventory", "sku_id", skuID, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "inventory created successfully", "inventory_id", inventory.ID, "sku_id", skuID)
	return inventory, nil
}

// GetInventory 获取指定SKU的库存记录。
// ctx: 上下文。
// skuID: SKU ID。
// 返回Inventory实体和可能发生的错误。
func (s *InventoryService) GetInventory(ctx context.Context, skuID uint64) (*entity.Inventory, error) {
	return s.repo.GetBySkuID(ctx, skuID)
}

// AddStock 增加指定SKU的库存数量。
// ctx: 上下文。
// skuID: SKU ID。
// quantity: 增加的数量。
// reason: 增加库存的原因。
// 返回可能发生的错误。
func (s *InventoryService) AddStock(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	// 获取库存实体。
	inventory, err := s.repo.GetBySkuID(ctx, skuID)
	if err != nil {
		return err
	}
	if inventory == nil {
		return errors.New("inventory not found")
	}

	// 调用实体方法增加库存。
	if err := inventory.Add(quantity, reason); err != nil {
		return err
	}

	// 通过仓储接口保存更新后的库存。
	if err := s.repo.Save(ctx, inventory); err != nil {
		s.logger.ErrorContext(ctx, "failed to add stock", "sku_id", skuID, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "stock added successfully", "sku_id", skuID, "quantity", quantity, "reason", reason)
	return nil
}

// DeductStock 扣减指定SKU的库存数量。
// ctx: 上下文。
// skuID: SKU ID。
// quantity: 扣减的数量。
// reason: 扣减库存的原因。
// 返回可能发生的错误。
func (s *InventoryService) DeductStock(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	// 获取库存实体。
	inventory, err := s.repo.GetBySkuID(ctx, skuID)
	if err != nil {
		return err
	}
	if inventory == nil {
		return errors.New("inventory not found")
	}

	// 调用实体方法扣减库存。
	if err := inventory.Deduct(quantity, reason); err != nil {
		return err
	}

	// 通过仓储接口保存更新后的库存。
	if err := s.repo.Save(ctx, inventory); err != nil {
		s.logger.ErrorContext(ctx, "failed to deduct stock", "sku_id", skuID, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "stock deducted successfully", "sku_id", skuID, "quantity", quantity, "reason", reason)
	return nil
}

// LockStock 锁定指定SKU的库存数量。
// ctx: 上下文。
// skuID: SKU ID。
// quantity: 锁定的数量。
// reason: 锁定库存的原因。
// 返回可能发生的错误。
func (s *InventoryService) LockStock(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	// 获取库存实体。
	inventory, err := s.repo.GetBySkuID(ctx, skuID)
	if err != nil {
		return err
	}
	if inventory == nil {
		return errors.New("inventory not found")
	}

	// 调用实体方法锁定库存。
	if err := inventory.Lock(quantity, reason); err != nil {
		return err
	}

	// 通过仓储接口保存更新后的库存。
	if err := s.repo.Save(ctx, inventory); err != nil {
		s.logger.ErrorContext(ctx, "failed to lock stock", "sku_id", skuID, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "stock locked successfully", "sku_id", skuID, "quantity", quantity, "reason", reason)
	return nil
}

// UnlockStock 解锁指定SKU的库存数量。
// ctx: 上下文。
// skuID: SKU ID。
// quantity: 解锁的数量。
// reason: 解锁库存的原因。
// 返回可能发生的错误。
func (s *InventoryService) UnlockStock(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	// 获取库存实体。
	inventory, err := s.repo.GetBySkuID(ctx, skuID)
	if err != nil {
		return err
	}
	if inventory == nil {
		return errors.New("inventory not found")
	}

	// 调用实体方法解锁库存。
	if err := inventory.Unlock(quantity, reason); err != nil {
		return err
	}

	// 通过仓储接口保存更新后的库存。
	if err := s.repo.Save(ctx, inventory); err != nil {
		s.logger.ErrorContext(ctx, "failed to unlock stock", "sku_id", skuID, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "stock unlocked successfully", "sku_id", skuID, "quantity", quantity, "reason", reason)
	return nil
}

// ConfirmDeduction 确认扣减指定SKU的库存数量。
// ctx: 上下文。
// skuID: SKU ID。
// quantity: 确认扣减的数量。
// reason: 确认扣减的原因。
// 返回可能发生的错误。
func (s *InventoryService) ConfirmDeduction(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	// 获取库存实体。
	inventory, err := s.repo.GetBySkuID(ctx, skuID)
	if err != nil {
		return err
	}
	if inventory == nil {
		return errors.New("inventory not found")
	}

	// 调用实体方法确认扣减。
	if err := inventory.ConfirmDeduction(quantity, reason); err != nil {
		return err
	}

	// 通过仓储接口保存更新后的库存。
	if err := s.repo.Save(ctx, inventory); err != nil {
		s.logger.ErrorContext(ctx, "failed to confirm deduction", "sku_id", skuID, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "deduction confirmed successfully", "sku_id", skuID, "quantity", quantity, "reason", reason)
	return nil
}

// ListInventories 获取库存列表。
// ctx: 上下文。
// page, pageSize: 分页参数。
// 返回库存列表、总数和可能发生的错误。
func (s *InventoryService) ListInventories(ctx context.Context, page, pageSize int) ([]*entity.Inventory, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, offset, pageSize)
}

// GetInventoryLogs 获取指定库存的日志列表。
// ctx: 上下文。
// inventoryID: 库存ID。
// page, pageSize: 分页参数。
// 返回库存日志列表、总数和可能发生的错误。
func (s *InventoryService) GetInventoryLogs(ctx context.Context, inventoryID uint64, page, pageSize int) ([]*entity.InventoryLog, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.GetLogs(ctx, inventoryID, offset, pageSize)
}

// AllocateStock 分配订单商品到仓库。
// 此方法利用仓库分配算法，根据用户位置和商品需求，计算最优的仓库分配方案。
// ctx: 上下文。
// userLat, userLon: 用户的地理位置。
// items: 订单中的商品列表。
// 返回分配结果列表和可能发生的错误。
func (s *InventoryService) AllocateStock(ctx context.Context, userLat, userLon float64, items []algorithm.OrderItem) ([]algorithm.AllocationResult, error) {
	// 1. 获取所有仓库信息。
	warehouses, err := s.warehouseRepo.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	// 2. 获取订单商品对应的库存信息。
	skuIDs := make([]uint64, len(items))
	for i, item := range items {
		skuIDs[i] = item.SkuID
	}
	inventories, err := s.repo.GetBySkuIDs(ctx, skuIDs)
	if err != nil {
		return nil, err
	}

	// 3. 准备分配器所需的数据结构。
	// warehouseMap: map[warehouseID]map[skuID]*algorithm.WarehouseInfo
	warehouseMap := make(map[uint64]map[uint64]*algorithm.WarehouseInfo)

	// 辅助函数：根据ID查找仓库实体。
	findWarehouse := func(id uint64) *entity.Warehouse {
		for _, w := range warehouses {
			// 注意：gorm.Model的ID字段类型是uint，这里需要进行类型转换。
			if w.ID == uint(id) {
				return w
			}
		}
		return nil
	}

	// 填充 warehouseMap。
	for _, inv := range inventories {
		w := findWarehouse(inv.WarehouseID)
		if w == nil {
			continue // 数据一致性问题，应记录日志或报错。
		}

		if _, exists := warehouseMap[inv.WarehouseID]; !exists {
			warehouseMap[inv.WarehouseID] = make(map[uint64]*algorithm.WarehouseInfo)
		}

		warehouseMap[inv.WarehouseID][inv.SkuID] = &algorithm.WarehouseInfo{
			ID:       uint64(w.ID),
			Lat:      w.Lat,
			Lon:      w.Lon,
			Stock:    inv.AvailableStock, // 使用可用的库存。
			Priority: w.Priority,
			ShipCost: w.ShipCost,
		}
	}

	// 4. 调用仓库分配算法计算分配方案。
	results := s.allocator.AllocateOptimal(userLat, userLon, items, warehouseMap)

	// 5. 根据分配结果进行库存预留（简化）。
	// TODO: 在实际系统中，此处应在事务中锁定库存，以防止超卖。
	// 当前方法仅返回分配计划，不执行实际的库存锁定操作。
	return results, nil
}
