package domain

import "time"

const (
	OrderOptimizedEventType        = "orderoptimization.optimized"
	OrderMergedEventType           = "orderoptimization.merged"
	OrderSplitEventType            = "orderoptimization.split"
	OrderAllocationPlanCreatedType = "orderoptimization.allocation.plan.created"
)

// OrderOptimizedEvent 订单优化完成事件（通用）。
type OrderOptimizedEvent struct {
	OrderID   uint64    `json:"order_id"`
	UserID    uint64    `json:"user_id"`
	Strategy  string    `json:"strategy"`
	Benefit   string    `json:"benefit"`
	Timestamp time.Time `json:"timestamp"`
}

// OrderMergedEvent 合并订单事件。
type OrderMergedEvent struct {
	MergedOrderID    uint64    `json:"merged_order_id"`
	UserID           uint64    `json:"user_id"`
	OriginalOrderIDs []uint64  `json:"original_order_ids"`
	Timestamp        time.Time `json:"timestamp"`
}

// OrderSplitEvent 拆单事件。
type OrderSplitEvent struct {
	OriginalOrderID uint64    `json:"original_order_id"`
	SplitOrderIDs   []uint64  `json:"split_order_ids"`
	Timestamp       time.Time `json:"timestamp"`
}

// OrderAllocationPlanCreatedEvent 仓库分配计划创建事件。
type OrderAllocationPlanCreatedEvent struct {
	OrderID   uint64    `json:"order_id"`
	PlanID    uint64    `json:"plan_id"`
	Timestamp time.Time `json:"timestamp"`
}
