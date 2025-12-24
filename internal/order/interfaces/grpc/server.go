package grpc

import (
	"context"
	"fmt"
	"strconv" // 导入字符串转换工具。

	pb "github.com/wyfcoding/ecommerce/goapi/order/v1"         // 导入订单模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/order/application" // 导入订单模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/order/domain"

	// 导入订单模块的领域实体。
	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 OrderService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedOrderServer                           // 嵌入生成的UnimplementedOrderServer，确保前向兼容性。
	app                         *application.OrderService // 依赖Order应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 Order gRPC 服务端实例。
func NewServer(app *application.OrderService) *Server {
	return &Server{app: app}
}

// CreateOrder 处理创建订单的gRPC请求。
// req: 包含用户ID、商品列表和收货地址的请求体。
// 返回created successfully的订单信息响应和可能发生的gRPC错误。
func (s *Server) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.OrderInfo, error) {
	// 将Proto的OrderItem列表转换为领域实体所需的 OrderItem 列表。
	items := make([]*domain.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = &domain.OrderItem{
			ProductID: item.ProductId,
			SkuID:     item.SkuId,
			Quantity:  item.Quantity,
			// 注意：Proto请求中的OrderItem可能只包含ProductId, SkuId, Quantity。
			// ProductName, SkuName, ProductImageURL, Price, TotalPrice等字段需要：
			// 1. 在此接口层通过调用Product服务来填充。
			// 2. 在应用服务层（application.OrderService）处理时填充。
			// 3. 假设这些信息在创建订单时暂时不需要，或有其他机制补充。
			// 当前实现直接传递，这意味着应用服务层需要处理这些缺失的信息（例如，通过查询商品服务）。
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

	// 调用应用服务层创建订单。
	order, err := s.app.CreateOrder(ctx, req.UserId, items, shippingAddr)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create order: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return s.toProto(order), nil
}

// GetOrderByID 处理根据订单ID获取订单信息的gRPC请求。
// req: 包含订单ID的请求体。
// 返回订单信息响应和可能发生的gRPC错误。
func (s *Server) GetOrderByID(ctx context.Context, req *pb.GetOrderByIDRequest) (*pb.OrderInfo, error) {
	order, err := s.app.GetOrder(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get order: %v", err))
	}
	if order == nil {
		return nil, status.Error(codes.NotFound, "order not found")
	}
	// TODO: 如果需要，此处可以添加权限检查，确保只有订单所属用户或管理员才能查看订单。

	return s.toProto(order), nil
}

// UpdateOrderStatus 处理更新订单状态的gRPC请求。
// 此方法通过映射Proto的状态值到应用服务层的具体业务方法来实现订单状态流转。
// req: 包含订单ID、新状态、操作人、备注的请求体。
// 返回更新后的订单信息响应和可能发生的gRPC错误。
func (s *Server) UpdateOrderStatus(ctx context.Context, req *pb.UpdateOrderStatusRequest) (*pb.OrderInfo, error) {
	var err error
	// 根据请求中的新状态调用应用服务层的对应方法。
	switch req.NewStatus {
	case pb.OrderStatus_PAID:
		err = s.app.PayOrder(ctx, req.Id, "Manual/Admin") // 假设通过此API进行的支付操作是手动或管理员触发。
	case pb.OrderStatus_SHIPPED:
		err = s.app.ShipOrder(ctx, req.Id, req.Operator)
	case pb.OrderStatus_DELIVERED:
		err = s.app.DeliverOrder(ctx, req.Id, req.Operator)
	case pb.OrderStatus_COMPLETED:
		err = s.app.CompleteOrder(ctx, req.Id, req.Operator)
	case pb.OrderStatus_CANCELLED:
		err = s.app.CancelOrder(ctx, req.Id, req.Operator, req.Remark)
	default:
		// 对于不支持通过此API直接转换的状态，返回InvalidArgument错误。
		return nil, status.Error(codes.InvalidArgument, "unsupported status transition via this API")
	}

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update order status: %v", err))
	}

	// 状态更新成功后，重新获取订单详情并返回。
	return s.GetOrderByID(ctx, &pb.GetOrderByIDRequest{Id: req.Id})
}

// CancelOrder 处理取消订单的gRPC请求。
// req: 包含订单ID、用户ID和取消原因的请求体。
// 返回取消后的订单信息响应和可能发生的gRPC错误。
func (s *Server) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.OrderInfo, error) {
	// 调用应用服务层取消订单。操作人使用用户ID的字符串表示。
	err := s.app.CancelOrder(ctx, req.Id, strconv.FormatUint(req.UserId, 10), req.Reason)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to cancel order: %v", err))
	}
	// 取消成功后，重新获取订单详情并返回。
	return s.GetOrderByID(ctx, &pb.GetOrderByIDRequest{Id: req.Id})
}

