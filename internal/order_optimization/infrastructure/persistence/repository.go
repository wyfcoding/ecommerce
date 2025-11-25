package persistence

import (
	"context"
	"errors"
	"github.com/wyfcoding/ecommerce/internal/order_optimization/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/order_optimization/domain/repository"

	"gorm.io/gorm"
)

type orderOptimizationRepository struct {
	db *gorm.DB
}

func NewOrderOptimizationRepository(db *gorm.DB) repository.OrderOptimizationRepository {
	return &orderOptimizationRepository{db: db}
}

// 合并订单
func (r *orderOptimizationRepository) SaveMergedOrder(ctx context.Context, order *entity.MergedOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *orderOptimizationRepository) GetMergedOrder(ctx context.Context, id uint64) (*entity.MergedOrder, error) {
	var order entity.MergedOrder
	if err := r.db.WithContext(ctx).First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// 拆分订单
func (r *orderOptimizationRepository) SaveSplitOrder(ctx context.Context, order *entity.SplitOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *orderOptimizationRepository) ListSplitOrders(ctx context.Context, originalOrderID uint64) ([]*entity.SplitOrder, error) {
	var list []*entity.SplitOrder
	if err := r.db.WithContext(ctx).Where("original_order_id = ?", originalOrderID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// 仓库分配
func (r *orderOptimizationRepository) SaveAllocationPlan(ctx context.Context, plan *entity.WarehouseAllocationPlan) error {
	return r.db.WithContext(ctx).Save(plan).Error
}

func (r *orderOptimizationRepository) GetAllocationPlan(ctx context.Context, orderID uint64) (*entity.WarehouseAllocationPlan, error) {
	var plan entity.WarehouseAllocationPlan
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&plan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &plan, nil
}
