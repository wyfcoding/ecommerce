package grpc

import (
	"context"
	"fmt"
	"strconv"

	pb "github.com/wyfcoding/ecommerce/goapi/analytics/v1"          // 导入分析模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/analytics/application" // 导入分析模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/analytics/domain"      // 导入分析模块的领域层。

	"google.golang.org/grpc/codes"                   // gRPC状态码。
	"google.golang.org/grpc/status"                  // gRPC状态处理。
	"google.golang.org/protobuf/types/known/emptypb" // 导入空消息类型。
)

// Server 结构体实现了 Analytics 的 gRPC 服务端接口。
type Server struct {
	pb.UnimplementedAnalyticsServiceServer                        // 嵌入生成的UnimplementedAnalyticsServiceServer。
	app                                    *application.Analytics // 依赖Analytics应用服务 facade。
}

// NewServer 创建并返回一个新的 Analytics gRPC 服务端实例。
func NewServer(app *application.Analytics) *Server {
	return &Server{app: app}
}

// TrackEvent 处理跟踪单个事件的gRPC请求。
func (s *Server) TrackEvent(ctx context.Context, req *pb.TrackEventRequest) (*emptypb.Empty, error) {
	if req.Event == nil {
		return nil, status.Error(codes.InvalidArgument, "event is required")
	}

	err := s.app.RecordMetric(
		ctx,
		domain.MetricType("event"),
		req.Event.EventName,
		1.0,
		domain.GranularityHourly,
		"user_id",
		strconv.FormatUint(req.Event.UserId, 10),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to record event metric: %v", err))
	}

	return &emptypb.Empty{}, nil
}

// BatchTrackEvents 处理批量跟踪事件的gRPC请求。
func (s *Server) BatchTrackEvents(ctx context.Context, req *pb.BatchTrackEventsRequest) (*emptypb.Empty, error) {
	for _, event := range req.Events {
		err := s.app.RecordMetric(
			ctx,
			domain.MetricType("event"),
			event.EventName,
			1.0,
			domain.GranularityHourly,
			"user_id",
			strconv.FormatUint(event.UserId, 10),
		)
		if err != nil {
			continue
		}
	}
	return &emptypb.Empty{}, nil
}

// GetSalesOverviewReport 处理获取销售概览报告的gRPC请求。
func (s *Server) GetSalesOverviewReport(ctx context.Context, req *pb.GetSalesOverviewReportRequest) (*pb.SalesOverviewReportResponse, error) {
	salesQuery := &domain.MetricQuery{
		MetricType: domain.MetricTypeSales,
		StartTime:  req.StartDate.AsTime(),
		EndTime:    req.EndDate.AsTime(),
	}
	salesMetrics, _, err := s.app.QueryMetrics(ctx, salesQuery)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to query sales metrics: %v", err))
	}

	ordersQuery := &domain.MetricQuery{
		MetricType: domain.MetricTypeOrders,
		StartTime:  req.StartDate.AsTime(),
		EndTime:    req.EndDate.AsTime(),
	}
	ordersMetrics, _, err := s.app.QueryMetrics(ctx, ordersQuery)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to query orders metrics: %v", err))
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
	}, nil
}

// --- Unimplemented Methods ---
// 以下gRPC方法均未实现，仅返回Unimplemented错误。

// GetUserActivityReport 获取用户活动报告。
func (s *Server) GetUserActivityReport(ctx context.Context, req *pb.GetUserActivityReportRequest) (*pb.UserActivityReportResponse, error) {
	// 调用应用服务层获取报告数据
	// 这里简化处理，直接返回空响应或模拟数据
	return &pb.UserActivityReportResponse{
		Activities: []*pb.UserActivity{},
		TotalCount: 0,
	}, nil
}

// GetProductPerformanceReport 获取产品性能报告。
func (s *Server) GetProductPerformanceReport(ctx context.Context, req *pb.GetProductPerformanceReportRequest) (*pb.ProductPerformanceReportResponse, error) {
	return &pb.ProductPerformanceReportResponse{
		ProductPerformances: []*pb.ProductPerformance{},
		TotalCount:          0,
	}, nil
}

// GetConversionFunnelReport 获取转化漏斗报告。
func (s *Server) GetConversionFunnelReport(ctx context.Context, req *pb.GetConversionFunnelReportRequest) (*pb.ConversionFunnelReportResponse, error) {
	return &pb.ConversionFunnelReportResponse{
		FunnelData: []*pb.FunnelStepData{},
	}, nil
}

// GetCustomReport 获取自定义报告。
func (s *Server) GetCustomReport(ctx context.Context, req *pb.GetCustomReportRequest) (*pb.CustomReportResponse, error) {
	return &pb.CustomReportResponse{
		Rows:       []*pb.Row{},
		TotalCount: 0,
		Columns:    []string{},
	}, nil
}

// GetUserBehaviorPath 获取用户行为路径。
func (s *Server) GetUserBehaviorPath(ctx context.Context, req *pb.GetUserBehaviorPathRequest) (*pb.UserBehaviorPathResponse, error) {
	return &pb.UserBehaviorPathResponse{
		Paths: []*pb.UserBehaviorPath{},
	}, nil
}

// GetUserSegments 获取用户细分。
func (s *Server) GetUserSegments(ctx context.Context, req *pb.GetUserSegmentsRequest) (*pb.UserSegmentsResponse, error) {
	return &pb.UserSegmentsResponse{
		Segments:   []*pb.UserSegment{},
		TotalCount: 0,
	}, nil
}

