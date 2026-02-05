// 变更说明：引入事件溯源写入与版本控制，确保订单写模型与事件流一致。
// 假设：事件流以订单ID作为聚合根ID，读模型由事件消费异步更新，存量订单不回补历史事件。
package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	advancedcouponv1 "github.com/wyfcoding/ecommerce/goapi/advancedcoupon/v1"
	inventoryv1 "github.com/wyfcoding/ecommerce/goapi/inventory/v1"
	orderv1 "github.com/wyfcoding/ecommerce/goapi/order/v1"
	paymentv1 "github.com/wyfcoding/ecommerce/goapi/payment/v1"
	productv1 "github.com/wyfcoding/ecommerce/goapi/product/v1"
	warehousev1 "github.com/wyfcoding/ecommerce/goapi/warehouse/v1"
	"github.com/wyfcoding/ecommerce/internal/order/domain"
	positionv1 "github.com/wyfcoding/financialtrading/go-api/position/v1"
	"github.com/wyfcoding/pkg/dtm"
	"github.com/wyfcoding/pkg/eventsourcing"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/security/risk"
	"github.com/wyfcoding/pkg/tracing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shopspring/decimal"
)

// OrderCommandService 处理所有订单相关的写入操作。
type OrderCommandService struct {
	repo              domain.OrderRepository
	eventStore        domain.OrderEventStore
	idGen             idgen.Generator
	publisher         domain.EventPublisher
	logger            *slog.Logger
	dtmServer         string
	warehouseGrpcAddr string
	orderSvcURL       string // 本服务地址，供 DTM 回调
	riskEvaluator     risk.Evaluator
	inventoryCli      inventoryv1.InventoryServiceClient
	paymentCli        paymentv1.PaymentServiceClient
	positionCli       positionv1.PositionServiceClient
	productCli        productv1.ProductServiceClient

	// 指标统计
	orderCreatedCounter *prometheus.CounterVec
}

// NewOrderCommandService 构造函数。
func NewOrderCommandService(
	repo domain.OrderRepository,
	eventStore domain.OrderEventStore,
	idGen idgen.Generator,
	publisher domain.EventPublisher,
	logger *slog.Logger,
	dtmServer, warehouseGrpcAddr string,
	m *metrics.Metrics,
	riskEvaluator risk.Evaluator,
) *OrderCommandService {
	orderCreatedCounter := m.NewCounterVec(&prometheus.CounterOpts{
		Name: "order_created_total",
		Help: "Total number of orders created",
	}, []string{"status"})

	return &OrderCommandService{
		repo:                repo,
		eventStore:          eventStore,
		idGen:               idGen,
		publisher:           publisher,
		logger:              logger,
		dtmServer:           dtmServer,
		warehouseGrpcAddr:   warehouseGrpcAddr,
		riskEvaluator:       riskEvaluator,
		orderCreatedCounter: orderCreatedCounter,
	}
}

// SetClients 注入下游依赖。
func (s *OrderCommandService) SetClients(
	invCli inventoryv1.InventoryServiceClient,
	payCli paymentv1.PaymentServiceClient,
	posCli positionv1.PositionServiceClient,
	prodCli productv1.ProductServiceClient,
) {
	s.inventoryCli = invCli
	s.paymentCli = payCli
	s.positionCli = posCli
	s.productCli = prodCli
}

// SetSvcURL 设置回调地址。
func (s *OrderCommandService) SetSvcURL(url string) {
	s.orderSvcURL = url
}

