package grpc

import (
	"context"
	pb "ecommerce/api/order/v1"
	"ecommerce/internal/order/application"
	"ecommerce/internal/order/domain/entity"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedOrderServer
	app *application.OrderService
}

func NewServer(app *application.OrderService) *Server {
	return &Server{app: app}
}

func (s *Server) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.OrderInfo, error) {
	items := make([]*entity.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = &entity.OrderItem{
			ProductID: item.ProductId,
			SkuID:     item.SkuId,
			Quantity:  item.Quantity,
			// Name, Price, etc. should be fetched from Product Service or passed in.
			// Service CreateOrder expects entity.OrderItem which usually has price populated.
			// Assuming Service or Repo handles price lookup or we pass 0 and it's handled.
			// For this refactor, we pass basic info.
		}
	}

	shippingAddr := &entity.ShippingAddress{
		RecipientName:   req.ShippingAddress.RecipientName,
		PhoneNumber:     req.ShippingAddress.PhoneNumber,
		Province:        req.ShippingAddress.Province,
		City:            req.ShippingAddress.City,
		District:        req.ShippingAddress.District,
		DetailedAddress: req.ShippingAddress.DetailedAddress,
		PostalCode:      req.ShippingAddress.PostalCode,
	}

	order, err := s.app.CreateOrder(ctx, req.UserId, items, shippingAddr)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return s.toProto(order), nil
}

func (s *Server) GetOrderByID(ctx context.Context, req *pb.GetOrderByIDRequest) (*pb.OrderInfo, error) {
	order, err := s.app.GetOrder(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if order == nil {
		return nil, status.Error(codes.NotFound, "order not found")
	}
	// Check permission if needed (req.UserId)

	return s.toProto(order), nil
}

func (s *Server) UpdateOrderStatus(ctx context.Context, req *pb.UpdateOrderStatusRequest) (*pb.OrderInfo, error) {
	// Service has specific methods for status transitions (Pay, Ship, Deliver, Complete, Cancel).
	// Generic UpdateOrderStatus might need to map to these or use a generic update if available.
	// Service doesn't have a generic "UpdateStatus" that takes an enum.
	// We need to switch on new_status.

	var err error
	switch req.NewStatus {
	case pb.OrderStatus_PAID:
		err = s.app.PayOrder(ctx, req.Id, "Manual/Admin")
	case pb.OrderStatus_SHIPPED:
		err = s.app.ShipOrder(ctx, req.Id, req.Operator)
	case pb.OrderStatus_DELIVERED:
		err = s.app.DeliverOrder(ctx, req.Id, req.Operator)
	case pb.OrderStatus_COMPLETED:
		err = s.app.CompleteOrder(ctx, req.Id, req.Operator)
	case pb.OrderStatus_CANCELLED:
		err = s.app.CancelOrder(ctx, req.Id, req.Operator, req.Remark)
	default:
		return nil, status.Error(codes.InvalidArgument, "unsupported status transition via this API")
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return s.GetOrderByID(ctx, &pb.GetOrderByIDRequest{Id: req.Id})
}

func (s *Server) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.OrderInfo, error) {
	err := s.app.CancelOrder(ctx, req.Id, strconv.FormatUint(req.UserId, 10), req.Reason)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return s.GetOrderByID(ctx, &pb.GetOrderByIDRequest{Id: req.Id})
}

func (s *Server) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	var statusPtr *int
	if req.Status != pb.OrderStatus_ORDER_STATUS_UNSPECIFIED {
		st := int(req.Status)
		statusPtr = &st
	}

	orders, total, err := s.app.ListOrders(ctx, req.UserId, statusPtr, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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
	err := s.app.PayOrder(ctx, req.OrderId, req.PaymentMethod)
	if err != nil {
		return &pb.PaymentResult{
			OrderId: req.OrderId,
			Status:  pb.PaymentStatus_FAILED,
			Message: err.Error(),
		}, nil
	}

	return &pb.PaymentResult{
		OrderId:       req.OrderId,
		TransactionId: "mock-txn-" + strconv.FormatUint(req.OrderId, 10),
		Status:        pb.PaymentStatus_SUCCESS,
		PaidAt:        timestamppb.Now(),
	}, nil
}

func (s *Server) RequestRefund(ctx context.Context, req *pb.RequestRefundRequest) (*pb.OrderInfo, error) {
	return nil, status.Error(codes.Unimplemented, "RequestRefund not implemented")
}

func (s *Server) GetOrderItemsByOrderID(ctx context.Context, req *pb.GetOrderItemsByOrderIDRequest) (*pb.GetOrderItemsByOrderIDResponse, error) {
	order, err := s.app.GetOrder(ctx, req.OrderId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if order == nil {
		return nil, status.Error(codes.NotFound, "order not found")
	}

	items := make([]*pb.OrderItem, len(order.Items))
	for i, item := range order.Items {
		items[i] = s.itemToProto(item)
	}

	return &pb.GetOrderItemsByOrderIDResponse{
		Items: items,
	}, nil
}

func (s *Server) UpdateOrderShippingStatus(ctx context.Context, req *pb.UpdateOrderShippingStatusRequest) (*pb.OrderInfo, error) {
	// Service methods map to status updates.
	// If shipping status maps to OrderStatus (e.g. Shipped, Delivered), we can reuse.
	// If it's separate, we might need a new service method.
	// Assuming mapping to OrderStatus for now.

	var err error
	switch req.NewShippingStatus {
	case pb.ShippingStatus_SHIPPING_SHIPPED:
		err = s.app.ShipOrder(ctx, req.OrderId, req.Operator)
	case pb.ShippingStatus_SHIPPING_DELIVERED:
		err = s.app.DeliverOrder(ctx, req.OrderId, req.Operator)
	default:
		return nil, status.Error(codes.Unimplemented, "shipping status not mapped")
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return s.GetOrderByID(ctx, &pb.GetOrderByIDRequest{Id: req.OrderId})
}

func (s *Server) toProto(o *entity.Order) *pb.OrderInfo {
	items := make([]*pb.OrderItem, len(o.Items))
	for i, item := range o.Items {
		items[i] = s.itemToProto(item)
	}

	return &pb.OrderInfo{
		Id:           uint64(o.ID),
		OrderNo:      o.OrderNo,
		UserId:       o.UserID,
		Status:       pb.OrderStatus(o.Status),
		TotalAmount:  int64(o.TotalAmount * 100), // Assuming float to cents
		ActualAmount: int64(o.TotalAmount * 100), // Simplified
		CreatedAt:    timestamppb.New(o.CreatedAt),
		UpdatedAt:    timestamppb.New(o.UpdatedAt),
		Items:        items,
		ShippingAddress: &pb.ShippingAddress{
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

func (s *Server) itemToProto(item *entity.OrderItem) *pb.OrderItem {
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
