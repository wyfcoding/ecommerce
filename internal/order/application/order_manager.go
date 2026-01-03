package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	advancedcouponv1 "github.com/wyfcoding/ecommerce/goapi/advancedcoupon/v1"
	inventoryv1 "github.com/wyfcoding/ecommerce/goapi/inventory/v1"
	paymentv1 "github.com/wyfcoding/ecommerce/goapi/payment/v1"
	warehousev1 "github.com/wyfcoding/ecommerce/goapi/warehouse/v1"
	"github.com/wyfcoding/ecommerce/internal/order/domain"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/wyfcoding/pkg/dtm"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/messagequeue/kafka"
	"github.com/wyfcoding/pkg/messagequeue/outbox"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/security/risk"
	"gorm.io/gorm"
)

// OrderManager 负责处理 Order 相关的写操作和业务逻辑。
type OrderManager struct {
	repo              domain.OrderRepository
	idGen             idgen.Generator
	producer          *kafka.Producer
	outboxMgr         *outbox.Manager
	logger            *slog.Logger
	dtmServer         string
	warehouseGrpcAddr string
	orderSvcURL       string // 本服务地址，供 DTM 回调
	riskEvaluator     risk.Evaluator
	inventoryCli      inventoryv1.InventoryServiceClient
	paymentCli        paymentv1.PaymentServiceClient

	// 指标统计
	orderCreatedCounter *prometheus.CounterVec
}

// NewOrderManager 负责处理 NewOrder 相关的写操作和业务逻辑。
func NewOrderManager(
	repo domain.OrderRepository,
	idGen idgen.Generator,
	producer *kafka.Producer,
	outboxMgr *outbox.Manager,
	logger *slog.Logger,
	dtmServer, warehouseGrpcAddr string,
	m *metrics.Metrics,
	riskEvaluator risk.Evaluator,
) *OrderManager {
	orderCreatedCounter := m.NewCounterVec(prometheus.CounterOpts{
		Name: "order_created_total",
		Help: "订单创建总数",
	}, []string{"status"})

	return &OrderManager{
		repo:                repo,
		idGen:               idGen,
		producer:            producer,
		outboxMgr:           outboxMgr,
		logger:              logger,
		dtmServer:           dtmServer,
		warehouseGrpcAddr:   warehouseGrpcAddr,
		riskEvaluator:       riskEvaluator,
		orderCreatedCounter: orderCreatedCounter,
	}
}

func (s *OrderManager) SetClients(invCli inventoryv1.InventoryServiceClient, payCli paymentv1.PaymentServiceClient) {
	s.inventoryCli = invCli
	s.paymentCli = payCli
}

func (s *OrderManager) SetSvcURL(url string) {
	s.orderSvcURL = url
}