// CreateOrder 创建订单逻辑。
func (s *OrderCommandService) CreateOrder(ctx context.Context, cmd *CreateOrderCommand) (*domain.Order, error) {
	// 1. 获取最新价格 (Price Safety)
	var items []*domain.OrderItem
	var totalAmount int64
	for _, it := range cmd.Items {
		sku, err := s.productCli.GetSKUByID(ctx, &productv1.GetSKUByIDRequest{Id: it.SkuID})
		if err != nil {
			return nil, fmt.Errorf("invalid SKU %d: %w", it.SkuID, err)
		}
		item := &domain.OrderItem{
			SkuID:       it.SkuID,
			ProductID:   it.ProductID,
			Quantity:    it.Quantity,
			Price:       sku.Price,
			ProductName: sku.Name,
			SkuName:     sku.Name,
			TotalPrice:  sku.Price * int64(it.Quantity),
		}
		items = append(items, item)
		totalAmount += item.TotalPrice
	}

	// 2. 风控评价 (使用 DTO 转换以避免在业务代码中直接操作 map[string]any)
	riskAssessment, err := s.assessRisk(ctx, cmd.UserID, totalAmount, cmd.ClientIP, cmd.DeviceID)
	if err != nil {
		s.logger.ErrorContext(ctx, "risk assessment error", "error", err)
		return nil, err
	}
	if riskAssessment.Level == risk.Reject {
		return nil, fmt.Errorf("risk rejected: %s", riskAssessment.Reason)
	}

	orderID := s.idGen.Generate()
	orderNo := fmt.Sprintf("%s%d", time.Now().Format("20060102"), orderID)
	order := domain.NewOrder(orderNo, cmd.UserID, items, cmd.ShippingAddress)
	order.Status = orderv1.OrderStatus_ALLOCATING
	if cmd.Remark != "" {
		order.Remark = cmd.Remark
	}
	if cmd.PaymentMethod != "" {
		order.PaymentMethod = cmd.PaymentMethod
	}

	// 3. 预锁库存 (乐观)
	for _, item := range items {
		_, err := s.inventoryCli.LockStock(ctx, &inventoryv1.LockStockRequest{
			SkuId:    item.SkuID,
			Quantity: int32(item.Quantity),
			Reason:   "Order " + orderNo,
		})
		if err != nil {
			return nil, fmt.Errorf("insufficient stock for SKU %d", item.SkuID)
		}
	}

	// 4. 事务：保存订单并发布事件 (EDA)
	if err := s.repo.WithTx(ctx, cmd.UserID, func(tx any) error {
		// 初始化版本号 (事件溯源起点)
		order.Version = 0
		if err := s.repo.SaveInTx(ctx, tx, order); err != nil {
			return err
		}

		// 事件溯源：记录订单创建事件
		createdPayload := &domain.OrderCreatedPayload{
			OrderID:              uint64(order.ID),
			OrderNo:              order.OrderNo,
			UserID:               order.UserID,
			Status:               order.Status,
			TotalAmount:          order.TotalAmount,
			ActualAmount:         order.ActualAmount,
			ShippingFee:          order.ShippingFee,
			DiscountAmount:       order.DiscountAmount,
			PaymentMethod:        order.PaymentMethod,
			PaymentTransactionID: order.PaymentTransactionID,
			Remark:               order.Remark,
			TrackingNumber:       order.TrackingNumber,
			LogisticsCompany:     order.LogisticsCompany,
			RefundAmount:         order.RefundAmount,
			RefundReason:         order.RefundReason,
			ShippingAddress:      order.ShippingAddress,
			Items:                order.Items,
			CreatedAt:            order.CreatedAt,
			InitLog:              buildEventLogFromOrder(order),
		}
		if err := s.appendEventInTx(ctx, tx, order, domain.OrderEventTypeCreated, createdPayload); err != nil {
			return err
		}

		// 回写版本号到 MySQL（保持事件流与写模型一致）
		if err := s.repo.UpdateInTx(ctx, tx, order); err != nil {
			return err
		}

		// 发布领域事件
		event := &domain.OrderCreatedEvent{
			OrderID:     uint64(order.ID),
			OrderNo:     order.OrderNo,
			UserID:      order.UserID,
			TotalAmount: order.TotalAmount,
			Status:      order.Status,
			Timestamp:   time.Now(),
		}
		if err := s.publisher.PublishInTx(ctx, tx, "order.created", order.OrderNo, event); err != nil {
			return err
		}

		// 发布超时预警 (延迟任务的基础)
		timeoutEvent := &domain.OrderPaymentTimeoutEvent{
			OrderID:   uint64(order.ID),
			OrderNo:   order.OrderNo,
			UserID:    order.UserID,
			ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
			Timestamp: time.Now(),
		}
		return s.publisher.PublishInTx(ctx, tx, "order.payment.timeout", order.OrderNo, timeoutEvent)
	}); err != nil {
		return nil, err
	}

	s.orderCreatedCounter.WithLabelValues(order.Status.String()).Inc()

	// 5. 分布式事务 Saga 编排
	saga := dtm.NewSaga(ctx, s.dtmServer, orderNo)
	orderGrpcPrefix := s.orderSvcURL + "/api.order.v1.OrderService"
	warehouseGrpcPrefix := s.warehouseGrpcAddr + "/api.warehouse.v1.WarehouseService"

	saga.Add(orderGrpcPrefix+"/SagaConfirmOrder", orderGrpcPrefix+"/SagaCancelOrder", &orderv1.SagaOrderRequest{
		UserId:  cmd.UserID,
		OrderId: uint64(order.ID),
	})
	for _, it := range items {
		saga.Add(warehouseGrpcPrefix+"/DeductStock", warehouseGrpcPrefix+"/RevertStock", &warehousev1.DeductStockRequest{
			OrderId:     uint64(order.ID),
			SkuId:       it.SkuID,
			Quantity:    it.Quantity,
			WarehouseId: 1,
		})
	}
	if cmd.CouponCode != "" {
		saga.Add("advancedcoupon:50051/api.advancedcoupon.v1.AdvancedCouponService/UseCoupon", "", &advancedcouponv1.UseCouponRequest{
			UserId:  cmd.UserID,
			Code:    cmd.CouponCode,
			OrderId: uint64(order.ID),
		})
	}

	if err := saga.Submit(ctx); err != nil {
		s.logger.ErrorContext(ctx, "saga submission failed", "order_no", orderNo, "error", err)
		return nil, fmt.Errorf("distributed transaction submission failed: %w", err)
	}

	// 6. 异步发起支付请求 (记录完整异常日志)
	if s.paymentCli != nil {
		go func() {
			reqCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			resp, err := s.paymentCli.InitiatePayment(reqCtx, &paymentv1.InitiatePaymentRequest{
				OrderId:       uint64(order.ID),
				UserId:        cmd.UserID,
				PaymentMethod: "WECHAT",
				Amount:        order.TotalAmount,
				ClientIp:      cmd.ClientIP,
			})
			if err != nil {
				s.logger.Error("failed to initiate payment in background", "order_id", order.ID, "error", err)
				return
			}
			s.logger.Info("background payment initiation success", "order_id", order.ID, "transaction_no", resp.TransactionNo)
		}()
	}

	return order, nil
}

