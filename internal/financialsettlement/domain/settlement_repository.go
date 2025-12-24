package domain

import (
	"context"
	"time"
)

// SettlementRepository 是财务结算模块的仓储接口。
type SettlementRepository interface {
	// --- 结算管理 (Settlement methods) ---

	// SaveSettlement 将结算单实体保存到数据存储中。
	SaveSettlement(ctx context.Context, settlement *Settlement) error
	// GetSettlement 根据ID获取结算单实体。
	GetSettlement(ctx context.Context, settlementID uint64) (*Settlement, error)
	// ListSellerSettlements 列出指定卖家的所有结算单实体，支持分页。
	ListSellerSettlements(ctx context.Context, sellerID uint64, offset, limit int) ([]*Settlement, int64, error)
	// GetSettlementsByPeriod 获取指定时间范围内的所有结算单实体.
	GetSettlementsByPeriod(ctx context.Context, startDate, endDate time.Time) ([]*Settlement, error)

	// --- 订单管理 (SettlementOrder methods) ---

	// SaveSettlementOrder 将结算订单实体保存到数据存储中.
	SaveSettlementOrder(ctx context.Context, order *SettlementOrder) error
	// GetSettlementOrders 获取指定结算单ID的所有结算订单实体.
	GetSettlementOrders(ctx context.Context, settlementID uint64) ([]*SettlementOrder, error)

	// --- 支付管理 (SettlementPayment methods) ---

	// SaveSettlementPayment 将结算支付实体保存到数据存储中.
	SaveSettlementPayment(ctx context.Context, payment *SettlementPayment) error
	// GetSettlementPayment 根据支付ID获取结算支付实体.
	GetSettlementPayment(ctx context.Context, paymentID uint64) (*SettlementPayment, error)
	// GetSettlementPaymentBySettlementID 根据结算单ID获取对应的结算支付实体.
	GetSettlementPaymentBySettlementID(ctx context.Context, settlementID uint64) (*SettlementPayment, error)

	// --- 统计 (Statistics methods) ---

	// GetSettlementStatistics 获取指定时间范围内的结算统计数据.
	GetSettlementStatistics(ctx context.Context, startDate, endDate time.Time) (*SettlementStatistics, error)
}
