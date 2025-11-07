package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	v1 "ecommerce/api/order/v1"
	"ecommerce/internal/order/model"
	"ecommerce/internal/order/repository"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// OrderService 封装了订单相关的业务逻辑，实现了 order.proto 中定义的 OrderServer 接口。
type OrderService struct {
	// 嵌入 v1.UnimplementedOrderServer 以确保向前兼容性
	v1.UnimplementedOrderServer
	orderRepo       repository.OrderRepo
	orderItemRepo   repository.OrderItemRepo
	shippingAddrRepo repository.ShippingAddressRepo
	orderLogRepo    repository.OrderLogRepo
	validator       *validator.Validate
	defaultPageSize int32
	maxPageSize     int32
	orderExpiration time.Duration

	// 假设这里有对其他服务的客户端，例如 ProductService 客户端
	// productClient v1_product.ProductClient
}

// NewOrderService 是 OrderService 的构造函数。
func NewOrderService(
	orderRepo repository.OrderRepo,
	orderItemRepo repository.OrderItemRepo,
	shippingAddrRepo repository.ShippingAddressRepo,
	orderLogRepo repository.OrderLogRepo,
	defaultPageSize, maxPageSize int32,
	orderExpirationMinutes int,
) *OrderService {
	return &OrderService{
		orderRepo:        orderRepo,
		orderItemRepo:    orderItemRepo,
		shippingAddrRepo: shippingAddrRepo,
		orderLogRepo:     orderLogRepo,
		validator:        validator.New(),
		defaultPageSize:  defaultPageSize,
		maxPageSize:      maxPageSize,
		orderExpiration:  time.Duration(orderExpirationMinutes) * time.Minute,
	}
}

// --- 订单核心接口实现 ---