// PayOrder 支付。
func (s *OrderCommandService) PayOrder(ctx context.Context, cmd *PayOrderCommand) error {
	return s.repo.WithTx(ctx, cmd.UserID, func(tx any) error {
		order, err := s.repo.FindByID(ctx, cmd.UserID, uint64(cmd.OrderID))
		if err != nil || order == nil {
			return errors.New("order not found")
		}
		if cmd.Amount > 0 && cmd.Amount != order.ActualAmount {
			return fmt.Errorf("payment amount mismatch: expect %d, got %d", order.ActualAmount, cmd.Amount)
		}
		oldStatus := order.Status
		if cmd.TransactionID != "" {
			order.PaymentTransactionID = cmd.TransactionID
		}
		if err := order.Pay(ctx, cmd.PaymentMethod, "User"); err != nil {
			return err
		}
		paidAt := time.Now()
		if order.PaidAt != nil {
			paidAt = *order.PaidAt
		}
		paidPayload := &domain.OrderPaidPayload{
			OrderID:              uint64(order.ID),
			OrderNo:              order.OrderNo,
			UserID:               order.UserID,
			PaymentMethod:        order.PaymentMethod,
			PaymentTransactionID: order.PaymentTransactionID,
			OldStatus:            oldStatus,
			NewStatus:            order.Status,
			PaidAt:               paidAt,
			Log:                  buildEventLogFromOrder(order),
		}
		if err := s.appendEventInTx(ctx, tx, order, domain.OrderEventTypePaid, paidPayload); err != nil {
			return err
		}
		if err := s.repo.UpdateInTx(ctx, tx, order); err != nil {
			return err
		}
		return s.publisher.PublishInTx(ctx, tx, "order.paid", order.OrderNo, &domain.OrderPaidEvent{
			OrderID:              uint64(order.ID),
			OrderNo:              order.OrderNo,
			UserID:               order.UserID,
			ActualAmount:         order.ActualAmount,
			PaymentMethod:        order.PaymentMethod,
			PaymentTransactionID: order.PaymentTransactionID,
			PaidAt:               *order.PaidAt,
			Timestamp:            time.Now(),
		})
	})
}

