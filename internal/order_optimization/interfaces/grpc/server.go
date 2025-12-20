package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/go-api/order_optimization/v1"
	"github.com/wyfcoding/ecommerce/internal/order_optimization/application"
	"github.com/wyfcoding/ecommerce/internal/order_optimization/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedOrderOptimizationServiceServer
	app *application.OrderOptimizationService
}

func NewServer(app *application.OrderOptimizationService) *Server {
	return &Server{app: app}
}

func (s *Server) MergeOrders(ctx context.Context, req *pb.MergeOrdersRequest) (*pb.MergeOrdersResponse, error) {
	mergedOrder, err := s.app.MergeOrders(ctx, req.UserId, req.OrderIds)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to merge orders: %v", err))
	}

	return &pb.MergeOrdersResponse{
		MergedOrder: convertMergedOrderToProto(mergedOrder),
	}, nil
}

func (s *Server) SplitOrder(ctx context.Context, req *pb.SplitOrderRequest) (*pb.SplitOrderResponse, error) {
	splitOrders, err := s.app.SplitOrder(ctx, req.OrderId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to split order: %v", err))
	}

	pbSplitOrders := make([]*pb.SplitOrder, len(splitOrders))
	for i, o := range splitOrders {
		pbSplitOrders[i] = convertSplitOrderToProto(o)
	}

	return &pb.SplitOrderResponse{
		SplitOrders: pbSplitOrders,
	}, nil
}

func (s *Server) AllocateWarehouse(ctx context.Context, req *pb.AllocateWarehouseRequest) (*pb.AllocateWarehouseResponse, error) {
	plan, err := s.app.AllocateWarehouse(ctx, req.OrderId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to allocate warehouse: %v", err))
	}

	return &pb.AllocateWarehouseResponse{
		Plan: convertAllocationPlanToProto(plan),
	}, nil
}

func convertMergedOrderToProto(o *domain.MergedOrder) *pb.MergedOrder {
	if o == nil {
		return nil
	}
	items := make([]*pb.OrderItem, len(o.Items))
	for i, item := range o.Items {
		items[i] = &pb.OrderItem{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	return &pb.MergedOrder{
		Id:               uint64(o.ID),
		UserId:           o.UserID,
		OriginalOrderIds: o.OriginalOrderIDs,
		Items:            items,
		TotalAmount:      o.TotalAmount,
		DiscountAmount:   o.DiscountAmount,
		FinalAmount:      o.FinalAmount,
		ShippingAddress: &pb.ShippingAddress{
			Name:     o.ShippingAddress.Name,
			Phone:    o.ShippingAddress.Phone,
			Province: o.ShippingAddress.Province,
			City:     o.ShippingAddress.City,
			District: o.ShippingAddress.District,
			Address:  o.ShippingAddress.Address,
		},
		Status:    o.Status,
		CreatedAt: timestamppb.New(o.CreatedAt),
		UpdatedAt: timestamppb.New(o.UpdatedAt),
	}
}

func convertSplitOrderToProto(o *domain.SplitOrder) *pb.SplitOrder {
	if o == nil {
		return nil
	}
	items := make([]*pb.OrderItem, len(o.Items))
	for i, item := range o.Items {
		items[i] = &pb.OrderItem{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	return &pb.SplitOrder{
		Id:              uint64(o.ID),
		OriginalOrderId: o.OriginalOrderID,
		SplitIndex:      o.SplitIndex,
		Items:           items,
		Amount:          o.Amount,
		WarehouseId:     o.WarehouseID,
		ShippingAddress: &pb.ShippingAddress{
			Name:     o.ShippingAddress.Name,
			Phone:    o.ShippingAddress.Phone,
			Province: o.ShippingAddress.Province,
			City:     o.ShippingAddress.City,
			District: o.ShippingAddress.District,
			Address:  o.ShippingAddress.Address,
		},
		Status:    o.Status,
		CreatedAt: timestamppb.New(o.CreatedAt),
		UpdatedAt: timestamppb.New(o.UpdatedAt),
	}
}

func convertAllocationPlanToProto(p *domain.WarehouseAllocationPlan) *pb.WarehouseAllocationPlan {
	if p == nil {
		return nil
	}
	allocations := make([]*pb.WarehouseAllocation, len(p.Allocations))
	for i, a := range p.Allocations {
		allocations[i] = &pb.WarehouseAllocation{
			ProductId:   a.ProductID,
			Quantity:    a.Quantity,
			WarehouseId: a.WarehouseID,
			Distance:    a.Distance,
		}
	}

	return &pb.WarehouseAllocationPlan{
		Id:          uint64(p.ID),
		OrderId:     p.OrderID,
		Allocations: allocations,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}
}
