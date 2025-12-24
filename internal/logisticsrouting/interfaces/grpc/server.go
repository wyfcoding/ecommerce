package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/goapi/logisticsrouting/v1"
	"github.com/wyfcoding/ecommerce/internal/logisticsrouting/application"
	"github.com/wyfcoding/ecommerce/internal/logisticsrouting/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server 结构体实现了 LogisticsRoutingService 的 gRPC 服务端接口。
type Server struct {
	pb.UnimplementedLogisticsRoutingServiceServer
	app *application.LogisticsRoutingService
}

// NewServer 创建并返回一个新的 LogisticsRouting gRPC 服务端实例。
func NewServer(app *application.LogisticsRoutingService) *Server {
	return &Server{app: app}
}

// RegisterCarrier 处理注册配送商的gRPC请求。
func (s *Server) RegisterCarrier(ctx context.Context, req *pb.RegisterCarrierRequest) (*emptypb.Empty, error) {
	carrier := &domain.Carrier{
		Name:              req.Name,
		Type:              req.Type,
		BaseCost:          req.BaseCost,
		WeightRate:        req.WeightRate,
		DistanceRate:      req.DistanceRate,
		BaseDeliveryTime:  req.BaseDeliveryTime,
		SupportedRegions:  req.SupportedRegions,
		AvailableCapacity: req.AvailableCapacity,
		Rating:            5.0,
		IsActive:          true,
	}

	if err := s.app.RegisterCarrier(ctx, carrier); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to register carrier: %v", err))
	}

	return &emptypb.Empty{}, nil
}

// OptimizeRoute 处理优化配送路线的gRPC请求。
func (s *Server) OptimizeRoute(ctx context.Context, req *pb.OptimizeRouteRequest) (*pb.OptimizeRouteResponse, error) {
	route, err := s.app.OptimizeRoute(ctx, req.OrderIds)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to optimize route: %v", err))
	}

	return &pb.OptimizeRouteResponse{
		Route: convertRouteToProto(route),
	}, nil
}

// GetRoute 处理获取优化路由详情的gRPC请求。
func (s *Server) GetRoute(ctx context.Context, req *pb.GetRouteRequest) (*pb.GetRouteResponse, error) {
	route, err := s.app.GetRoute(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("route not found: %v", err))
	}

	return &pb.GetRouteResponse{
		Route: convertRouteToProto(route),
	}, nil
}

// ListCarriers 处理列出配送商的gRPC请求。
func (s *Server) ListCarriers(ctx context.Context, _ *emptypb.Empty) (*pb.ListCarriersResponse, error) {
	carriers, err := s.app.ListCarriers(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list carriers: %v", err))
	}

	pbCarriers := make([]*pb.Carrier, len(carriers))
	for i, c := range carriers {
		pbCarriers[i] = convertCarrierToProto(c)
	}

	return &pb.ListCarriersResponse{
		Carriers: pbCarriers,
	}, nil
}

func convertCarrierToProto(c *domain.Carrier) *pb.Carrier {
	if c == nil {
		return nil
	}
	return &pb.Carrier{
		Id:                uint64(c.ID),
		Name:              c.Name,
		Type:              c.Type,
		BaseCost:          c.BaseCost,
		WeightRate:        c.WeightRate,
		DistanceRate:      c.DistanceRate,
		BaseDeliveryTime:  c.BaseDeliveryTime,
		SupportedRegions:  c.SupportedRegions,
		AvailableCapacity: c.AvailableCapacity,
		Rating:            c.Rating,
		IsActive:          c.IsActive,
		CreatedAt:         timestamppb.New(c.CreatedAt),
		UpdatedAt:         timestamppb.New(c.UpdatedAt),
	}
}

func convertRouteToProto(r *domain.OptimizedRoute) *pb.OptimizedRoute {
	if r == nil {
		return nil
	}
	orders := make([]*pb.RouteOrder, len(r.Orders))
	for i, o := range r.Orders {
		orders[i] = &pb.RouteOrder{
			OrderId:       o.OrderID,
			CarrierId:     o.CarrierID,
			CarrierName:   o.CarrierName,
			EstimatedCost: o.EstimatedCost,
			EstimatedTime: o.EstimatedTime,
		}
	}
	return &pb.OptimizedRoute{
		Id:          uint64(r.ID),
		Orders:      orders,
		OrderCount:  r.OrderCount,
		TotalCost:   r.TotalCost,
		AverageCost: r.AverageCost,
		CreatedAt:   timestamppb.New(r.CreatedAt),
		UpdatedAt:   timestamppb.New(r.UpdatedAt),
	}
}