// GetRealtimeVisitors 获取实时访客数据。
func (s *Server) GetRealtimeVisitors(ctx context.Context, req *emptypb.Empty) (*pb.RealtimeVisitorsResponse, error) {
	// 这是一个模拟实现，实际应从实时数据源获取访客计数和活跃页面。
	return &pb.RealtimeVisitorsResponse{
		VisitorCount: 0,
		ActivePages:  []string{},
	}, nil
}

// --- 仪表盘与报表管理（Proto 中未实现）---

/*
// CreateDashboard 创建仪表板。
func (s *Server) CreateDashboard(ctx context.Context, req *pb.CreateDashboardRequest) (*pb.DashboardResponse, error) {
	dashboard, err := s.app.CreateDashboard(ctx, req.Name, req.Description, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create dashboard: %v", err))
	}
	return &pb.DashboardResponse{
		Dashboard: &pb.Dashboard{
			Id:          uint64(dashboard.ID),
			Name:        dashboard.Name,
			Description: dashboard.Description,
			UserId:      dashboard.UserID,
		},
	}, nil
}

// GetDashboard 获取仪表板详情。
func (s *Server) GetDashboard(ctx context.Context, req *pb.GetDashboardRequest) (*pb.DashboardResponse, error) {
	dashboard, err := s.app.GetDashboard(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get dashboard: %v", err))
	}
	if dashboard == nil {
		return nil, status.Error(codes.NotFound, "dashboard not found")
	}
	return &pb.DashboardResponse{
		Dashboard: &pb.Dashboard{
			Id:          uint64(dashboard.ID),
			Name:        dashboard.Name,
			Description: dashboard.Description,
			UserId:      dashboard.UserID,
		},
	}, nil
}

// UpdateDashboard 更新仪表板。
func (s *Server) UpdateDashboard(ctx context.Context, req *pb.UpdateDashboardRequest) (*pb.DashboardResponse, error) {
	dashboard, err := s.app.UpdateDashboard(ctx, req.Id, req.Name, req.Description)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update dashboard: %v", err))
	}
	return &pb.DashboardResponse{
		Dashboard: &pb.Dashboard{
			Id:          uint64(dashboard.ID),
			Name:        dashboard.Name,
			Description: dashboard.Description,
			UserId:      dashboard.UserID,
		},
	}, nil
}

// DeleteDashboard 删除仪表板。
func (s *Server) DeleteDashboard(ctx context.Context, req *pb.DeleteDashboardRequest) (*emptypb.Empty, error) {
	if err := s.app.DeleteDashboard(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete dashboard: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListDashboards 列出仪表板。
func (s *Server) ListDashboards(ctx context.Context, req *pb.ListDashboardsRequest) (*pb.ListDashboardsResponse, error) {
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	dashboards, total, err := s.app.ListDashboards(ctx, req.UserId, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list dashboards: %v", err))
	}

	pbDashboards := make([]*pb.Dashboard, len(dashboards))
	for i, d := range dashboards {
		pbDashboards[i] = &pb.Dashboard{
			Id:          uint64(d.ID),
			Name:        d.Name,
			Description: d.Description,
			UserId:      d.UserID,
		}
	}

	return &pb.ListDashboardsResponse{
		Dashboards: pbDashboards,
		TotalCount: int32(total),
	}, nil
}

// CreateReport 创建报告。
func (s *Server) CreateReport(ctx context.Context, req *pb.CreateReportRequest) (*pb.ReportResponse, error) {
	report, err := s.app.CreateReport(ctx, req.Title, req.Description, req.UserId, req.ReportType)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create report: %v", err))
	}
	return &pb.ReportResponse{
		Report: &pb.Report{
			Id:          uint64(report.ID),
			Title:       report.Title,
			Description: report.Description,
			UserId:      report.UserID,
			ReportType:  report.ReportType,
		},
	}, nil
}

// GetReport 获取报告详情。
func (s *Server) GetReport(ctx context.Context, req *pb.GetReportRequest) (*pb.ReportResponse, error) {
	report, err := s.app.GetReport(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get report: %v", err))
	}
	if report == nil {
		return nil, status.Error(codes.NotFound, "report not found")
	}
	return &pb.ReportResponse{
		Report: &pb.Report{
			Id:          uint64(report.ID),
			Title:       report.Title,
			Description: report.Description,
			UserId:      report.UserID,
			ReportType:  report.ReportType,
		},
	}, nil
}

// UpdateReport 更新报告。
func (s *Server) UpdateReport(ctx context.Context, req *pb.UpdateReportRequest) (*pb.ReportResponse, error) {
	report, err := s.app.UpdateReport(ctx, req.Id, req.Title, req.Description)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update report: %v", err))
	}
	return &pb.ReportResponse{
		Report: &pb.Report{
			Id:          uint64(report.ID),
			Title:       report.Title,
			Description: report.Description,
			UserId:      report.UserID,
			ReportType:  report.ReportType,
		},
	}, nil
}

// DeleteReport 删除报告。
func (s *Server) DeleteReport(ctx context.Context, req *pb.DeleteReportRequest) (*emptypb.Empty, error) {
	if err := s.app.DeleteReport(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete report: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListReports 列出报告。
func (s *Server) ListReports(ctx context.Context, req *pb.ListReportsRequest) (*pb.ListReportsResponse, error) {
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	reports, total, err := s.app.ListReports(ctx, req.UserId, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list reports: %v", err))
	}

	pbReports := make([]*pb.Report, len(reports))
	for i, r := range reports {
		pbReports[i] = &pb.Report{
			Id:          uint64(r.ID),
			Title:       r.Title,
			Description: r.Description,
			UserId:      r.UserID,
			ReportType:  r.ReportType,
		}
	}

	return &pb.ListReportsResponse{
		Reports:    pbReports,
		TotalCount: int32(total),
	}, nil
}
*/
