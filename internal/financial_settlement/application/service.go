package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/financial_settlement/domain/entity"     // 导入财务结算领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/financial_settlement/domain/repository" // 导入财务结算领域的仓储接口。

	"log/slog" // 导入结构化日志库。
)

// FinancialSettlementService 结构体定义了财务结算相关的应用服务。
// 它协调领域层和基础设施层，处理卖家结算单的生成、审核、支付等业务逻辑。
type FinancialSettlementService struct {
	repo   repository.SettlementRepository // 依赖SettlementRepository接口，用于数据持久化操作。
	logger *slog.Logger                    // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewFinancialSettlementService 创建并返回一个新的 FinancialSettlementService 实例。
func NewFinancialSettlementService(repo repository.SettlementRepository, logger *slog.Logger) *FinancialSettlementService {
	return &FinancialSettlementService{
		repo:   repo,
		logger: logger,
	}
}

// CreateSettlement 创建一个新的结算单。
// ctx: 上下文。
// sellerID: 关联的卖家ID。
// period: 结算周期描述（例如，“2023-01”）。
// startDate, endDate: 结算周期的起始和结束日期。
// 返回创建成功的Settlement实体和可能发生的错误。
func (s *FinancialSettlementService) CreateSettlement(ctx context.Context, sellerID uint64, period string, startDate, endDate time.Time) (*entity.Settlement, error) {
	// TODO: 在实际场景中，此处应根据订单数据计算各项金额。
	// 当前实现创建了一个带有模拟金额的占位结算单。
	settlement := &entity.Settlement{
		SellerID:         sellerID,
		Period:           period,
		StartDate:        startDate,
		EndDate:          endDate,
		TotalSalesAmount: 100000,                         // 模拟：总销售额。
		CommissionAmount: 5000,                           // 模拟：佣金。
		RebateAmount:     1000,                           // 模拟：返利。
		OtherFees:        500,                            // 模拟：其他费用。
		FinalAmount:      95500,                          // 模拟：最终结算金额。
		Status:           entity.SettlementStatusPending, // 初始状态为待处理。
	}

	// 通过仓储接口保存结算单。
	if err := s.repo.SaveSettlement(ctx, settlement); err != nil {
		s.logger.Error("failed to save settlement", "error", err)
		return nil, err
	}

	return settlement, nil
}

// ApproveSettlement 审核批准一个结算单。
// ctx: 上下文。
// id: 结算单ID。
// approvedBy: 批准操作人员的名称。
// 返回可能发生的错误。
func (s *FinancialSettlementService) ApproveSettlement(ctx context.Context, id uint64, approvedBy string) error {
	// 获取结算单实体。
	settlement, err := s.repo.GetSettlement(ctx, id)
	if err != nil {
		return err
	}

	// 更新结算单状态为已批准。
	settlement.Status = entity.SettlementStatusApproved
	settlement.ApprovedBy = approvedBy
	now := time.Now()
	settlement.ApprovedAt = &now // 记录批准时间。

	// 通过仓储接口保存更新后的结算单。
	return s.repo.SaveSettlement(ctx, settlement)
}

// RejectSettlement 审核拒绝一个结算单。
// ctx: 上下文。
// id: 结算单ID。
// reason: 拒绝的原因。
// 返回可能发生的错误。
func (s *FinancialSettlementService) RejectSettlement(ctx context.Context, id uint64, reason string) error {
	// 获取结算单实体。
	settlement, err := s.repo.GetSettlement(ctx, id)
	if err != nil {
		return err
	}

	// 更新结算单状态为已拒绝。
	settlement.Status = entity.SettlementStatusRejected
	settlement.RejectionReason = reason // 记录拒绝原因。

	// 通过仓储接口保存更新后的结算单。
	return s.repo.SaveSettlement(ctx, settlement)
}

// GetSettlement 获取指定ID的结算单详情。
// ctx: 上下文。
// id: 结算单ID。
// 返回Settlement实体和可能发生的错误。
func (s *FinancialSettlementService) GetSettlement(ctx context.Context, id uint64) (*entity.Settlement, error) {
	return s.repo.GetSettlement(ctx, id)
}

// ListSettlements 获取结算单列表。
// ctx: 上下文。
// sellerID: 筛选结算单的卖家ID。
// page, pageSize: 分页参数。
// 返回结算单列表、总数和可能发生的错误。
func (s *FinancialSettlementService) ListSettlements(ctx context.Context, sellerID uint64, page, pageSize int) ([]*entity.Settlement, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListSellerSettlements(ctx, sellerID, offset, pageSize)
}

// ProcessPayment 处理结算单的支付流程。
// ctx: 上下文。
// settlementID: 待支付的结算单ID。
// 返回支付记录实体和可能发生的错误。
func (s *FinancialSettlementService) ProcessPayment(ctx context.Context, settlementID uint64) (*entity.SettlementPayment, error) {
	// 获取结算单实体。
	settlement, err := s.repo.GetSettlement(ctx, settlementID)
	if err != nil {
		return nil, err
	}

	// 检查结算单状态，只有已批准的结算单才能进行支付。
	if settlement.Status != entity.SettlementStatusApproved {
		return nil, ProcessError{Message: "Settlement not approved"}
	}

	// 创建支付记录实体。
	payment := &entity.SettlementPayment{
		SettlementID:  settlementID,
		SellerID:      settlement.SellerID,
		Amount:        settlement.FinalAmount,
		Status:        entity.PaymentStatusProcessing,               // 初始状态为处理中。
		TransactionID: "TXN-" + time.Now().Format("20060102150405"), // 模拟生成交易ID。
	}

	// 保存支付记录。
	if err := s.repo.SaveSettlementPayment(ctx, payment); err != nil {
		return nil, err
	}

	// Simulate payment completion: 模拟支付完成。
	payment.Status = entity.PaymentStatusCompleted // 更新支付状态为已完成。
	now := time.Now()
	payment.CompletedAt = &now                 // 记录完成时间。
	s.repo.SaveSettlementPayment(ctx, payment) // 保存更新后的支付记录。

	// 更新结算单状态为已完成。
	settlement.Status = entity.SettlementStatusCompleted
	s.repo.SaveSettlement(ctx, settlement) // 保存更新后的结算单。

	return payment, nil
}

// ProcessError 是一个自定义错误类型，用于表示处理流程中的错误。
type ProcessError struct {
	Message string
}

// Error 方法实现了Go的 `error` 接口。
func (e ProcessError) Error() string {
	return e.Message
}
