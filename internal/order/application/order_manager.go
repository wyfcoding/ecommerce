package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"log/slog"

	warehousev1 "github.com/wyfcoding/ecommerce/go-api/warehouse/v1"
	"github.com/wyfcoding/ecommerce/internal/order/domain"

	"github.com/dtm-labs/client/dtmgrpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/messagequeue/kafka"
	"github.com/wyfcoding/pkg/metrics"
)

// OrderManager 负责处理 Order 相关的写操作和业务逻辑。
type OrderManager struct {
	repo              domain.OrderRepository
	idGen             idgen.Generator
	producer          *kafka.Producer
	logger            *slog.Logger
	dtmServer         string
	warehouseGrpcAddr string

	// Metrics
	orderCreatedCounter *prometheus.CounterVec
}

// NewOrderManager 负责处理 NewOrder 相关的写操作和业务逻辑。
func NewOrderManager(
	repo domain.OrderRepository,
	idGen idgen.Generator,
	producer *kafka.Producer,
	logger *slog.Logger,
	dtmServer, warehouseGrpcAddr string,
	m *metrics.Metrics,
) *OrderManager {
	orderCreatedCounter := m.NewCounterVec(prometheus.CounterOpts{
		Name: "order_created_total",
		Help: "Total number of orders created",
	}, []string{"status"})

	return &OrderManager{
		repo:                repo,
		idGen:               idGen,
		producer:            producer,
		logger:              logger,
		dtmServer:           dtmServer,
		warehouseGrpcAddr:   warehouseGrpcAddr,
		orderCreatedCounter: orderCreatedCounter,
	}
}

// CreateOrder 创建订单。
func (s *OrderManager) CreateOrder(ctx context.Context, userID uint64, items []*domain.OrderItem, shippingAddr *domain.ShippingAddress) (*domain.Order, error) {
	orderID := s.idGen.Generate()
	orderNo := fmt.Sprintf("%s%d", time.Now().Format("20060102"), orderID)

	order := domain.NewOrder(orderNo, userID, items, shippingAddr)
	order.Status = domain.PendingPayment

	if err := s.repo.Save(ctx, order); err != nil {
		s.logger.ErrorContext(ctx, "failed to create order", "error", err)
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
		_ = order.Cancel("System", "Saga Submit Failed")
		_ = s.repo.Save(ctx, order)
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
