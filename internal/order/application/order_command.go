package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
	advancedcouponv1 "github.com/wyfcoding/ecommerce/goapi/advancedcoupon/v1"
	inventoryv1 "github.com/wyfcoding/ecommerce/goapi/inventory/v1"
	orderv1 "github.com/wyfcoding/ecommerce/goapi/order/v1"
	paymentv1 "github.com/wyfcoding/ecommerce/goapi/payment/v1"
	productv1 "github.com/wyfcoding/ecommerce/goapi/product/v1"
	warehousev1 "github.com/wyfcoding/ecommerce/goapi/warehouse/v1"
	"github.com/wyfcoding/ecommerce/internal/order/domain"
	positionv1 "github.com/wyfcoding/financialtrading/go-api/position/v1"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/wyfcoding/pkg/dtm"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/security/risk"
)

type OrderCommandService struct {
	repo              domain.OrderRepository
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

func NewOrderCommandService(
	repo domain.OrderRepository,
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
		idGen:               idGen,
		publisher:           publisher,
		logger:              logger,
		dtmServer:           dtmServer,
		warehouseGrpcAddr:   warehouseGrpcAddr,
		riskEvaluator:       riskEvaluator,
		orderCreatedCounter: orderCreatedCounter,
	}
}

// SetClients 注入下游服务客户端依赖。
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

// SetSvcURL 设置当前服务的访问地址，用于 DTM 回调。
func (s *OrderCommandService) SetSvcURL(url string) {
	s.orderSvcURL = url
}

