package grpc

import (
	"context"
	"log/slog"

	pb "github.com/wyfcoding/ecommerce/go-api/reportcenter/v1"
	"github.com/wyfcoding/ecommerce/internal/reportcenter/application"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	pb.UnimplementedReportCenterServiceServer
	svc    *application.ReportService
	logger *slog.Logger
}

func NewServer(svc *application.ReportService, logger *slog.Logger) pb.ReportCenterServiceServer {
	return &Server{
		svc:    svc,
		logger: logger.With("component", "report_grpc"),
	}
}

func (s *Server) GetSalesReport(ctx context.Context, req *pb.GetSalesReportRequest) (*pb.SalesReport, error) {
	return s.svc.GetSalesReport(ctx, req)
}

func (s *Server) GetInventoryHealth(ctx context.Context, req *pb.GetInventoryHealthRequest) (*pb.InventoryHealthReport, error) {
	return s.svc.GetInventoryHealth(ctx)
}

func (s *Server) GetLowStockAlerts(ctx context.Context, _ *emptypb.Empty) (*pb.LowStockAlertsResponse, error) {
	alerts, err := s.svc.GetLowStockAlerts(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.LowStockAlertsResponse{Alerts: alerts}, nil
}

func (s *Server) GetFinancialSummary(ctx context.Context, req *pb.GetFinancialSummaryRequest) (*pb.FinancialSummary, error) {
	return s.svc.GetFinancialSummary(ctx, req)
}

func (s *Server) CreateCustomReport(ctx context.Context, req *pb.CreateCustomReportRequest) (*pb.CustomReport, error) {
	report, err := s.svc.CreateCustomReport(ctx, req)
	if err != nil {
		return nil, err
	}
	// 简单映射
	return &pb.CustomReport{
		Id:         report.ID,
		Name:       report.Name,
		ReportType: report.Type,
		Status:     report.Status,
	}, nil
}
