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

// NewSettlementService 创建并返回一个新的 SettlementService 实例。
func NewSettlementService(manager *SettlementManager, query *SettlementQuery) *SettlementService {
	return &SettlementService{
		manager: manager,
		query:   query,
	}
}

// CreateSettlement 创建一个新的结算单。
func (s *SettlementService) CreateSettlement(ctx context.Context, sellerID uint64, period string, startDate, endDate time.Time) (*domain.Settlement, error) {
	return s.manager.CreateSettlement(ctx, sellerID, period, startDate, endDate)
}

// ApproveSettlement 审核批准一个结算单。
func (s *SettlementService) ApproveSettlement(ctx context.Context, id uint64, approvedBy string) error {
	return s.manager.ApproveSettlement(ctx, id, approvedBy)
}

// RejectSettlement 审核拒绝一个结算单。
func (s *SettlementService) RejectSettlement(ctx context.Context, id uint64, reason string) error {
	return s.manager.RejectSettlement(ctx, id, reason)
}

// GetSettlement 获取指定ID的结算单详情。
func (s *SettlementService) GetSettlement(ctx context.Context, id uint64) (*domain.Settlement, error) {
	return s.query.GetSettlement(ctx, id)
}

// ListSettlements 获取结算单列表。
func (s *SettlementService) ListSettlements(ctx context.Context, sellerID uint64, page, pageSize int) ([]*domain.Settlement, int64, error) {
	offset := (page - 1) * pageSize
	return s.query.ListSellerSettlements(ctx, sellerID, offset, pageSize)
}

// ProcessPayment 处理结算单的支付流程。
func (s *SettlementService) ProcessPayment(ctx context.Context, settlementID uint64) (*domain.SettlementPayment, error) {
	return s.manager.ProcessPayment(ctx, settlementID)
}

// GetSettlementOrders 获取结算单明细。
func (s *SettlementService) GetSettlementOrders(ctx context.Context, settlementID uint64) ([]*domain.SettlementOrder, error) {
	return s.query.GetSettlementOrders(ctx, settlementID)
}

// GetSettlementPayment 获取结算支付详情。
func (s *SettlementService) GetSettlementPayment(ctx context.Context, settlementID uint64) (*domain.SettlementPayment, error) {
	return s.query.GetSettlementPayment(ctx, settlementID)
}

// GetStatistics 获取结算统计。
func (s *SettlementService) GetStatistics(ctx context.Context, startDate, endDate time.Time) (*domain.SettlementStatistics, error) {
	return s.query.GetStatistics(ctx, startDate, endDate)
}
