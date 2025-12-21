package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/financial_settlement/domain"
)

// SettlementService 结构体定义了财务结算模块的应用服务标题。
// 它是一个门面（Facade），将复杂的结算逻辑委托给 Manager 和 Query 处理。
type SettlementService struct {
	manager *SettlementManager
	query   *SettlementQuery
}

// NewSettlementService 创建财务结算服务门面实例。
func NewSettlementService(manager *SettlementManager, query *SettlementQuery) *SettlementService {
	return &SettlementService{
		manager: manager,
		query:   query,
	}
}

// CreateSettlement 为指定商家创建一个新的周期性结算单。
func (s *SettlementService) CreateSettlement(ctx context.Context, sellerID uint64, period string, startDate, endDate time.Time) (*domain.Settlement, error) {
	return s.manager.CreateSettlement(ctx, sellerID, period, startDate, endDate)
}

// ApproveSettlement 财务审核：批准指定的结算单。
func (s *SettlementService) ApproveSettlement(ctx context.Context, id uint64, approvedBy string) error {
	return s.manager.ApproveSettlement(ctx, id, approvedBy)
}

// RejectSettlement 财务审核：驳回并拒绝指定的结算单。
func (s *SettlementService) RejectSettlement(ctx context.Context, id uint64, reason string) error {
	return s.manager.RejectSettlement(ctx, id, reason)
}

// GetSettlement 获取指定ID的结算单详细信息。
func (s *SettlementService) GetSettlement(ctx context.Context, id uint64) (*domain.Settlement, error) {
	return s.query.GetSettlement(ctx, id)
}

// ListSettlements 分页获取指定商家的结算单列表。
func (s *SettlementService) ListSettlements(ctx context.Context, sellerID uint64, page, pageSize int) ([]*domain.Settlement, int64, error) {
	offset := (page - 1) * pageSize
	return s.query.ListSellerSettlements(ctx, sellerID, offset, pageSize)
}

// ProcessPayment 触发结算单的打款/支付流程。
func (s *SettlementService) ProcessPayment(ctx context.Context, settlementID uint64) (*domain.SettlementPayment, error) {
	return s.manager.ProcessPayment(ctx, settlementID)
}

// GetSettlementOrders 获取结算单中包含的所有订单明细流水。
func (s *SettlementService) GetSettlementOrders(ctx context.Context, settlementID uint64) ([]*domain.SettlementOrder, error) {
	return s.query.GetSettlementOrders(ctx, settlementID)
}

// GetSettlementPayment 获取结算单对应的支付打款记录详情。
func (s *SettlementService) GetSettlementPayment(ctx context.Context, settlementID uint64) (*domain.SettlementPayment, error) {
	return s.query.GetSettlementPayment(ctx, settlementID)
}

// GetStatistics 获取指定时间段内的全局结算数据统计。
func (s *SettlementService) GetStatistics(ctx context.Context, startDate, endDate time.Time) (*domain.SettlementStatistics, error) {
	return s.query.GetStatistics(ctx, startDate, endDate)
}
