package application

import (
	"context"
	"fmt"
	"log/slog"

	settlementv1 "github.com/wyfcoding/ecommerce/go-api/settlement/v1"
	"github.com/wyfcoding/ecommerce/internal/payment/domain"
	"github.com/wyfcoding/ecommerce/pkg/idgen"
)

// PaymentApplicationService 支付应用服务
// 负责协调领域对象、基础设施（网关、仓储、风控）以完成支付业务用例
type PaymentApplicationService struct {
	paymentRepo   domain.PaymentRepository                     // 支付仓储
	refundRepo    domain.RefundRepository                      // 退款仓储
	channelRepo   domain.ChannelRepository                     // 渠道配置仓储
	riskService   domain.RiskService                           // 风控服务
	idGenerator   idgen.Generator                              // ID生成器
	gateways      map[domain.GatewayType]domain.PaymentGateway // 网关策略集
	settlementCli settlementv1.SettlementServiceClient
	logger        *slog.Logger // 日志组件
}

// NewPaymentApplicationService 构造函数
func NewPaymentApplicationService(
	paymentRepo domain.PaymentRepository,
	refundRepo domain.RefundRepository,
	channelRepo domain.ChannelRepository,
	riskService domain.RiskService,
	idGenerator idgen.Generator,
	gateways map[domain.GatewayType]domain.PaymentGateway,
	settlementCli settlementv1.SettlementServiceClient,
	logger *slog.Logger,
) *PaymentApplicationService {
	return &PaymentApplicationService{
		paymentRepo:   paymentRepo,
		refundRepo:    refundRepo,
		channelRepo:   channelRepo,
		riskService:   riskService,
		idGenerator:   idGenerator,
		gateways:      gateways,
		settlementCli: settlementCli,
		logger:        logger,
	}
}

// InitiatePayment 发起支付交易
// 核心流程：风控检查 -> 渠道路由 -> 创建/获取单据 -> 调用三方网关
func (s *PaymentApplicationService) InitiatePayment(ctx context.Context, orderID uint64, userID uint64, amount int64, paymentMethodStr string) (*domain.Payment, *domain.PaymentGatewayResponse, error) {
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
	_ = s.riskService.RecordTransaction(ctx, &domain.RiskContext{UserID: userID, Amount: amount})

	s.logger.InfoContext(ctx, "payment initiated", "payment_id", payment.ID, "gateway", gatewayType)
	return payment, gatewayResp, nil
}

// HandlePaymentCallback 处理支付结果异步回调
func (s *PaymentApplicationService) HandlePaymentCallback(ctx context.Context, paymentNo string, success bool, transactionID, thirdPartyNo string, callbackData map[string]string) error {
	s.logger.InfoContext(ctx, "handling payment callback", "payment_no", paymentNo, "success", success)

	payment, err := s.paymentRepo.FindByPaymentNo(ctx, paymentNo)
	if err != nil || payment == nil {
		s.logger.ErrorContext(ctx, "payment not found", "payment_no", paymentNo)
		return fmt.Errorf("payment not found")
	}

	// 验证签名
	gateway, ok := s.gateways[payment.GatewayType]
	if ok {
		valid, err := gateway.VerifyCallback(ctx, callbackData)
		if err != nil || !valid {
			s.logger.WarnContext(ctx, "invalid callback signature", "payment_no", paymentNo)
			return fmt.Errorf("invalid signature")
		}
	}

	// 领域层处理状态变更
	if err := payment.Process(success, transactionID, thirdPartyNo); err != nil {
		s.logger.ErrorContext(ctx, "failed to process state transition", "error", err)
		return err
	}

	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return err
	}

	return nil
}

// RequestRefund 发起退款申请
func (s *PaymentApplicationService) RequestRefund(ctx context.Context, paymentID uint64, amount int64, reason string) (*domain.Refund, error) {
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
	// Currently paymentRepo.Save handles aggregates? Assuming yes for simplicity or strict DDD.

	s.logger.InfoContext(ctx, "refund processed", "refund_no", refund.RefundNo, "success", success)
	return refund, nil
}

// GetPaymentStatus 获取支付详情
func (s *PaymentApplicationService) GetPaymentStatus(ctx context.Context, id uint64) (*domain.Payment, error) {
	return s.paymentRepo.FindByID(ctx, id)
}

// ---------------- Helper Methods ----------------

func (s *PaymentApplicationService) checkRisk(ctx context.Context, orderID, userID uint64, amount int64, method string) error {
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

func (s *PaymentApplicationService) routeChannel(ctx context.Context, method string) (domain.GatewayType, *domain.ChannelConfig) {
	cfg, err := s.channelRepo.FindByCode(ctx, method)
	if err == nil && cfg != nil {
		return domain.GatewayType(cfg.Type), cfg
	}
	return s.resolveGatewayType(method), nil
}

func (s *PaymentApplicationService) resolveGatewayType(method string) domain.GatewayType {
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

func (s *PaymentApplicationService) buildGatewayRequest(p *domain.Payment, cfg *domain.ChannelConfig) *domain.PaymentGatewayRequest {
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
