package persistence

import (
	"context"
	"errors"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/inventory/domain"
	"gorm.io/gorm"
)

type warehouseRepository struct {
	db *gorm.DB
}

// NewWarehouseRepository 创建并返回一个新的 warehouseRepository 实例。
// db: GORM数据库连接实例。
func NewWarehouseRepository(db *gorm.DB) domain.WarehouseRepository {
	return &warehouseRepository{db: db}
}

// Save 将仓库实体保存到数据库。
// 如果仓库已存在（通过ID），则更新其信息；如果不存在，则创建。
func (r *warehouseRepository) Save(ctx context.Context, warehouse *domain.Warehouse) error {
	if warehouse == nil {
		return nil
	}
	model := toWarehouseModel(warehouse)
	if model == nil {
		return nil
	}
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	if synced := toDomainWarehouse(model); synced != nil {
		*warehouse = *synced
	}
	return nil
}

// GetByID 根据ID从数据库获取仓库记录。
func (r *warehouseRepository) GetByID(ctx context.Context, id uint64) (*domain.Warehouse, error) {
	var warehouse WarehouseModel
	if err := r.db.WithContext(ctx).First(&warehouse, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainWarehouse(&warehouse), nil
}

// ListAll 从数据库列出所有仓库记录。
func (r *warehouseRepository) SaveWarehouse(ctx context.Context, warehouse *domain.Warehouse) error {
	return r.Save(ctx, warehouse)
}

func (r *warehouseRepository) GetWarehouse(ctx context.Context, warehouseID string) (*domain.Warehouse, error) {
	id, _ := strconv.ParseUint(warehouseID, 10, 64)
	return r.GetByID(ctx, id)
}

func (r *warehouseRepository) GetWarehouseByCode(ctx context.Context, warehouseCode string) (*domain.Warehouse, error) {
	var model WarehouseModel
	if err := r.db.WithContext(ctx).Where("warehouse_code = ?", warehouseCode).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainWarehouse(&model), nil
}

func (r *warehouseRepository) GetActiveWarehouses(ctx context.Context) ([]*domain.Warehouse, error) {
	return r.ListAll(ctx)
}

func (r *warehouseRepository) ListAll(ctx context.Context) ([]*domain.Warehouse, error) {
	var list []WarehouseModel
	if err := r.db.WithContext(ctx).Find(&list).Error; err != nil {
		return nil, err
	}
	result := make([]*domain.Warehouse, 0, len(list))
	for i := range list {
		w := toDomainWarehouse(&list[i])
		if w != nil {
			result = append(result, w)
		}
	}
	return result, nil
}

func (r *warehouseRepository) GetWarehousesByType(ctx context.Context, warehouseType domain.WarehouseType) ([]*domain.Warehouse, error) {
	var list []WarehouseModel
	if err := r.db.WithContext(ctx).Where("warehouse_type = ?", warehouseType).Find(&list).Error; err != nil {
		return nil, err
	}
	result := make([]*domain.Warehouse, 0, len(list))
	for i := range list {
		w := toDomainWarehouse(&list[i])
		if w != nil {
			result = append(result, w)
		}
	}
	return result, nil
}

func (r *warehouseRepository) UpdateWarehouse(ctx context.Context, warehouse *domain.Warehouse) error {
	return r.Save(ctx, warehouse)
}

func (r *warehouseRepository) DeleteWarehouse(ctx context.Context, warehouseID string) error {
	id, _ := strconv.ParseUint(warehouseID, 10, 64)
	return r.db.WithContext(ctx).Delete(&WarehouseModel{}, id).Error
}
