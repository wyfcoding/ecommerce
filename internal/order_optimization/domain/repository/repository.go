package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/order_optimization/domain/entity" // 导入订单优化领域的实体定义。
)

// OrderOptimizationRepository 是订单优化模块的仓储接口。
// 它定义了对合并订单、拆分订单和仓库分配计划实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type OrderOptimizationRepository interface {
	// --- 合并订单 (MergedOrder methods) ---

	// SaveMergedOrder 将合并订单实体保存到数据存储中。
	// 如果实体已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// order: 待保存的合并订单实体。
	SaveMergedOrder(ctx context.Context, order *entity.MergedOrder) error
	// GetMergedOrder 根据ID获取合并订单实体。
	GetMergedOrder(ctx context.Context, id uint64) (*entity.MergedOrder, error)

	// --- 拆分订单 (SplitOrder methods) ---

	// SaveSplitOrder 将拆分订单实体保存到数据存储中。
	SaveSplitOrder(ctx context.Context, order *entity.SplitOrder) error
	// ListSplitOrders 列出指定原始订单ID的所有拆分订单实体。
	ListSplitOrders(ctx context.Context, originalOrderID uint64) ([]*entity.SplitOrder, error)

	// --- 仓库分配 (WarehouseAllocationPlan methods) ---

	// SaveAllocationPlan 将仓库分配计划实体保存到数据存储中。
	SaveAllocationPlan(ctx context.Context, plan *entity.WarehouseAllocationPlan) error
	// GetAllocationPlan 根据订单ID获取仓库分配计划实体。
	GetAllocationPlan(ctx context.Context, orderID uint64) (*entity.WarehouseAllocationPlan, error)
}