// ShipOrder 发货。
func (s *OrderCommandService) ShipOrder(ctx context.Context, cmd *ShipOrderCommand) error {
	return s.repo.WithTx(ctx, cmd.UserID, func(tx any) error {
		order, err := s.repo.FindByID(ctx, cmd.UserID, uint64(cmd.OrderID))
		if err != nil || order == nil {
			return errors.New("order not found")
		}
		oldStatus := order.Status
		if cmd.TrackingNumber != "" {
			order.TrackingNumber = cmd.TrackingNumber
		}
		if cmd.LogisticsCompany != "" {
			order.LogisticsCompany = cmd.LogisticsCompany
		}
		if err := order.Ship(ctx, cmd.Operator); err != nil {
			return err
		}
		shippedAt := time.Now()
		if order.ShippedAt != nil {
			shippedAt = *order.ShippedAt
		}
		shippedPayload := &domain.OrderShippedPayload{
			OrderID:          uint64(order.ID),
			OrderNo:          order.OrderNo,
			UserID:           order.UserID,
			OldStatus:        oldStatus,
			NewStatus:        order.Status,
			TrackingNumber:   order.TrackingNumber,
			LogisticsCompany: order.LogisticsCompany,
			ShippedAt:        shippedAt,
			Log:              buildEventLogFromOrder(order),
		}
		if err := s.appendEventInTx(ctx, tx, order, domain.OrderEventTypeShipped, shippedPayload); err != nil {
			return err
		}
		if err := s.repo.UpdateInTx(ctx, tx, order); err != nil {
			return err
		}
		return s.publisher.PublishInTx(ctx, tx, "order.shipped", order.OrderNo, &domain.OrderShippedEvent{
			OrderID:          uint64(order.ID),
			OrderNo:          order.OrderNo,
			UserID:           order.UserID,
			TrackingNumber:   order.TrackingNumber,
			LogisticsCompany: order.LogisticsCompany,
			ShippedAt:        *order.ShippedAt,
			Timestamp:        time.Now(),
		})
	})
}

// DeliverOrder 送达。
func (s *OrderCommandService) DeliverOrder(ctx context.Context, cmd *DeliverOrderCommand) error {
	return s.repo.WithTx(ctx, cmd.UserID, func(tx any) error {
		order, err := s.repo.FindByID(ctx, cmd.UserID, uint64(cmd.OrderID))
		if err != nil || order == nil {
			return errors.New("order not found")
		}
		oldStatus := order.Status
		if cmd.TrackingNumber != "" {
			order.TrackingNumber = cmd.TrackingNumber
		}
		if cmd.LogisticsCompany != "" {
			order.LogisticsCompany = cmd.LogisticsCompany
		}
		if err := order.Deliver(ctx, cmd.Operator); err != nil {
			return err
		}
		deliveredAt := time.Now()
		if order.DeliveredAt != nil {
			deliveredAt = *order.DeliveredAt
		}
		deliveredPayload := &domain.OrderDeliveredPayload{
			OrderID:     uint64(order.ID),
			OrderNo:     order.OrderNo,
			UserID:      order.UserID,
			OldStatus:   oldStatus,
			NewStatus:   order.Status,
			DeliveredAt: deliveredAt,
			Log:         buildEventLogFromOrder(order),
		}
		if err := s.appendEventInTx(ctx, tx, order, domain.OrderEventTypeDelivered, deliveredPayload); err != nil {
			return err
		}
		if err := s.repo.UpdateInTx(ctx, tx, order); err != nil {
			return err
		}
		return s.publisher.PublishInTx(ctx, tx, "order.delivered", order.OrderNo, &domain.OrderDeliveredEvent{
			OrderID:     uint64(order.ID),
			OrderNo:     order.OrderNo,
			UserID:      order.UserID,
			DeliveredAt: *order.DeliveredAt,
			Timestamp:   time.Now(),
		})
	})
}