// CreateOrder 创建订单。
func (s *OrderCommandService) CreateOrder(ctx context.Context, cmd *CreateOrderCommand) (*domain.Order, error) {
	// --- 1. 验证并获取商品信息 (Price Safety) ---
	var items []*domain.OrderItem
	var totalAmount int64
	for _, it := range cmd.Items {
		// 从商品服务获取最新价格与名称 (防止前端篡改价格)
		sku, err := s.productCli.GetSKUByID(ctx, &productv1.GetSKUByIDRequest{Id: it.SkuID})
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to fetch SKU info", "sku_id", it.SkuID, "error", err)
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

	// --- 2. 深度风控拦截 ---
	riskData := map[string]any{
		"user_id":      cmd.UserID,
		"amount":       totalAmount,
		"item_count":   len(items),
		"client_ip":    cmd.ClientIP,
		"device_id":    cmd.DeviceID,
		"is_real_name": true,
	}

	// 引入交易资产价值增强风控精度
	if s.positionCli != nil {
		posResp, err := s.positionCli.GetPositions(ctx, &positionv1.GetPositionsRequest{
			UserId: fmt.Sprintf("%d", cmd.UserID),
		})
		if err == nil {
			var totalEquity decimal.Decimal
			for _, p := range posResp.GetPositions() {
				val, _ := decimal.NewFromString(p.Quantity)
				price, _ := decimal.NewFromString(p.CurrentPrice)
				totalEquity = totalEquity.Add(val.Mul(price))
			}
			riskData["trading_asset_value"], _ = totalEquity.Float64()
		}
	}

	riskAssessment, err := s.riskEvaluator.Assess(ctx, "order.create", riskData)
	if err != nil {
		s.logger.ErrorContext(ctx, "risk assessment failed, fail-open applied", "error", err)
	} else {
		switch riskAssessment.Level {
		case risk.Reject:
			s.logger.WarnContext(ctx, "order rejected by risk control",
				"user_id", cmd.UserID, "code", riskAssessment.Code, "reason", riskAssessment.Reason)
			return nil, fmt.Errorf("transaction security risk: %s", riskAssessment.Reason)
		case risk.Review:
			s.logger.InfoContext(ctx, "order needs risk review",
				"user_id", cmd.UserID, "code", riskAssessment.Code)
		}
	}

	orderID := s.idGen.Generate()
	orderNo := fmt.Sprintf("%s%d", time.Now().Format("20060102"), orderID)

	order := domain.NewOrder(orderNo, cmd.UserID, items, cmd.ShippingAddress)
	order.Status = orderv1.OrderStatus_ALLOCATING

	// --- 3. 预同步锁定库存 (Optimistic Lock) ---
	for _, item := range items {
		_, err := s.inventoryCli.LockStock(ctx, &inventoryv1.LockStockRequest{
			SkuId:    item.SkuID,
			Quantity: int32(item.Quantity),
			Reason:   "Order " + orderNo,
		})
		if err != nil {
			s.logger.ErrorContext(ctx, "stock lock failed", "sku_id", item.SkuID, "error", err)
			return nil, fmt.Errorf("insufficient stock for SKU %d", item.SkuID)
		}
	}

	// --- 4. 本地事务：落地并发布 Outbox (EDA) ---
	err = s.repo.Transaction(ctx, cmd.UserID, func(tx any) error {
		txRepo := s.repo.WithTx(tx)
		if err := txRepo.Save(ctx, order); err != nil {
			return err
		}

		// 4.1 发布订单创建事件
		event := map[string]any{
			"order_id": order.ID,
			"order_no": order.OrderNo,
			"user_id":  order.UserID,
			"amount":   order.TotalAmount,
			"status":   order.Status.String(),
		}
		if err := s.publisher.PublishInTx(ctx, tx, "order.created", orderNo, event); err != nil {
			return err
		}

		// 4.2 发布超时自动取消任务 (EDA + Delay Task)
		timeoutEvent := map[string]any{
			"order_id":   order.ID,
			"order_no":   order.OrderNo,
			"user_id":    order.UserID,
			"expires_at": time.Now().Add(15 * time.Minute).Unix(),
		}
		return s.publisher.PublishInTx(ctx, tx, "order.payment.timeout", orderNo, timeoutEvent)
	})
	if err != nil {
		return nil, err
	}

	s.orderCreatedCounter.WithLabelValues(order.Status.String()).Inc()

	// --- 5. 提交 DTM Saga 分布式事务 (Final Consistency) ---
	saga := dtm.NewSaga(ctx, s.dtmServer, orderNo)
	orderGrpcPrefix := s.orderSvcURL + "/api.order.v1.OrderService"
	warehouseGrpcPrefix := s.warehouseGrpcAddr + "/api.warehouse.v1.WarehouseService"

	saga.Add(orderGrpcPrefix+"/SagaConfirmOrder", orderGrpcPrefix+"/SagaCancelOrder", &orderv1.SagaOrderRequest{
		UserId:  cmd.UserID,
		OrderId: uint64(order.ID),
	})

	for _, item := range items {
		saga.Add(warehouseGrpcPrefix+"/DeductStock", warehouseGrpcPrefix+"/RevertStock", &warehousev1.DeductStockRequest{
			OrderId:     uint64(order.ID),
			SkuId:       item.SkuID,
			Quantity:    item.Quantity,
			WarehouseId: 1,
		})
	}

	if cmd.CouponCode != "" {
		couponSvcAddr := "advancedcoupon:50051"
		saga.Add(couponSvcAddr+"/api.advancedcoupon.v1.AdvancedCouponService/UseCoupon", "", &advancedcouponv1.UseCouponRequest{
			UserId:  cmd.UserID,
			Code:    cmd.CouponCode,
			OrderId: uint64(order.ID),
		})
	}

	if err := saga.Submit(ctx); err != nil {
		s.logger.ErrorContext(ctx, "saga submission failed", "order_no", orderNo, "error", err)
		return nil, fmt.Errorf("distributed transaction submission failed: %w", err)
	}

	// --- 6. 同步发起支付请求 (User Experience) ---
	if s.paymentCli != nil {
		payResp, err := s.paymentCli.InitiatePayment(ctx, &paymentv1.InitiatePaymentRequest{
			OrderId:       uint64(order.ID),
			UserId:        cmd.UserID,
			PaymentMethod: "WECHAT",
			Amount:        order.TotalAmount,
			ClientIp:      cmd.ClientIP,
		})
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to initiate payment", "order_no", orderNo, "error", err)
		} else {
			s.logger.InfoContext(ctx, "payment initiated", "order_no", orderNo, "transaction_no", payResp.TransactionNo)
		}
	}

	return order, nil
}

