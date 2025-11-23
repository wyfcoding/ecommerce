package repository

import (
	"context"
	"ecommerce/internal/financial_settlement/domain/entity"
	"time"
)

// SettlementRepository 结算仓储接口
type SettlementRepository interface {
	// 结算管理
	SaveSettlement(ctx context.Context, settlement *entity.Settlement) error
	GetSettlement(ctx context.Context, settlementID uint64) (*entity.Settlement, error)
	ListSellerSettlements(ctx context.Context, sellerID uint64, offset, limit int) ([]*entity.Settlement, int64, error)
	GetSettlementsByPeriod(ctx context.Context, startDate, endDate time.Time) ([]*entity.Settlement, error)

	// 订单管理
	SaveSettlementOrder(ctx context.Context, order *entity.SettlementOrder) error
	GetSettlementOrders(ctx context.Context, settlementID uint64) ([]*entity.SettlementOrder, error)

	// 支付管理
	SaveSettlementPayment(ctx context.Context, payment *entity.SettlementPayment) error
	GetSettlementPayment(ctx context.Context, paymentID uint64) (*entity.SettlementPayment, error)
	GetSettlementPaymentBySettlementID(ctx context.Context, settlementID uint64) (*entity.SettlementPayment, error)

	// 统计
	GetSettlementStatistics(ctx context.Context, startDate, endDate time.Time) (*entity.SettlementStatistics, error)
}
