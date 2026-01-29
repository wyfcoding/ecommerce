package persistence

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/warehouse/domain"
	"github.com/wyfcoding/pkg/dtm"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type warehouseRepository struct {
	db *gorm.DB
}

// NewWarehouseRepository 创建仓库仓储实现。
func NewWarehouseRepository(db *gorm.DB) domain.WarehouseRepository {
	return &warehouseRepository{db: db}
}

func (r *warehouseRepository) Save(ctx context.Context, w *domain.Warehouse) error {
	return r.db.WithContext(ctx).Save(w).Error
}

func (r *warehouseRepository) FindByID(ctx context.Context, id uint64) (*domain.Warehouse, error) {
	var w domain.Warehouse
	if err := r.db.WithContext(ctx).First(&w, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &w, nil
}

func (r *warehouseRepository) FindByCode(ctx context.Context, code string) (*domain.Warehouse, error) {
	var w domain.Warehouse
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&w).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &w, nil
}

func (r *warehouseRepository) List(ctx context.Context, offset, limit int) ([]*domain.Warehouse, int64, error) {
	var list []*domain.Warehouse
	var total int64
	db := r.db.WithContext(ctx).Model(&domain.Warehouse{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(limit).Order("priority desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *warehouseRepository) GetStock(ctx context.Context, warehouseID, skuID uint64) (*domain.WarehouseStock, error) {
	var s domain.WarehouseStock
	if err := r.db.WithContext(ctx).Where("warehouse_id = ? AND sku_id = ?", warehouseID, skuID).First(&s).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *warehouseRepository) GetStockWithLock(ctx context.Context, tx any, warehouseID, skuID uint64) (*domain.WarehouseStock, error) {
	var s domain.WarehouseStock
	gormTx := tx.(*gorm.DB).WithContext(ctx)
	// 使用悲观锁进行行级锁定
	if err := gormTx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("warehouse_id = ? AND sku_id = ?", warehouseID, skuID).
		First(&s).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *warehouseRepository) SaveStock(ctx context.Context, s *domain.WarehouseStock) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *warehouseRepository) SaveStockInTx(ctx context.Context, tx any, s *domain.WarehouseStock) error {
	return tx.(*gorm.DB).WithContext(ctx).Save(s).Error
}

func (r *warehouseRepository) UpdateStock(ctx context.Context, s *domain.WarehouseStock) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *warehouseRepository) UpdateStockInTx(ctx context.Context, tx any, s *domain.WarehouseStock) error {
	return tx.(*gorm.DB).WithContext(ctx).Save(s).Error
}

func (r *warehouseRepository) SaveTransfer(ctx context.Context, t *domain.StockTransfer) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *warehouseRepository) FindTransferByNo(ctx context.Context, no string) (*domain.StockTransfer, error) {
	var t domain.StockTransfer
	if err := r.db.WithContext(ctx).Where("transfer_no = ?", no).First(&t).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *warehouseRepository) UpdateTransfer(ctx context.Context, t *domain.StockTransfer) error {
	return r.db.WithContext(ctx).Save(t).Error
}

// 事务方法实现
func (r *warehouseRepository) BeginTx(ctx context.Context) any {
	return r.db.WithContext(ctx).Begin()
}

func (r *warehouseRepository) CommitTx(tx any) error {
	return tx.(*gorm.DB).Commit().Error
}

func (r *warehouseRepository) RollbackTx(tx any) error {
	return tx.(*gorm.DB).Rollback().Error
}

func (r *warehouseRepository) WithTx(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

func (r *warehouseRepository) WithBarrier(ctx context.Context, barrier any, fn func(tx any) error) error {
	if barrier == nil {
		return r.WithTx(ctx, fn)
	}

	// 使用 pkg/dtm 提供的反射助手，解耦 GORM 与 DTM Barrier 的集成
	return dtm.CallWithGorm(ctx, barrier, r.db.WithContext(ctx), func(tx *gorm.DB) error {
		return fn(tx)
	})
}
