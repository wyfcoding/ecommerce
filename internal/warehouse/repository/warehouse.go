package repository

import (
	"context"

	"gorm.io/gorm"

	"ecommerce/internal/warehouse/model"
)

// WarehouseRepo 仓库仓储接口
type WarehouseRepo interface {
	// 仓库管理
	CreateWarehouse(ctx context.Context, warehouse *model.Warehouse) error
	UpdateWarehouse(ctx context.Context, warehouse *model.Warehouse) error
	GetWarehouseByID(ctx context.Context, id uint64) (*model.Warehouse, error)
	GetWarehouseByCode(ctx context.Context, code string) (*model.Warehouse, error)
	ListWarehouses(ctx context.Context, status string, pageSize, pageNum int32) ([]*model.Warehouse, int64, error)
	
	// 仓库库存
	GetWarehouseStock(ctx context.Context, warehouseID, skuID uint64) (*model.WarehouseStock, error)
	ListWarehouseStocks(ctx context.Context, warehouseID uint64, pageSize, pageNum int32) ([]*model.WarehouseStock, int64, error)
	UpdateWarehouseStock(ctx context.Context, warehouseID, skuID uint64, quantity int32) error
	DeductWarehouseStock(ctx context.Context, warehouseID, skuID uint64, quantity int32) error
	AddWarehouseStock(ctx context.Context, warehouseID, skuID uint64, quantity int32) error
	AdjustWarehouseStock(ctx context.Context, warehouseID, skuID uint64, quantity int32) error
	
	// 库存分配
	CreateStockAllocation(ctx context.Context, allocation *model.StockAllocation) error
	GetStockAllocation(ctx context.Context, orderID uint64) (*model.StockAllocation, error)
	
	// 库存调拨
	CreateStockTransfer(ctx context.Context, transfer *model.StockTransfer) error
	UpdateStockTransfer(ctx context.Context, transfer *model.StockTransfer) error
	GetStockTransferByID(ctx context.Context, id uint64) (*model.StockTransfer, error)
	GetStockTransferByNo(ctx context.Context, transferNo string) (*model.StockTransfer, error)
	ListStockTransfers(ctx context.Context, status string, pageSize, pageNum int32) ([]*model.StockTransfer, int64, error)
	
	// 库存盘点
	CreateStocktaking(ctx context.Context, stocktaking *model.Stocktaking) error
	UpdateStocktaking(ctx context.Context, stocktaking *model.Stocktaking) error
	GetStocktakingByID(ctx context.Context, id uint64) (*model.Stocktaking, error)
	GetStocktakingByNo(ctx context.Context, stockNo string) (*model.Stocktaking, error)
	CreateStocktakingItem(ctx context.Context, item *model.StocktakingItem) error
	ListStocktakingItems(ctx context.Context, stocktakingID uint64) ([]*model.StocktakingItem, error)
	
	// 仓库区域
	CreateWarehouseArea(ctx context.Context, area *model.WarehouseArea) error
	ListWarehouseAreas(ctx context.Context, warehouseID uint64) ([]*model.WarehouseArea, error)
	
	// 事务
	InTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type warehouseRepo struct {
	db *gorm.DB
}

// NewWarehouseRepo 创建仓库仓储实例
func NewWarehouseRepo(db *gorm.DB) WarehouseRepo {
	return &warehouseRepo{db: db}
}

// CreateWarehouse 创建仓库
func (r *warehouseRepo) CreateWarehouse(ctx context.Context, warehouse *model.Warehouse) error {
	return r.db.WithContext(ctx).Create(warehouse).Error
}

// UpdateWarehouse 更新仓库
func (r *warehouseRepo) UpdateWarehouse(ctx context.Context, warehouse *model.Warehouse) error {
	return r.db.WithContext(ctx).Save(warehouse).Error
}

// GetWarehouseByID 根据ID获取仓库
func (r *warehouseRepo) GetWarehouseByID(ctx context.Context, id uint64) (*model.Warehouse, error) {
	var warehouse model.Warehouse
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&warehouse).Error
	if err != nil {
		return nil, err
	}
	return &warehouse, nil
}

// GetWarehouseByCode 根据编码获取仓库
func (r *warehouseRepo) GetWarehouseByCode(ctx context.Context, code string) (*model.Warehouse, error) {
	var warehouse model.Warehouse
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&warehouse).Error
	if err != nil {
		return nil, err
	}
	return &warehouse, nil
}