// ListOrders 处理列出订单的gRPC请求。
// req: 包含用户ID、状态和分页参数的请求体。
// 返回订单列表响应和可能发生的gRPC错误。
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
// req: 包含订单ID和支付方式的请求体。
// 返回支付结果响应和可能发生的gRPC错误。
func (s *Server) ProcessPayment(ctx context.Context, req *pb.ProcessPaymentRequest) (*pb.PaymentResult, error) {
	err := s.app.PayOrder(ctx, req.OrderId, req.PaymentMethod)
	if err != nil {
		// 支付失败时返回相应的错误状态和消息。
		return &pb.PaymentResult{
			OrderId: req.OrderId,
			Status:  pb.PaymentStatus_FAILED,
			Message: err.Error(),
		}, nil
	}

	// 支付成功时返回成功结果。
	return &pb.PaymentResult{
		OrderId:       req.OrderId,
		TransactionId: "mock-txn-" + strconv.FormatUint(req.OrderId, 10), // 模拟交易ID。
		Status:        pb.PaymentStatus_SUCCESS,
		PaidAt:        timestamppb.Now(),
	}, nil
}

// RequestRefund 处理退款请求的gRPC请求。
// req: 包含订单ID的请求体。
// 目前未实现，返回 Unimplemented 状态码。
func (s *Server) RequestRefund(ctx context.Context, req *pb.RequestRefundRequest) (*pb.OrderInfo, error) {
	// TODO: 实现退款请求逻辑，调用应用服务层的退款方法。
	return nil, status.Error(codes.Unimplemented, "RequestRefund not implemented")
}

// GetOrderItemsByOrderID 处理根据订单ID获取订单项列表的gRPC请求。
// req: 包含订单ID的请求体。
// 返回订单项列表响应和可能发生的gRPC错误。
func (s *Server) GetOrderItemsByOrderID(ctx context.Context, req *pb.GetOrderItemsByOrderIDRequest) (*pb.GetOrderItemsByOrderIDResponse, error) {
	order, err := s.app.GetOrder(ctx, req.OrderId)
	if err != nil {
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
// 此方法通过映射Proto的配送状态值到应用服务层的具体业务方法来实现。
// req: 包含订单ID、新配送状态和操作人的请求体。
// 返回更新后的订单信息响应和可能发生的gRPC错误。
func (s *Server) UpdateOrderShippingStatus(ctx context.Context, req *pb.UpdateOrderShippingStatusRequest) (*pb.OrderInfo, error) {
	var err error
	// 根据请求中的新配送状态调用应用服务层的对应方法。
	switch req.NewShippingStatus {
	case pb.ShippingStatus_SHIPPING_SHIPPED:
		err = s.app.ShipOrder(ctx, req.OrderId, req.Operator)
	case pb.ShippingStatus_SHIPPING_DELIVERED:
		err = s.app.DeliverOrder(ctx, req.OrderId, req.Operator)
	default:
		// 对于不支持通过此API直接转换的配送状态，返回Unimplemented错误。
		return nil, status.Error(codes.Unimplemented, "shipping status not mapped")
	}

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update order shipping status: %v", err))
	}

	// 配送状态更新成功后，重新获取订单详情并返回。
	return s.GetOrderByID(ctx, &pb.GetOrderByIDRequest{Id: req.OrderId})
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
