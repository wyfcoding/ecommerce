package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/order/v1"
	"github.com/wyfcoding/ecommerce/internal/order/application"
	"github.com/wyfcoding/ecommerce/internal/order/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server implements the Order gRPC service.
type Server struct {
	pb.UnimplementedOrderServiceServer
	cmdService   *application.OrderCommandService
	queryService *application.OrderQueryService
}

// NewServer Creates a new Order gRPC server.
func NewServer(cmd *application.OrderCommandService, query *application.OrderQueryService) *Server {
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
		Lat:             req.ShippingAddress.Lat,
		Lon:             req.ShippingAddress.Lon,
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
		Remark:          req.Remark,
		PaymentMethod:   req.PaymentMethod,
		ClientIP:        clientIPFromContext(ctx),
		DeviceID:        deviceIDFromContext(ctx),
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
	case pb.OrderStatus_REFUNDED:
		err = s.cmdService.ApproveRefund(ctx, &application.ApproveRefundCommand{UserID: req.UserId, OrderID: req.Id, Operator: req.Operator})
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
	offset := (page - 1) * pageSize

	var statusPtr *int
	if req.Status != pb.OrderStatus_ORDER_STATUS_UNSPECIFIED {
		st := int(req.Status)
		statusPtr = &st
	}

	var startTime *time.Time
	if req.StartTime != nil {
		t := req.StartTime.AsTime()
		startTime = &t
	}
	var endTime *time.Time
	if req.EndTime != nil {
		t := req.EndTime.AsTime()
		endTime = &t
	}

	var (
		orders []*domain.Order
		total  int64
		err    error
	)
	if req.UserId > 0 {
		orders, total, err = s.queryService.ListUserOrders(ctx, req.UserId, statusPtr, offset, pageSize, startTime, endTime, req.SortBy)
	} else {
		orders, total, err = s.queryService.ListOrders(ctx, statusPtr, offset, pageSize, startTime, endTime, req.SortBy)
	}
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

	paymentStatus := req.PaymentStatus
	if paymentStatus == pb.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED || paymentStatus == pb.PaymentStatus_SUCCESS {
		err := s.cmdService.PayOrder(ctx, &application.PayOrderCommand{
			UserID:        req.UserId,
			OrderID:       req.OrderId,
			PaymentMethod: req.PaymentMethod,
			Amount:        req.Amount,
			TransactionID: req.TransactionId,
		})
		if err != nil {
			slog.Error("gRPC ProcessPayment failed", "order_id", req.OrderId, "user_id", req.UserId, "error", err, "duration", time.Since(start))
			return &pb.PaymentResult{OrderId: req.OrderId, Status: pb.PaymentStatus_FAILED, Message: err.Error()}, nil
		}
		slog.Info("gRPC ProcessPayment successful", "order_id", req.OrderId, "user_id", req.UserId, "duration", time.Since(start))
		transactionID := req.TransactionId
		if transactionID == "" {
			transactionID = "mock-txn-" + strconv.FormatUint(req.OrderId, 10)
		}
		return &pb.PaymentResult{
			OrderId:       req.OrderId,
			TransactionId: transactionID,
			Status:        pb.PaymentStatus_SUCCESS,
			PaidAt:        timestamppb.Now(),
		}, nil
	}

	switch paymentStatus {
	case pb.PaymentStatus_PROCESSING, pb.PaymentStatus_FAILED, pb.PaymentStatus_REFUND_FAILED:
		err := s.cmdService.UpdatePaymentStatus(ctx, &application.UpdatePaymentStatusCommand{
			UserID:        req.UserId,
			OrderID:       req.OrderId,
			Operator:      "System",
			Status:        paymentStatus,
			PaymentMethod: req.PaymentMethod,
			TransactionID: req.TransactionId,
			Remark:        req.Remark,
		})
		if err != nil {
			slog.Error("gRPC ProcessPayment failed", "order_id", req.OrderId, "user_id", req.UserId, "error", err, "duration", time.Since(start))
			return &pb.PaymentResult{OrderId: req.OrderId, Status: pb.PaymentStatus_FAILED, Message: err.Error()}, nil
		}
		slog.Info("gRPC ProcessPayment successful", "order_id", req.OrderId, "user_id", req.UserId, "duration", time.Since(start))
		return &pb.PaymentResult{
			OrderId:       req.OrderId,
			TransactionId: req.TransactionId,
			Status:        paymentStatus,
			Message:       req.Remark,
		}, nil
	default:
		return nil, status.Error(codes.InvalidArgument, "unsupported payment status")
	}
}

