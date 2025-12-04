package application

import (
	"context"
	"errors" // 导入标准错误处理库。
	"fmt"    // 导入格式化库。
	"time"   // 导入时间库。

	warehousev1 "github.com/wyfcoding/ecommerce/api/warehouse/v1"     // 导入仓库服务的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/order/domain/entity"     // 导入订单领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/order/domain/repository" // 导入订单领域的仓储接口。

	"github.com/wyfcoding/ecommerce/pkg/idgen"              // 导入ID生成器。
	"github.com/wyfcoding/ecommerce/pkg/messagequeue/kafka" // 导入Kafka消息生产者。

	"github.com/dtm-labs/client/dtmgrpc"             // 导入DTM分布式事务库（gRPC模式）。
	"github.com/prometheus/client_golang/prometheus" // 导入Prometheus客户端库。
	"github.com/wyfcoding/ecommerce/pkg/metrics"     // 导入自定义指标库。

	"log/slog" // 导入结构化日志库。
)

// OrderService 结构体定义了订单管理相关的应用服务。
// 它协调领域层和基础设施层，处理订单的创建、状态变更、查询等业务逻辑，并支持分布式事务。
type OrderService struct {
	repo              repository.OrderRepository // 依赖OrderRepository接口，用于数据持久化操作。
	idGen             idgen.Generator            // ID生成器，用于生成订单ID。
	producer          *kafka.Producer            // Kafka消息生产者，用于发布领域事件。
	logger            *slog.Logger               // 日志记录器，用于记录服务运行时的信息和错误。
	dtmServer         string                     // DTM服务器地址，用于分布式事务协调。
	warehouseGrpcAddr string                     // 仓库服务的gRPC地址，用于调用库存扣减接口。

	// Metrics
	orderCreatedCounter *prometheus.CounterVec // Prometheus计数器，用于统计订单创建数量。
}

