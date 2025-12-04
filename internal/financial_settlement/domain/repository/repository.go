package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/financial_settlement/domain/entity" // 导入财务结算领域的实体定义。
	"time"                                                                       // 导入时间包，用于查询条件。
)

// SettlementRepository 是财务结算模块的仓储接口。
// 它定义了对结算单、结算订单、结算支付和结算统计等实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type SettlementRepository interface {
	// --- 结算管理 (Settlement methods) ---

	// SaveSettlement 将结算单实体保存到数据存储中。
	// 如果结算单已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// settlement: 待保存的结算单实体。
	SaveSettlement(ctx context.Context, settlement *entity.Settlement) error
	// GetSettlement 根据ID获取结算单实体。
	GetSettlement(ctx context.Context, settlementID uint64) (*entity.Settlement, error)
	// ListSellerSettlements 列出指定卖家的所有结算单实体，支持分页。
	ListSellerSettlements(ctx context.Context, sellerID uint64, offset, limit int) ([]*entity.Settlement, int64, error)
	// GetSettlementsByPeriod 获取指定时间范围内的所有结算单实体。
	GetSettlementsByPeriod(ctx context.Context, startDate, endDate time.Time) ([]*entity.Settlement, error)

	// --- 订单管理 (SettlementOrder methods) ---

	// SaveSettlementOrder 将结算订单实体保存到数据存储中。
	SaveSettlementOrder(ctx context.Context, order *entity.SettlementOrder) error
	// GetSettlementOrders 获取指定结算单ID的所有结算订单实体。
	GetSettlementOrders(ctx context.Context, settlementID uint64) ([]*entity.SettlementOrder, error)

	// --- 支付管理 (SettlementPayment methods) ---

	// SaveSettlementPayment 将结算支付实体保存到数据存储中。
	SaveSettlementPayment(ctx context.Context, payment *entity.SettlementPayment) error
	// GetSettlementPayment 根据支付ID获取结算支付实体。
	GetSettlementPayment(ctx context.Context, paymentID uint64) (*entity.SettlementPayment, error)
	// GetSettlementPaymentBySettlementID 根据结算单ID获取对应的结算支付实体。
	GetSettlementPaymentBySettlementID(ctx context.Context, settlementID uint64) (*entity.SettlementPayment, error)

	// --- 统计 (Statistics methods) ---

	// GetSettlementStatistics 获取指定时间范围内的结算统计数据。
	GetSettlementStatistics(ctx context.Context, startDate, endDate time.Time) (*entity.SettlementStatistics, error)
}
