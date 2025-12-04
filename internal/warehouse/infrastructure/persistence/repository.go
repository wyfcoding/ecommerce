package persistence

import (
	"context"
	"errors" // 导入标准错误处理库。

	"github.com/wyfcoding/ecommerce/internal/warehouse/domain/entity"     // 导入仓库领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/warehouse/domain/repository" // 导入仓库领域的仓储接口。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

type warehouseRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewWarehouseRepository 创建并返回一个新的 warehouseRepository 实例。
func NewWarehouseRepository(db *gorm.DB) repository.WarehouseRepository {
	return &warehouseRepository{db: db}
}

// --- 仓库管理 (Warehouse methods) ---

// SaveWarehouse 将仓库实体保存到数据库。
// 如果实体已存在，则更新；如果不存在，则创建。
func (r *warehouseRepository) SaveWarehouse(ctx context.Context, warehouse *entity.Warehouse) error {
	return r.db.WithContext(ctx).Save(warehouse).Error
}

// GetWarehouse 根据ID从数据库获取仓库记录。
// 如果记录未找到，则返回nil。
func (r *warehouseRepository) GetWarehouse(ctx context.Context, id uint64) (*entity.Warehouse, error) {
	var warehouse entity.Warehouse
	if err := r.db.WithContext(ctx).First(&warehouse, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &warehouse, nil
}

// GetWarehouseByCode 根据仓库代码从数据库获取仓库记录。
// 如果记录未找到，则返回nil。
func (r *warehouseRepository) GetWarehouseByCode(ctx context.Context, code string) (*entity.Warehouse, error) {
	var warehouse entity.Warehouse
	// 按仓库代码过滤。
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&warehouse).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &warehouse, nil
}

// ListWarehouses 从数据库列出所有仓库记录，支持通过状态过滤和分页。
func (r *warehouseRepository) ListWarehouses(ctx context.Context, status *entity.WarehouseStatus, offset, limit int) ([]*entity.Warehouse, int64, error) {
	var list []*entity.Warehouse
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Warehouse{})
	if status != nil { // 如果提供了状态，则按状态过滤。
		db = db.Where("status = ?", *status)
	}

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序（按ID降序）。
	if err := db.Offset(offset).Limit(limit).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// --- 库存管理 (WarehouseStock methods) ---

// SaveStock 将仓库库存实体保存到数据库。
func (r *warehouseRepository) SaveStock(ctx context.Context, stock *entity.WarehouseStock) error {
	return r.db.WithContext(ctx).Save(stock).Error
}

// GetStock 获取指定仓库ID和SKUID的库存记录。
// 如果记录未找到，则返回nil。
func (r *warehouseRepository) GetStock(ctx context.Context, warehouseID, skuID uint64) (*entity.WarehouseStock, error) {
	var stock entity.WarehouseStock
	// 按仓库ID和SKUID过滤。
	if err := r.db.WithContext(ctx).Where("warehouse_id = ? AND sku_id = ?", warehouseID, skuID).First(&stock).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &stock, nil
}

// ListStocks 从数据库列出指定仓库ID的所有库存记录，支持分页。
func (r *warehouseRepository) ListStocks(ctx context.Context, warehouseID uint64, offset, limit int) ([]*entity.WarehouseStock, int64, error) {
	var list []*entity.WarehouseStock
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.WarehouseStock{}).Where("warehouse_id = ?", warehouseID)
	if err := db.Count(&total).Error; err != nil { // 统计总记录数。
		return nil, 0, err
	}

	// 应用分页和排序（按ID降序）。
	if err := db.Offset(offset).Limit(limit).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// --- 调拨管理 (StockTransfer methods) ---

// SaveTransfer 将库存调拨实体保存到数据库。
func (r *warehouseRepository) SaveTransfer(ctx context.Context, transfer *entity.StockTransfer) error {
	return r.db.WithContext(ctx).Save(transfer).Error
}

// GetTransfer 根据ID从数据库获取库存调拨记录。
// 如果记录未找到，则返回nil。
func (r *warehouseRepository) GetTransfer(ctx context.Context, id uint64) (*entity.StockTransfer, error) {
	var transfer entity.StockTransfer
	if err := r.db.WithContext(ctx).First(&transfer, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &transfer, nil
}

// ListTransfers 从数据库列出库存调拨记录，支持通过调出/调入仓库ID、状态过滤和分页。
func (r *warehouseRepository) ListTransfers(ctx context.Context, fromWarehouseID, toWarehouseID uint64, status *entity.StockTransferStatus, offset, limit int) ([]*entity.StockTransfer, int64, error) {
	var list []*entity.StockTransfer
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.StockTransfer{})
	if fromWarehouseID > 0 { // 如果提供了调出仓库ID，则按此过滤。
		db = db.Where("from_warehouse_id = ?", fromWarehouseID)
	}
	if toWarehouseID > 0 { // 如果提供了调入仓库ID，则按此过滤。
		db = db.Where("to_warehouse_id = ?", toWarehouseID)
	}
	if status != nil { // 如果提供了状态，则按状态过滤。
		db = db.Where("status = ?", *status)
	}

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序（按ID降序）。
	if err := db.Offset(offset).Limit(limit).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
