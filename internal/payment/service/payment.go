package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"ecommerce/internal/payment/model"
	"ecommerce/internal/payment/repository"
	// 伪代码: 模拟支付网关和消息队列
	// "ecommerce/pkg/gateway/alipay"
	// "ecommerce/pkg/gateway/stripe"
	// "ecommerce/pkg/mq"
)

// PaymentService 定义了支付服务的业务逻辑接口
type PaymentService interface {
	CreatePayment(ctx context.Context, userID uint, orderSN string, amount float64, gateway string) (string, error)
	HandleWebhook(ctx context.Context, gateway string, payload []byte, headers http.Header) error
	CreateRefund(ctx context.Context, orderSN, reason string, amount float64) (*model.RefundTransaction, error)
}

// paymentService 是接口的具体实现
type paymentService struct {
	repo   repository.PaymentRepository
	logger *zap.Logger
	// alipayGateway alipay.Client
	// stripeGateway stripe.Client
	// mqProducer    mq.Producer
}

// NewPaymentService 创建一个新的 paymentService 实例
func NewPaymentService(repo repository.PaymentRepository, logger *zap.Logger) PaymentService {
	return &paymentService{repo: repo, logger: logger}
}

// CreatePayment 为订单创建支付，并返回支付凭证 (如支付URL或二维码数据)
func (s *paymentService) CreatePayment(ctx context.Context, userID uint, orderSN string, amount float64, gateway string) (string, error) {
	s.logger.Info("CreatePayment started", zap.String("orderSN", orderSN), zap.String("gateway", gateway))

	// 1. 创建内部支付流水记录
	tx := &model.PaymentTransaction{
		TransactionSN: uuid.New().String(),
		OrderSN:       orderSN,
		UserID:        userID,
		Amount:        amount,
		Gateway:       gateway,
		Status:        model.PaymentStatusPending,
	}
	if err := s.repo.CreatePaymentTransaction(ctx, tx); err != nil {
		s.logger.Error("Failed to create payment transaction in DB", zap.Error(err))
		return "", fmt.Errorf("创建支付流水失败: %w", err)
	}

	// 2. 根据选择的网关，调用对应的支付接口
	var paymentCredential string
	var err error

	switch gateway {
	case "stripe":
		// paymentCredential, err = s.stripeGateway.CreatePaymentIntent(tx.TransactionSN, int64(amount*100))
		paymentCredential = "https://stripe.com/pay/dummy_url"
	case "alipay":
		// paymentCredential, err = s.alipayGateway.CreatePagePayment(tx.TransactionSN, "商品标题", fmt.Sprintf("%.2f", amount))
		paymentCredential = "https://alipay.com/pay/dummy_url"
	default:
		err = fmt.Errorf("不支持的支付网关: %s", gateway)
	}

	if err != nil {
		s.logger.Error("Failed to create payment with gateway", zap.String("gateway", gateway), zap.Error(err))
		// 更新流水状态为失败
		tx.Status = model.PaymentStatusFailed
		tx.FailureReason = err.Error()
		s.repo.UpdatePaymentTransaction(context.Background(), tx)
		return "", err
	}

	s.logger.Info("Payment credential created", zap.String("orderSN", orderSN))
	return paymentCredential, nil
}

// HandleWebhook 处理来自第三方支付网关的异步回调通知
func (s *paymentService) HandleWebhook(ctx context.Context, gateway string, payload []byte, headers http.Header) error {
	s.logger.Info("HandleWebhook started", zap.String("gateway", gateway))

	var orderSN, gatewaySN, paymentMethod string
	var isSuccess bool
	var err error

	// 1. 验证并解析回调
	switch gateway {
	case "stripe":
		// event, err := s.stripeGateway.ParseWebhook(payload, headers.Get("Stripe-Signature"))
		// if event.Type == "payment_intent.succeeded" { ... }
		orderSN, gatewaySN, paymentMethod, isSuccess, err = "DUMMY_ORDER_SN", "stripe_sn_123", "card", true, nil
	case "alipay":
		// params, err := s.alipayGateway.ParseWebhook(payload)
		// if params.Get("trade_status") == "TRADE_SUCCESS" { ... }
		orderSN, gatewaySN, paymentMethod, isSuccess, err = "DUMMY_ORDER_SN", "alipay_sn_123", "alipay", true, nil
	default:
		err = fmt.Errorf("不支持的网关回调: %s", gateway)
	}

	if err != nil {
		s.logger.Error("Failed to parse webhook", zap.String("gateway", gateway), zap.Error(err))
		return err
	}

	// 2. 根据订单号找到对应的支付流水
	tx, err := s.repo.GetPaymentByOrderSN(ctx, orderSN)
	if err != nil || tx == nil {
		return fmt.Errorf("找不到对应的支付流水, orderSN: %s", orderSN)
	}

	// 3. 幂等性检查：如果已处理，则直接返回成功
	if tx.Status == model.PaymentStatusSuccess {
		s.logger.Info("Webhook already processed", zap.String("orderSN", orderSN))
		return nil
	}

	// 4. 更新支付流水状态
	if isSuccess {
		tx.Status = model.PaymentStatusSuccess
		tx.GatewaySN = gatewaySN
	} else {
		tx.Status = model.PaymentStatusFailed
		// tx.FailureReason = ... 从回调中获取失败原因
	}
	if err := s.repo.UpdatePaymentTransaction(ctx, tx); err != nil {
		return fmt.Errorf("更新支付流水状态失败: %w", err)
	}

	// 5. 如果支付成功，发送消息通知订单服务更新订单状态
	if isSuccess {
		// eventPayload := map[string]string{"orderSN": orderSN, "paymentSN": tx.TransactionSN, "paymentMethod": paymentMethod}
		// if err := s.mqProducer.Publish("payment.success", eventPayload); err != nil {
		// 	 s.logger.Error("Failed to publish payment success event", zap.Error(err))
		// 	 // 这里需要有重试或补偿机制
		// }
		s.logger.Info("Payment success event published", zap.String("orderSN", orderSN))
	}

	return nil
}

// CreateRefund 发起退款
func (s *paymentService) CreateRefund(ctx context.Context, orderSN, reason string, amount float64) (*model.RefundTransaction, error) {
	// 1. 找到原始支付记录
	// 2. 创建内部退款流水
	// 3. 调用第三方网关的退款接口
	// 4. 处理退款结果，更新退款流水状态
	// 5. (可选) 发送消息通知订单服务更新订单状态为已退款
	panic("implement me")
}