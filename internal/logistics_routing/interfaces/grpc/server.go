package grpc

import (
	"context"
	pb "ecommerce/api/logistics_routing/v1"
	"ecommerce/internal/logistics_routing/application"
	"ecommerce/internal/logistics_routing/domain/entity"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedLogisticsRoutingServiceServer
	app *application.LogisticsRoutingService
}

func NewServer(app *application.LogisticsRoutingService) *Server {
	return &Server{app: app}
}

func (s *Server) RegisterCarrier(ctx context.Context, req *pb.RegisterCarrierRequest) (*emptypb.Empty, error) {
	carrier := &entity.Carrier{
		Name:              req.Name,
		Type:              req.Type,
		BaseCost:          req.BaseCost,
		WeightRate:        req.WeightRate,
		DistanceRate:      req.DistanceRate,
		BaseDeliveryTime:  req.BaseDeliveryTime,
		SupportedRegions:  req.SupportedRegions,
		AvailableCapacity: req.AvailableCapacity,
		Rating:            5.0,  // Default
		IsActive:          true, // Default
	}

	if err := s.app.RegisterCarrier(ctx, carrier); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) OptimizeRoute(ctx context.Context, req *pb.OptimizeRouteRequest) (*pb.OptimizeRouteResponse, error) {
	route, err := s.app.OptimizeRoute(ctx, req.OrderIds)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.OptimizeRouteResponse{
		Route: convertRouteToProto(route),
	}, nil
}

func (s *Server) GetRoute(ctx context.Context, req *pb.GetRouteRequest) (*pb.GetRouteResponse, error) {
	route, err := s.app.GetRoute(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &pb.GetRouteResponse{
		Route: convertRouteToProto(route),
	}, nil
}

func (s *Server) ListCarriers(ctx context.Context, _ *emptypb.Empty) (*pb.ListCarriersResponse, error) {
	carriers, err := s.app.ListCarriers(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbCarriers := make([]*pb.Carrier, len(carriers))
	for i, c := range carriers {
		pbCarriers[i] = convertCarrierToProto(c)
	}

	return &pb.ListCarriersResponse{
		Carriers: pbCarriers,
	}, nil
}

func convertCarrierToProto(c *entity.Carrier) *pb.Carrier {
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

func convertRouteToProto(r *entity.OptimizedRoute) *pb.OptimizedRoute {
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
