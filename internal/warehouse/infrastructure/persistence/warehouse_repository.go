package persistence

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/warehouse/domain"

	"gorm.io/gorm"
)

type warehouseRepository struct {
	db *gorm.DB
}

// NewWarehouseRepository 创建并返回一个新的 warehouseRepository 实例。
func NewWarehouseRepository(db *gorm.DB) domain.WarehouseRepository {
	return &warehouseRepository{db: db}
}

// --- 仓库管理 (Warehouse methods) ---

func (r *warehouseRepository) SaveWarehouse(ctx context.Context, warehouse *domain.Warehouse) error {
	return r.db.WithContext(ctx).Save(warehouse).Error
}

func (r *warehouseRepository) GetWarehouse(ctx context.Context, id uint64) (*domain.Warehouse, error) {
	var warehouse domain.Warehouse
	if err := r.db.WithContext(ctx).First(&warehouse, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &warehouse, nil
}

func (r *warehouseRepository) GetWarehouseByCode(ctx context.Context, code string) (*domain.Warehouse, error) {
	var warehouse domain.Warehouse
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&warehouse).Error; err != nil {
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

	db := r.db.WithContext(ctx).Model(&domain.Warehouse{})
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
	return r.db.WithContext(ctx).Save(stock).Error
}

func (r *warehouseRepository) GetStock(ctx context.Context, warehouseID, skuID uint64) (*domain.WarehouseStock, error) {
	var stock domain.WarehouseStock
	if err := r.db.WithContext(ctx).Where("warehouse_id = ? AND sku_id = ?", warehouseID, skuID).First(&stock).Error; err != nil {
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

	db := r.db.WithContext(ctx).Model(&domain.WarehouseStock{}).Where("warehouse_id = ?", warehouseID)
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
	return r.db.WithContext(ctx).Save(transfer).Error
}

func (r *warehouseRepository) GetTransfer(ctx context.Context, id uint64) (*domain.StockTransfer, error) {
	var transfer domain.StockTransfer
	if err := r.db.WithContext(ctx).First(&transfer, id).Error; err != nil {
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

	db := r.db.WithContext(ctx).Model(&domain.StockTransfer{})
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