// CreateOrder 创建订单。
func (s *OrderManager) CreateOrder(ctx context.Context, userID uint64, items []*domain.OrderItem, shippingAddr *domain.ShippingAddress, couponCode string) (*domain.Order, error) {
	// --- 架构增强：内联风控拦截 (Inline Risk Control) ---
	var totalAmount int64
	for _, it := range items {
		totalAmount += it.Price * int64(it.Quantity)
	}

	riskAssessment, err := s.riskEvaluator.Assess(ctx, "order.create", map[string]any{
		"user_id":      userID,
		"amount":       totalAmount,
		"item_count":   len(items),
		"client_ip":    ctx.Value("client_ip"),
		"device_id":    ctx.Value("device_id"),
		"is_real_name": true,
	})

	if err != nil {
		s.logger.ErrorContext(ctx, "risk assessment failed, fail-open applied", "error", err)
	} else {
		switch riskAssessment.Level {
		case risk.Reject:
			s.logger.WarnContext(ctx, "order rejected by risk control",
				"user_id", userID, "code", riskAssessment.Code, "reason", riskAssessment.Reason)
			return nil, fmt.Errorf("transaction security risk: %s", riskAssessment.Reason)
		case risk.Review:
			s.logger.InfoContext(ctx, "order needs risk review",
				"user_id", userID, "code", riskAssessment.Code)
		}
	}

	orderID := s.idGen.Generate()
	orderNo := fmt.Sprintf("%s%d", time.Now().Format("20060102"), orderID)

	order := domain.NewOrder(orderNo, userID, items, shippingAddr)
	order.Status = domain.Allocating // 切换到“分配中”状态，表示正在执行分布式事务

	// --- 架构增强：预同步锁定库存 (Internal Service Interaction) ---
	for _, item := range items {
		_, err := s.inventoryCli.LockStock(ctx, &inventoryv1.LockStockRequest{
			SkuId:    item.SkuID,
			Quantity: int32(item.Quantity),
			Reason:   "Order " + orderNo,
		})
		if err != nil {
			s.logger.ErrorContext(ctx, "synchronous stock locking failed", "sku_id", item.SkuID, "error", err)
			return nil, fmt.Errorf("insufficient stock for SKU %d", item.SkuID)
		}
	}

	// 1. 本地事务：保存订单并写入 Outbox
	err = s.repo.Transaction(ctx, userID, func(tx any) error {
		txRepo := s.repo.WithTx(tx)
		if err := txRepo.Save(ctx, order); err != nil {
			return err
		}

		event := map[string]any{
			"order_id": order.ID,
			"order_no": order.OrderNo,
			"user_id":  order.UserID,
			"amount":   order.TotalAmount,
			"status":   order.Status.String(),
		}

		gormTx := tx.(*gorm.DB)
		
		// 1.1 发布订单创建事件
		if err := s.outboxMgr.PublishInTx(ctx, gormTx, "order.created", orderNo, event); err != nil {
			return err
		}

		// 1.2 发布超时自动取消/释放库存任务 (延迟任务占位)
		// 在实际生产中，Outbox Processor 或 Kafka 延迟队列会处理此消息。
		timeoutEvent := map[string]any{
			"order_id":   order.ID,
			"order_no":   order.OrderNo,
			"user_id":    order.UserID,
			"items":      items, // 包含 SKU 和数量用于释放
			"expires_at": time.Now().Add(15 * time.Minute).Unix(),
		}
		return s.outboxMgr.PublishInTx(ctx, gormTx, "order.payment.timeout", orderNo, timeoutEvent)
	})
	if err != nil {
		return nil, err
	}

	s.orderCreatedCounter.WithLabelValues(order.Status.String()).Inc()

	// --- 2. 启动 DTM Saga 分布式事务 ---
	s.logger.InfoContext(ctx, "submitting saga transaction via pkg/dtm", "gid", orderNo)
	saga := dtm.NewSaga(ctx, s.dtmServer, orderNo)

	orderGrpcPrefix := s.orderSvcURL + "/api.order.v1.OrderService"
	warehouseGrpcPrefix := s.warehouseGrpcAddr + "/api.warehouse.v1.WarehouseService"

	// 2.1 注册订单确认作为第一步 (补偿为取消)
	saga.Add(
		orderGrpcPrefix+"/SagaConfirmOrder",
		orderGrpcPrefix+"/SagaCancelOrder",
		&advancedcouponv1.UseCouponRequest{ // 临时借用结构体，实际应为对应接口类型
			UserId:  userID,
			OrderId: uint64(order.ID),
		},
	)

	// 2.2 为每个商品添加库存扣减步骤
	for _, item := range items {
		saga.Add(
			warehouseGrpcPrefix+"/DeductStock",
			warehouseGrpcPrefix+"/RevertStock",
			&warehousev1.DeductStockRequest{
				OrderId:     uint64(order.ID),
				SkuId:       item.SkuID,
				Quantity:    item.Quantity,
				WarehouseId: 1, // 实际场景应由库存分配算法决定
			},
		)
	}

	// 2.3 如果使用了优惠券，添加核销步骤
	if couponCode != "" {
		couponSvcAddr := "advancedcoupon:50051"
		saga.Add(
			couponSvcAddr+"/api.advancedcoupon.v1.AdvancedCouponService/UseCoupon",
			"",
			&advancedcouponv1.UseCouponRequest{
				UserId:  userID,
				Code:    couponCode,
				OrderId: uint64(order.ID),
			},
		)
	}

	// 提交 Saga
	if err := saga.Submit(); err != nil {
		slog.ErrorContext(ctx, "failed to submit saga transaction", "order_no", orderNo, "error", err)
		return nil, fmt.Errorf("transaction submission failed: %w", err)
	}

	s.logger.InfoContext(ctx, "order created and saga submitted", "order_no", orderNo)

	// --- 架构增强：同步发起支付 (Internal Service Interaction) ---
	if s.paymentCli != nil {
		payResp, err := s.paymentCli.InitiatePayment(ctx, &paymentv1.InitiatePaymentRequest{
			OrderId:       uint64(order.ID),
			UserId:        userID,
			PaymentMethod: "WECHAT", // 默认微信，实际应由前端传参
			Amount:        order.TotalAmount,
			ClientIp:      "127.0.0.1",
		})
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to initiate payment", "order_no", orderNo, "error", err)
			// 注意：此时订单已创建且 Saga 已提交，支付失败不应导致订单回滚，
			// 用户可以在订单列表重新发起支付。
		} else {
			s.logger.InfoContext(ctx, "payment initiated successfully", "order_no", orderNo, "payment_url", payResp.PaymentUrl)
			// 在实际场景中，可能会将 payment_url 返回给前端
		}
	}

	return order, nil
}

