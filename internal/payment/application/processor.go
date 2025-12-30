package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
	"github.com/wyfcoding/pkg/idgen"
)

// PaymentProcessor 结构体定义。
type PaymentProcessor struct {
	paymentRepo domain.PaymentRepository
	channelRepo domain.ChannelRepository
	riskService domain.RiskService
	idGenerator idgen.Generator
	gateways    map[domain.GatewayType]domain.PaymentGateway
	logger      *slog.Logger
}

// NewPaymentProcessor 创建支付处理服务实例。
func NewPaymentProcessor(
	paymentRepo domain.PaymentRepository,
	channelRepo domain.ChannelRepository,
	riskService domain.RiskService,
	idGenerator idgen.Generator,
	gateways map[domain.GatewayType]domain.PaymentGateway,
	logger *slog.Logger,
) *PaymentProcessor {
	return &PaymentProcessor{
		paymentRepo: paymentRepo,
		channelRepo: channelRepo,
		riskService: riskService,
		idGenerator: idGenerator,
		gateways:    gateways,
		logger:      logger,
	}
}

// InitiatePayment 发起支付交易。
func (s *PaymentProcessor) InitiatePayment(ctx context.Context, orderID uint64, userID uint64, amount int64, paymentMethodStr string) (*domain.Payment, *domain.PaymentGatewayResponse, error) {
	// 1. 风控前置检查
	if err := s.checkRisk(ctx, orderID, userID, amount, paymentMethodStr); err != nil {
		return nil, nil, err
	}

	// 2. 渠道/网关路由
	gatewayType, channelConfig := s.routeChannel(ctx, paymentMethodStr)
	gateway, ok := s.gateways[gatewayType]
	if !ok {
		s.logger.ErrorContext(ctx, "unsupported gateway", "method", paymentMethodStr, "type", gatewayType)
		return nil, nil, fmt.Errorf("unsupported gateway: %s", gatewayType)
	}

	// 3. 幂等性检查与单据创建
	existingPayment, err := s.paymentRepo.FindByOrderID(ctx, orderID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to query existing payment", "order_id", orderID, "error", err)
		return nil, nil, err
	}

	var payment *domain.Payment
	if existingPayment != nil {
		payment = existingPayment
		// 若允许重新发起（如之前的失败或待支付），更新支付方式
		if payment.Status == domain.PaymentPending || payment.Status == domain.PaymentFailed {
			payment.PaymentMethod = paymentMethodStr
			payment.GatewayType = gatewayType
		}
	} else {
		payment = domain.NewPayment(orderID, fmt.Sprintf("%d", orderID), userID, amount, paymentMethodStr, gatewayType)
		payment.ID = uint64(s.idGenerator.Generate())
	}

	// 4. 构建网关请求参数
	gatewayReq := s.buildGatewayRequest(payment, channelConfig)

	// 5. 调用第三方网关
	gatewayResp, err := gateway.Pay(ctx, gatewayReq)
	if err != nil {
		s.logger.ErrorContext(ctx, "gateway pay failed", "payment_no", payment.PaymentNo, "error", err)
		return nil, nil, err
	}

	// 6. 保存状态
	// 部分网关同步返回 TransactionID，需更新
	if gatewayResp.TransactionID != "" {
		payment.TransactionID = gatewayResp.TransactionID
	}

	if existingPayment != nil {
		if err := s.paymentRepo.Update(ctx, payment); err != nil {
			return nil, nil, err
		}
	} else {
		if err := s.paymentRepo.Save(ctx, payment); err != nil {
			return nil, nil, err
		}
	}

	// 7. 记录交易数据供风控后续分析（异步或同步）
	if err := s.riskService.RecordTransaction(ctx, &domain.RiskContext{UserID: userID, Amount: amount}); err != nil {
		s.logger.ErrorContext(ctx, "failed to record transaction for risk analysis", "user_id", userID, "amount", amount, "error", err)
	}

	s.logger.InfoContext(ctx, "payment initiated", "payment_id", payment.ID, "gateway", gatewayType)
	return payment, gatewayResp, nil
}

// checkRisk 内部风控检查。
func (s *PaymentProcessor) checkRisk(ctx context.Context, orderID, userID uint64, amount int64, method string) error {
	riskRes, err := s.riskService.CheckPrePayment(ctx, &domain.RiskContext{
		UserID:        userID,
		Amount:        amount,
		PaymentMethod: method,
		OrderID:       orderID,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "risk check error", "error", err)
		return err
	}
	if riskRes.Action == domain.RiskActionBlock {
		s.logger.WarnContext(ctx, "payment blocked by risk", "order_id", orderID, "reason", riskRes.Reason)
		return fmt.Errorf("risk_rejected: %s", riskRes.Description)
	}
	return nil
}

// routeChannel 路由到支付渠道。
func (s *PaymentProcessor) routeChannel(ctx context.Context, method string) (domain.GatewayType, *domain.ChannelConfig) {
	cfg, err := s.channelRepo.FindByCode(ctx, method)
	if err == nil && cfg != nil {
		return domain.GatewayType(cfg.Type), cfg
	}
	return s.resolveGatewayType(method), nil
}

// resolveGatewayType 解析网关类型。
func (s *PaymentProcessor) resolveGatewayType(method string) domain.GatewayType {
	switch method {
	case "alipay":
		return domain.GatewayTypeAlipay
	case "wechat":
		return domain.GatewayTypeWechat
	case "stripe":
		return domain.GatewayTypeStripe
	default:
		return domain.GatewayTypeMock
	}
}

// buildGatewayRequest 构建网关请求。
func (s *PaymentProcessor) buildGatewayRequest(p *domain.Payment, cfg *domain.ChannelConfig) *domain.PaymentGatewayRequest {
	req := &domain.PaymentGatewayRequest{
		OrderID:     p.PaymentNo,
		Amount:      p.Amount,
		Currency:    "CNY",
		Description: "Order " + p.OrderNo,
		NotifyURL:   "http://localhost/callback", // 实际应从配置读取
		ReturnURL:   "http://localhost/return",
		ExtraData:   make(map[string]string),
	}
	if cfg != nil {
		req.ExtraData["config_json"] = cfg.ConfigJSON
	}
	return req
}
