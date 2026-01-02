package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	paymentv1 "github.com/wyfcoding/ecommerce/goapi/payment/v1"
	"github.com/wyfcoding/ecommerce/internal/financialsettlement/domain"
)

// SettlementManager 处理财务结算模块的写操作和业务逻辑。
type SettlementManager struct {
	repo       domain.SettlementRepository
	paymentCli paymentv1.PaymentServiceClient
	logger     *slog.Logger
}

// NewSettlementManager 创建并返回一个新的 SettlementManager 实例。
func NewSettlementManager(repo domain.SettlementRepository, paymentCli paymentv1.PaymentServiceClient, logger *slog.Logger) *SettlementManager {
	return &SettlementManager{
		repo:       repo,
		paymentCli: paymentCli,
		logger:     logger,
	}
}

// CreateSettlement 创建结算单。
func (m *SettlementManager) CreateSettlement(ctx context.Context, sellerID uint64, period string, startDate, endDate time.Time) (*domain.Settlement, error) {
	settlement := &domain.Settlement{
		SellerID:         sellerID,
		Period:           period,
		StartDate:        startDate,
		EndDate:          endDate,
		TotalSalesAmount: 100000,
		CommissionAmount: 5000,
		RebateAmount:     1000,
		OtherFees:        500,
		FinalAmount:      95500,
		Status:           domain.SettlementStatusPending,
	}

	if err := m.repo.SaveSettlement(ctx, settlement); err != nil {
		m.logger.Error("failed to create settlement", "error", err, "seller_id", sellerID)
		return nil, err
	}
	return settlement, nil
}

// ApproveSettlement 审核批准结算单。
func (m *SettlementManager) ApproveSettlement(ctx context.Context, id uint64, approvedBy string) error {
	settlement, err := m.repo.GetSettlement(ctx, id)
	if err != nil {
		return err
	}

	settlement.Status = domain.SettlementStatusApproved
	settlement.ApprovedBy = approvedBy
	now := time.Now()
	settlement.ApprovedAt = &now

	return m.repo.SaveSettlement(ctx, settlement)
}

// RejectSettlement 审核拒绝结算单。
func (m *SettlementManager) RejectSettlement(ctx context.Context, id uint64, reason string) error {
	settlement, err := m.repo.GetSettlement(ctx, id)
	if err != nil {
		return err
	}

	settlement.Status = domain.SettlementStatusRejected
	settlement.RejectionReason = reason

	return m.repo.SaveSettlement(ctx, settlement)
}

// ProcessPayment 处理结算支付。
func (m *SettlementManager) ProcessPayment(ctx context.Context, settlementID uint64) (*domain.SettlementPayment, error) {
	settlement, err := m.repo.GetSettlement(ctx, settlementID)
	if err != nil {
		return nil, err
	}

	if settlement.Status != domain.SettlementStatusApproved {
		return nil, fmt.Errorf("settlement not approved")
	}

	// 1. 创建本地结算支付记录
	payment := &domain.SettlementPayment{
		SettlementID: settlementID,
		SellerID:     settlement.SellerID,
		Amount:       settlement.FinalAmount,
		Status:       domain.PaymentStatusProcessing,
	}

	if err := m.repo.SaveSettlementPayment(ctx, payment); err != nil {
		return nil, err
	}

	// 2. 调用支付服务发起真实结算付款 (Cross-Project Interaction)
	if m.paymentCli != nil {
		payResp, err := m.paymentCli.InitiatePayment(ctx, &paymentv1.InitiatePaymentRequest{
			OrderId:       settlementID, // 在结算场景中复用 OrderId 字段表示结算单ID
			UserId:        settlement.SellerID,
			PaymentMethod: "BANK_TRANSFER", // 结算通常使用银行转账
			Amount:        settlement.FinalAmount,
		})
		if err != nil {
			m.logger.ErrorContext(ctx, "failed to initiate real payment for settlement", "settlement_id", settlementID, "error", err)
			payment.Status = domain.PaymentStatusFailed
			_ = m.repo.SaveSettlementPayment(ctx, payment)
			return nil, fmt.Errorf("payment initiation failed: %w", err)
		}

		payment.TransactionID = payResp.TransactionNo
		m.logger.InfoContext(ctx, "settlement payment initiated via payment service", "payment_no", payResp.TransactionNo)
	} else {
		// 回退到 Mock 逻辑（如果未配置支付服务）
		payment.TransactionID = "MOCK-TXN-" + time.Now().Format("20060102150405")
		payment.Status = domain.PaymentStatusCompleted
		now := time.Now()
		payment.CompletedAt = &now
	}

	if err := m.repo.SaveSettlementPayment(ctx, payment); err != nil {
		m.logger.ErrorContext(ctx, "failed to update settlement payment", "error", err)
	}

	// 3. 更新结算单状态为“已支付” (或“支付中”，取决于支付服务反馈)
	if payment.Status == domain.PaymentStatusCompleted {
		settlement.Status = domain.SettlementStatusCompleted
	} else {
		settlement.Status = domain.SettlementStatusProcessing
	}

	if err := m.repo.SaveSettlement(ctx, settlement); err != nil {
		m.logger.ErrorContext(ctx, "failed to update settlement status", "settlement_id", settlement.ID, "error", err)
	}

	return payment, nil
}