// CompleteOrder 完成。
func (s *OrderCommandService) CompleteOrder(ctx context.Context, cmd *CompleteOrderCommand) error {
	return s.repo.WithTx(ctx, cmd.UserID, func(tx any) error {
		order, err := s.repo.FindByID(ctx, cmd.UserID, uint64(cmd.OrderID))
		if err != nil || order == nil {
			return errors.New("order not found")
		}
		oldStatus := order.Status
		if err := order.Complete(ctx, cmd.Operator); err != nil {
			return err
		}
		completedAt := time.Now()
		if order.CompletedAt != nil {
			completedAt = *order.CompletedAt
		}
		completedPayload := &domain.OrderCompletedPayload{
			OrderID:     uint64(order.ID),
			OrderNo:     order.OrderNo,
			UserID:      order.UserID,
			OldStatus:   oldStatus,
			NewStatus:   order.Status,
			CompletedAt: completedAt,
			Log:         buildEventLogFromOrder(order),
		}
		if err := s.appendEventInTx(ctx, tx, order, domain.OrderEventTypeCompleted, completedPayload); err != nil {
			return err
		}
		if err := s.repo.UpdateInTx(ctx, tx, order); err != nil {
			return err
		}
		return s.publisher.PublishInTx(ctx, tx, "order.completed", order.OrderNo, &domain.OrderCompletedEvent{
			OrderID:     uint64(order.ID),
			OrderNo:     order.OrderNo,
			UserID:      order.UserID,
			CompletedAt: *order.CompletedAt,
			Timestamp:   time.Now(),
		})
	})
}

// CancelOrder 取消。
func (s *OrderCommandService) CancelOrder(ctx context.Context, cmd *CancelOrderCommand) error {
	return s.repo.WithTx(ctx, cmd.UserID, func(tx any) error {
		order, err := s.repo.FindByID(ctx, cmd.UserID, uint64(cmd.OrderID))
		if err != nil || order == nil {
			return errors.New("order not found")
		}
		oldStatus := order.Status
		if err := order.Cancel(ctx, cmd.Operator, cmd.Reason); err != nil {
			return err
		}
		cancelledAt := time.Now()
		if order.CancelledAt != nil {
			cancelledAt = *order.CancelledAt
		}
		cancelledPayload := &domain.OrderCancelledPayload{
			OrderID:     uint64(order.ID),
			OrderNo:     order.OrderNo,
			UserID:      order.UserID,
			OldStatus:   oldStatus,
			NewStatus:   order.Status,
			Reason:      cmd.Reason,
			CancelledAt: cancelledAt,
			Log:         buildEventLogFromOrder(order),
		}
		if err := s.appendEventInTx(ctx, tx, order, domain.OrderEventTypeCancelled, cancelledPayload); err != nil {
			return err
		}
		if err := s.repo.UpdateInTx(ctx, tx, order); err != nil {
			return err
		}
		return s.publisher.PublishInTx(ctx, tx, "order.cancelled", order.OrderNo, &domain.OrderCancelledEvent{
			OrderID:     uint64(order.ID),
			OrderNo:     order.OrderNo,
			UserID:      order.UserID,
			Reason:      cmd.Reason,
			CancelledAt: *order.CancelledAt,
			Timestamp:   time.Now(),
		})
	})
}

