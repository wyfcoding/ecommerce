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

func (s *RefundService) RequestRefund(ctx context.Context, userID, paymentID uint64, amount int64, reason string) (*domain.Refund, error) {
	payment, err := s.paymentRepo.FindByID(ctx, userID, paymentID)
	if err != nil || payment == nil {
		return nil, fmt.Errorf("payment not found")
	}

	// 1. 调用网关退款
	gateway, ok := s.gateways[payment.GatewayType]
	if !ok {
		return nil, fmt.Errorf("gateway not found: %s", payment.GatewayType)
	}

	if err := gateway.Refund(ctx, payment.TransactionID, amount); err != nil {
		return nil, err
	}

	var refund *domain.Refund
	// 2. 事务处理
	err = s.paymentRepo.Transaction(ctx, userID, func(tx any) error {
		txPaymentRepo := s.paymentRepo.WithTx(tx)
		txRefundRepo := s.refundRepo.WithTx(tx)

		// 重新加载支付单状态
		p, err := txPaymentRepo.FindByID(ctx, userID, paymentID)
		if err != nil {
			return err
		}

		// 状态机处理 (退款申请)
		if err := p.Trigger(ctx, "REFUND_REQ", reason); err != nil {
			return err
		}

		// 创建退款单
		refund = &domain.Refund{
			RefundNo:     fmt.Sprintf("REF%d", s.idGenerator.Generate()),
			PaymentID:    uint64(p.ID),
			PaymentNo:    p.PaymentNo,
			OrderID:      p.OrderID,
			OrderNo:      p.OrderNo,
			UserID:       p.UserID,
			RefundAmount: amount,
			Reason:       reason,
			Status:       p.Status,
		}

		// 状态机处理 (退款完成)
		if err := p.Trigger(ctx, "REFUND_FINISH", "Refund completed"); err != nil {
			return err
		}

		refund.Status = p.Status
		now := time.Now()
		refund.RefundedAt = &now

		// 保存支付单和退款单
		if err := txPaymentRepo.Update(ctx, p); err != nil {
			return err
		}
		return txRefundRepo.Save(ctx, refund)
	})
	if err != nil {
		return nil, err
	}

	return refund, nil
}
