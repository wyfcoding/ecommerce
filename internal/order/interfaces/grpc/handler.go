package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	pb "github.com/wyfcoding/ecommerce/goapi/order/v1"
	"github.com/wyfcoding/ecommerce/internal/order/application"
	"github.com/wyfcoding/ecommerce/internal/order/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server implements the Order gRPC service.
type Server struct {
	pb.UnimplementedOrderServiceServer
	cmdService   *application.OrderCommandService
	queryService *application.OrderQuery
}

// NewServer Creates a new Order gRPC server.
func NewServer(cmd *application.OrderCommandService, query *application.OrderQuery) *Server {
	return &Server{
		cmdService:   cmd,
		queryService: query,
	}
}

// CreateOrder handles order creation.
func (s *Server) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.OrderInfo, error) {
	start := time.Now()
	slog.Info("gRPC CreateOrder received", "user_id", req.UserId, "items_count", len(req.Items))

	// Convert items
	var items []*application.CreateOrderItemCommand
	for _, item := range req.Items {
		items = append(items, &application.CreateOrderItemCommand{
			ProductID: item.ProductId,
			SkuID:     item.SkuId,
			Quantity:  item.Quantity,
			Price:     item.Price, // Use price from request for verification, service will re-fetch for safety
		})
	}

	shippingAddr := &domain.ShippingAddress{
		RecipientName:   req.ShippingAddress.RecipientName,
		PhoneNumber:     req.ShippingAddress.PhoneNumber,
		Province:        req.ShippingAddress.Province,
		City:            req.ShippingAddress.City,
		District:        req.ShippingAddress.District,
		DetailedAddress: req.ShippingAddress.DetailedAddress,
		PostalCode:      req.ShippingAddress.PostalCode,
	}

	var couponCode string
	if req.CouponCode != nil {
		couponCode = req.CouponCode.Value
	}

	cmd := &application.CreateOrderCommand{
		UserID:          req.UserId,
		Items:           items,
		ShippingAddress: shippingAddr,
		CouponCode:      couponCode,
		ClientIP:        "127.0.0.1",
		DeviceID:        "unknown",
	}

	order, err := s.cmdService.CreateOrder(ctx, cmd)
	if err != nil {
		slog.Error("gRPC CreateOrder failed", "user_id", req.UserId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create order: %v", err))
	}

	slog.Info("gRPC CreateOrder successful", "order_id", order.ID, "user_id", req.UserId, "duration", time.Since(start))
	return s.toProto(order), nil
}

func (s *Server) GetOrderByID(ctx context.Context, req *pb.GetOrderByIDRequest) (*pb.OrderInfo, error) {
	order, err := s.queryService.GetOrder(ctx, req.UserId, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get order: %v", err))
	}
	if order == nil {
		return nil, status.Error(codes.NotFound, "order not found")
	}
	return s.toProto(order), nil
}

func (s *Server) UpdateOrderStatus(ctx context.Context, req *pb.UpdateOrderStatusRequest) (*pb.OrderInfo, error) {
	start := time.Now()
	slog.Info("gRPC UpdateOrderStatus received", "id", req.Id, "user_id", req.UserId, "new_status", req.NewStatus)

	var err error
	switch req.NewStatus {
	case pb.OrderStatus_PAID:
		err = s.cmdService.PayOrder(ctx, &application.PayOrderCommand{UserID: req.UserId, OrderID: req.Id, PaymentMethod: "Manual/Admin"})
	case pb.OrderStatus_SHIPPED:
		err = s.cmdService.ShipOrder(ctx, &application.ShipOrderCommand{UserID: req.UserId, OrderID: req.Id, Operator: req.Operator})
	case pb.OrderStatus_DELIVERED:
		err = s.cmdService.DeliverOrder(ctx, &application.DeliverOrderCommand{UserID: req.UserId, OrderID: req.Id, Operator: req.Operator})
	case pb.OrderStatus_COMPLETED:
		err = s.cmdService.CompleteOrder(ctx, &application.CompleteOrderCommand{UserID: req.UserId, OrderID: req.Id, Operator: req.Operator})
	case pb.OrderStatus_CANCELLED:
		err = s.cmdService.CancelOrder(ctx, &application.CancelOrderCommand{UserID: req.UserId, OrderID: req.Id, Operator: req.Operator, Reason: req.Remark})
	default:
		return nil, status.Error(codes.InvalidArgument, "unsupported status transition via this API")
	}

	if err != nil {
		slog.Error("gRPC UpdateOrderStatus failed", "id", req.Id, "user_id", req.UserId, "action", req.NewStatus, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update status: %v", err))
	}

	slog.Info("gRPC UpdateOrderStatus successful", "id", req.Id, "user_id", req.UserId, "duration", time.Since(start))
	return s.GetOrderByID(ctx, &pb.GetOrderByIDRequest{Id: req.Id, UserId: req.UserId})
}

func (s *Server) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.OrderInfo, error) {
	start := time.Now()
	slog.Info("gRPC CancelOrder received", "id", req.Id, "user_id", req.UserId)

	err := s.cmdService.CancelOrder(ctx, &application.CancelOrderCommand{
		UserID:   req.UserId,
		OrderID:  req.Id,
		Operator: strconv.FormatUint(req.UserId, 10),
		Reason:   req.Reason,
	})
	if err != nil {
		slog.Error("gRPC CancelOrder failed", "id", req.Id, "user_id", req.UserId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to cancel order: %v", err))
	}

	slog.Info("gRPC CancelOrder successful", "id", req.Id, "user_id", req.UserId, "duration", time.Since(start))
	return s.GetOrderByID(ctx, &pb.GetOrderByIDRequest{Id: req.Id, UserId: req.UserId})
}

func (s *Server) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	page := max(int(req.Page), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	var statusPtr *int
	if req.Status != pb.OrderStatus_ORDER_STATUS_UNSPECIFIED {
		st := int(req.Status)
		statusPtr = &st
	}

	orders, total, err := s.queryService.ListOrders(ctx, req.UserId, statusPtr, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list orders: %v", err))
	}

	pbOrders := make([]*pb.OrderInfo, len(orders))
	for i, o := range orders {
		pbOrders[i] = s.toProto(o)
	}

	return &pb.ListOrdersResponse{
		Orders:   pbOrders,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

func (s *Server) ProcessPayment(ctx context.Context, req *pb.ProcessPaymentRequest) (*pb.PaymentResult, error) {
	start := time.Now()
	slog.Info("gRPC ProcessPayment received", "order_id", req.OrderId, "user_id", req.UserId, "method", req.PaymentMethod)

	err := s.cmdService.PayOrder(ctx, &application.PayOrderCommand{
		UserID:        req.UserId,
		OrderID:       req.OrderId,
		PaymentMethod: req.PaymentMethod,
	})
	if err != nil {
		slog.Error("gRPC ProcessPayment failed", "order_id", req.OrderId, "user_id", req.UserId, "error", err, "duration", time.Since(start))
		return &pb.PaymentResult{OrderId: req.OrderId, Status: pb.PaymentStatus_FAILED, Message: err.Error()}, nil
	}
	slog.Info("gRPC ProcessPayment successful", "order_id", req.OrderId, "user_id", req.UserId, "duration", time.Since(start))
	return &pb.PaymentResult{
		OrderId:       req.OrderId,
		TransactionId: "mock-txn-" + strconv.FormatUint(req.OrderId, 10),
		Status:        pb.PaymentStatus_SUCCESS,
		PaidAt:        timestamppb.Now(),
	}, nil
}

func (s *Server) RequestRefund(ctx context.Context, req *pb.RequestRefundRequest) (*pb.OrderInfo, error) {
	err := s.cmdService.CancelOrder(ctx, &application.CancelOrderCommand{
		UserID:   req.UserId,
		OrderID:  req.OrderId,
		Operator: "User",
		Reason:   req.Reason,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return s.GetOrderByID(ctx, &pb.GetOrderByIDRequest{Id: req.OrderId, UserId: req.UserId})
}

func (s *Server) GetOrderItemsByOrderID(ctx context.Context, req *pb.GetOrderItemsByOrderIDRequest) (*pb.GetOrderItemsByOrderIDResponse, error) {
	// Support sharding using UserId
	order, err := s.queryService.GetOrder(ctx, req.UserId, req.OrderId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get order: %v", err))
	}
	if order == nil {
		return nil, status.Error(codes.NotFound, "order not found")
	}
	items := make([]*pb.OrderItem, len(order.Items))
	for i, item := range order.Items {
		items[i] = s.itemToProto(item)
	}
	return &pb.GetOrderItemsByOrderIDResponse{Items: items}, nil
}

func (s *Server) UpdateOrderShippingStatus(ctx context.Context, req *pb.UpdateOrderShippingStatusRequest) (*pb.OrderInfo, error) {
	start := time.Now()
	slog.Info("gRPC UpdateOrderShippingStatus received", "order_id", req.OrderId, "user_id", req.UserId, "new_status", req.NewShippingStatus)

	var err error
	switch req.NewShippingStatus {
	case pb.ShippingStatus_SHIPPING_SHIPPED:
		err = s.cmdService.ShipOrder(ctx, &application.ShipOrderCommand{UserID: req.UserId, OrderID: req.OrderId, Operator: req.Operator})
	case pb.ShippingStatus_SHIPPING_DELIVERED:
		err = s.cmdService.DeliverOrder(ctx, &application.DeliverOrderCommand{UserID: req.UserId, OrderID: req.OrderId, Operator: req.Operator})
	default:
		return nil, status.Error(codes.Unimplemented, "shipping status mapping not found")
	}
	if err != nil {
		slog.Error("gRPC UpdateOrderShippingStatus failed", "order_id", req.OrderId, "status", req.NewShippingStatus, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update shipping status: %v", err))
	}
	slog.Info("gRPC UpdateOrderShippingStatus successful", "order_id", req.OrderId, "duration", time.Since(start))
	return s.GetOrderByID(ctx, &pb.GetOrderByIDRequest{Id: req.OrderId, UserId: req.UserId})
}

func (s *Server) SagaConfirmOrder(ctx context.Context, req *pb.SagaOrderRequest) (*pb.SagaOrderResponse, error) {
	if err := s.cmdService.SagaConfirmOrder(ctx, req.UserId, req.OrderId); err != nil {
		return nil, status.Errorf(codes.Internal, "SagaConfirmOrder failed: %v", err)
	}
	return &pb.SagaOrderResponse{Success: true}, nil
}

func (s *Server) SagaCancelOrder(ctx context.Context, req *pb.SagaOrderRequest) (*pb.SagaOrderResponse, error) {
	if err := s.cmdService.SagaCancelOrder(ctx, req.UserId, req.OrderId, req.Reason); err != nil {
		return nil, status.Errorf(codes.Internal, "SagaCancelOrder failed: %v", err)
	}
	return &pb.SagaOrderResponse{Success: true}, nil
}

// Helpers
func (s *Server) toProto(o *domain.Order) *pb.OrderInfo {
	if o == nil {
		return nil
	}
	items := make([]*pb.OrderItem, len(o.Items))
	for i, item := range o.Items {
		items[i] = s.itemToProto(item)
	}
	var sa *pb.ShippingAddress
	if o.ShippingAddress != nil {
		sa = &pb.ShippingAddress{
			RecipientName:   o.ShippingAddress.RecipientName,
			PhoneNumber:     o.ShippingAddress.PhoneNumber,
			Province:        o.ShippingAddress.Province,
			City:            o.ShippingAddress.City,
			District:        o.ShippingAddress.District,
			DetailedAddress: o.ShippingAddress.DetailedAddress,
			PostalCode:      o.ShippingAddress.PostalCode,
			Lat:             o.ShippingAddress.Lat,
			Lon:             o.ShippingAddress.Lon,
		}
	}
	return &pb.OrderInfo{
		Id:              uint64(o.ID),
		OrderNo:         o.OrderNo,
		UserId:          o.UserID,
		Status:          o.Status,
		TotalAmount:     o.TotalAmount,
		ActualAmount:    o.ActualAmount,
		CreatedAt:       timestamppb.New(o.CreatedAt),
		UpdatedAt:       timestamppb.New(o.UpdatedAt),
		Items:           items,
		ShippingAddress: sa,
	}
}

func (s *Server) itemToProto(item *domain.OrderItem) *pb.OrderItem {
	if item == nil {
		return nil
	}
	return &pb.OrderItem{
		Id:          uint64(item.ID),
		OrderId:     item.OrderID,
		ProductId:   item.ProductID,
		SkuId:       item.SkuID,
		ProductName: item.ProductName,
		SkuName:     item.SkuName,
		Price:       item.Price,
		Quantity:    item.Quantity,
		TotalPrice:  item.Price * int64(item.Quantity),
	}
}