// PayOrder 支付订单 (EDA).
func (s *OrderCommandService) PayOrder(ctx context.Context, cmd *PayOrderCommand) error {
	return s.repo.Transaction(ctx, cmd.UserID, func(tx any) error {
		txRepo := s.repo.WithTx(tx)
		order, err := txRepo.FindByID(ctx, cmd.UserID, uint(cmd.OrderID))
		if err != nil || order == nil {
			return errors.New("order not found")
		}

		if err := order.Pay(ctx, cmd.PaymentMethod, "User"); err != nil {
			return err
		}

		if err := txRepo.Save(ctx, order); err != nil {
			return err
		}

		// 发布支付成功事件 (EDA)
		event := map[string]any{
			"order_id": order.ID,
			"order_no": order.OrderNo,
			"user_id":  cmd.UserID,
			"amount":   order.ActualAmount,
			"paid_at":  time.Now().Unix(),
		}
		return s.publisher.PublishInTx(ctx, tx, "order.paid", order.OrderNo, event)
	})
}

// ShipOrder 发货订单 (EDA).
func (s *OrderCommandService) ShipOrder(ctx context.Context, cmd *ShipOrderCommand) error {
	return s.repo.Transaction(ctx, cmd.UserID, func(tx any) error {
		txRepo := s.repo.WithTx(tx)
		order, err := txRepo.FindByID(ctx, cmd.UserID, uint(cmd.OrderID))
		if err != nil || order == nil {
			return errors.New("order not found")
		}
		if err := order.Ship(ctx, cmd.Operator); err != nil {
			return err
		}
		if err := txRepo.Save(ctx, order); err != nil {
			return err
		}

		event := map[string]any{
			"order_id":   order.ID,
			"order_no":   order.OrderNo,
			"user_id":    order.UserID,
			"shipped_at": time.Now().Unix(),
		}
		return s.publisher.PublishInTx(ctx, tx, "order.shipped", order.OrderNo, event)
	})
}

// DeliverOrder 送达订单 (EDA).
func (s *OrderCommandService) DeliverOrder(ctx context.Context, cmd *DeliverOrderCommand) error {
	return s.repo.Transaction(ctx, cmd.UserID, func(tx any) error {
		txRepo := s.repo.WithTx(tx)
		order, err := txRepo.FindByID(ctx, cmd.UserID, uint(cmd.OrderID))
		if err != nil || order == nil {
			return errors.New("order not found")
		}
		if err := order.Deliver(ctx, cmd.Operator); err != nil {
			return err
		}
		if err := txRepo.Save(ctx, order); err != nil {
			return err
		}

		event := map[string]any{
			"order_id":     order.ID,
			"order_no":     order.OrderNo,
			"user_id":      order.UserID,
			"delivered_at": time.Now().Unix(),
		}
		return s.publisher.PublishInTx(ctx, tx, "order.delivered", order.OrderNo, event)
	})
}

// CompleteOrder 完成订单 (EDA).
func (s *OrderCommandService) CompleteOrder(ctx context.Context, cmd *CompleteOrderCommand) error {
	return s.repo.Transaction(ctx, cmd.UserID, func(tx any) error {
		txRepo := s.repo.WithTx(tx)
		order, err := txRepo.FindByID(ctx, cmd.UserID, uint(cmd.OrderID))
		if err != nil || order == nil {
			return errors.New("order not found")
		}
		if err := order.Complete(ctx, cmd.Operator); err != nil {
			return err
		}
		if err := txRepo.Save(ctx, order); err != nil {
			return err
		}

		event := map[string]any{
			"order_id":     order.ID,
			"order_no":     order.OrderNo,
			"user_id":      order.UserID,
			"completed_at": time.Now().Unix(),
		}
		return s.publisher.PublishInTx(ctx, tx, "order.completed", order.OrderNo, event)
	})
}