// ListWarehouses 获取仓库列表
func (r *warehouseRepo) ListWarehouses(ctx context.Context, status string, pageSize, pageNum int32) ([]*model.Warehouse, int64, error) {
	var warehouses []*model.Warehouse
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Warehouse{})
	
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	err := query.Offset(int(offset)).Limit(int(pageSize)).Order("priority DESC, created_at DESC").Find(&warehouses).Error
	if err != nil {
		return nil, 0, err
	}

	return warehouses, total, nil
}

// GetWarehouseStock 获取仓库库存
func (r *warehouseRepo) GetWarehouseStock(ctx context.Context, warehouseID, skuID uint64) (*model.WarehouseStock, error) {
	var stock model.WarehouseStock
	err := r.db.WithContext(ctx).
		Where("warehouse_id = ? AND sku_id = ?", warehouseID, skuID).
		First(&stock).Error
	if err != nil {
		return nil, err
	}
	return &stock, nil
}

// ListWarehouseStocks 获取仓库库存列表
func (r *warehouseRepo) ListWarehouseStocks(ctx context.Context, warehouseID uint64, pageSize, pageNum int32) ([]*model.WarehouseStock, int64, error) {
	var stocks []*model.WarehouseStock
	var total int64

	query := r.db.WithContext(ctx).Model(&model.WarehouseStock{}).Where("warehouse_id = ?", warehouseID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	err := query.Offset(int(offset)).Limit(int(pageSize)).Find(&stocks).Error
	if err != nil {
		return nil, 0, err
	}

	return stocks, total, nil
}

// UpdateWarehouseStock 更新仓库库存
func (r *warehouseRepo) UpdateWarehouseStock(ctx context.Context, warehouseID, skuID uint64, quantity int32) error {
	return r.db.WithContext(ctx).Model(&model.WarehouseStock{}).
		Where("warehouse_id = ? AND sku_id = ?", warehouseID, skuID).
		Update("stock", quantity).Error
}

// DeductWarehouseStock 扣减仓库库存
func (r *warehouseRepo) DeductWarehouseStock(ctx context.Context, warehouseID, skuID uint64, quantity int32) error {
	return r.db.WithContext(ctx).Model(&model.WarehouseStock{}).
		Where("warehouse_id = ? AND sku_id = ? AND stock >= ?", warehouseID, skuID, quantity).
		Update("stock", gorm.Expr("stock - ?", quantity)).Error
}

// AddWarehouseStock 增加仓库库存
func (r *warehouseRepo) AddWarehouseStock(ctx context.Context, warehouseID, skuID uint64, quantity int32) error {
	// 先尝试更新
	result := r.db.WithContext(ctx).Model(&model.WarehouseStock{}).
		Where("warehouse_id = ? AND sku_id = ?", warehouseID, skuID).
		Update("stock", gorm.Expr("stock + ?", quantity))

	if result.RowsAffected == 0 {
		// 如果不存在，创建新记录
		stock := &model.WarehouseStock{
			WarehouseID: warehouseID,
			SkuID:       skuID,
			Stock:       quantity,
		}
		return r.db.WithContext(ctx).Create(stock).Error
	}

	return result.Error
}

// AdjustWarehouseStock 调整仓库库存（可正可负）
func (r *warehouseRepo) AdjustWarehouseStock(ctx context.Context, warehouseID, skuID uint64, quantity int32) error {
	return r.db.WithContext(ctx).Model(&model.WarehouseStock{}).
		Where("warehouse_id = ? AND sku_id = ?", warehouseID, skuID).
		Update("stock", gorm.Expr("stock + ?", quantity)).Error
}

// CreateStockAllocation 创建库存分配
func (r *warehouseRepo) CreateStockAllocation(ctx context.Context, allocation *model.StockAllocation) error {
	return r.db.WithContext(ctx).Create(allocation).Error
}

// GetStockAllocation 获取库存分配
func (r *warehouseRepo) GetStockAllocation(ctx context.Context, orderID uint64) (*model.StockAllocation, error) {
	var allocation model.StockAllocation
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&allocation).Error
	if err != nil {
		return nil, err
	}
	return &allocation, nil
}

