package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

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
func (s *OrderManager) CreateOrder(ctx context.Context, userID uint64, items []*domain.OrderItem, shippingAddr *domain.ShippingAddress) (*domain.Order, error) {
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
		"is_real_name": true, // 假设从 Context 或 Auth 获取
	})

	if err != nil {
		// Fail-Open 策略：风控引擎故障时默认放行，保证业务连续性
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
			// 此处可以抛出特定错误让前端触发 MFA (多因素认证)
		}
	}
	// --- 内联风控结束 ---

	orderID := s.idGen.Generate()
	orderNo := fmt.Sprintf("%s%d", time.Now().Format("20060102"), orderID)

	order := domain.NewOrder(orderNo, userID, items, shippingAddr)
	order.Status = domain.PendingPayment

	// 顶级架构实践：利用本地事务确保业务数据与 Outbox 消息的强一致性
	err = s.repo.Transaction(ctx, userID, func(tx any) error {
		// 1. 使用事务中的仓储
		txRepo := s.repo.WithTx(tx)
		if err := txRepo.Save(ctx, order); err != nil {
			return err
		}

		// 2. 在同一事务中写入 Outbox 消息
		event := map[string]any{
			"order_id": order.ID,
			"order_no": order.OrderNo,
			"user_id":  order.UserID,
			"amount":   order.TotalAmount,
			"status":   order.Status.String(),
		}

		gormTx := tx.(*gorm.DB)
		if err := s.outboxMgr.PublishInTx(gormTx, "order.created", orderNo, event); err != nil {
			s.logger.ErrorContext(ctx, "failed to publish order created event to outbox", "error", err)
			return err
		}

		return nil
	})

	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create order in transaction", "error", err)
		return nil, err
	}

	s.orderCreatedCounter.WithLabelValues(order.Status.String()).Inc()

	// --- 启动 DTM Saga 分布式事务 ---
	// 在电商场景中，创建订单和扣减库存通常不在同一个数据库或服务中。
	// 为了保证最终一致性，我们使用 Saga 模式：
	// 1. 正向操作 (Action)：DeductStock (扣减库存)
	// 2. 逆向补偿 (Compensate)：RevertStock (回滚库存，用于事务失败时的补偿)
	gid := orderNo
	saga := dtmgrpc.NewSagaGrpc(s.dtmServer, gid).
		Add(
			s.warehouseGrpcAddr+"/api.warehouse.v1.WarehouseService/DeductStock", // 调用库存服务的扣减接口
			s.warehouseGrpcAddr+"/api.warehouse.v1.WarehouseService/RevertStock", // 注册补偿接口
			&warehousev1.DeductStockRequest{
				OrderId:     uint64(order.ID),
				SkuId:       items[0].SkuID,
				Quantity:    items[0].Quantity,
				WarehouseId: 1,
			},
		)

	// 提交 Saga 事务给 DTM 服务器。
	// DTM 会负责根据定义的步骤进行编排，并在发生故障时执行补偿逻辑。
	if err := saga.Submit(); err != nil {
		s.logger.ErrorContext(ctx, "failed to submit saga", "error", err)
		// 如果 Saga 提交失败，本地订单需要标记为取消。
		if cancelErr := order.Cancel("System", "Saga Submit Failed"); cancelErr != nil {
			s.logger.ErrorContext(ctx, "failed to cancel order after saga submit failure", "error", cancelErr)
		}
		if saveErr := s.repo.Save(ctx, order); saveErr != nil {
			s.logger.ErrorContext(ctx, "failed to save cancelled order", "error", saveErr)
		}
		return nil, fmt.Errorf("failed to submit distributed transaction: %w", err)
	}

	s.logger.InfoContext(ctx, "order created successfully", "order_id", order.ID, "order_no", order.OrderNo)
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