// CancelOrder 取消订单 (EDA).
func (s *OrderCommandService) CancelOrder(ctx context.Context, cmd *CancelOrderCommand) error {
	return s.repo.Transaction(ctx, cmd.UserID, func(tx any) error {
		txRepo := s.repo.WithTx(tx)
		order, err := txRepo.FindByID(ctx, cmd.UserID, uint(cmd.OrderID))
		if err != nil || order == nil {
			return errors.New("order not found")
		}
		if err := order.Cancel(ctx, cmd.Operator, cmd.Reason); err != nil {
			return err
		}
		if err := txRepo.Save(ctx, order); err != nil {
			return err
		}

		event := map[string]any{
			"order_id":     order.ID,
			"order_no":     order.OrderNo,
			"user_id":      order.UserID,
			"reason":       cmd.Reason,
			"cancelled_at": time.Now().Unix(),
		}
		return s.publisher.PublishInTx(ctx, tx, "order.cancelled", order.OrderNo, event)
	})
}

func (s *OrderCommandService) SagaConfirmOrder(ctx context.Context, userID, orderID uint64) error {
	return s.repo.Transaction(ctx, userID, func(tx any) error {
		txRepo := s.repo.WithTx(tx)
		order, err := txRepo.FindByID(ctx, userID, uint(orderID))
		if err != nil || order == nil {
			return fmt.Errorf("order not found: %d", orderID)
		}
		if order.Status != orderv1.OrderStatus_ALLOCATING {
			return nil
		}

		if err := order.Trigger(ctx, "CONFIRM", "System", "Saga verification success"); err != nil {
			return err
		}

		if err := txRepo.Save(ctx, order); err != nil {
			return err
		}

		// 发布确认事件 (EDA)
		event := map[string]any{
			"order_id": order.ID,
			"order_no": order.OrderNo,
			"user_id":  order.UserID,
			"status":   "PENDING_PAYMENT",
		}
		return s.publisher.PublishInTx(ctx, tx, "order.confirmed", order.OrderNo, event)
	})
}

func (s *OrderCommandService) SagaCancelOrder(ctx context.Context, userID, orderID uint64, reason string) error {
	return s.repo.Transaction(ctx, userID, func(tx any) error {
		txRepo := s.repo.WithTx(tx)
		order, err := txRepo.FindByID(ctx, userID, uint(orderID))
		if err != nil || order == nil {
			return fmt.Errorf("order not found: %d", orderID)
		}
		if order.Status == orderv1.OrderStatus_CANCELLED {
			return nil
		}

		if err := order.Cancel(ctx, "System", reason); err != nil {
			return err
		}

		if err := txRepo.Save(ctx, order); err != nil {
			return err
		}

		event := map[string]any{
			"order_id": order.ID,
			"order_no": order.OrderNo,
			"reason":   reason,
		}
		return s.publisher.PublishInTx(ctx, tx, "order.cancelled", order.OrderNo, event)
	})
}

func (s *OrderCommandService) HandleFlashsaleOrder(ctx context.Context, orderID, userID, productID, skuID uint64, quantity int32, price int64) error {
	s.logger.InfoContext(ctx, "Flashsale persistence", "order_id", orderID, "user_id", userID)

	orderNo := fmt.Sprintf("FS%d", orderID)
	items := []*domain.OrderItem{
		{
			SkuID:       skuID,
			ProductID:   productID,
			Quantity:    quantity,
			Price:       price,
			ProductName: "Flashsale Product",
			SkuName:     "Flashsale SKU",
			TotalPrice:  price * int64(quantity),
		},
	}

	order := domain.NewOrder(orderNo, userID, items, nil)
	order.ID = uint(orderID)
	order.Status = orderv1.OrderStatus_PENDING_PAYMENT

	if err := s.repo.Save(ctx, order); err != nil {
		return err
	}

	s.orderCreatedCounter.WithLabelValues(order.Status.String()).Inc()
	return nil
}