// CreateStockTransfer 创建库存调拨
func (r *warehouseRepo) CreateStockTransfer(ctx context.Context, transfer *model.StockTransfer) error {
	return r.db.WithContext(ctx).Create(transfer).Error
}

// UpdateStockTransfer 更新库存调拨
func (r *warehouseRepo) UpdateStockTransfer(ctx context.Context, transfer *model.StockTransfer) error {
	return r.db.WithContext(ctx).Save(transfer).Error
}

// GetStockTransferByID 根据ID获取库存调拨
func (r *warehouseRepo) GetStockTransferByID(ctx context.Context, id uint64) (*model.StockTransfer, error) {
	var transfer model.StockTransfer
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&transfer).Error
	if err != nil {
		return nil, err
	}
	return &transfer, nil
}

// GetStockTransferByNo 根据调拨单号获取库存调拨
func (r *warehouseRepo) GetStockTransferByNo(ctx context.Context, transferNo string) (*model.StockTransfer, error) {
	var transfer model.StockTransfer
	err := r.db.WithContext(ctx).Where("transfer_no = ?", transferNo).First(&transfer).Error
	if err != nil {
		return nil, err
	}
	return &transfer, nil
}

// ListStockTransfers 获取库存调拨列表
func (r *warehouseRepo) ListStockTransfers(ctx context.Context, status string, pageSize, pageNum int32) ([]*model.StockTransfer, int64, error) {
	var transfers []*model.StockTransfer
	var total int64

	query := r.db.WithContext(ctx).Model(&model.StockTransfer{})
	
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	err := query.Offset(int(offset)).Limit(int(pageSize)).Order("created_at DESC").Find(&transfers).Error
	if err != nil {
		return nil, 0, err
	}

	return transfers, total, nil
}

// CreateStocktaking 创建库存盘点
func (r *warehouseRepo) CreateStocktaking(ctx context.Context, stocktaking *model.Stocktaking) error {
	return r.db.WithContext(ctx).Create(stocktaking).Error
}

// UpdateStocktaking 更新库存盘点
func (r *warehouseRepo) UpdateStocktaking(ctx context.Context, stocktaking *model.Stocktaking) error {
	return r.db.WithContext(ctx).Save(stocktaking).Error
}

// GetStocktakingByID 根据ID获取库存盘点
func (r *warehouseRepo) GetStocktakingByID(ctx context.Context, id uint64) (*model.Stocktaking, error) {
	var stocktaking model.Stocktaking
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&stocktaking).Error
	if err != nil {
		return nil, err
	}
	return &stocktaking, nil
}

// GetStocktakingByNo 根据盘点单号获取库存盘点
func (r *warehouseRepo) GetStocktakingByNo(ctx context.Context, stockNo string) (*model.Stocktaking, error) {
	var stocktaking model.Stocktaking
	err := r.db.WithContext(ctx).Where("stock_no = ?", stockNo).First(&stocktaking).Error
	if err != nil {
		return nil, err
	}
	return &stocktaking, nil
}

// CreateStocktakingItem 创建盘点明细
func (r *warehouseRepo) CreateStocktakingItem(ctx context.Context, item *model.StocktakingItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// ListStocktakingItems 获取盘点明细列表
func (r *warehouseRepo) ListStocktakingItems(ctx context.Context, stocktakingID uint64) ([]*model.StocktakingItem, error) {
	var items []*model.StocktakingItem
	err := r.db.WithContext(ctx).
		Where("stocktaking_id = ?", stocktakingID).
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

// CreateWarehouseArea 创建仓库区域
func (r *warehouseRepo) CreateWarehouseArea(ctx context.Context, area *model.WarehouseArea) error {
	return r.db.WithContext(ctx).Create(area).Error
}

// ListWarehouseAreas 获取仓库区域列表
func (r *warehouseRepo) ListWarehouseAreas(ctx context.Context, warehouseID uint64) ([]*model.WarehouseArea, error) {
	var areas []*model.WarehouseArea
	err := r.db.WithContext(ctx).
		Where("warehouse_id = ? AND is_active = ?", warehouseID, true).
		Order("created_at ASC").
		Find(&areas).Error
	if err != nil {
		return nil, err
	}
	return areas, nil
}

// InTx 在事务中执行
func (r *warehouseRepo) InTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, "tx", tx))
	})
}
