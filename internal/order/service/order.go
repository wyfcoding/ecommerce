package service

import (
	"context"
	"strconv"
	"errors"

	v1 "ecommerce/api/order/v1"
	"ecommerce/internal/order/biz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type OrderService struct {
	v1.UnimplementedOrderServer
	uc *biz.OrderUsecase
}

// NewOrderService 是 OrderService 的构造函数。
func NewOrderService(uc *biz.OrderUsecase) *OrderService {
	return &OrderService{uc: uc}
}

// getUserIDFromContext 从 gRPC 上下文的 metadata 中提取用户ID。
func getUserIDFromContext(ctx context.Context) (uint64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Errorf(codes.Unauthenticated, "无法获取元数据")
	}
	// 兼容 gRPC-Gateway 在 HTTP 请求时注入的用户ID
	values := md.Get("x-md-global-user-id")
	if len(values) == 0 {
		// 兼容直接 gRPC 调用时注入的用户ID
		values = md.Get("x-user-id")
		if len(values) == 0 {
			return 0, status.Errorf(codes.Unauthenticated, "请求头中缺少 x-user-id 信息")
		}
	}
	userID, err := strconv.ParseUint(values[0], 10, 64)
	if err != nil {
		return 0, status.Errorf(codes.Unauthenticated, "x-user-id 格式无效")
	}
	return userID, nil
}

// bizOrderToProto 将 biz.Order 领域模型转换为 v1.OrderInfo API 模型。
func bizOrderToProto(order *biz.Order) *v1.OrderInfo {
	if order == nil {
		return nil
	}
	return &v1.OrderInfo{
		OrderId:       order.ID,
		UserId:        order.UserID,
		TotalAmount:   order.TotalAmount,
		PaymentAmount: order.PaymentAmount,
		Status:        order.Status,
	}
}

// CreateOrder 实现了创建订单的 RPC。
func (s *OrderService) CreateOrder(ctx context.Context, req *v1.CreateOrderRequest) (*v1.CreateOrderResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if len(req.Items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "items are required")
	}

	bizItems := make([]*biz.CreateOrderRequestItem, 0, len(req.Items))
	for _, item := range req.Items {
		bizItems = append(bizItems, &biz.CreateOrderRequestItem{
			SkuID:    item.SkuId,
			Quantity: item.Quantity,
		})
	}

	createdOrder, err := s.uc.CreateOrder(ctx, userID, bizItems, req.AddressId, req.Remark, req.ShippingAddress, req.PaymentAmount)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "创建订单失败: %v", err)
	}

	return &v1.CreateOrderResponse{
		OrderId: createdOrder.ID,
	}, nil
}

// GetOrderDetail 实现了获取订单详情的 RPC。
func (s *OrderService) GetOrderDetail(ctx context.Context, req *v1.GetOrderDetailRequest) (*v1.GetOrderDetailResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.OrderId == 0 {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}

	order, err := s.uc.GetOrder(ctx, req.OrderId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { // 假设 GetOrder 会返回 gorm.ErrRecordNotFound
			return nil, status.Errorf(codes.NotFound, "订单未找到")
		}
		return nil, status.Errorf(codes.Internal, "获取订单详情失败: %v", err)
	}

	// 检查订单是否属于当前用户
	if order.UserID != userID {
		return nil, status.Errorf(codes.PermissionDenied, "无权访问此订单")
	}

	// TODO: 获取订单商品项
	// 目前 biz.OrderUsecase 中没有获取 OrderItems 的方法，需要添加
	// 假设 s.uc.GetOrderItems(ctx, order.ID) 可以获取到 []*biz.OrderItem

	// 为了不阻塞流程，暂时返回一个简化的响应
	return &v1.GetOrderDetailResponse{
		Order: bizOrderToProto(order),
		Items: []*v1.OrderItem{}, // 暂时为空
	}, nil
}

// GetPaymentURL 实现了获取支付链接的 RPC。
func (s *OrderService) GetPaymentURL(ctx context.Context, req *v1.GetPaymentURLRequest) (*v1.GetPaymentURLResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.OrderId == 0 {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}

	// TODO: 实现获取支付链接的业务逻辑
	// 1. 根据 orderID 获取订单信息
	// 2. 检查订单状态，只有待支付的订单才能获取支付链接
	// 3. 调用支付服务生成支付链接
	// 4. 返回支付链接

	return nil, status.Errorf(codes.Unimplemented, "method GetPaymentURL not implemented")
}

// ProcessPaymentNotification 实现了处理支付回调的 RPC。
func (s *OrderService) ProcessPaymentNotification(ctx context.Context, req *v1.ProcessPaymentNotificationRequest) (*v1.ProcessPaymentNotificationResponse, error) {
	// TODO: 实现支付回调的业务逻辑
	// 1. 验证支付通知的合法性（签名等）
	// 2. 解析通知数据，获取订单ID、支付状态等信息
	// 3. 根据支付状态更新订单状态
	// 4. 处理库存、积分等后续业务逻辑

	return nil, status.Errorf(codes.Unimplemented, "method ProcessPaymentNotification not implemented")
}
