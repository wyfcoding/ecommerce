package grpc

import (
	"context"
	pb "ecommerce/api/audit/v1"
	"ecommerce/internal/audit/application"
	"ecommerce/internal/audit/domain/entity"
	"ecommerce/internal/audit/domain/repository"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedAuditServiceServer
	app *application.AuditService
}

func NewServer(app *application.AuditService) *Server {
	return &Server{app: app}
}

func (s *Server) LogEvent(ctx context.Context, req *pb.LogEventRequest) (*emptypb.Empty, error) {
	var opts []application.LogOption
	if req.ResourceType != "" || req.ResourceId != "" {
		opts = append(opts, application.WithResource(req.ResourceType, req.ResourceId))
	}
	if req.OldValue != "" || req.NewValue != "" {
		opts = append(opts, application.WithChange(req.OldValue, req.NewValue))
	}
	if req.Ip != "" || req.UserAgent != "" {
		opts = append(opts, application.WithClientInfo(req.Ip, req.UserAgent))
	}
	if req.Duration > 0 {
		opts = append(opts, application.WithDuration(req.Duration))
	}
	if req.ErrorMsg != "" {
		opts = append(opts, application.WithError(req.ErrorMsg))
	}

	err := s.app.LogEvent(
		ctx,
		req.UserId,
		req.Username,
		entity.AuditEventType(req.EventType),
		req.Module,
		req.Action,
		opts...,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) QueryLogs(ctx context.Context, req *pb.QueryLogsRequest) (*pb.QueryLogsResponse, error) {
	query := &repository.AuditLogQuery{
		UserID:    req.UserId,
		Module:    req.Module,
		EventType: entity.AuditEventType(req.EventType),
		StartTime: req.StartTime.AsTime(),
		EndTime:   req.EndTime.AsTime(),
		Page:      int(req.PageNum),
		PageSize:  int(req.PageSize),
	}
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 10
	}

	logs, total, err := s.app.QueryLogs(ctx, query)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbLogs := make([]*pb.AuditLog, len(logs))
	for i, l := range logs {
		pbLogs[i] = convertAuditLogToProto(l)
	}

	return &pb.QueryLogsResponse{
		Logs:       pbLogs,
		TotalCount: uint64(total),
	}, nil
}

func (s *Server) CreatePolicy(ctx context.Context, req *pb.CreatePolicyRequest) (*pb.CreatePolicyResponse, error) {
	policy, err := s.app.CreatePolicy(ctx, req.Name, req.Description)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.CreatePolicyResponse{
		Policy: convertAuditPolicyToProto(policy),
	}, nil
}

func (s *Server) UpdatePolicy(ctx context.Context, req *pb.UpdatePolicyRequest) (*emptypb.Empty, error) {
	if err := s.app.UpdatePolicy(ctx, req.Id, req.EventTypes, req.Modules, req.Enabled); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListPolicies(ctx context.Context, req *pb.ListPoliciesRequest) (*pb.ListPoliciesResponse, error) {
	page := int(req.PageNum)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	policies, total, err := s.app.ListPolicies(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbPolicies := make([]*pb.AuditPolicy, len(policies))
	for i, p := range policies {
		pbPolicies[i] = convertAuditPolicyToProto(p)
	}

	return &pb.ListPoliciesResponse{
		Policies:   pbPolicies,
		TotalCount: uint64(total),
	}, nil
}

func (s *Server) CreateReport(ctx context.Context, req *pb.CreateReportRequest) (*pb.CreateReportResponse, error) {
	report, err := s.app.CreateReport(ctx, req.Title, req.Description)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.CreateReportResponse{
		Report: convertAuditReportToProto(report),
	}, nil
}

func (s *Server) GenerateReport(ctx context.Context, req *pb.GenerateReportRequest) (*emptypb.Empty, error) {
	if err := s.app.GenerateReport(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListReports(ctx context.Context, req *pb.ListReportsRequest) (*pb.ListReportsResponse, error) {
	page := int(req.PageNum)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	reports, total, err := s.app.ListReports(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbReports := make([]*pb.AuditReport, len(reports))
	for i, r := range reports {
		pbReports[i] = convertAuditReportToProto(r)
	}

	return &pb.ListReportsResponse{
		Reports:    pbReports,
		TotalCount: uint64(total),
	}, nil
}

func convertAuditLogToProto(l *entity.AuditLog) *pb.AuditLog {
	if l == nil {
		return nil
	}
	return &pb.AuditLog{
		Id:           uint64(l.ID),
		AuditNo:      l.AuditNo,
		UserId:       l.UserID,
		Username:     l.Username,
		EventType:    string(l.EventType),
		Level:        string(l.Level),
		Module:       l.Module,
		Action:       l.Action,
		ResourceType: l.ResourceType,
		ResourceId:   l.ResourceID,
		OldValue:     l.OldValue,
		NewValue:     l.NewValue,
		Ip:           l.IP,
		UserAgent:    l.UserAgent,
		Status:       l.Status,
		ErrorMsg:     l.ErrorMsg,
		Duration:     l.Duration,
		Timestamp:    timestamppb.New(l.Timestamp),
	}
}

func convertAuditPolicyToProto(p *entity.AuditPolicy) *pb.AuditPolicy {
	if p == nil {
		return nil
	}
	return &pb.AuditPolicy{
		Id:            uint64(p.ID),
		Name:          p.Name,
		Description:   p.Description,
		EventTypes:    p.EventTypes,
		Modules:       p.Modules,
		Enabled:       p.Enabled,
		RetentionDays: p.RetentionDays,
	}
}

func convertAuditReportToProto(r *entity.AuditReport) *pb.AuditReport {
	if r == nil {
		return nil
	}
	resp := &pb.AuditReport{
		Id:          uint64(r.ID),
		ReportNo:    r.ReportNo,
		Title:       r.Title,
		Description: r.Description,
		Status:      r.Status,
		CreatedAt:   timestamppb.New(r.CreatedAt),
	}
	if r.GeneratedAt != nil {
		resp.GeneratedAt = timestamppb.New(*r.GeneratedAt)
	}
	return resp
}