// CreateOrder 实现了创建订单的 RPC 方法。
// 这是一个复杂的业务流程，涉及商品校验、库存扣减（预扣）、价格计算、订单生成、日志记录等。
func (s *OrderService) CreateOrder(ctx context.Context, req *v1.CreateOrderRequest) (*v1.OrderInfo, error) {
	zap.S().Infof("CreateOrder request received for user %d", req.UserId)

	// 1. 参数校验
	if err := s.validator.Struct(req); err != nil {
		zap.S().Warnf("CreateOrder request validation failed: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid argument: %v", err)
	}
	if len(req.Items) == 0 {
		zap.S().Warn("CreateOrder request with no items")
		return nil, status.Errorf(codes.InvalidArgument, "order must contain at least one item")
	}

	// 2. 校验商品信息和库存 (此处仅为模拟，实际应调用 ProductService 和 InventoryService)
	var totalAmount int64 = 0
	var orderItems []*model.OrderItem
	for _, itemReq := range req.Items {
		// 假设这里调用 ProductService 获取商品和SKU详情，并校验库存
		// product, err := s.productClient.GetProductByID(ctx, &v1_product.GetProductByIDRequest{Id: itemReq.ProductId})
		// sku, err := s.productClient.GetSKUByID(ctx, &v1_product.GetSKUByIDRequest{Id: itemReq.SkuId})
		// if sku == nil || sku.StockQuantity < itemReq.Quantity {
		// 	return nil, status.Errorf(codes.ResourceExhausted, "SKU %d stock not enough", itemReq.SkuId)
		// }
		// 模拟商品数据
		productName := fmt.Sprintf("Product-%d", itemReq.ProductId)
		skuName := fmt.Sprintf("SKU-%d-Spec", itemReq.SkuId)
		price := int64(10000) // 模拟价格，100.00元

		if itemReq.Quantity <= 0 {
			return nil, status.Errorf(codes.InvalidArgument, "item quantity must be positive")
		}

		itemTotalPrice := price * int64(itemReq.Quantity)
		totalAmount += itemTotalPrice

		orderItems = append(orderItems, &model.OrderItem{
			ProductID:       itemReq.ProductId,
			SkuID:           itemReq.SkuId,
			ProductName:     productName,
			SkuName:         skuName,
			ProductImageURL: "http://example.com/image.jpg", // 模拟图片URL
			Price:           price,
			Quantity:        itemReq.Quantity,
			TotalPrice:      itemTotalPrice,
		})
	}

	// 3. 计算最终金额 (此处简化，不考虑优惠券和运费)
	shippingFee := int64(0) // 模拟运费
	discountAmount := int64(0) // 模拟优惠金额
	actualAmount := totalAmount + shippingFee - discountAmount

	// 4. 构建订单模型
	order := &model.Order{
		UserID:         req.UserId,
		Status:         model.PendingPayment,
		PaymentStatus:  model.Unpaid,
		ShippingStatus: model.PendingShipment,
		TotalAmount:    totalAmount,
		ActualAmount:   actualAmount,
		ShippingFee:    shippingFee,
		DiscountAmount: discountAmount,
		PaymentMethod:  req.PaymentMethod,
		Remark:         req.Remark,
		Items:          orderItems,
	}

	// 自动生成订单编号 (示例: O + 时间戳 + 随机数)
	order.OrderNo = fmt.Sprintf("O%d%s", time.Now().UnixNano()/int64(time.Millisecond), strconv.Itoa(int(time.Now().Nanosecond()%1000)))

	// 构建收货地址
	if req.ShippingAddress != nil {
		order.ShippingAddress = model.ShippingAddress{
			RecipientName:   req.ShippingAddress.RecipientName,
			PhoneNumber:     req.ShippingAddress.PhoneNumber,
			Province:        req.ShippingAddress.Province,
			City:            req.ShippingAddress.City,
			District:        req.ShippingAddress.District,
			DetailedAddress: req.ShippingAddress.DetailedAddress,
			PostalCode:      req.ShippingAddress.PostalCode,
		}
	}

	// 记录订单创建日志
	order.Logs = append(order.Logs, &model.OrderLog{
		Operator:  fmt.Sprintf("User-%d", req.UserId),
		Action:    "Order Created",
		NewStatus: model.PendingPayment.String(),
		Remark:    "Initial order creation",
	})

	// 5. 调用仓库层创建订单 (包含订单项和收货地址)
	createdOrder, err := s.orderRepo.CreateOrder(ctx, order)
	if err != nil {
		zap.S().Errorf("failed to create order for user %d: %v", req.UserId, err)
		return nil, status.Errorf(codes.Internal, "failed to create order")
	}

	// 6. 预扣库存 (此处仅为模拟，实际应调用 InventoryService)
	// for _, item := range createdOrder.Items {
	// 	_ = s.inventoryClient.PreDeductStock(ctx, item.SkuID, item.Quantity)
	// }

	zap.S().Infof("Order %d (OrderNo: %s) created successfully for user %d", createdOrder.ID, createdOrder.OrderNo, req.UserId)
	return s.bizOrderToProto(createdOrder), nil
}

