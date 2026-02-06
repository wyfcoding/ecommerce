package mysql

import (
	"context"
	"errors"

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
	if w == nil {
		return nil
	}
	model := toWarehouseModel(w)
	if model == nil {
		return nil
	}
	if model.ID == 0 {
		if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
			return err
		}
	} else {
		if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
			return err
		}
	}
	if synced := toDomainWarehouse(model); synced != nil {
		*w = *synced
	}
	return nil
}

func (r *warehouseRepository) FindByID(ctx context.Context, id uint64) (*domain.Warehouse, error) {
	var w WarehouseModel
	if err := r.db.WithContext(ctx).First(&w, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainWarehouse(&w), nil
}

func (r *warehouseRepository) FindByCode(ctx context.Context, code string) (*domain.Warehouse, error) {
	var w WarehouseModel
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&w).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainWarehouse(&w), nil
}

func (r *warehouseRepository) List(ctx context.Context, offset, limit int) ([]*domain.Warehouse, int64, error) {
	var list []*WarehouseModel
	var total int64
	db := r.db.WithContext(ctx).Model(&WarehouseModel{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(limit).Order("priority desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	result := make([]*domain.Warehouse, 0, len(list))
	for _, m := range list {
		result = append(result, toDomainWarehouse(m))
	}
	return result, total, nil
}

func (r *warehouseRepository) GetStock(ctx context.Context, warehouseID, skuID uint64) (*domain.WarehouseStock, error) {
	var s WarehouseStockModel
	if err := r.db.WithContext(ctx).Where("warehouse_id = ? AND sku_id = ?", warehouseID, skuID).First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainWarehouseStock(&s), nil
}

func (r *warehouseRepository) GetStockWithLock(ctx context.Context, tx any, warehouseID, skuID uint64) (*domain.WarehouseStock, error) {
	var s WarehouseStockModel
	gormTx := tx.(*gorm.DB).WithContext(ctx)
	if err := gormTx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("warehouse_id = ? AND sku_id = ?", warehouseID, skuID).
		First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainWarehouseStock(&s), nil
}

func (r *warehouseRepository) SaveStock(ctx context.Context, s *domain.WarehouseStock) error {
	if s == nil {
		return nil
	}
	model := toWarehouseStockModel(s)
	if model.ID == 0 {
		if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
			return err
		}
	} else {
		if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
			return err
		}
	}
	if synced := toDomainWarehouseStock(model); synced != nil {
		*s = *synced
	}
	return nil
}

func (r *warehouseRepository) SaveStockInTx(ctx context.Context, tx any, s *domain.WarehouseStock) error {
	if s == nil {
		return nil
	}
	model := toWarehouseStockModel(s)
	gormTx := tx.(*gorm.DB).WithContext(ctx)
	if model.ID == 0 {
		if err := gormTx.Create(model).Error; err != nil {
			return err
		}
	} else {
		if err := gormTx.Save(model).Error; err != nil {
			return err
		}
	}
	if synced := toDomainWarehouseStock(model); synced != nil {
		*s = *synced
	}
	return nil
}

func (r *warehouseRepository) UpdateStock(ctx context.Context, s *domain.WarehouseStock) error {
	return r.SaveStock(ctx, s)
}

func (r *warehouseRepository) UpdateStockInTx(ctx context.Context, tx any, s *domain.WarehouseStock) error {
	return r.SaveStockInTx(ctx, tx, s)
}

func (r *warehouseRepository) SaveTransfer(ctx context.Context, t *domain.StockTransfer) error {
	return r.saveTransferWithTx(ctx, r.db, t)
}

func (r *warehouseRepository) SaveTransferInTx(ctx context.Context, tx any, t *domain.StockTransfer) error {
	if tx == nil {
		return r.SaveTransfer(ctx, t)
	}
	return r.saveTransferWithTx(ctx, tx.(*gorm.DB), t)
}

func (r *warehouseRepository) FindTransferByID(ctx context.Context, id uint64) (*domain.StockTransfer, error) {
	var t StockTransferModel
	if err := r.db.WithContext(ctx).First(&t, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainStockTransfer(&t), nil
}

func (r *warehouseRepository) FindTransferByNo(ctx context.Context, no string) (*domain.StockTransfer, error) {
	var t StockTransferModel
	if err := r.db.WithContext(ctx).Where("transfer_no = ?", no).First(&t).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainStockTransfer(&t), nil
}

func (r *warehouseRepository) UpdateTransfer(ctx context.Context, t *domain.StockTransfer) error {
	return r.SaveTransfer(ctx, t)
}

func (r *warehouseRepository) UpdateTransferInTx(ctx context.Context, tx any, t *domain.StockTransfer) error {
	return r.SaveTransferInTx(ctx, tx, t)
}

func (r *warehouseRepository) ListTransfers(ctx context.Context, fromID, toID uint64, status *string, offset, limit int) ([]*domain.StockTransfer, int64, error) {
	var list []*StockTransferModel
	var total int64
	db := r.db.WithContext(ctx).Model(&StockTransferModel{})
	if fromID > 0 {
		db = db.Where("from_warehouse_id = ?", fromID)
	}
	if toID > 0 {
		db = db.Where("to_warehouse_id = ?", toID)
	}
	if status != nil && *status != "" {
		db = db.Where("status = ?", *status)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(limit).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	result := make([]*domain.StockTransfer, 0, len(list))
	for _, m := range list {
		result = append(result, toDomainStockTransfer(m))
	}
	return result, total, nil
}

func (r *warehouseRepository) saveTransferWithTx(ctx context.Context, tx *gorm.DB, t *domain.StockTransfer) error {
	if t == nil {
		return nil
	}
	model := toStockTransferModel(t)
	gormTx := tx.WithContext(ctx)
	if model.ID == 0 {
		if err := gormTx.Create(model).Error; err != nil {
			return err
		}
	} else {
		if err := gormTx.Save(model).Error; err != nil {
			return err
		}
	}
	if synced := toDomainStockTransfer(model); synced != nil {
		*t = *synced
	}
	return nil
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