// PayOrder 支付订单。
func (s *OrderManager) PayOrder(ctx context.Context, userID, id uint64, paymentMethod string) error {
	order, err := s.repo.FindByID(ctx, userID, uint(id))
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	if err := order.Pay(ctx, paymentMethod, "User"); err != nil {
		return err
	}

	return s.repo.Save(ctx, order)
}

// ShipOrder 发货订单。
func (s *OrderManager) ShipOrder(ctx context.Context, userID, id uint64, operator string) error {
	order, err := s.repo.FindByID(ctx, userID, uint(id))
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	if err := order.Ship(ctx, operator); err != nil {
		return err
	}

	return s.repo.Save(ctx, order)
}

// DeliverOrder 送达订单。
func (s *OrderManager) DeliverOrder(ctx context.Context, userID, id uint64, operator string) error {
	order, err := s.repo.FindByID(ctx, userID, uint(id))
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	if err := order.Deliver(ctx, operator); err != nil {
		return err
	}

	return s.repo.Save(ctx, order)
}

// CompleteOrder 完成订单。
func (s *OrderManager) CompleteOrder(ctx context.Context, userID, id uint64, operator string) error {
	order, err := s.repo.FindByID(ctx, userID, uint(id))
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	if err := order.Complete(ctx, operator); err != nil {
		return err
	}

	return s.repo.Save(ctx, order)
}

// SagaConfirmOrder Saga 正向: 确认订单 (Allocating -> PendingPayment)
func (s *OrderManager) SagaConfirmOrder(ctx context.Context, userID, orderID uint64) error {
	order, err := s.repo.FindByID(ctx, userID, uint(orderID))
	if err != nil || order == nil {
		return fmt.Errorf("order not found: %d", orderID)
	}
	if order.Status != domain.Allocating {
		return nil // 幂等
	}
	order.Status = domain.PendingPayment
	order.AddLog("System", "Saga Confirmed", domain.Allocating.String(), domain.PendingPayment.String(), "Inventory and logic verified")
	return s.repo.Save(ctx, order)
}

