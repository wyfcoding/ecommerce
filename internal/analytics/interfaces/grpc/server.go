package grpc

import (
	"context"
	pb "ecommerce/api/analytics/v1"
	"ecommerce/internal/analytics/application"
	"ecommerce/internal/analytics/domain/entity"
	"ecommerce/internal/analytics/domain/repository"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	pb.UnimplementedAnalyticsServiceServer
	app *application.AnalyticsService
}

func NewServer(app *application.AnalyticsService) *Server {
	return &Server{app: app}
}

// --- Event Tracking ---

func (s *Server) TrackEvent(ctx context.Context, req *pb.TrackEventRequest) (*emptypb.Empty, error) {
	if req.Event == nil {
		return nil, status.Error(codes.InvalidArgument, "event is required")
	}

	// Map Event to Metric
	// We treat each event as a metric record with value 1.
	// MetricType: "event"
	// Name: event_name
	// Dimension: "user_id"
	// DimensionVal: user_id

	err := s.app.RecordMetric(
		ctx,
		entity.MetricType("event"), // Custom type for raw events
		req.Event.EventName,
		1.0,
		entity.GranularityHourly, // Default granularity
		"user_id",
		strconv.FormatUint(req.Event.UserId, 10),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) BatchTrackEvents(ctx context.Context, req *pb.BatchTrackEventsRequest) (*emptypb.Empty, error) {
	for _, event := range req.Events {
		err := s.app.RecordMetric(
			ctx,
			entity.MetricType("event"),
			event.EventName,
			1.0,
			entity.GranularityHourly,
			"user_id",
			strconv.FormatUint(event.UserId, 10),
		)
		if err != nil {
			// Log error but continue? Or return error?
			// For batch, maybe best effort.
			continue
		}
	}
	return &emptypb.Empty{}, nil
}

// --- Reports ---

func (s *Server) GetSalesOverviewReport(ctx context.Context, req *pb.GetSalesOverviewReportRequest) (*pb.SalesOverviewReportResponse, error) {
	// Query Sales Metric
	salesQuery := &repository.MetricQuery{
		MetricType: entity.MetricTypeSales,
		StartTime:  req.StartDate.AsTime(),
		EndTime:    req.EndDate.AsTime(),
	}
	salesMetrics, _, err := s.app.QueryMetrics(ctx, salesQuery)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Query Orders Metric
	ordersQuery := &repository.MetricQuery{
		MetricType: entity.MetricTypeOrders,
		StartTime:  req.StartDate.AsTime(),
		EndTime:    req.EndDate.AsTime(),
	}
	ordersMetrics, _, err := s.app.QueryMetrics(ctx, ordersQuery)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var totalSales float64
	for _, m := range salesMetrics {
		totalSales += m.Value
	}

	var totalOrders uint64
	for _, m := range ordersMetrics {
		totalOrders += uint64(m.Value)
	}

	avgOrderValue := 0.0
	if totalOrders > 0 {
		avgOrderValue = totalSales / float64(totalOrders)
	}

	return &pb.SalesOverviewReportResponse{
		TotalSalesAmount:  totalSales,
		TotalOrders:       totalOrders,
		AverageOrderValue: avgOrderValue,
		// Trends not implemented for now
	}, nil
}

// --- Unimplemented Methods ---

func (s *Server) GetUserActivityReport(ctx context.Context, req *pb.GetUserActivityReportRequest) (*pb.UserActivityReportResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetUserActivityReport not implemented")
}

func (s *Server) GetProductPerformanceReport(ctx context.Context, req *pb.GetProductPerformanceReportRequest) (*pb.ProductPerformanceReportResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetProductPerformanceReport not implemented")
}

func (s *Server) GetConversionFunnelReport(ctx context.Context, req *pb.GetConversionFunnelReportRequest) (*pb.ConversionFunnelReportResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetConversionFunnelReport not implemented")
}

func (s *Server) GetCustomReport(ctx context.Context, req *pb.GetCustomReportRequest) (*pb.CustomReportResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetCustomReport not implemented")
}

func (s *Server) GetUserBehaviorPath(ctx context.Context, req *pb.GetUserBehaviorPathRequest) (*pb.UserBehaviorPathResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetUserBehaviorPath not implemented")
}

func (s *Server) GetUserSegments(ctx context.Context, req *pb.GetUserSegmentsRequest) (*pb.UserSegmentsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetUserSegments not implemented")
}

func (s *Server) GetRealtimeVisitors(ctx context.Context, req *emptypb.Empty) (*pb.RealtimeVisitorsResponse, error) {
	// Mock implementation or Unimplemented
	return &pb.RealtimeVisitorsResponse{
		VisitorCount: 0,
		ActivePages:  []string{},
	}, nil
}