// RequestRefund 申请退款。
func (s *OrderCommandService) RequestRefund(ctx context.Context, cmd *RequestRefundCommand) error {
	return s.repo.WithTx(ctx, cmd.UserID, func(tx any) error {
		order, err := s.repo.FindByID(ctx, cmd.UserID, uint64(cmd.OrderID))
		if err != nil || order == nil {
			return errors.New("order not found")
		}
		oldStatus := order.Status
		if err := order.RequestRefund(ctx, cmd.Operator, cmd.Reason); err != nil {
			return err
		}
		if cmd.RefundAmount > 0 {
			order.RefundAmount = cmd.RefundAmount
		}
		if cmd.Reason != "" {
			order.RefundReason = cmd.Reason
		}
		requestedAt := time.Now()
		refundPayload := &domain.OrderRefundRequestedPayload{
			OrderID:      uint64(order.ID),
			OrderNo:      order.OrderNo,
			UserID:       order.UserID,
			OldStatus:    oldStatus,
			NewStatus:    order.Status,
			RefundAmount: order.RefundAmount,
			RefundReason: order.RefundReason,
			RequestedAt:  requestedAt,
			Log:          buildEventLogFromOrder(order),
		}
		if err := s.appendEventInTx(ctx, tx, order, domain.OrderEventTypeRefundRequested, refundPayload); err != nil {
			return err
		}
		if err := s.repo.UpdateInTx(ctx, tx, order); err != nil {
			return err
		}
		return s.publisher.PublishInTx(ctx, tx, "order.refund.requested", order.OrderNo, &domain.OrderRefundRequestedEvent{
			OrderID:      uint64(order.ID),
			OrderNo:      order.OrderNo,
			UserID:       order.UserID,
			RefundAmount: order.RefundAmount,
			RefundReason: order.RefundReason,
			Timestamp:    time.Now(),
		})
	})
}

// ApproveRefund 审核并确认退款完成。
func (s *OrderCommandService) ApproveRefund(ctx context.Context, cmd *ApproveRefundCommand) error {
	return s.repo.WithTx(ctx, cmd.UserID, func(tx any) error {
		order, err := s.repo.FindByID(ctx, cmd.UserID, uint64(cmd.OrderID))
		if err != nil || order == nil {
			return errors.New("order not found")
		}
		oldStatus := order.Status
		if err := order.ApproveRefund(ctx, cmd.Operator); err != nil {
			return err
		}
		refundedAt := time.Now()
		refundPayload := &domain.OrderRefundApprovedPayload{
			OrderID:    uint64(order.ID),
			OrderNo:    order.OrderNo,
			UserID:     order.UserID,
			OldStatus:  oldStatus,
			NewStatus:  order.Status,
			RefundedAt: refundedAt,
			Log:        buildEventLogFromOrder(order),
		}
		if err := s.appendEventInTx(ctx, tx, order, domain.OrderEventTypeRefundApproved, refundPayload); err != nil {
			return err
		}
		if err := s.repo.UpdateInTx(ctx, tx, order); err != nil {
			return err
		}
		return s.publisher.PublishInTx(ctx, tx, "order.refund.approved", order.OrderNo, &domain.OrderRefundApprovedEvent{
			OrderID:   uint64(order.ID),
			OrderNo:   order.OrderNo,
			UserID:    order.UserID,
			Timestamp: time.Now(),
		})
	})
}