// NewOrderService 创建并返回一个新的 OrderService 实例。
func NewOrderService(repo repository.OrderRepository, idGen idgen.Generator, producer *kafka.Producer, logger *slog.Logger, dtmServer, warehouseGrpcAddr string, m *metrics.Metrics) *OrderService {
	// 初始化订单创建计数器。
	orderCreatedCounter := m.NewCounterVec(prometheus.CounterOpts{
		Name: "order_created_total",
		Help: "Total number of orders created",
	}, []string{"status"}) // 按照订单状态进行维度划分。

	return &OrderService{
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
// 此方法采用Saga分布式事务模式，确保订单创建和库存扣减的原子性。
// ctx: 上下文。
// userID: 购买用户ID。
// items: 订单商品列表。
// shippingAddr: 收货地址。
// 返回创建成功的Order实体和可能发生的错误。
func (s *OrderService) CreateOrder(ctx context.Context, userID uint64, items []*entity.OrderItem, shippingAddr *entity.ShippingAddress) (*entity.Order, error) {
	// 1. 生成订单ID和订单号。
	orderID := s.idGen.Generate()
	orderNo := fmt.Sprintf("%s%d", time.Now().Format("20060102"), orderID)

	order := entity.NewOrder(orderNo, userID, items, shippingAddr)
	// 订单初始状态：待支付。
	// 在实际业务中，订单可能先进入“待分配库存”或“待确认”状态，通过Saga模式来确保库存预留。
	order.Status = entity.PendingPayment

	// 2. 本地保存订单（作为Saga的第一个分支的局部事务）。
	if err := s.repo.Save(ctx, order); err != nil {
		s.logger.ErrorContext(ctx, "failed to create order", "error", err)
		return nil, err
	}

	// 记录订单创建指标。
	s.orderCreatedCounter.WithLabelValues(order.Status.String()).Inc()

	// 3. 构建Saga分布式事务。
	// 使用订单号作为全局事务ID。
	gid := orderNo
	saga := dtmgrpc.NewSagaGrpc(s.dtmServer, gid).
		// 添加库存扣减分支。
		// 第一个参数是正向操作的gRPC地址及方法，第二个参数是补偿操作的gRPC地址及方法。
		// 第三个参数是正向操作的请求体。
		Add(
			s.warehouseGrpcAddr+"/api.warehouse.v1.WarehouseService/DeductStock", // 库存扣减服务接口。
			s.warehouseGrpcAddr+"/api.warehouse.v1.WarehouseService/RevertStock", // 库存回滚服务接口。
			&warehousev1.DeductStockRequest{ // 扣减库存请求体。
				OrderId:     uint64(order.ID),
				SkuId:       items[0].SkuID,    // 简化处理，假设只有一个商品。
				Quantity:    items[0].Quantity, // 简化处理，假设只有一个商品。
				WarehouseId: 1,                 // 简化处理，假设只有一个仓库。
			},
		)

	// 注意：对于多商品订单，应该为每个商品添加一个Saga分支。
	// dtmgrpc.Add 方法接受一个payload作为参数。
	// 为简化示例，这里假设订单只有一个商品。
	// 如果有多个商品，则需要循环添加多个分支：
	/*
		for _, item := range items {
			saga.Add(..., &warehousev1.DeductStockRequest{...})
		}
	*/

	// 4. 提交Saga事务。
	// Submit通常会立即返回，如果DTM服务器成功接收到事务。
	// 但如果提交本身失败（例如网络问题），我们需要处理。
	if err := saga.Submit(); err != nil {
		s.logger.ErrorContext(ctx, "failed to submit saga", "error", err)
		// 如果提交Saga失败，我们应该将本地已保存的订单标记为取消或失败。
		// 这里将订单状态更新为已取消。
		_ = order.Cancel("System", "Saga Submit Failed")
		_ = s.repo.Save(ctx, order) // 保存取消后的订单状态。
		return nil, fmt.Errorf("failed to submit saga: %w", err)
	}

	s.logger.InfoContext(ctx, "order created successfully", "order_id", order.ID, "order_no", order.OrderNo)
	return order, nil
}

// GetOrder 获取指定ID的订单详情。
// ctx: 上下文。
// id: 订单ID。
// 返回Order实体和可能发生的错误。
func (s *OrderService) GetOrder(ctx context.Context, id uint64) (*entity.Order, error) {
	return s.repo.GetByID(ctx, id)
}

// PayOrder 支付订单。
// ctx: 上下文。
// id: 订单ID。
// paymentMethod: 支付方式。
// 返回可能发生的错误。
func (s *OrderService) PayOrder(ctx context.Context, id uint64, paymentMethod string) error {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	// 调用订单实体方法进行支付操作。
	if err := order.Pay(paymentMethod, "User"); err != nil {
		return err
	}

	return s.repo.Save(ctx, order) // 保存更新后的订单状态。
}

// ShipOrder 发货订单。
// ctx: 上下文。
// id: 订单ID。
// operator: 操作人员。
// 返回可能发生的错误。
func (s *OrderService) ShipOrder(ctx context.Context, id uint64, operator string) error {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	// 调用订单实体方法进行发货操作。
	if err := order.Ship(operator); err != nil {
		return err
	}

	return s.repo.Save(ctx, order) // 保存更新后的订单状态。
}

// DeliverOrder 送达订单。
// ctx: 上下文。
// id: 订单ID。
// operator: 操作人员。
// 返回可能发生的错误。
func (s *OrderService) DeliverOrder(ctx context.Context, id uint64, operator string) error {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	// 调用订单实体方法进行送达操作。
	if err := order.Deliver(operator); err != nil {
		return err
	}

	return s.repo.Save(ctx, order) // 保存更新后的订单状态。
}

// CompleteOrder 完成订单。
// ctx: 上下文。
// id: 订单ID。
// operator: 操作人员。
// 返回可能发生的错误。
func (s *OrderService) CompleteOrder(ctx context.Context, id uint64, operator string) error {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	// 调用订单实体方法进行完成操作。
	if err := order.Complete(operator); err != nil {
		return err
	}

	return s.repo.Save(ctx, order) // 保存更新后的订单状态。
}

// CancelOrder 取消订单。
// ctx: 上下文。
// id: 订单ID。
// operator: 操作人员。
// reason: 取消原因。
// 返回可能发生的错误。
func (s *OrderService) CancelOrder(ctx context.Context, id uint64, operator, reason string) error {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	// 调用订单实体方法进行取消操作。
	if err := order.Cancel(operator, reason); err != nil {
		return err
	}

	return s.repo.Save(ctx, order) // 保存更新后的订单状态。
}

// ListOrders 获取订单列表。
// ctx: 上下文。
// userID: 用户ID，用于过滤订单。
// status: 订单状态，用于过滤订单。
// page, pageSize: 分页参数。
// 返回订单列表、总数和可能发生的错误。
func (s *OrderService) ListOrders(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*entity.Order, int64, error) {
	offset := (page - 1) * pageSize
	var orderStatus *entity.OrderStatus
	if status != nil {
		s := entity.OrderStatus(*status)
		orderStatus = &s
	}
	return s.repo.List(ctx, userID, orderStatus, offset, pageSize)
}

// --- Saga 事务分支处理器 ---

// HandleInventoryReserved 处理库存已预留事件（Saga补偿事务的正向操作）。
// 当库存服务成功预留库存后，DTM会调用此方法。
func (s *OrderService) HandleInventoryReserved(ctx context.Context, orderID uint64) error {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	// 更新订单状态为 PendingPayment（待支付），表示订单已准备好支付。
	order.Status = entity.PendingPayment
	order.AddLog("System", "Inventory Reserved", entity.Allocating.String(), entity.PendingPayment.String(), "Inventory reserved successfully")

	return s.repo.Save(ctx, order)
}

// HandleInventoryReservationFailed 处理库存预留失败事件（Saga补偿事务的逆向操作）。
// 当库存服务预留库存失败时，DTM会调用此方法。
func (s *OrderService) HandleInventoryReservationFailed(ctx context.Context, orderID uint64, reason string) error {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	// 取消订单。
	return order.Cancel("System", fmt.Sprintf("Inventory reservation failed: %s", reason))
}

// HandlePaymentProcessed 处理支付已完成事件。
// 当支付服务成功处理支付后，DTM或其他事件机制会通知此方法。
func (s *OrderService) HandlePaymentProcessed(ctx context.Context, orderID uint64) error {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	// 调用订单实体方法进行支付操作，并标记为“Online”支付方式和“System”操作者。
	return order.Pay("Online", "System")
}

// HandlePaymentFailed 处理支付失败事件。
// 当支付服务处理支付失败后，DTM或其他事件机制会通知此方法。
func (s *OrderService) HandlePaymentFailed(ctx context.Context, orderID uint64, reason string) error {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	// 取消订单。
	if err := order.Cancel("System", fmt.Sprintf("Payment failed: %s", reason)); err != nil {
		return err
	}

	// TODO: 发布 ReleaseInventory 事件以释放之前预留的库存。
	// 当前的pkg/messagequeue/kafka/kafka.go中Producer结构体的设计是与单个Topic绑定的。
	// 要发布到不同的Topic（例如“inventory-release”），需要修改Producer实现以支持动态Topic，
	// 或者创建多个Producer实例。
	// 这里仅记录日志以模拟应发布事件。
	s.logger.InfoContext(ctx, "should publish ReleaseInventory event", "order_id", orderID, "reason", reason)
	return nil
}
