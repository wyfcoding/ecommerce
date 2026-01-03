package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"strconv" // 导入字符串转换工具。
	"time"

	pb "github.com/wyfcoding/ecommerce/goapi/order/v1"          // 导入订单模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/order/application" // 导入订单模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/order/domain"

	// 导入订单模块的领域实体。
	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 Order 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedOrderServiceServer                           // 嵌入生成的UnimplementedOrderServiceServer，确保前向兼容性。
	app                                *application.OrderService // 依赖Order应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 Order gRPC 服务端实例。
func NewServer(app *application.OrderService) *Server {
	return &Server{app: app}
}

// CreateOrder 处理创建订单的gRPC请求。
// req: 包含用户ID、商品列表和收货地址的请求体。
// 返回created successfully的订单信息响应和可能发生的gRPC错误。
func (s *Server) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.OrderInfo, error) {
	start := time.Now()
	slog.Info("gRPC CreateOrder received", "user_id", req.UserId, "items_count", len(req.Items))

	// 将Proto的OrderItem列表转换为领域实体所需的 OrderItem 列表。
	items := make([]*domain.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = &domain.OrderItem{
			ProductID: item.ProductId,
			SkuID:     item.SkuId,
			Quantity:  item.Quantity,
		}
	}

	// 将Proto的ShippingAddress转换为领域实体所需的 ShippingAddress 值对象。
	shippingAddr := &domain.ShippingAddress{
		RecipientName:   req.ShippingAddress.RecipientName,
		PhoneNumber:     req.ShippingAddress.PhoneNumber,
		Province:        req.ShippingAddress.Province,
		City:            req.ShippingAddress.City,
		District:        req.ShippingAddress.District,
		DetailedAddress: req.ShippingAddress.DetailedAddress,
		PostalCode:      req.ShippingAddress.PostalCode,
	}

	// 获取优惠券码
	var couponCode string
	if req.CouponCode != nil {
		couponCode = req.CouponCode.Value
	}

	// 调用应用服务层创建订单。
	order, err := s.app.CreateOrder(ctx, req.UserId, items, shippingAddr, couponCode)
	if err != nil {
		slog.Error("gRPC CreateOrder failed", "user_id", req.UserId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create order: %v", err))
	}

	slog.Info("gRPC CreateOrder successful", "order_id", order.ID, "user_id", req.UserId, "duration", time.Since(start))
	// 将领域实体转换为protobuf响应格式。
	return s.toProto(order), nil
}

// GetOrderByID 处理根据订单ID获取订单信息的gRPC请求。
func (s *Server) GetOrderByID(ctx context.Context, req *pb.GetOrderByIDRequest) (*pb.OrderInfo, error) {
	order, err := s.app.GetOrder(ctx, req.UserId, req.Id)
	if err != nil {
		slog.Error("gRPC GetOrderByID failed", "id", req.Id, "user_id", req.UserId, "error", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get order: %v", err))
	}
	if order == nil {
		return nil, status.Error(codes.NotFound, "order not found")
	}

	return s.toProto(order), nil
}

// UpdateOrderStatus 处理更新订单状态的gRPC请求。
func (s *Server) UpdateOrderStatus(ctx context.Context, req *pb.UpdateOrderStatusRequest) (*pb.OrderInfo, error) {
	start := time.Now()
	slog.Info("gRPC UpdateOrderStatus received", "id", req.Id, "user_id", req.UserId, "new_status", req.NewStatus)

	var err error
	// 根据请求中的新状态调用应用服务层的对应方法。
	switch req.NewStatus {
	case pb.OrderStatus_PAID:
		err = s.app.PayOrder(ctx, req.UserId, req.Id, "Manual/Admin")
	case pb.OrderStatus_SHIPPED:
		err = s.app.ShipOrder(ctx, req.UserId, req.Id, req.Operator)
	case pb.OrderStatus_DELIVERED:
		err = s.app.DeliverOrder(ctx, req.UserId, req.Id, req.Operator)
	case pb.OrderStatus_COMPLETED:
		err = s.app.CompleteOrder(ctx, req.UserId, req.Id, req.Operator)
	case pb.OrderStatus_CANCELLED:
		err = s.app.CancelOrder(ctx, req.UserId, req.Id, req.Operator, req.Remark)
	default:
		return nil, status.Error(codes.InvalidArgument, "unsupported status transition via this API")
	}

	if err != nil {
		slog.Error("gRPC UpdateOrderStatus failed", "id", req.Id, "user_id", req.UserId, "action", req.NewStatus, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update order status: %v", err))
	}

	slog.Info("gRPC UpdateOrderStatus successful", "id", req.Id, "user_id", req.UserId, "duration", time.Since(start))
	// 状态更新成功后，重新获取订单详情并返回。
	return s.GetOrderByID(ctx, &pb.GetOrderByIDRequest{Id: req.Id, UserId: req.UserId})
}

// CancelOrder 处理取消订单的gRPC请求。
func (s *Server) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.OrderInfo, error) {
	start := time.Now()
	slog.Info("gRPC CancelOrder received", "id", req.Id, "user_id", req.UserId)

	// 调用应用服务层取消订单。操作人使用用户ID的字符串表示。
	err := s.app.CancelOrder(ctx, req.UserId, req.Id, strconv.FormatUint(req.UserId, 10), req.Reason)
	if err != nil {
		slog.Error("gRPC CancelOrder failed", "id", req.Id, "user_id", req.UserId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to cancel order: %v", err))
	}

	slog.Info("gRPC CancelOrder successful", "id", req.Id, "user_id", req.UserId, "duration", time.Since(start))
	// 取消成功后，重新获取订单详情并返回。
	return s.GetOrderByID(ctx, &pb.GetOrderByIDRequest{Id: req.Id, UserId: req.UserId})
}

// ListOrders 处理列出订单的gRPC请求。
func (s *Server) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	// 获取分页参数。
	page := max(int(req.Page), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 根据Proto的状态值构建过滤器。
	var statusPtr *int
	if req.Status != pb.OrderStatus_ORDER_STATUS_UNSPECIFIED { // 检查是否提供了特定状态。
		st := int(req.Status)
		statusPtr = &st // 将Proto状态转换为int类型指针。
	}

	// 调用应用服务层获取订单列表。
	orders, total, err := s.app.ListOrders(ctx, req.UserId, statusPtr, page, pageSize)
	if err != nil {
		slog.Error("gRPC ListOrders failed", "user_id", req.UserId, "error", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list orders: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbOrders := make([]*pb.OrderInfo, len(orders))
	for i, o := range orders {
		pbOrders[i] = s.toProto(o)
	}

	return &pb.ListOrdersResponse{
		Orders:   pbOrders,
		Total:    int32(total),    // 总记录数。
		Page:     int32(page),     // 当前页码。
		PageSize: int32(pageSize), // 每页大小。
	}, nil
}

// ProcessPayment 处理支付请求的gRPC请求。
func (s *Server) ProcessPayment(ctx context.Context, req *pb.ProcessPaymentRequest) (*pb.PaymentResult, error) {
	start := time.Now()
	slog.Info("gRPC ProcessPayment received", "order_id", req.OrderId, "user_id", req.UserId, "method", req.PaymentMethod)

	err := s.app.PayOrder(ctx, req.UserId, req.OrderId, req.PaymentMethod)
	if err != nil {
		slog.Error("gRPC ProcessPayment failed", "order_id", req.OrderId, "user_id", req.UserId, "error", err, "duration", time.Since(start))
		// 支付失败时返回相应的错误状态和消息。
		return &pb.PaymentResult{
			OrderId: req.OrderId,
			Status:  pb.PaymentStatus_FAILED,
			Message: err.Error(),
		}, nil
	}

	slog.Info("gRPC ProcessPayment successful", "order_id", req.OrderId, "user_id", req.UserId, "duration", time.Since(start))
	// 支付成功时返回成功结果。
	return &pb.PaymentResult{
		OrderId:       req.OrderId,
		TransactionId: "mock-txn-" + strconv.FormatUint(req.OrderId, 10), // 模拟交易ID。
		Status:        pb.PaymentStatus_SUCCESS,
		PaidAt:        timestamppb.Now(),
	}, nil
}

// RequestRefund 处理退款请求的gRPC请求。
func (s *Server) RequestRefund(ctx context.Context, req *pb.RequestRefundRequest) (*pb.OrderInfo, error) {
	slog.Info("gRPC RequestRefund received", "order_id", req.OrderId, "user_id", req.UserId, "amount", req.RefundAmount)

	// 调用应用层处理退款申请逻辑
	err := s.app.CancelOrder(ctx, req.UserId, req.OrderId, "User", req.Reason)
	if err != nil {
		slog.Error("gRPC RequestRefund failed", "order_id", req.OrderId, "error", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return s.GetOrderByID(ctx, &pb.GetOrderByIDRequest{Id: req.OrderId, UserId: req.UserId})
}

// GetOrderItemsByOrderID 处理根据订单ID获取订单项列表的gRPC请求。
func (s *Server) GetOrderItemsByOrderID(ctx context.Context, req *pb.GetOrderItemsByOrderIDRequest) (*pb.GetOrderItemsByOrderIDResponse, error) {
	// TODO: GetOrderItemsByOrderIDRequest should include user_id for sharded lookup
	order, err := s.app.GetOrder(ctx, 0, req.OrderId)
	if err != nil {
		slog.Error("gRPC GetOrderItemsByOrderID failed", "order_id", req.OrderId, "error", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get order: %v", err))
	}
	if order == nil {
		return nil, status.Error(codes.NotFound, "order not found")
	}

	// 将领域实体中的 OrderItem 列表转换为protobuf的 OrderItem 列表。
	items := make([]*pb.OrderItem, len(order.Items))
	for i, item := range order.Items {
		items[i] = s.itemToProto(item)
	}

	return &pb.GetOrderItemsByOrderIDResponse{
		Items: items,
	}, nil
}

// UpdateOrderShippingStatus 处理更新订单配送状态的gRPC请求。
func (s *Server) UpdateOrderShippingStatus(ctx context.Context, req *pb.UpdateOrderShippingStatusRequest) (*pb.OrderInfo, error) {
	start := time.Now()
	slog.Info("gRPC UpdateOrderShippingStatus received", "order_id", req.OrderId, "new_status", req.NewShippingStatus)

	var err error
	// 根据请求中的新配送状态调用应用服务层的对应方法。
	// TODO: UpdateOrderShippingStatusRequest should include user_id for sharded lookup
	switch req.NewShippingStatus {
	case pb.ShippingStatus_SHIPPING_SHIPPED:
		err = s.app.ShipOrder(ctx, 0, req.OrderId, req.Operator)
	case pb.ShippingStatus_SHIPPING_DELIVERED:
		err = s.app.DeliverOrder(ctx, 0, req.OrderId, req.Operator)
	default:
		return nil, status.Error(codes.Unimplemented, "shipping status not mapped")
	}

	if err != nil {
		slog.Error("gRPC UpdateOrderShippingStatus failed", "order_id", req.OrderId, "status", req.NewShippingStatus, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update order shipping status: %v", err))
	}

	slog.Info("gRPC UpdateOrderShippingStatus successful", "order_id", req.OrderId, "duration", time.Since(start))
	// 配送状态更新成功后，重新获取订单详情并返回。
	return s.GetOrderByID(ctx, &pb.GetOrderByIDRequest{Id: req.OrderId, UserId: 0})
}

// SagaConfirmOrder Saga 正向: 确认订单
func (s *Server) SagaConfirmOrder(ctx context.Context, req *pb.SagaOrderRequest) (*pb.SagaOrderResponse, error) {
	if err := s.app.SagaConfirmOrder(ctx, req.UserId, req.OrderId); err != nil {
		return nil, status.Errorf(codes.Internal, "SagaConfirmOrder failed: %v", err)
	}
	return &pb.SagaOrderResponse{Success: true}, nil
}

// SagaCancelOrder Saga 补偿: 取消订单
func (s *Server) SagaCancelOrder(ctx context.Context, req *pb.SagaOrderRequest) (*pb.SagaOrderResponse, error) {
	if err := s.app.SagaCancelOrder(ctx, req.UserId, req.OrderId, req.Reason); err != nil {
		return nil, status.Errorf(codes.Internal, "SagaCancelOrder failed: %v", err)
	}
	return &pb.SagaOrderResponse{Success: true}, nil
}

// toProto 是一个辅助函数，将领域层的 Order 实体转换为 protobuf 的 OrderInfo 消息。
func (s *Server) toProto(o *domain.Order) *pb.OrderInfo {
	if o == nil {
		return nil
	}
	// 转换订单项列表。
	items := make([]*pb.OrderItem, len(o.Items))
	for i, item := range o.Items {
		items[i] = s.itemToProto(item)
	}

	return &pb.OrderInfo{
		Id:           uint64(o.ID),                 // 订单ID。
		OrderNo:      o.OrderNo,                    // 订单编号。
		UserId:       o.UserID,                     // 用户ID。
		Status:       pb.OrderStatus(o.Status),     // 订单状态。
		TotalAmount:  o.TotalAmount,                // 订单总金额。
		ActualAmount: o.ActualAmount,               // 实际支付金额。
		CreatedAt:    timestamppb.New(o.CreatedAt), // 创建时间。
		UpdatedAt:    timestamppb.New(o.UpdatedAt), // 更新时间。
		Items:        items,                        // 订单项列表。
		ShippingAddress: &pb.ShippingAddress{ // 收货地址信息。
			RecipientName:   o.ShippingAddress.RecipientName,
			PhoneNumber:     o.ShippingAddress.PhoneNumber,
			Province:        o.ShippingAddress.Province,
			City:            o.ShippingAddress.City,
			District:        o.ShippingAddress.District,
			DetailedAddress: o.ShippingAddress.DetailedAddress,
			PostalCode:      o.ShippingAddress.PostalCode,
			Lat:             o.ShippingAddress.Lat,
			Lon:             o.ShippingAddress.Lon,
		},
	}
}

// itemToProto 是一个辅助函数，将领域层的 OrderItem 实体转换为 protobuf 的 OrderItem 消息。
func (s *Server) itemToProto(item *domain.OrderItem) *pb.OrderItem {
	if item == nil {
		return nil
	}
	return &pb.OrderItem{
		Id:          uint64(item.ID),                   // 订单项ID。
		OrderId:     item.OrderID,                      // 订单ID。
		ProductId:   item.ProductID,                    // 商品ID。
		SkuId:       item.SkuID,                        // SKU ID。
		ProductName: item.ProductName,                  // 商品名称。
		SkuName:     item.SkuName,                      // SKU名称。
		Price:       item.Price,                        // 单价。
		Quantity:    item.Quantity,                     // 数量。
		TotalPrice:  item.Price * int64(item.Quantity), // 总价。
	}
}