// SagaConfirmOrder Saga 确认。
func (s *OrderCommandService) SagaConfirmOrder(ctx context.Context, userID, orderID uint64) error {
	return s.repo.WithTx(ctx, userID, func(tx any) error {
		order, err := s.repo.FindByID(ctx, userID, orderID)
		if err != nil || order == nil {
			return fmt.Errorf("order not found: %d", orderID)
		}
		if order.Status != orderv1.OrderStatus_ALLOCATING {
			return nil
		}
		oldStatus := order.Status
		if err := order.Trigger(ctx, "CONFIRM", "System", "Saga success"); err != nil {
			return err
		}
		confirmedPayload := &domain.OrderConfirmedPayload{
			OrderID:   uint64(order.ID),
			OrderNo:   order.OrderNo,
			UserID:    order.UserID,
			OldStatus: oldStatus,
			NewStatus: order.Status,
			Confirmed: time.Now(),
			Log:       buildEventLogFromOrder(order),
		}
		if err := s.appendEventInTx(ctx, tx, order, domain.OrderEventTypeConfirmed, confirmedPayload); err != nil {
			return err
		}
		if err := s.repo.UpdateInTx(ctx, tx, order); err != nil {
			return err
		}
		return s.publisher.PublishInTx(ctx, tx, "order.confirmed", order.OrderNo, &domain.OrderConfirmedEvent{
			OrderID:   uint64(order.ID),
			OrderNo:   order.OrderNo,
			UserID:    order.UserID,
			Timestamp: time.Now(),
		})
	})
}

// SagaCancelOrder Saga 取消。
func (s *OrderCommandService) SagaCancelOrder(ctx context.Context, userID, orderID uint64, reason string) error {
	return s.repo.WithTx(ctx, userID, func(tx any) error {
		order, err := s.repo.FindByID(ctx, userID, orderID)
		if err != nil || order == nil {
			return nil
		}
		if order.Status == orderv1.OrderStatus_CANCELLED {
			return nil
		}
		oldStatus := order.Status
		if err := order.Cancel(ctx, "System", reason); err != nil {
			return err
		}
		cancelledPayload := &domain.OrderCancelledPayload{
			OrderID:     uint64(order.ID),
			OrderNo:     order.OrderNo,
			UserID:      order.UserID,
			OldStatus:   oldStatus,
			NewStatus:   order.Status,
			Reason:      reason,
			CancelledAt: time.Now(),
			Log:         buildEventLogFromOrder(order),
		}
		if err := s.appendEventInTx(ctx, tx, order, domain.OrderEventTypeCancelled, cancelledPayload); err != nil {
			return err
		}
		if err := s.repo.UpdateInTx(ctx, tx, order); err != nil {
			return err
		}
		return s.publisher.PublishInTx(ctx, tx, "order.cancelled", order.OrderNo, &domain.OrderCancelledEvent{
			OrderID:     uint64(order.ID),
			OrderNo:     order.OrderNo,
			UserID:      order.UserID,
			Reason:      reason,
			CancelledAt: *order.CancelledAt,
			Timestamp:   time.Now(),
		})
	})
}

// assessRisk 封装风控逻辑，集中处理 map[string]any 转换并严格处理错误。
func (s *OrderCommandService) assessRisk(ctx context.Context, userID uint64, amount int64, ip, deviceID string) (*risk.Assessment, error) {
	data := map[string]any{
		"user_id":   userID,
		"amount":    amount,
		"client_ip": ip,
		"device_id": deviceID,
	}

	if s.positionCli != nil {
		posResp, err := s.positionCli.GetPositions(ctx, &positionv1.GetPositionsRequest{UserId: fmt.Sprintf("%d", userID)})
		if err != nil {
			s.logger.WarnContext(ctx, "failed to get positions for risk assessment", "error", err)
			// 不作为致命错误，继续评估
		} else {
			var totalEquity decimal.Decimal
			for _, p := range posResp.GetPositions() {
				q, err := decimal.NewFromString(p.Quantity)
				if err != nil {
					continue
				}
				pr, err := decimal.NewFromString(p.CurrentPrice)
				if err != nil {
					continue
				}
				totalEquity = totalEquity.Add(q.Mul(pr))
			}
			val, ok := totalEquity.Float64()
			if !ok {
				s.logger.WarnContext(ctx, "equity float conversion is not exact", "total_equity", totalEquity.String())
			}
			data["trading_asset_value"] = val
		}
	}

	assessment, err := s.riskEvaluator.Assess(ctx, "order.create", data)
	if err != nil {
		return nil, fmt.Errorf("risk assessment call fail: %w", err)
	}
	return assessment, nil
}

