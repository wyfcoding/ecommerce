package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/warehouse/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/warehouse/domain/repository"
	"errors"

	"gorm.io/gorm"
)

type warehouseRepository struct {
	db *gorm.DB
}

func NewWarehouseRepository(db *gorm.DB) repository.WarehouseRepository {
	return &warehouseRepository{db: db}
}

// 仓库管理
func (r *warehouseRepository) SaveWarehouse(ctx context.Context, warehouse *entity.Warehouse) error {
	return r.db.WithContext(ctx).Save(warehouse).Error
}

func (r *warehouseRepository) GetWarehouse(ctx context.Context, id uint64) (*entity.Warehouse, error) {
	var warehouse entity.Warehouse
	if err := r.db.WithContext(ctx).First(&warehouse, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &warehouse, nil
}

func (r *warehouseRepository) GetWarehouseByCode(ctx context.Context, code string) (*entity.Warehouse, error) {
	var warehouse entity.Warehouse
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&warehouse).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &warehouse, nil
}

func (r *warehouseRepository) ListWarehouses(ctx context.Context, status *entity.WarehouseStatus, offset, limit int) ([]*entity.Warehouse, int64, error) {
	var list []*entity.Warehouse
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Warehouse{})
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

// 库存管理
func (r *warehouseRepository) SaveStock(ctx context.Context, stock *entity.WarehouseStock) error {
	return r.db.WithContext(ctx).Save(stock).Error
}

func (r *warehouseRepository) GetStock(ctx context.Context, warehouseID, skuID uint64) (*entity.WarehouseStock, error) {
	var stock entity.WarehouseStock
	if err := r.db.WithContext(ctx).Where("warehouse_id = ? AND sku_id = ?", warehouseID, skuID).First(&stock).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &stock, nil
}

func (r *warehouseRepository) ListStocks(ctx context.Context, warehouseID uint64, offset, limit int) ([]*entity.WarehouseStock, int64, error) {
	var list []*entity.WarehouseStock
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.WarehouseStock{}).Where("warehouse_id = ?", warehouseID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// 调拨管理
func (r *warehouseRepository) SaveTransfer(ctx context.Context, transfer *entity.StockTransfer) error {
	return r.db.WithContext(ctx).Save(transfer).Error
}

func (r *warehouseRepository) GetTransfer(ctx context.Context, id uint64) (*entity.StockTransfer, error) {
	var transfer entity.StockTransfer
	if err := r.db.WithContext(ctx).First(&transfer, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &transfer, nil
}

func (r *warehouseRepository) ListTransfers(ctx context.Context, fromWarehouseID, toWarehouseID uint64, status *entity.StockTransferStatus, offset, limit int) ([]*entity.StockTransfer, int64, error) {
	var list []*entity.StockTransfer
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.StockTransfer{})
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