func (s *Server) RequestRefund(ctx context.Context, req *pb.RequestRefundRequest) (*pb.OrderInfo, error) {
	err := s.cmdService.RequestRefund(ctx, &application.RequestRefundCommand{
		UserID:       req.UserId,
		OrderID:      req.OrderId,
		Operator:     "User",
		RefundAmount: req.RefundAmount,
		Reason:       req.Reason,
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
		err = s.cmdService.ShipOrder(ctx, &application.ShipOrderCommand{
			UserID:           req.UserId,
			OrderID:          req.OrderId,
			Operator:         req.Operator,
			TrackingNumber:   req.TrackingNumber,
			LogisticsCompany: req.LogisticsCompany,
		})
	case pb.ShippingStatus_SHIPPING_DELIVERED:
		err = s.cmdService.DeliverOrder(ctx, &application.DeliverOrderCommand{
			UserID:           req.UserId,
			OrderID:          req.OrderId,
			Operator:         req.Operator,
			TrackingNumber:   req.TrackingNumber,
			LogisticsCompany: req.LogisticsCompany,
		})
	case pb.ShippingStatus_SHIPPING_STATUS_UNSPECIFIED:
		return nil, status.Error(codes.InvalidArgument, "shipping status is required")
	default:
		err = s.cmdService.UpdateShippingStatus(ctx, &application.UpdateShippingStatusCommand{
			UserID:           req.UserId,
			OrderID:          req.OrderId,
			Operator:         req.Operator,
			NewStatus:        req.NewShippingStatus,
			TrackingNumber:   req.TrackingNumber,
			LogisticsCompany: req.LogisticsCompany,
			Remark:           req.Remark,
		})
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
	logs := make([]*pb.OrderLog, 0, len(o.Logs))
	for _, log := range o.Logs {
		if log == nil {
			continue
		}
		logs = append(logs, &pb.OrderLog{
			Id:        uint64(log.ID),
			OrderId:   log.OrderID,
			Operator:  log.Operator,
			Action:    log.Action,
			OldStatus: log.OldStatus,
			NewStatus: log.NewStatus,
			Remark:    log.Remark,
			CreatedAt: timestamppb.New(log.CreatedAt),
		})
	}

	info := &pb.OrderInfo{
		Id:                   uint64(o.ID),
		OrderNo:              o.OrderNo,
		UserId:               o.UserID,
		Status:               o.Status,
		PaymentStatus:        paymentStatusFromOrder(o),
		ShippingStatus:       shippingStatusFromOrder(o),
		TotalAmount:          o.TotalAmount,
		ActualAmount:         o.ActualAmount,
		ShippingFee:          o.ShippingFee,
		DiscountAmount:       o.DiscountAmount,
		PaymentMethod:        o.PaymentMethod,
		PaymentTransactionId: o.PaymentTransactionID,
		Remark:               o.Remark,
		TrackingNumber:       o.TrackingNumber,
		LogisticsCompany:     o.LogisticsCompany,
		RefundAmount:         o.RefundAmount,
		RefundReason:         o.RefundReason,
		CreatedAt:            timestamppb.New(o.CreatedAt),
		UpdatedAt:            timestamppb.New(o.UpdatedAt),
		Items:                items,
		Logs:                 logs,
		ShippingAddress:      sa,
	}
	if o.PaidAt != nil {
		info.PaidAt = timestamppb.New(*o.PaidAt)
	}
	if o.ShippedAt != nil {
		info.ShippedAt = timestamppb.New(*o.ShippedAt)
	}
	if o.DeliveredAt != nil {
		info.DeliveredAt = timestamppb.New(*o.DeliveredAt)
	}
	if o.CompletedAt != nil {
		info.CompletedAt = timestamppb.New(*o.CompletedAt)
	}
	if o.CancelledAt != nil {
		info.CancelledAt = timestamppb.New(*o.CancelledAt)
	}
	return info
}

func (s *Server) itemToProto(item *domain.OrderItem) *pb.OrderItem {
	if item == nil {
		return nil
	}
	return &pb.OrderItem{
		Id:              uint64(item.ID),
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

func paymentStatusFromOrder(o *domain.Order) pb.PaymentStatus {
	if o == nil {
		return pb.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED
	}
	if o.PaymentStatus != pb.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED {
		return o.PaymentStatus
	}
	switch o.Status {
	case pb.OrderStatus_REFUND_REQUESTED:
		return pb.PaymentStatus_REFUNDING
	case pb.OrderStatus_REFUNDED:
		return pb.PaymentStatus_REFUND_SUCCESS
	case pb.OrderStatus_CANCELLED:
		if o.PaidAt != nil {
			return pb.PaymentStatus_REFUNDING
		}
		return pb.PaymentStatus_UNPAID
	case pb.OrderStatus_PAID, pb.OrderStatus_SHIPPED, pb.OrderStatus_DELIVERED, pb.OrderStatus_COMPLETED:
		return pb.PaymentStatus_SUCCESS
	case pb.OrderStatus_PENDING_PAYMENT, pb.OrderStatus_ALLOCATING:
		return pb.PaymentStatus_UNPAID
	case pb.OrderStatus_CLOSED:
		return pb.PaymentStatus_FAILED
	default:
		return pb.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED
	}
}

func shippingStatusFromOrder(o *domain.Order) pb.ShippingStatus {
	if o == nil {
		return pb.ShippingStatus_SHIPPING_STATUS_UNSPECIFIED
	}
	if o.ShippingStatus != pb.ShippingStatus_SHIPPING_STATUS_UNSPECIFIED {
		return o.ShippingStatus
	}
	switch o.Status {
	case pb.OrderStatus_SHIPPED:
		return pb.ShippingStatus_SHIPPING_SHIPPED
	case pb.OrderStatus_DELIVERED, pb.OrderStatus_COMPLETED:
		return pb.ShippingStatus_SHIPPING_DELIVERED
	case pb.OrderStatus_CANCELLED, pb.OrderStatus_REFUND_REQUESTED, pb.OrderStatus_REFUNDED:
		return pb.ShippingStatus_EXCEPTION
	case pb.OrderStatus_PENDING_PAYMENT, pb.OrderStatus_ALLOCATING, pb.OrderStatus_PAID:
		return pb.ShippingStatus_PENDING_SHIPMENT
	default:
		return pb.ShippingStatus_SHIPPING_STATUS_UNSPECIFIED
	}
}

func clientIPFromContext(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for _, key := range []string{"x-forwarded-for", "x-real-ip", "x-client-ip", "client-ip"} {
			if ip := firstNonEmpty(md.Get(key)); ip != "" {
				if parsed := extractFirstIP(ip); parsed != "" {
					return parsed
				}
			}
		}
	}

	if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
		if host, _, err := net.SplitHostPort(p.Addr.String()); err == nil {
			return host
		}
		return p.Addr.String()
	}
	return "unknown"
}

func deviceIDFromContext(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for _, key := range []string{"x-device-id", "device-id", "device_id"} {
			if v := firstNonEmpty(md.Get(key)); v != "" {
				return v
			}
		}
		if ua := firstNonEmpty(md.Get("user-agent")); ua != "" {
			return ua
		}
	}
	return "unknown"
}

func firstNonEmpty(values []string) string {
	for _, v := range values {
		val := strings.TrimSpace(v)
		if val != "" {
			return val
		}
	}
	return ""
}

func extractFirstIP(value string) string {
	for part := range strings.SplitSeq(value, ",") {
		ip := strings.TrimSpace(part)
		if ip == "" {
			continue
		}
		if host, _, err := net.SplitHostPort(ip); err == nil {
			ip = host
		}
		return ip
	}
	return ""
}