// HandleFlashsaleOrder 秒杀持久化 (EDA).
func (s *OrderCommandService) HandleFlashsaleOrder(ctx context.Context, orderID, userID, productID, skuID uint64, quantity int32, price int64) error {
	orderNo := fmt.Sprintf("FS%d", orderID)
	items := []*domain.OrderItem{{
		SkuID: skuID, ProductID: productID, Quantity: quantity, Price: price, TotalPrice: price * int64(quantity),
		ProductName: "Flashsale", SkuName: "Flashsale",
	}}
	order := domain.NewOrder(orderNo, userID, items, nil)
	order.ID = uint(orderID)
	order.Status = orderv1.OrderStatus_PENDING_PAYMENT
	return s.repo.WithTx(ctx, userID, func(tx any) error {
		if err := s.repo.SaveInTx(ctx, tx, order); err != nil {
			return err
		}
		createdPayload := &domain.OrderCreatedPayload{
			OrderID:         uint64(order.ID),
			OrderNo:         order.OrderNo,
			UserID:          order.UserID,
			Status:          order.Status,
			TotalAmount:     order.TotalAmount,
			ActualAmount:    order.ActualAmount,
			ShippingFee:     order.ShippingFee,
			DiscountAmount:  order.DiscountAmount,
			PaymentMethod:   order.PaymentMethod,
			Remark:          order.Remark,
			ShippingAddress: order.ShippingAddress,
			Items:           order.Items,
			CreatedAt:       time.Now(),
			InitLog:         buildEventLogFromOrder(order),
		}
		if err := s.appendEventInTx(ctx, tx, order, domain.OrderEventTypeCreated, createdPayload); err != nil {
			return err
		}
		if err := s.repo.UpdateInTx(ctx, tx, order); err != nil {
			return err
		}
		return s.publisher.PublishInTx(ctx, tx, "order.created", order.OrderNo, &domain.OrderCreatedEvent{
			OrderID:     uint64(order.ID),
			OrderNo:     order.OrderNo,
			UserID:      order.UserID,
			TotalAmount: order.TotalAmount,
			Status:      order.Status,
			Timestamp:   time.Now(),
		})
	})
}

// appendEventInTx 在同一事务中追加事件并推进版本号。
func (s *OrderCommandService) appendEventInTx(ctx context.Context, tx any, order *domain.Order, eventType string, payload any) error {
	if s.eventStore == nil {
		return errors.New("event store is not configured")
	}

	nextVersion := order.Version + 1
	aggregateID := fmt.Sprintf("%d", order.ID)

	metadata := eventsourcing.Metadata{
		UserID:  fmt.Sprintf("%d", order.UserID),
		TraceID: tracing.GetTraceID(ctx),
		Extra:   map[string]string{"service": "order"},
	}

	base := eventsourcing.NewBaseEventWithMetadata(eventType, aggregateID, nextVersion, metadata)
	base.Data = payload

	if err := s.eventStore.SaveInTx(ctx, tx, order.UserID, aggregateID, []eventsourcing.DomainEvent{&base}, order.Version); err != nil {
		return err
	}

	order.Version = nextVersion
	return nil
}

// buildEventLogFromOrder 从订单日志中构建事件溯源日志载荷。
func buildEventLogFromOrder(order *domain.Order) *domain.OrderEventLog {
	if len(order.Logs) == 0 {
		return nil
	}
	last := order.Logs[len(order.Logs)-1]
	return &domain.OrderEventLog{
		Operator:  last.Operator,
		Action:    last.Action,
		OldStatus: last.OldStatus,
		NewStatus: last.NewStatus,
		Remark:    last.Remark,
		LoggedAt:  time.Now(),
	}
}
