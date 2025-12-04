package persistence

import (
	"context"
	"errors"
	"github.com/wyfcoding/ecommerce/internal/order_optimization/domain/entity"     // 导入订单优化模块的领域实体定义。
	"github.com/wyfcoding/ecommerce/internal/order_optimization/domain/repository" // 导入订单优化模块的领域仓储接口。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

type orderOptimizationRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewOrderOptimizationRepository 创建并返回一个新的 orderOptimizationRepository 实例。
// db: GORM数据库连接实例。
func NewOrderOptimizationRepository(db *gorm.DB) repository.OrderOptimizationRepository {
	return &orderOptimizationRepository{db: db}
}

// --- 合并订单 (MergedOrder methods) ---

// SaveMergedOrder 将合并订单实体保存到数据库。
// 如果实体已存在（通过ID），则更新其信息；如果不存在，则创建。
func (r *orderOptimizationRepository) SaveMergedOrder(ctx context.Context, order *entity.MergedOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

// GetMergedOrder 根据ID从数据库获取合并订单记录。
// 如果记录未找到，则返回nil。
func (r *orderOptimizationRepository) GetMergedOrder(ctx context.Context, id uint64) (*entity.MergedOrder, error) {
	var order entity.MergedOrder
	if err := r.db.WithContext(ctx).First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &order, nil
}

// --- 拆分订单 (SplitOrder methods) ---

// SaveSplitOrder 将拆分订单实体保存到数据库。
func (r *orderOptimizationRepository) SaveSplitOrder(ctx context.Context, order *entity.SplitOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

// ListSplitOrders 从数据库列出指定原始订单ID的所有拆分订单记录。
func (r *orderOptimizationRepository) ListSplitOrders(ctx context.Context, originalOrderID uint64) ([]*entity.SplitOrder, error) {
	var list []*entity.SplitOrder
	if err := r.db.WithContext(ctx).Where("original_order_id = ?", originalOrderID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// --- 仓库分配 (WarehouseAllocationPlan methods) ---

// SaveAllocationPlan 将仓库分配计划实体保存到数据库。
func (r *orderOptimizationRepository) SaveAllocationPlan(ctx context.Context, plan *entity.WarehouseAllocationPlan) error {
	return r.db.WithContext(ctx).Save(plan).Error
}

// GetAllocationPlan 根据订单ID从数据库获取仓库分配计划记录。
// 如果记录未找到，则返回nil。
func (r *orderOptimizationRepository) GetAllocationPlan(ctx context.Context, orderID uint64) (*entity.WarehouseAllocationPlan, error) {
	var plan entity.WarehouseAllocationPlan
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&plan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &plan, nil
}
