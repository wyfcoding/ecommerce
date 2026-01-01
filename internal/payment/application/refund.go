package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
	"github.com/wyfcoding/pkg/idgen"
)

type RefundService struct {
	paymentRepo domain.PaymentRepository
	refundRepo  domain.RefundRepository
	idGenerator idgen.Generator
	gateways    map[domain.GatewayType]domain.PaymentGateway
	logger      *slog.Logger
}

func NewRefundService(
	paymentRepo domain.PaymentRepository,
	refundRepo domain.RefundRepository,
	idGenerator idgen.Generator,
	gateways map[domain.GatewayType]domain.PaymentGateway,
	logger *slog.Logger,
) *RefundService {
	return &RefundService{
		paymentRepo: paymentRepo,
		refundRepo:  refundRepo,
		idGenerator: idGenerator,
		gateways:    gateways,
		logger:      logger,
	}
}

func (s *RefundService) RequestRefund(ctx context.Context, paymentID uint64, amount int64, reason string) (*domain.Refund, error) {
	payment, err := s.paymentRepo.FindByID(ctx, paymentID)
	if err != nil || payment == nil {
		return nil, fmt.Errorf("payment not found")
	}

	// 1. 调用网关退款
	gateway, ok := s.gateways[payment.GatewayType]
	if !ok {
		return nil, fmt.Errorf("gateway not found: %s", payment.GatewayType)
	}

	err = gateway.Refund(ctx, payment.TransactionID, amount)
	if err != nil {
		return nil, err
	}

	// 2. 状态机处理 (退款申请)
	if err := payment.Trigger(ctx, "REFUND_REQ", reason); err != nil {
		return nil, err
	}

	// 3. 创建退款单
	refund := &domain.Refund{
		RefundNo:     fmt.Sprintf("REF%d", s.idGenerator.Generate()),
		PaymentID:    uint64(payment.ID),
		PaymentNo:    payment.PaymentNo,
		OrderID:      payment.OrderID,
		OrderNo:      payment.OrderNo,
		UserID:       payment.UserID,
		RefundAmount: amount,
		Reason:       reason,
		Status:       payment.Status,
	}

	// 4. 退款完成 (这里简化为同步完成)
	if err := payment.Trigger(ctx, "REFUND_FINISH", "Refund completed"); err != nil {
		return nil, err
	}
	refund.Status = payment.Status
	now := time.Now()
	refund.RefundedAt = &now

	// 5. 保存
	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return nil, err
	}
	// TODO: s.refundRepo.Save(ctx, refund)

	return refund, nil
}
