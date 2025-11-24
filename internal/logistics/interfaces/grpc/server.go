package grpc

import (
	"context"
	pb "ecommerce/api/logistics/v1"
	"ecommerce/internal/logistics/application"
	"ecommerce/internal/logistics/domain/entity"
	"ecommerce/pkg/algorithm"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedLogisticsServiceServer
	app *application.LogisticsService
}

func NewServer(app *application.LogisticsService) *Server {
	return &Server{app: app}
}

func (s *Server) CreateLogistics(ctx context.Context, req *pb.CreateLogisticsRequest) (*pb.CreateLogisticsResponse, error) {
	logistics, err := s.app.CreateLogistics(
		ctx,
		req.OrderId,
		req.OrderNo,
		req.TrackingNo,
		req.Carrier,
		req.CarrierCode,
		req.SenderName,
		req.SenderPhone,
		req.SenderAddress,
		req.SenderLat,
		req.SenderLon,
		req.ReceiverName,
		req.ReceiverPhone,
		req.ReceiverAddress,
		req.ReceiverLat,
		req.ReceiverLon,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateLogisticsResponse{
		Logistics: convertLogisticsToProto(logistics),
	}, nil
}

func (s *Server) GetLogistics(ctx context.Context, req *pb.GetLogisticsRequest) (*pb.GetLogisticsResponse, error) {
	logistics, err := s.app.GetLogistics(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &pb.GetLogisticsResponse{
		Logistics: convertLogisticsToProto(logistics),
	}, nil
}

func (s *Server) GetLogisticsByTrackingNo(ctx context.Context, req *pb.GetLogisticsByTrackingNoRequest) (*pb.GetLogisticsResponse, error) {
	logistics, err := s.app.GetLogisticsByTrackingNo(ctx, req.TrackingNo)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &pb.GetLogisticsResponse{
		Logistics: convertLogisticsToProto(logistics),
	}, nil
}

func (s *Server) UpdateStatus(ctx context.Context, req *pb.UpdateStatusRequest) (*emptypb.Empty, error) {
	if err := s.app.UpdateStatus(ctx, req.Id, entity.LogisticsStatus(req.Status), req.Location, req.Description); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) AddTrace(ctx context.Context, req *pb.AddTraceRequest) (*emptypb.Empty, error) {
	if err := s.app.AddTrace(ctx, req.Id, req.Location, req.Description, req.Status); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) SetEstimatedTime(ctx context.Context, req *pb.SetEstimatedTimeRequest) (*emptypb.Empty, error) {
	if err := s.app.SetEstimatedTime(ctx, req.Id, req.EstimatedTime.AsTime()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListLogistics(ctx context.Context, req *pb.ListLogisticsRequest) (*pb.ListLogisticsResponse, error) {
	page := int(req.PageNum)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	logisticsList, total, err := s.app.ListLogistics(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbLogistics := make([]*pb.Logistics, len(logisticsList))
	for i, l := range logisticsList {
		pbLogistics[i] = convertLogisticsToProto(l)
	}

	return &pb.ListLogisticsResponse{
		Logistics:  pbLogistics,
		TotalCount: uint64(total),
	}, nil
}

func (s *Server) OptimizeDeliveryRoute(ctx context.Context, req *pb.OptimizeDeliveryRouteRequest) (*pb.OptimizeDeliveryRouteResponse, error) {
	destinations := make([]algorithm.Location, len(req.Destinations))
	for i, d := range req.Destinations {
		destinations[i] = algorithm.Location{
			ID:  d.Id,
			Lat: d.Lat,
			Lon: d.Lon,
		}
	}

	route, err := s.app.OptimizeDeliveryRoute(ctx, req.LogisticsId, destinations)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.OptimizeDeliveryRouteResponse{
		Route: convertRouteToProto(route),
	}, nil
}

func convertLogisticsToProto(l *entity.Logistics) *pb.Logistics {
	if l == nil {
		return nil
	}
	traces := make([]*pb.LogisticsTrace, len(l.Traces))
	for i, t := range l.Traces {
		traces[i] = convertTraceToProto(t)
	}
	resp := &pb.Logistics{
		Id:              uint64(l.ID),
		OrderId:         l.OrderID,
		OrderNo:         l.OrderNo,
		TrackingNo:      l.TrackingNo,
		Carrier:         l.Carrier,
		CarrierCode:     l.CarrierCode,
		SenderName:      l.SenderName,
		SenderPhone:     l.SenderPhone,
		SenderAddress:   l.SenderAddress,
		SenderLat:       l.SenderLat,
		SenderLon:       l.SenderLon,
		ReceiverName:    l.ReceiverName,
		ReceiverPhone:   l.ReceiverPhone,
		ReceiverAddress: l.ReceiverAddress,
		ReceiverLat:     l.ReceiverLat,
		ReceiverLon:     l.ReceiverLon,
		Status:          int32(l.Status),
		CurrentLocation: l.CurrentLocation,
		Traces:          traces,
		Route:           convertRouteToProto(l.Route),
		CreatedAt:       timestamppb.New(l.CreatedAt),
		UpdatedAt:       timestamppb.New(l.UpdatedAt),
	}
	if l.EstimatedTime != nil {
		resp.EstimatedTime = timestamppb.New(*l.EstimatedTime)
	}
	if l.DeliveredAt != nil {
		resp.DeliveredAt = timestamppb.New(*l.DeliveredAt)
	}
	return resp
}

func convertTraceToProto(t *entity.LogisticsTrace) *pb.LogisticsTrace {
	if t == nil {
		return nil
	}
	return &pb.LogisticsTrace{
		Id:          uint64(t.ID),
		LogisticsId: t.LogisticsID,
		TrackingNo:  t.TrackingNo,
		Location:    t.Location,
		Description: t.Description,
		Status:      t.Status,
		CreatedAt:   timestamppb.New(t.CreatedAt),
	}
}

func convertRouteToProto(r *entity.DeliveryRoute) *pb.DeliveryRoute {
	if r == nil {
		return nil
	}
	return &pb.DeliveryRoute{
		Id:          uint64(r.ID),
		LogisticsId: r.LogisticsID,
		RouteData:   r.RouteData,
		Distance:    r.Distance,
	}
}
