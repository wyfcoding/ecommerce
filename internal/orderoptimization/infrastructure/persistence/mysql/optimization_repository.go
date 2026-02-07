package mysql

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/orderoptimization/domain"
	"gorm.io/gorm"
)

type orderOptimizationRepository struct {
	db *gorm.DB
}

// NewOrderOptimizationRepository 创建并返回一个新的 orderOptimizationRepository 实例。
func NewOrderOptimizationRepository(db *gorm.DB) domain.OrderOptimizationRepository {
	return &orderOptimizationRepository{db: db}
}

// --- tx helpers ---

func (r *orderOptimizationRepository) BeginTx(ctx context.Context) any {
	return r.db.WithContext(ctx).Begin()
}

func (r *orderOptimizationRepository) CommitTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Commit().Error
}

func (r *orderOptimizationRepository) RollbackTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Rollback().Error
}

func (r *orderOptimizationRepository) WithTx(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// --- 合并订单 (MergedOrder methods) ---

func (r *orderOptimizationRepository) SaveMergedOrder(ctx context.Context, order *domain.MergedOrder) error {
	return r.saveMergedOrderWithTx(ctx, r.db, order)
}

func (r *orderOptimizationRepository) SaveMergedOrderInTx(ctx context.Context, tx any, order *domain.MergedOrder) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveMergedOrderWithTx(ctx, gormTx, order)
}

func (r *orderOptimizationRepository) GetMergedOrder(ctx context.Context, id uint64) (*domain.MergedOrder, error) {
	var order MergedOrderModel
	if err := r.db.WithContext(ctx).First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toMergedOrder(&order), nil
}

// --- 拆分订单 (SplitOrder methods) ---

func (r *orderOptimizationRepository) SaveSplitOrder(ctx context.Context, order *domain.SplitOrder) error {
	return r.saveSplitOrderWithTx(ctx, r.db, order)
}

func (r *orderOptimizationRepository) SaveSplitOrderInTx(ctx context.Context, tx any, order *domain.SplitOrder) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveSplitOrderWithTx(ctx, gormTx, order)
}

func (r *orderOptimizationRepository) ListSplitOrders(ctx context.Context, originalOrderID uint64) ([]*domain.SplitOrder, error) {
	var list []*SplitOrderModel
	if err := r.db.WithContext(ctx).Where("original_order_id = ?", originalOrderID).Order("created_at asc").Find(&list).Error; err != nil {
		return nil, err
	}
	result := make([]*domain.SplitOrder, len(list))
	for i, item := range list {
		result[i] = toSplitOrder(item)
	}
	return result, nil
}

// --- 仓库分配 (WarehouseAllocationPlan methods) ---

func (r *orderOptimizationRepository) SaveAllocationPlan(ctx context.Context, plan *domain.WarehouseAllocationPlan) error {
	return r.saveAllocationPlanWithTx(ctx, r.db, plan)
}

func (r *orderOptimizationRepository) SaveAllocationPlanInTx(ctx context.Context, tx any, plan *domain.WarehouseAllocationPlan) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveAllocationPlanWithTx(ctx, gormTx, plan)
}

func (r *orderOptimizationRepository) GetAllocationPlan(ctx context.Context, orderID uint64) (*domain.WarehouseAllocationPlan, error) {
	var plan WarehouseAllocationPlanModel
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&plan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toAllocationPlan(&plan), nil
}

// --- internal helpers ---

func (r *orderOptimizationRepository) saveMergedOrderWithTx(ctx context.Context, tx *gorm.DB, order *domain.MergedOrder) error {
	if order == nil {
		return nil
	}
	model := toMergedOrderModel(order)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	if synced := toMergedOrder(model); synced != nil {
		*order = *synced
	}
	return nil
}

func (r *orderOptimizationRepository) saveSplitOrderWithTx(ctx context.Context, tx *gorm.DB, order *domain.SplitOrder) error {
	if order == nil {
		return nil
	}
	model := toSplitOrderModel(order)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	if synced := toSplitOrder(model); synced != nil {
		*order = *synced
	}
	return nil
}

func (r *orderOptimizationRepository) saveAllocationPlanWithTx(ctx context.Context, tx *gorm.DB, plan *domain.WarehouseAllocationPlan) error {
	if plan == nil {
		return nil
	}
	model := toAllocationPlanModel(plan)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	if synced := toAllocationPlan(model); synced != nil {
		*plan = *synced
	}
	return nil
}
