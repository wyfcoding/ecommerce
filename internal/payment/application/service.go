package application

import (
	"context"
	"errors"   // 导入标准错误处理库。
	"fmt"      // 导入格式化库。
	"log/slog" // 导入结构化日志库。

	"github.com/wyfcoding/ecommerce/internal/payment/domain" // 导入支付领域的领域层接口和实体。
	"github.com/wyfcoding/ecommerce/pkg/idgen"               // 导入ID生成器。
)

// PaymentApplicationService 结构体定义了支付管理相关的应用服务。
// 它协调领域层和基础设施层，处理支付的发起、回调处理、退款请求和查询等业务逻辑。
type PaymentApplicationService struct {
	paymentRepo domain.PaymentRepository // 依赖PaymentRepository接口，用于支付数据持久化操作。
	idGenerator idgen.Generator          // ID生成器，用于生成支付和退款ID。
	logger      *slog.Logger             // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewPaymentApplicationService 创建并返回一个新的 PaymentApplicationService 实例。
func NewPaymentApplicationService(paymentRepo domain.PaymentRepository, idGenerator idgen.Generator, logger *slog.Logger) *PaymentApplicationService {
	return &PaymentApplicationService{
		paymentRepo: paymentRepo,
		idGenerator: idGenerator,
		logger:      logger,
	}
}

// InitiatePayment 发起支付。
// 它会检查是否已存在该订单的支付记录，并根据情况返回现有支付或创建新的支付。
// ctx: 上下文。
// orderID: 关联的订单ID。
// userID: 支付用户ID。
// amount: 支付金额（单位：分）。
// paymentMethod: 支付方式。
// 返回支付实体和可能发生的错误。
func (s *PaymentApplicationService) InitiatePayment(ctx context.Context, orderID uint64, userID uint64, amount int64, paymentMethod string) (*domain.Payment, error) {
	// 1. 检查该订单是否已存在支付记录。
	existingPayment, err := s.paymentRepo.FindByOrderID(ctx, orderID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to find existing payment", "order_id", orderID, "error", err)
		return nil, err
	}
	if existingPayment != nil {
		// 如果已存在成功支付的记录，则直接返回错误。
		if existingPayment.Status == domain.PaymentSuccess {
			return nil, errors.New("order already paid")
		}
		// 如果已存在待处理的支付记录，则返回该记录（可能需要重新尝试支付或等待结果）。
		if existingPayment.Status == domain.PaymentPending {
			return existingPayment, nil
		}
	}

	// 2. 创建新的支付实体。
	payment := domain.NewPayment(orderID, fmt.Sprintf("%d", orderID), userID, amount, paymentMethod) // paymentNo 暂时使用 orderID 的字符串形式。
	payment.ID = uint64(s.idGenerator.Generate())                                                    // 生成支付ID。

	// 3. 保存新的支付实体。
	if err := s.paymentRepo.Save(ctx, payment); err != nil {
		s.logger.ErrorContext(ctx, "failed to save payment", "order_id", orderID, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "payment initiated successfully", "payment_id", payment.ID, "order_id", orderID)

	return payment, nil
}

// HandlePaymentCallback 处理支付回调。
// 这是由第三方支付平台通知支付结果的接口。
// ctx: 上下文。
// paymentNo: 支付单号。
// success: 支付是否成功。
// transactionID: 第三方支付平台返回的交易ID。
// thirdPartyNo: 第三方支付平台返回的支付流水号。
// 返回可能发生的错误。
func (s *PaymentApplicationService) HandlePaymentCallback(ctx context.Context, paymentNo string, success bool, transactionID, thirdPartyNo string) error {
	// 1. 根据支付单号查找支付记录。
	payment, err := s.paymentRepo.FindByPaymentNo(ctx, paymentNo)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to find payment by no", "payment_no", paymentNo, "error", err)
		return err
	}
	if payment == nil {
		return errors.New("payment not found")
	}

	// 2. 调用支付实体的方法处理支付结果。
	if err := payment.Process(success, transactionID, thirdPartyNo); err != nil {
		s.logger.ErrorContext(ctx, "failed to process payment", "payment_no", paymentNo, "error", err)
		return err
	}

	// 3. 更新支付记录。
	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		s.logger.ErrorContext(ctx, "failed to update payment", "payment_no", paymentNo, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "payment processed successfully", "payment_no", paymentNo, "success", success)
	return nil
}

// GetPaymentStatus 获取支付状态。
// ctx: 上下文。
// paymentID: 支付ID。
// 返回支付实体和可能发生的错误。
func (s *PaymentApplicationService) GetPaymentStatus(ctx context.Context, paymentID uint64) (*domain.Payment, error) {
	return s.paymentRepo.FindByID(ctx, paymentID)
}

// RequestRefund 请求退款。
// ctx: 上下文。
// paymentID: 关联的支付ID。
// amount: 退款金额（单位：分）。
// reason: 退款原因。
// 返回退款实体和可能发生的错误。
func (s *PaymentApplicationService) RequestRefund(ctx context.Context, paymentID uint64, amount int64, reason string) (*domain.Refund, error) {
	// 1. 查找关联的支付记录。
	payment, err := s.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to find payment", "payment_id", paymentID, "error", err)
		return nil, err
	}
	if payment == nil {
		return nil, errors.New("payment not found")
	}

	// 2. 调用支付实体的方法创建退款记录。
	refund, err := payment.CreateRefund(amount, reason)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create refund", "payment_id", paymentID, "error", err)
		return nil, err
	}
	refund.ID = uint64(s.idGenerator.Generate()) // 生成退款ID。

	// 3. 更新支付记录（包含新的退款信息）。
	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		s.logger.ErrorContext(ctx, "failed to update payment for refund", "payment_id", paymentID, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "refund requested successfully", "payment_id", paymentID, "refund_amount", amount)

	return refund, nil
}

// HandleRefundCallback 处理退款回调。
// 这是由第三方支付平台通知退款结果的接口。
// ctx: 上下文。
// refundNo: 退款单号。
// success: 退款是否成功。
// 返回可能发生的错误。
func (s *PaymentApplicationService) HandleRefundCallback(ctx context.Context, refundNo string, success bool) error {
	// TODO: 目前该方法未完全实现，因为缺少通过 refundNo 查找退款记录的仓储方法。
	// 理想情况下，需要 PaymentRepository 支持 FindRefundByRefundNo 或创建一个独立的 RefundRepository。
	// 或者，退款回调中应包含 PaymentID，以便能找到对应的 Payment 聚合根。
	return errors.New("HandleRefundCallback not fully implemented due to missing repo method")
}

// GetRefundStatus 获取退款状态。
// ctx: 上下文。
// refundID: 退款ID。
// 返回退款实体和可能发生的错误。
func (s *PaymentApplicationService) GetRefundStatus(ctx context.Context, refundID uint64) (*domain.Refund, error) {
	// TODO: 目前该方法未完全实现，因为缺少通过 refundID 查找退款记录的仓储方法。
	// 理想情况下，需要 PaymentRepository 支持 FindRefundByID 或 GetRefundsByPaymentID。
	return nil, errors.New("GetRefundStatus not fully implemented")
}
