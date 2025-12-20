package application

import (
	"context"
	"fmt"
	"log/slog"

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

// RequestRefund 发起退款申请
func (s *RefundService) RequestRefund(ctx context.Context, paymentID uint64, amount int64, reason string) (*domain.Refund, error) {
	payment, err := s.paymentRepo.FindByID(ctx, paymentID)
	if err != nil || payment == nil {
		return nil, fmt.Errorf("payment not found")
	}

	// 领域层：创建退款单
	refund, err := payment.CreateRefund(amount, reason)
	if err != nil {
		return nil, err
	}
	refund.ID = uint64(s.idGenerator.Generate())

	// 调用网关退款
	gateway, ok := s.gateways[payment.GatewayType]
	if !ok {
		return nil, fmt.Errorf("gateway not found: %s", payment.GatewayType)
	}

	gwReq := &domain.RefundGatewayRequest{
		PaymentID:     payment.PaymentNo,
		TransactionID: payment.TransactionID,
		RefundID:      refund.RefundNo,
		Amount:        amount,
		Reason:        reason,
	}
	gwResp, err := gateway.Refund(ctx, gwReq)
	if err != nil {
		s.logger.ErrorContext(ctx, "gateway refund failed", "payment_id", payment.ID, "error", err)
		// 记录失败日志，但不阻断，可能需要人工介入
		// 这里简单处理：直接返回错误
		return nil, err
	}

	// 更新退款结果（假设网关同步返回结果）
	// 注意：很多网关退款也是异步的，这里做了简化
	success := true
	if gwResp.Status == "FAILED" {
		success = false
	}
	_ = payment.ProcessRefund(refund.RefundNo, success, gwResp.RefundID)

	// 事务性保存：同时更新 Payment 和保存 Refund (在 Repository 实现中应由事务保证)
	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return nil, err
	}
	// TODO: Save refund separately if needed, depends on repo implementation.
	// 目前 paymentRepo.Save 处理聚合吗？为简单起见或严格 DDD，假设是的。

	s.logger.InfoContext(ctx, "refund processed", "refund_no", refund.RefundNo, "success", success)
	return refund, nil
}