// GetOrderByID 实现了根据ID获取订单详情的 RPC 方法。
func (s *OrderService) GetOrderByID(ctx context.Context, req *v1.GetOrderByIDRequest) (*v1.OrderInfo, error) {
	zap.S().Infof("GetOrderByID request received for order ID: %d, user ID: %d", req.Id, req.UserId)

	// 1. 参数校验
	if req.Id == 0 || req.UserId == 0 {
		zap.S().Warn("GetOrderByID request with zero order ID or user ID")
		return nil, status.Errorf(codes.InvalidArgument, "order ID and user ID cannot be zero")
	}

	// 2. 调用仓库层获取订单
	order, err := s.orderRepo.GetOrderByID(ctx, req.Id)
	if err != nil {
		zap.S().Errorf("failed to get order by ID %d: %v", req.Id, err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve order")
	}
	if order == nil {
		zap.S().Warnf("order with ID %d not found", req.Id)
		return nil, status.Errorf(codes.NotFound, "order not found")
	}

	// 3. 权限校验：确保用户只能查看自己的订单
	if order.UserID != req.UserId {
		zap.S().Warnf("user %d attempted to access order %d belonging to user %d", req.UserId, req.Id, order.UserID)
		return nil, status.Errorf(codes.PermissionDenied, "access denied")
	}

	zap.S().Infof("Order %d retrieved successfully for user %d", req.Id, req.UserId)
	return s.bizOrderToProto(order), nil
}

// UpdateOrderStatus 实现了更新订单状态的 RPC 方法。
// 这是一个关键业务逻辑，需要严格的状态流转校验。
func (s *OrderService) UpdateOrderStatus(ctx context.Context, req *v1.UpdateOrderStatusRequest) (*v1.OrderInfo, error) {
	zap.S().Infof("UpdateOrderStatus request received for order ID: %d, new status: %s", req.Id, req.NewStatus.String())

	// 1. 参数校验
	if req.Id == 0 || req.NewStatus == v1.OrderStatus_ORDER_STATUS_UNSPECIFIED {
		zap.S().Warn("UpdateOrderStatus request with invalid order ID or new status")
		return nil, status.Errorf(codes.InvalidArgument, "invalid order ID or new status")
	}

	// 2. 获取现有订单信息
	existingOrder, err := s.orderRepo.GetOrderByID(ctx, req.Id)
	if err != nil {
		zap.S().Errorf("failed to get existing order %d for status update: %v", req.Id, err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve order for update")
	}
	if existingOrder == nil {
		zap.S().Warnf("order with ID %d not found for status update", req.Id)
		return nil, status.Errorf(codes.NotFound, "order not found")
	}

	oldStatus := existingOrder.Status
	newStatus := model.OrderStatus(req.NewStatus)

	// 3. 状态流转校验 (复杂业务逻辑)
	if !isValidStatusTransition(oldStatus, newStatus) {
		zap.S().Warnf("invalid order status transition from %s to %s for order %d", oldStatus.String(), newStatus.String(), req.Id)
		return nil, status.Errorf(codes.FailedPrecondition, "invalid status transition from %s to %s", oldStatus.String(), newStatus.String())
	}

	// 4. 更新订单状态和相关时间戳
	existingOrder.Status = newStatus
	switch newStatus {
	case model.Paid:
		now := time.Now()
		existingOrder.PaymentStatus = model.Success
		existingOrder.PaidAt = &now
	case model.Shipped:
		now := time.Now()
		existingOrder.ShippingStatus = model.ShippingShipped
		existingOrder.ShippedAt = &now
	case model.Delivered:
		now := time.Now()
		existingOrder.ShippingStatus = model.ShippingDelivered
		existingOrder.DeliveredAt = &now
	case model.Completed:
		now := time.Now()
		existingOrder.CompletedAt = &now
	case model.Cancelled, model.Closed:
		now := time.Now()
		existingOrder.CancelledAt = &now
		// 如果是取消或关闭，且已支付，可能需要触发退款流程
		if existingOrder.PaymentStatus == model.Success {
			zap.S().Warnf("order %d cancelled/closed after payment, refund needed", req.Id)
			// 实际应触发退款流程
		}
	}

	// 5. 记录订单日志
	logEntry := &model.OrderLog{
		OrderID:   req.Id,
		Operator:  req.Operator,
		Action:    fmt.Sprintf("Status Updated to %s", newStatus.String()),
		OldStatus: oldStatus.String(),
		NewStatus: newStatus.String(),
		Remark:    req.Remark,
	}
	if _, err := s.orderLogRepo.CreateOrderLog(ctx, logEntry); err != nil {
		zap.S().Errorf("failed to create order log for order %d: %v", req.Id, err)
		// 不阻断主流程，但记录错误
	}

	// 6. 调用仓库层更新订单
	updatedOrder, err := s.orderRepo.UpdateOrder(ctx, existingOrder)
	if err != nil {
		zap.S().Errorf("failed to update order %d status to %s: %v", req.Id, newStatus.String(), err)
		return nil, status.Errorf(codes.Internal, "failed to update order status")
	}

	zap.S().Infof("Order %d status updated from %s to %s successfully", req.Id, oldStatus.String(), newStatus.String())
	return s.bizOrderToProto(updatedOrder), nil
}

// CancelOrder 实现了取消订单的 RPC 方法。
func (s *OrderService) CancelOrder(ctx context.Context, req *v1.CancelOrderRequest) (*v1.OrderInfo, error) {
	zap.S().Infof("CancelOrder request received for order ID: %d, user ID: %d", req.Id, req.UserId)

	// 1. 参数校验
	if req.Id == 0 || req.UserId == 0 {
		zap.S().Warn("CancelOrder request with invalid order ID or user ID")
		return nil, status.Errorf(codes.InvalidArgument, "order ID and user ID cannot be zero")
	}

	// 2. 获取现有订单信息
	existingOrder, err := s.orderRepo.GetOrderByID(ctx, req.Id)
	if err != nil {
		zap.S().Errorf("failed to get existing order %d for cancellation: %v", req.Id, err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve order for cancellation")
	}
	if existingOrder == nil {
		zap.S().Warnf("order with ID %d not found for cancellation", req.Id)
		return nil, status.Errorf(codes.NotFound, "order not found")
	}

	// 3. 权限校验
	if existingOrder.UserID != req.UserId {
		zap.S().Warnf("user %d attempted to cancel order %d belonging to user %d", req.UserId, req.Id, existingOrder.UserID)
		return nil, status.Errorf(codes.PermissionDenied, "access denied")
	}

	// 4. 状态校验：只有特定状态的订单才能取消
	if existingOrder.Status != model.PendingPayment && existingOrder.Status != model.Paid {
		zap.S().Warnf("order %d in status %s cannot be cancelled", req.Id, existingOrder.Status.String())
		return nil, status.Errorf(codes.FailedPrecondition, "order in status %s cannot be cancelled", existingOrder.Status.String())
	}

	// 5. 更新订单状态为已取消
	oldStatus := existingOrder.Status
	existingOrder.Status = model.Cancelled
	existingOrder.CancelledAt = func() *time.Time { t := time.Now(); return &t }()

	// 6. 记录订单日志
	logEntry := &model.OrderLog{
		OrderID:   req.Id,
		Operator:  fmt.Sprintf("User-%d", req.UserId),
		Action:    "Order Cancelled",
		OldStatus: oldStatus.String(),
		NewStatus: model.Cancelled.String(),
		Remark:    req.Reason,
	}
	if _, err := s.orderLogRepo.CreateOrderLog(ctx, logEntry); err != nil {
		zap.S().Errorf("failed to create order log for order %d cancellation: %v", req.Id, err)
	}

	// 7. 调用仓库层更新订单
	updatedOrder, err := s.orderRepo.UpdateOrder(ctx, existingOrder)
	if err != nil {
		zap.S().Errorf("failed to cancel order %d: %v", req.Id, err)
		return nil, status.Errorf(codes.Internal, "failed to cancel order")
	}

	// 8. 释放库存 (此处仅为模拟，实际应调用 InventoryService)
	// for _, item := range updatedOrder.Items {
	// 	_ = s.inventoryClient.ReleaseStock(ctx, item.SkuID, item.Quantity)
	// }

	zap.S().Infof("Order %d cancelled successfully by user %d", req.Id, req.UserId)
	return s.bizOrderToProto(updatedOrder), nil
}

// ListOrders 实现了分页列出订单的 RPC 方法。
func (s *OrderService) ListOrders(ctx context.Context, req *v1.ListOrdersRequest) (*v1.ListOrdersResponse, error) {
	zap.S().Infof("ListOrders request received: user_id=%d, status=%s, page=%d, page_size=%d",
		req.UserId, req.Status.String(), req.Page, req.PageSize)

	// 1. 参数校验与默认值设置
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = s.defaultPageSize
	}
	if pageSize > s.maxPageSize {
		pageSize = s.maxPageSize
	}
	page := req.Page
	if page == 0 {
		page = 1
	}

	query := &repository.OrderListQuery{
		Page:     page,
		PageSize: pageSize,
		UserID:   req.UserId,
		Status:   model.OrderStatus(req.Status),
		SortBy:   req.SortBy,
	}
	if req.StartTime != nil {
		startTime := req.StartTime.AsTime()
		query.StartTime = &startTime
	}
	if req.EndTime != nil {
		endTime := req.EndTime.AsTime()
		query.EndTime = &endTime
	}

	// 2. 调用仓库层查询订单列表
	orders, total, err := s.orderRepo.ListOrders(ctx, query)
	if err != nil {
		zap.S().Errorf("failed to list orders for user %d: %v", req.UserId, err)
		return nil, status.Errorf(codes.Internal, "failed to list orders")
	}

	// 3. 转换为Protobuf类型
	protoOrders := make([]*v1.OrderInfo, len(orders))
	for i, o := range orders {
		protoOrders[i] = s.bizOrderToProto(o)
	}

	zap.S().Infof("Listed %d orders (total: %d) for user %d", len(orders), total, req.UserId)
	return &v1.ListOrdersResponse{
		Orders:   protoOrders,
		Total:    int32(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// --- 支付相关接口实现 ---

// ProcessPayment 实现了模拟支付处理的 RPC 方法。
// 实际系统中会与支付网关集成，这里仅模拟状态更新。
func (s *OrderService) ProcessPayment(ctx context.Context, req *v1.ProcessPaymentRequest) (*v1.PaymentResult, error) {
	zap.S().Infof("ProcessPayment request received for order ID: %d, amount: %d", req.OrderId, req.Amount)

	// 1. 参数校验
	if req.OrderId == 0 || req.Amount <= 0 || req.UserId == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payment request parameters")
	}

	// 2. 获取订单信息
	order, err := s.orderRepo.GetOrderByID(ctx, req.OrderId)
	if err != nil {
		zap.S().Errorf("failed to get order %d for payment processing: %v", req.OrderId, err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve order")
	}
	if order == nil {
		zap.S().Warnf("order %d not found for payment processing", req.OrderId)
		return nil, status.Errorf(codes.NotFound, "order not found")
	}

	// 3. 权限校验
	if order.UserID != req.UserId {
		zap.S().Warnf("user %d attempted to pay for order %d belonging to user %d", req.UserId, req.OrderId, order.UserID)
		return nil, status.Errorf(codes.PermissionDenied, "access denied")
	}

	// 4. 状态校验：只有待支付的订单才能进行支付
	if order.Status != model.PendingPayment || order.PaymentStatus != model.Unpaid {
		zap.S().Warnf("order %d is not in pending payment status (current status: %s, payment status: %s)",
			req.OrderId, order.Status.String(), order.PaymentStatus.String())
		return nil, status.Errorf(codes.FailedPrecondition, "order is not in pending payment status")
	}

	// 5. 金额校验
	if order.ActualAmount != req.Amount {
		zap.S().Warnf("payment amount mismatch for order %d: expected %d, got %d", req.OrderId, order.ActualAmount, req.Amount)
		return nil, status.Errorf(codes.InvalidArgument, "payment amount mismatch")
	}

	// 6. 模拟支付成功
	now := time.Now()
	order.Status = model.Paid
	order.PaymentStatus = model.Success
	order.PaidAt = &now
	order.PaymentMethod = req.PaymentMethod

	// 7. 记录订单日志
	logEntry := &model.OrderLog{
		OrderID:   req.OrderId,
		Operator:  fmt.Sprintf("User-%d", req.UserId),
		Action:    "Payment Processed",
		OldStatus: model.PendingPayment.String(),
		NewStatus: model.Paid.String(),
		Remark:    fmt.Sprintf("Payment successful via %s, transaction ID: %s", req.PaymentMethod, req.TransactionId),
	}
	if _, err := s.orderLogRepo.CreateOrderLog(ctx, logEntry); err != nil {
		zap.S().Errorf("failed to create order log for payment %d: %v", req.OrderId, err)
	}

	// 8. 调用仓库层更新订单
	updatedOrder, err := s.orderRepo.UpdateOrder(ctx, order)
	if err != nil {
		zap.S().Errorf("failed to update order %d after payment: %v", req.OrderId, err)
		return nil, status.Errorf(codes.Internal, "failed to update order after payment")
	}

	// 9. 扣减库存 (此处仅为模拟，实际应调用 InventoryService)
	// for _, item := range updatedOrder.Items {
	// 	_ = s.inventoryClient.DeductStock(ctx, item.SkuID, item.Quantity)
	// }

	zap.S().Infof("Order %d payment processed successfully", req.OrderId)
	return &v1.PaymentResult{
		OrderId:       updatedOrder.ID,
		TransactionId: req.TransactionId,
		Status:        v1.PaymentStatus_SUCCESS,
		Message:       "Payment successful",
		PaidAt:        timestamppb.New(*updatedOrder.PaidAt),
	}, nil
}

// RequestRefund 实现了申请退款的 RPC 方法。
func (s *OrderService) RequestRefund(ctx context.Context, req *v1.RequestRefundRequest) (*v1.OrderInfo, error) {
	zap.S().Infof("RequestRefund request received for order ID: %d, user ID: %d", req.OrderId, req.UserId)

	// 1. 参数校验
	if req.OrderId == 0 || req.UserId == 0 || req.RefundAmount <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid refund request parameters")
	}

	// 2. 获取订单信息
	order, err := s.orderRepo.GetOrderByID(ctx, req.OrderId)
	if err != nil {
		zap.S().Errorf("failed to get order %d for refund request: %v", req.OrderId, err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve order")
	}
	if order == nil {
		zap.S().Warnf("order %d not found for refund request", req.OrderId)
		return nil, status.Errorf(codes.NotFound, "order not found")
	}

	// 3. 权限校验
	if order.UserID != req.UserId {
		zap.S().Warnf("user %d attempted to request refund for order %d belonging to user %d", req.UserId, req.OrderId, order.UserID)
		return nil, status.Errorf(codes.PermissionDenied, "access denied")
	}

	// 4. 状态校验：只有已支付或已发货的订单才能申请退款
	if order.Status != model.Paid && order.Status != model.Shipped && order.Status != model.Delivered && order.Status != model.Completed {
		zap.S().Warnf("order %d in status %s cannot request refund", req.OrderId, order.Status.String())
		return nil, status.Errorf(codes.FailedPrecondition, "order in status %s cannot request refund", order.Status.String())
	}

	// 5. 金额校验
	if req.RefundAmount > order.ActualAmount {
		zap.S().Warnf("refund amount %d exceeds actual amount %d for order %d", req.RefundAmount, order.ActualAmount, req.OrderId)
		return nil, status.Errorf(codes.InvalidArgument, "refund amount exceeds actual payment")
	}

	// 6. 更新订单支付状态为退款中，订单状态不变或根据业务逻辑调整
	order.PaymentStatus = model.Refunding

	// 7. 记录订单日志
	logEntry := &model.OrderLog{
		OrderID:   req.OrderId,
		Operator:  fmt.Sprintf("User-%d", req.UserId),
		Action:    "Refund Requested",
		OldStatus: order.Status.String(),
		NewStatus: order.Status.String(), // 订单主状态可能不变
		Remark:    fmt.Sprintf("Refund amount: %d, Reason: %s", req.RefundAmount, req.Reason),
	}
	if _, err := s.orderLogRepo.CreateOrderLog(ctx, logEntry); err != nil {
		zap.S().Errorf("failed to create order log for refund request %d: %v", req.OrderId, err)
	}

	// 8. 调用仓库层更新订单
	updatedOrder, err := s.orderRepo.UpdateOrder(ctx, order)
	if err != nil {
		zap.S().Errorf("failed to update order %d after refund request: %v", req.OrderId, err)
		return nil, status.Errorf(codes.Internal, "failed to update order after refund request")
	}

	zap.S().Infof("Refund requested for order %d by user %d", req.OrderId, req.UserId)
	return s.bizOrderToProto(updatedOrder), nil
}

// --- 内部接口实现 ---

// GetOrderItemsByOrderID 实现了根据订单ID获取所有订单项的 RPC 方法。
func (s *OrderService) GetOrderItemsByOrderID(ctx context.Context, req *v1.GetOrderItemsByOrderIDRequest) (*v1.GetOrderItemsByOrderIDResponse, error) {
	zap.S().Infof("GetOrderItemsByOrderID request received for order ID: %d", req.OrderId)

	// 1. 参数校验
	if req.OrderId == 0 {
		zap.S().Warn("GetOrderItemsByOrderID request with zero order ID")
		return nil, status.Errorf(codes.InvalidArgument, "order ID cannot be zero")
	}

	// 2. 调用仓库层获取订单项
	items, err := s.orderItemRepo.GetOrderItemsByOrderID(ctx, req.OrderId)
	if err != nil {
		zap.S().Errorf("failed to get order items for order %d: %v", req.OrderId, err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve order items")
	}

	// 3. 转换为Protobuf类型
	protoItems := make([]*v1.OrderItem, len(items))
	for i, item := range items {
		protoItems[i] = s.bizOrderItemToProto(item)
	}

	zap.S().Infof("Retrieved %d order items for order %d", len(items), req.OrderId)
	return &v1.GetOrderItemsByOrderIDResponse{
		Items: protoItems,
	}, nil
}

// UpdateOrderShippingStatus 实现了更新订单配送状态的 RPC 方法。
func (s *OrderService) UpdateOrderShippingStatus(ctx context.Context, req *v1.UpdateOrderShippingStatusRequest) (*v1.OrderInfo, error) {
	zap.S().Infof("UpdateOrderShippingStatus request received for order ID: %d, new shipping status: %s", req.OrderId, req.NewShippingStatus.String())

	// 1. 参数校验
	if req.OrderId == 0 || req.NewShippingStatus == v1.ShippingStatus_SHIPPING_STATUS_UNSPECIFIED {
		return nil, status.Errorf(codes.InvalidArgument, "invalid shipping status update parameters")
	}

	// 2. 获取订单信息
	order, err := s.orderRepo.GetOrderByID(ctx, req.OrderId)
	if err != nil {
		zap.S().Errorf("failed to get order %d for shipping status update: %v", req.OrderId, err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve order")
	}
	if order == nil {
		zap.S().Warnf("order %d not found for shipping status update", req.OrderId)
		return nil, status.Errorf(codes.NotFound, "order not found")
	}

	oldShippingStatus := order.ShippingStatus
	newShippingStatus := model.ShippingStatus(req.NewShippingStatus)

	// 3. 配送状态流转校验 (简化)
	if !isValidShippingStatusTransition(oldShippingStatus, newShippingStatus) {
		zap.S().Warnf("invalid shipping status transition from %s to %s for order %d", oldShippingStatus.String(), newShippingStatus.String(), req.OrderId)
		return nil, status.Errorf(codes.FailedPrecondition, "invalid shipping status transition")
	}

	// 4. 更新订单配送状态和相关时间戳
	order.ShippingStatus = newShippingStatus
	switch newShippingStatus {
	case model.ShippingShipped:
		now := time.Now()
		order.ShippedAt = &now
	case model.ShippingDelivered:
		now := time.Now()
		order.DeliveredAt = &now
		// 如果已送达，可以将订单主状态更新为已送达
		if order.Status == model.Paid {
			order.Status = model.Delivered
		}
	}

	// 5. 记录订单日志
	logEntry := &model.OrderLog{
		OrderID:   req.OrderId,
		Operator:  req.Operator,
		Action:    fmt.Sprintf("Shipping Status Updated to %s", newShippingStatus.String()),
		OldStatus: oldShippingStatus.String(),
		NewStatus: newShippingStatus.String(),
		Remark:    fmt.Sprintf("Tracking: %s, Company: %s", req.TrackingNumber, req.LogisticsCompany),
	}
	if _, err := s.orderLogRepo.CreateOrderLog(ctx, logEntry); err != nil {
		zap.S().Errorf("failed to create order log for shipping status update %d: %v", req.OrderId, err)
	}

	// 6. 调用仓库层更新订单
	updatedOrder, err := s.orderRepo.UpdateOrder(ctx, order)
	if err != nil {
		zap.S().Errorf("failed to update order %d shipping status to %s: %v", req.OrderId, newShippingStatus.String(), err)
		return nil, status.Errorf(codes.Internal, "failed to update order shipping status")
	}

	zap.S().Infof("Order %d shipping status updated from %s to %s successfully", req.OrderId, oldShippingStatus.String(), newShippingStatus.String())
	return s.bizOrderToProto(updatedOrder), nil
}

// --- 辅助函数：模型转换 ---

// bizOrderToProto 将 model.Order 领域模型转换为 v1.OrderInfo API 模型。
func (s *OrderService) bizOrderToProto(o *model.Order) *v1.OrderInfo {
	if o == nil {
		return nil
	}

	protoItems := make([]*v1.OrderItem, len(o.Items))
	for i, item := range o.Items {
		protoItems[i] = s.bizOrderItemToProto(item)
	}

	protoLogs := make([]*v1.OrderLog, len(o.Logs))
	for i, log := range o.Logs {
		protoLogs[i] = s.bizOrderLogToProto(log)
	}

	return &v1.OrderInfo{
		Id:             o.ID,
		OrderNo:        o.OrderNo,
		UserId:         o.UserID,
		Status:         v1.OrderStatus(o.Status),
		PaymentStatus:  v1.PaymentStatus(o.PaymentStatus),
		ShippingStatus: v1.ShippingStatus(o.ShippingStatus),
		TotalAmount:    o.TotalAmount,
		ActualAmount:   o.ActualAmount,
		ShippingFee:    o.ShippingFee,
		DiscountAmount: o.DiscountAmount,
		PaymentMethod:  o.PaymentMethod,
		CreatedAt:      timestamppb.New(o.CreatedAt),
		UpdatedAt:      timestamppb.New(o.UpdatedAt),
		PaidAt:         timestamppb.New(o.PaidAt.Add(0)), // 处理指针类型
		ShippedAt:      timestamppb.New(o.ShippedAt.Add(0)),
		DeliveredAt:    timestamppb.New(o.DeliveredAt.Add(0)),
		CompletedAt:    timestamppb.New(o.CompletedAt.Add(0)),
		CancelledAt:    timestamppb.New(o.CancelledAt.Add(0)),
		Remark:         o.Remark,
		ShippingAddress: s.bizShippingAddressToProto(&o.ShippingAddress),
		Items:           protoItems,
		Logs:            protoLogs,
	}
}

// bizOrderItemToProto 将 model.OrderItem 领域模型转换为 v1.OrderItem API 模型。
func (s *OrderService) bizOrderItemToProto(item *model.OrderItem) *v1.OrderItem {
	if item == nil {
		return nil
	}
	return &v1.OrderItem{
		Id:              item.ID,
		OrderId:         item.OrderID,
		ProductId:       item.ProductID,
		SkuId:           item.SkuID,
		ProductName:     item.ProductName,
		SkuName:         item.SkuName,
		ProductImageUrl: item.ProductImageURL,
		Price:           item.Price,
		Quantity:        item.Quantity,
		TotalPrice:      item.TotalPrice,
	}
}

// bizShippingAddressToProto 将 model.ShippingAddress 领域模型转换为 v1.ShippingAddress API 模型。
func (s *OrderService) bizShippingAddressToProto(addr *model.ShippingAddress) *v1.ShippingAddress {
	if addr == nil {
		return nil
	}
	return &v1.ShippingAddress{
		RecipientName:   addr.RecipientName,
		PhoneNumber:     addr.PhoneNumber,
		Province:        addr.Province,
		City:            addr.City,
		District:        addr.District,
		DetailedAddress: addr.DetailedAddress,
		PostalCode:      addr.PostalCode,
	}
}

// bizOrderLogToProto 将 model.OrderLog 领域模型转换为 v1.OrderLog API 模型。
func (s *OrderService) bizOrderLogToProto(log *model.OrderLog) *v1.OrderLog {
	if log == nil {
		return nil
	}
	return &v1.OrderLog{
		Id:        log.ID,
		OrderId:   log.OrderID,
		Operator:  log.Operator,
		Action:    log.Action,
		OldStatus: log.OldStatus,
		NewStatus: log.NewStatus,
		CreatedAt: timestamppb.New(log.CreatedAt),
	}
}

// --- 内部辅助函数 ---

// isValidStatusTransition 校验订单状态流转是否合法。
// 这是一个简化的状态机，实际业务中可能更复杂。
func isValidStatusTransition(oldStatus, newStatus model.OrderStatus) bool {
	switch oldStatus {
	case model.PendingPayment:
		return newStatus == model.Paid || newStatus == model.Cancelled || newStatus == model.Closed
	case model.Paid:
		return newStatus == model.Shipped || newStatus == model.Cancelled || newStatus == model.RefundRequested
	case model.Shipped:
		return newStatus == model.Delivered || newStatus == model.RefundRequested
	case model.Delivered:
		return newStatus == model.Completed || newStatus == model.RefundRequested
	case model.Completed:
		return newStatus == model.RefundRequested // 完成后仍可退款
	case model.Cancelled, model.Refunded, model.Closed:
		return false // 最终状态，不可再流转
	case model.RefundRequested:
		return newStatus == model.Refunded || newStatus == model.Paid // 退款申请后可能退款成功或驳回回到已支付
	}
	return false
}

// isValidShippingStatusTransition 校验配送状态流转是否合法。
func isValidShippingStatusTransition(oldStatus, newStatus model.ShippingStatus) bool {
	switch oldStatus {
	case model.PendingShipment:
		return newStatus == model.ShippingShipped
	case model.ShippingShipped, model.InTransit:
		return newStatus == model.InTransit || newStatus == model.ShippingDelivered || newStatus == model.Exception
	case model.ShippingDelivered, model.Exception:
		return false // 最终状态
	}
	return false
}