package mysql

import (
	"github.com/wyfcoding/ecommerce/internal/orderoptimization/domain"
	"gorm.io/gorm"
)

// MergedOrderModel 合并订单写模型。
type MergedOrderModel struct {
	gorm.Model
	UserID           uint64                 `gorm:"column:user_id;not null;index;comment:用户ID"`
	OriginalOrderIDs domain.Uint64Array     `gorm:"column:original_order_ids;type:json;comment:原始订单ID列表"`
	Items            domain.OrderItemArray  `gorm:"column:items;type:json;comment:订单项"`
	TotalAmount      int64                  `gorm:"column:total_amount;not null;default:0;comment:总金额(分)"`
	DiscountAmount   int64                  `gorm:"column:discount_amount;not null;default:0;comment:优惠金额(分)"`
	FinalAmount      int64                  `gorm:"column:final_amount;not null;default:0;comment:最终金额(分)"`
	ShippingAddress  domain.ShippingAddress `gorm:"column:shipping_address;type:json;comment:配送地址"`
	Status           string                 `gorm:"column:status;type:varchar(32);not null;comment:状态"`
}

// SplitOrderModel 拆分订单写模型。
type SplitOrderModel struct {
	gorm.Model
	OriginalOrderID uint64                 `gorm:"column:original_order_id;not null;index;comment:原始订单ID"`
	SplitIndex      int32                  `gorm:"column:split_index;not null;comment:拆分序号"`
	Items           domain.OrderItemArray  `gorm:"column:items;type:json;comment:订单项"`
	Amount          int64                  `gorm:"column:amount;not null;default:0;comment:金额(分)"`
	WarehouseID     uint64                 `gorm:"column:warehouse_id;not null;comment:仓库ID"`
	ShippingAddress domain.ShippingAddress `gorm:"column:shipping_address;type:json;comment:配送地址"`
	Status          string                 `gorm:"column:status;type:varchar(32);not null;comment:状态"`
}

// WarehouseAllocationPlanModel 仓库分配计划写模型。
type WarehouseAllocationPlanModel struct {
	gorm.Model
	OrderID     uint64                          `gorm:"column:order_id;not null;index;comment:订单ID"`
	Allocations domain.WarehouseAllocationArray `gorm:"column:allocations;type:json;comment:分配详情"`
}

func (MergedOrderModel) TableName() string             { return "merged_orders" }
func (SplitOrderModel) TableName() string              { return "split_orders" }
func (WarehouseAllocationPlanModel) TableName() string { return "warehouse_allocation_plans" }

func toMergedOrderModel(order *domain.MergedOrder) *MergedOrderModel {
	if order == nil {
		return nil
	}
	return &MergedOrderModel{
		Model: gorm.Model{
			ID:        order.ID,
			CreatedAt: order.CreatedAt,
			UpdatedAt: order.UpdatedAt,
		},
		UserID:           order.UserID,
		OriginalOrderIDs: order.OriginalOrderIDs,
		Items:            order.Items,
		TotalAmount:      order.TotalAmount,
		DiscountAmount:   order.DiscountAmount,
		FinalAmount:      order.FinalAmount,
		ShippingAddress:  order.ShippingAddress,
		Status:           order.Status,
	}
}

func toMergedOrder(model *MergedOrderModel) *domain.MergedOrder {
	if model == nil {
		return nil
	}
	return &domain.MergedOrder{
		ID:               model.ID,
		CreatedAt:        model.CreatedAt,
		UpdatedAt:        model.UpdatedAt,
		UserID:           model.UserID,
		OriginalOrderIDs: model.OriginalOrderIDs,
		Items:            model.Items,
		TotalAmount:      model.TotalAmount,
		DiscountAmount:   model.DiscountAmount,
		FinalAmount:      model.FinalAmount,
		ShippingAddress:  model.ShippingAddress,
		Status:           model.Status,
	}
}

func toSplitOrderModel(order *domain.SplitOrder) *SplitOrderModel {
	if order == nil {
		return nil
	}
	return &SplitOrderModel{
		Model: gorm.Model{
			ID:        order.ID,
			CreatedAt: order.CreatedAt,
			UpdatedAt: order.UpdatedAt,
		},
		OriginalOrderID: order.OriginalOrderID,
		SplitIndex:      order.SplitIndex,
		Items:           order.Items,
		Amount:          order.Amount,
		WarehouseID:     order.WarehouseID,
		ShippingAddress: order.ShippingAddress,
		Status:          order.Status,
	}
}

func toSplitOrder(model *SplitOrderModel) *domain.SplitOrder {
	if model == nil {
		return nil
	}
	return &domain.SplitOrder{
		ID:              model.ID,
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
		OriginalOrderID: model.OriginalOrderID,
		SplitIndex:      model.SplitIndex,
		Items:           model.Items,
		Amount:          model.Amount,
		WarehouseID:     model.WarehouseID,
		ShippingAddress: model.ShippingAddress,
		Status:          model.Status,
	}
}

func toAllocationPlanModel(plan *domain.WarehouseAllocationPlan) *WarehouseAllocationPlanModel {
	if plan == nil {
		return nil
	}
	return &WarehouseAllocationPlanModel{
		Model: gorm.Model{
			ID:        plan.ID,
			CreatedAt: plan.CreatedAt,
			UpdatedAt: plan.UpdatedAt,
		},
		OrderID:     plan.OrderID,
		Allocations: plan.Allocations,
	}
}

func toAllocationPlan(model *WarehouseAllocationPlanModel) *domain.WarehouseAllocationPlan {
	if model == nil {
		return nil
	}
	return &domain.WarehouseAllocationPlan{
		ID:          model.ID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		OrderID:     model.OrderID,
		Allocations: model.Allocations,
	}
}
