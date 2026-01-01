package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	advancedcouponv1 "github.com/wyfcoding/ecommerce/goapi/advancedcoupon/v1"
	warehousev1 "github.com/wyfcoding/ecommerce/goapi/warehouse/v1"
	"github.com/wyfcoding/ecommerce/internal/order/domain"

	"github.com/dtm-labs/client/dtmgrpc"
	"github.com/prometheus/client_golang/prometheus"
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
	riskEvaluator     risk.Evaluator

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
		return s.outboxMgr.PublishInTx(gormTx, "order.created", orderNo, event)
	})
	if err != nil {
		return nil, err
	}

	s.orderCreatedCounter.WithLabelValues(order.Status.String()).Inc()

	// --- 2. 启动 DTM Saga 分布式事务 ---
	s.logger.InfoContext(ctx, "submitting saga transaction to DTM", "gid", orderNo)
	saga := dtmgrpc.NewSagaGrpc(s.dtmServer, orderNo)

	// 2.1 为每个商品添加库存扣减步骤
	for _, item := range items {
		saga.Add(
			s.warehouseGrpcAddr+"/api.warehouse.v1.WarehouseService/DeductStock",
			s.warehouseGrpcAddr+"/api.warehouse.v1.WarehouseService/RevertStock",
			&warehousev1.DeductStockRequest{
				OrderId:     uint64(order.ID),
				SkuId:       item.SkuID,
				Quantity:    item.Quantity,
				WarehouseId: 1, // 实际场景应由库存分配算法决定
			},
		)
	}

	// 2.2 如果使用了优惠券，添加核销步骤 (假设 AdvancedCoupon 服务的 gRPC 地址)
	if couponCode != "" {
		// 注意：此处需要 AdvancedCoupon 的 gRPC 地址，这里假设一个或通过配置注入
		couponSvcAddr := "advancedcoupon:50051" // 示例
		saga.Add(
			couponSvcAddr+"/api.advancedcoupon.v1.AdvancedCouponService/UseCoupon",
			"", // 假设 UseCoupon 是幂等的或不需要显式补偿（或者由内部逻辑处理）
			&advancedcouponv1.UseCouponRequest{
				UserId:  userID,
				Code:    couponCode,
				OrderId: uint64(order.ID),
			},
		)
	}

	// 提交 Saga
	if err := saga.Submit(); err != nil {
		s.logger.ErrorContext(ctx, "failed to submit saga to DTM", "gid", orderNo, "error", err)
		// 容错处理：本地标记为取消
		order.Status = domain.Cancelled
		_ = s.repo.Save(ctx, order)
		return nil, fmt.Errorf("distributed transaction failed: %w", err)
	}

	s.logger.InfoContext(ctx, "order created and saga submitted", "order_no", orderNo)
	return order, nil
}

// PayOrder 支付订单。
func (s *OrderManager) PayOrder(ctx context.Context, id uint64, paymentMethod string) error {
	order, err := s.repo.FindByID(ctx, uint(id))
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	if err := order.Pay(paymentMethod, "User"); err != nil {
		return err
	}

	return s.repo.Save(ctx, order)
}

// ShipOrder 发货订单。
func (s *OrderManager) ShipOrder(ctx context.Context, id uint64, operator string) error {
	order, err := s.repo.FindByID(ctx, uint(id))
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	if err := order.Ship(operator); err != nil {
		return err
	}

	return s.repo.Save(ctx, order)
}

// DeliverOrder 送达订单。
func (s *OrderManager) DeliverOrder(ctx context.Context, id uint64, operator string) error {
	order, err := s.repo.FindByID(ctx, uint(id))
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	if err := order.Deliver(operator); err != nil {
		return err
	}

	return s.repo.Save(ctx, order)
}

// CompleteOrder 完成订单。
func (s *OrderManager) CompleteOrder(ctx context.Context, id uint64, operator string) error {
	order, err := s.repo.FindByID(ctx, uint(id))
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	if err := order.Complete(operator); err != nil {
		return err
	}

	return s.repo.Save(ctx, order)
}

// CancelOrder 取消订单。
func (s *OrderManager) CancelOrder(ctx context.Context, id uint64, operator, reason string) error {
	order, err := s.repo.FindByID(ctx, uint(id))
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	if err := order.Cancel(operator, reason); err != nil {
		return err
	}

	return s.repo.Save(ctx, order)
}

// HandleInventoryReserved 处理库存已预留事件。
func (s *OrderManager) HandleInventoryReserved(ctx context.Context, orderID uint64) error {
	order, err := s.repo.FindByID(ctx, uint(orderID))
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
func (s *OrderManager) HandleInventoryReservationFailed(ctx context.Context, orderID uint64, reason string) error {
	order, err := s.repo.FindByID(ctx, uint(orderID))
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	return order.Cancel("System", fmt.Sprintf("Inventory reservation failed: %s", reason))
}

// HandlePaymentProcessed 处理支付已完成事件。
func (s *OrderManager) HandlePaymentProcessed(ctx context.Context, orderID uint64) error {
	order, err := s.repo.FindByID(ctx, uint(orderID))
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	return order.Pay("Online", "System")
}

// HandlePaymentFailed 处理支付失败事件。
func (s *OrderManager) HandlePaymentFailed(ctx context.Context, orderID uint64, reason string) error {
	order, err := s.repo.FindByID(ctx, uint(orderID))
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	if err := order.Cancel("System", fmt.Sprintf("Payment failed: %s", reason)); err != nil {
		return err
	}

	s.logger.InfoContext(ctx, "should publish ReleaseInventory event", "order_id", orderID, "reason", reason)
	return nil
}