// SagaCancelOrder Saga 补偿: 取消订单 (Allocating -> Cancelled)
func (s *OrderManager) SagaCancelOrder(ctx context.Context, userID, orderID uint64, reason string) error {
	order, err := s.repo.FindByID(ctx, userID, uint(orderID))
	if err != nil || order == nil {
		return fmt.Errorf("order not found: %d", orderID)
	}
	if order.Status == domain.Cancelled {
		return nil // 幂等
	}
	return order.Cancel(ctx, "System", reason)
}

// CancelOrder 取消订单。
func (s *OrderManager) CancelOrder(ctx context.Context, userID, id uint64, operator, reason string) error {
	order, err := s.repo.FindByID(ctx, userID, uint(id))
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	if err := order.Cancel(ctx, operator, reason); err != nil {
		return err
	}

	return s.repo.Save(ctx, order)
}

// HandleInventoryReserved 处理库存已预留事件。
func (s *OrderManager) HandleInventoryReserved(ctx context.Context, userID, orderID uint64) error {
	order, err := s.repo.FindByID(ctx, userID, uint(orderID))
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	order.Status = domain.PendingPayment
	order.AddLog("System", "Inventory Reserved", domain.Allocating.String(), domain.PendingPayment.String(), "Inventory reserved successfully")

	return s.repo.Save(ctx, order)
}

// HandleInventoryReservationFailed 处理库存预留失败事件。
func (s *OrderManager) HandleInventoryReservationFailed(ctx context.Context, userID, orderID uint64, reason string) error {
	order, err := s.repo.FindByID(ctx, userID, uint(orderID))
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	return order.Cancel(ctx, "System", fmt.Sprintf("Inventory reservation failed: %s", reason))
}

// HandlePaymentProcessed 处理支付已完成事件。
func (s *OrderManager) HandlePaymentProcessed(ctx context.Context, userID, orderID uint64) error {
	order, err := s.repo.FindByID(ctx, userID, uint(orderID))
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	return order.Pay(ctx, "Online", "System")
}

// HandlePaymentFailed 处理支付失败事件。
func (s *OrderManager) HandlePaymentFailed(ctx context.Context, userID, orderID uint64, reason string) error {
	order, err := s.repo.FindByID(ctx, userID, uint(orderID))
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	if err := order.Cancel(ctx, "System", fmt.Sprintf("Payment failed: %s", reason)); err != nil {
		return err
	}

	s.logger.InfoContext(ctx, "should publish ReleaseInventory event", "order_id", orderID, "reason", reason)
	return nil
}

// HandleFlashsaleOrder 处理秒杀订单落库。
func (s *OrderManager) HandleFlashsaleOrder(ctx context.Context, orderID, userID, productID, skuID uint64, quantity int32, price int64) error {
	s.logger.InfoContext(ctx, "handling flashsale order persistence", "order_id", orderID, "user_id", userID)

	orderNo := fmt.Sprintf("FS%d", orderID)
	items := []*domain.OrderItem{
		{
			SkuID:       skuID,
			ProductID:   productID,
			Quantity:    quantity,
			Price:       price,
			ProductName: "Flashsale Product", // 占位符，由后续详情补全或从缓存获取
			SkuName:     "Flashsale SKU",
		},
	}

	// 秒杀订单无需再校验库存或风控（Flashsale 服务已完成）
	// 直接创建并设置为待支付状态
	order := domain.NewOrder(orderNo, userID, items, nil)
	order.ID = uint(orderID)
	order.Status = domain.PendingPayment
	order.AddLog("System", "Flashsale Order Created", "", domain.PendingPayment.String(), "Asynchronous creation from flashsale event")

	if err := s.repo.Save(ctx, order); err != nil {
		s.logger.ErrorContext(ctx, "failed to save flashsale order", "order_id", orderID, "error", err)
		return err
	}

	s.orderCreatedCounter.WithLabelValues(order.Status.String()).Inc()
	return nil
}
