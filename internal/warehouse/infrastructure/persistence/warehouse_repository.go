package persistence

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/warehouse/domain"
	"github.com/wyfcoding/pkg/dtm"

	"gorm.io/gorm"
)

type warehouseRepository struct {
	db *gorm.DB
}

// NewWarehouseRepository 创建并返回一个新的 warehouseRepository 实例。
func NewWarehouseRepository(db *gorm.DB) domain.WarehouseRepository {
	return &warehouseRepository{db: db}
}

// getDB 尝试从 Context 获取事务 DB，否则返回默认 DB
func (r *warehouseRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value("tx_db").(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}

// --- 仓库管理 (Warehouse methods) ---

func (r *warehouseRepository) SaveWarehouse(ctx context.Context, warehouse *domain.Warehouse) error {
	return r.getDB(ctx).Save(warehouse).Error
}

func (r *warehouseRepository) GetWarehouse(ctx context.Context, id uint64) (*domain.Warehouse, error) {
	var warehouse domain.Warehouse
	if err := r.getDB(ctx).First(&warehouse, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &warehouse, nil
}

func (r *warehouseRepository) GetWarehouseByCode(ctx context.Context, code string) (*domain.Warehouse, error) {
	var warehouse domain.Warehouse
	if err := r.getDB(ctx).Where("code = ?", code).First(&warehouse).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &warehouse, nil
}

func (r *warehouseRepository) ListWarehouses(ctx context.Context, status *domain.WarehouseStatus, offset, limit int) ([]*domain.Warehouse, int64, error) {
	var list []*domain.Warehouse
	var total int64

	db := r.getDB(ctx).Model(&domain.Warehouse{})
	if status != nil {
		db = db.Where("status = ?", *status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// --- 库存管理 (WarehouseStock methods) ---

func (r *warehouseRepository) SaveStock(ctx context.Context, stock *domain.WarehouseStock) error {
	return r.getDB(ctx).Save(stock).Error
}

func (r *warehouseRepository) GetStock(ctx context.Context, warehouseID, skuID uint64) (*domain.WarehouseStock, error) {
	var stock domain.WarehouseStock
	if err := r.getDB(ctx).Where("warehouse_id = ? AND sku_id = ?", warehouseID, skuID).First(&stock).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &stock, nil
}

func (r *warehouseRepository) ListStocks(ctx context.Context, warehouseID uint64, offset, limit int) ([]*domain.WarehouseStock, int64, error) {
	var list []*domain.WarehouseStock
	var total int64

	db := r.getDB(ctx).Model(&domain.WarehouseStock{}).Where("warehouse_id = ?", warehouseID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// --- 调拨管理 (StockTransfer methods) ---

func (r *warehouseRepository) SaveTransfer(ctx context.Context, transfer *domain.StockTransfer) error {
	return r.getDB(ctx).Save(transfer).Error
}

func (r *warehouseRepository) GetTransfer(ctx context.Context, id uint64) (*domain.StockTransfer, error) {
	var transfer domain.StockTransfer
	if err := r.getDB(ctx).First(&transfer, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &transfer, nil
}

func (r *warehouseRepository) ListTransfers(ctx context.Context, fromWarehouseID, toWarehouseID uint64, status *domain.StockTransferStatus, offset, limit int) ([]*domain.StockTransfer, int64, error) {
	var list []*domain.StockTransfer
	var total int64

	db := r.getDB(ctx).Model(&domain.StockTransfer{})
	if fromWarehouseID > 0 {
		db = db.Where("from_warehouse_id = ?", fromWarehouseID)
	}
	if toWarehouseID > 0 {
		db = db.Where("to_warehouse_id = ?", toWarehouseID)
	}
	if status != nil {
		db = db.Where("status = ?", *status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *warehouseRepository) ListWarehousesWithStock(ctx context.Context, skuID uint64, minQty int32) ([]*domain.Warehouse, []int32, error) {
	var results []struct {
		domain.Warehouse
		AvailableStock int32
	}

	// 使用直接 Table("warehouses") 以便进行 Joins。
	err := r.getDB(ctx).Table("warehouses").
		Select("warehouses.*, (warehouse_stocks.stock - warehouse_stocks.locked_stock) as available_stock").
		Joins("JOIN warehouse_stocks ON warehouse_stocks.warehouse_id = warehouses.id").
		Where("warehouse_stocks.sku_id = ? AND (warehouse_stocks.stock - warehouse_stocks.locked_stock) >= ?", skuID, minQty).
		Where("warehouses.status = ? AND warehouses.deleted_at IS NULL", domain.WarehouseStatusActive).
		Find(&results).Error
	if err != nil {
		return nil, nil, err
	}

	warehouses := make([]*domain.Warehouse, len(results))
	stocks := make([]int32, len(results))
	for i, res := range results {
		w := res.Warehouse
		warehouses[i] = &w
		stocks[i] = res.AvailableStock
	}

	return warehouses, stocks, nil
}

func (r *warehouseRepository) ExecWithBarrier(ctx context.Context, barrier interface{}, fn func(ctx context.Context) error) error {
	return dtm.CallWithGorm(ctx, barrier, r.db, func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, "tx_db", tx)
		return fn(txCtx)
	})
}
