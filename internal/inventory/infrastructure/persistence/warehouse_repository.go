package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/inventory/domain/entity"     // 导入库存模块的领域实体定义。
	"github.com/wyfcoding/ecommerce/internal/inventory/domain/repository" // 导入库存模块的领域仓储接口。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

type warehouseRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewWarehouseRepository 创建并返回一个新的 warehouseRepository 实例。
// db: GORM数据库连接实例。
func NewWarehouseRepository(db *gorm.DB) repository.WarehouseRepository {
	return &warehouseRepository{db: db}
}

// Save 将仓库实体保存到数据库。
// 如果仓库已存在（通过ID），则更新其信息；如果不存在，则创建。
func (r *warehouseRepository) Save(ctx context.Context, warehouse *entity.Warehouse) error {
	return r.db.WithContext(ctx).Save(warehouse).Error
}

// GetByID 根据ID从数据库获取仓库记录。
func (r *warehouseRepository) GetByID(ctx context.Context, id uint64) (*entity.Warehouse, error) {
	var warehouse entity.Warehouse
	if err := r.db.WithContext(ctx).First(&warehouse, id).Error; err != nil {
		return nil, err
	}
	return &warehouse, nil
}

// ListAll 从数据库列出所有仓库记录。
func (r *warehouseRepository) ListAll(ctx context.Context) ([]*entity.Warehouse, error) {
	var list []*entity.Warehouse
	if err := r.db.WithContext(ctx).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
