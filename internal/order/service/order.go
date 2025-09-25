package service

import (
	"context"
	"encoding/json"
	"fmt"

	v1 "ecommerce/api/order/v1"
	"ecommerce/internal/order/biz"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// OrderService 是订单服务的 gRPC 实现。
type OrderService struct {
	v1.UnimplementedOrderServer
	orderUsecase *biz.OrderUsecase
	log          *zap.SugaredLogger
}

// NewOrderService 创建一个新的 OrderService。
func NewOrderService(orderUsecase *biz.OrderUsecase, logger *zap.SugaredLogger) *OrderService {
	return &OrderService{
		orderUsecase: orderUsecase,
		log:          logger,
	}
}

// CreateOrder 实现创建订单 RPC。
func (s *OrderService) CreateOrder(ctx context.Context, req *v1.CreateOrderRequest) (*v1.CreateOrderResponse, error) {
	// 转换请求中的商品项
	reqItems := make([]*biz.CreateOrderRequestItem, len(req.Items))
	for i, item := range req.Items {
		reqItems[i] = &biz.CreateOrderRequestItem{
			SkuID:    item.SkuId,
			Quantity: item.Quantity,
		}
	}

	// 调用业务逻辑创建订单
	order, err := s.orderUsecase.CreateOrder(ctx, req.UserId, reqItems, req.ShippingAddress, req.PaymentAmount)
	if err != nil {
		s.log.Errorf("CreateOrder: failed to create order: %v", err)
		return nil, status.Errorf(codes.Internal, "创建订单失败: %v", err)
	}

	return &v1.CreateOrderResponse{OrderId: order.ID}, nil
}

// GetOrderDetail 实现获取订单详情 RPC。
func (s *OrderService) GetOrderDetail(ctx context.Context, req *v1.GetOrderDetailRequest) (*v1.GetOrderDetailResponse, error) {
	order, err := s.orderUsecase.GetOrder(ctx, req.OrderId)
	if err != nil {
		s.log.Errorf("GetOrderDetail: failed to get order %d: %v", req.OrderId, err)
		return nil, status.Errorf(codes.Internal, "获取订单详情失败: %v", err)
	}

	// 转换 OrderInfo
	orderInfo := &v1.OrderInfo{
		OrderId:      order.ID,
		UserId:       order.UserID,
		TotalAmount:  order.TotalAmount,
		PaymentAmount: order.PaymentAmount,
		Status:       int32(order.Status),
		// ... 其他字段
	}

	// 转换 OrderItem
	orderItems := make([]*v1.OrderItem, len(order.Items))
	for i, item := range order.Items {
		orderItems[i] = &v1.OrderItem{
			SkuId:        item.SkuID,
			ProductTitle: item.ProductTitle,
			ProductImage: item.ProductImage,
			Price:        item.Price,
			Quantity:     item.Quantity,
		}
	}

	return &v1.GetOrderDetailResponse{Order: orderInfo, Items: orderItems}, nil
}

// GetPaymentURL 实现获取支付链接 RPC。
func (s *OrderService) GetPaymentURL(ctx context.Context, req *v1.GetPaymentURLRequest) (*v1.GetPaymentURLResponse, error) {
	// TODO: 实现获取支付链接的业务逻辑
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

// ProcessPaymentNotification 实现处理支付异步通知 RPC。
func (s *OrderService) ProcessPaymentNotification(ctx context.Context, req *v1.ProcessPaymentNotificationRequest) (*v1.ProcessPaymentNotificationResponse, error) {
	// TODO: 实现处理支付异步通知的业务逻辑
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

// CreateOrderForFlashSale 为秒杀活动创建订单 RPC。
func (s *OrderService) CreateOrderForFlashSale(ctx context.Context, req *v1.CreateOrderForFlashSaleRequest) (*v1.CreateOrderForFlashSaleResponse, error) {
	orderID, err := s.orderUsecase.CreateOrderForFlashSale(ctx, req.UserId, req.ProductId, req.Quantity, req.Price)
	if err != nil {
		s.log.Errorf("CreateOrderForFlashSale: failed to create order: %v", err)
		return nil, status.Errorf(codes.Internal, "为秒杀创建订单失败: %v", err)
	}
	return &v1.CreateOrderForFlashSaleResponse{OrderId: orderID}, nil
}

// CompensateCreateOrder 补偿创建订单 RPC。
func (s *OrderService) CompensateCreateOrder(ctx context.Context, req *v1.CompensateCreateOrderRequest) (*v1.CompensateCreateOrderResponse, error) {
	err := s.orderUsecase.CompensateCreateOrder(ctx, req.OrderId)
	if err != nil {
		s.log.Errorf("CompensateCreateOrder: failed to compensate order %s: %v", req.OrderId, err)
		return nil, status.Errorf(codes.Internal, "补偿创建订单失败: %v", err)
	}
	return &v1.CompensateCreateOrderResponse{Success: true, Message: "订单补偿成功"}, nil
}
