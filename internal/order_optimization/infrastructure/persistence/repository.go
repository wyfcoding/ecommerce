package persistence

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/order_optimization/domain"

	"gorm.io/gorm"
)

type orderOptimizationRepository struct {
	db *gorm.DB
}

// NewOrderOptimizationRepository 创建并返回一个新的 orderOptimizationRepository 实例。
func NewOrderOptimizationRepository(db *gorm.DB) domain.OrderOptimizationRepository {
	return &orderOptimizationRepository{db: db}
}

// --- 合并订单 (MergedOrder methods) ---

func (r *orderOptimizationRepository) SaveMergedOrder(ctx context.Context, order *domain.MergedOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *orderOptimizationRepository) GetMergedOrder(ctx context.Context, id uint64) (*domain.MergedOrder, error) {
	var order domain.MergedOrder
	if err := r.db.WithContext(ctx).First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// --- 拆分订单 (SplitOrder methods) ---

func (r *orderOptimizationRepository) SaveSplitOrder(ctx context.Context, order *domain.SplitOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *orderOptimizationRepository) ListSplitOrders(ctx context.Context, originalOrderID uint64) ([]*domain.SplitOrder, error) {
	var list []*domain.SplitOrder
	if err := r.db.WithContext(ctx).Where("original_order_id = ?", originalOrderID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// --- 仓库分配 (WarehouseAllocationPlan methods) ---

func (r *orderOptimizationRepository) SaveAllocationPlan(ctx context.Context, plan *domain.WarehouseAllocationPlan) error {
	return r.db.WithContext(ctx).Save(plan).Error
}

func (r *orderOptimizationRepository) GetAllocationPlan(ctx context.Context, orderID uint64) (*domain.WarehouseAllocationPlan, error) {
	var plan domain.WarehouseAllocationPlan
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&plan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &plan, nil
}
