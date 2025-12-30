package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/goapi/audit/v1"
	"github.com/wyfcoding/ecommerce/internal/audit/application"
	"github.com/wyfcoding/ecommerce/internal/audit/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server 结构体定义。
type Server struct {
	pb.UnimplementedAuditServiceServer
	app *application.Audit
}

// NewServer 创建并返回一个新的 Audit gRPC 服务端实例。
func NewServer(app *application.Audit) *Server {
	return &Server{app: app}
}

// LogEvent 处理记录审计事件的gRPC请求。
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
		domain.AuditEventType(req.EventType),
		req.Module,
		req.Action,
		opts...,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to log audit event: %v", err))
	}

	return &emptypb.Empty{}, nil
}

// QueryLogs 处理查询审计日志的gRPC请求。
func (s *Server) QueryLogs(ctx context.Context, req *pb.QueryLogsRequest) (*pb.QueryLogsResponse, error) {
	query := &domain.AuditLogQuery{
		UserID:    req.UserId,
		Module:    req.Module,
		EventType: domain.AuditEventType(req.EventType),
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
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to query audit logs: %v", err))
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

// CreatePolicy 处理创建审计策略的gRPC请求。
func (s *Server) CreatePolicy(ctx context.Context, req *pb.CreatePolicyRequest) (*pb.CreatePolicyResponse, error) {
	policy, err := s.app.CreatePolicy(ctx, req.Name, req.Description)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create audit policy: %v", err))
	}
	return &pb.CreatePolicyResponse{
		Policy: convertAuditPolicyToProto(policy),
	}, nil
}

// UpdatePolicy 处理更新审计策略的gRPC请求。
func (s *Server) UpdatePolicy(ctx context.Context, req *pb.UpdatePolicyRequest) (*emptypb.Empty, error) {
	if err := s.app.UpdatePolicy(ctx, req.Id, req.EventTypes, req.Modules, req.Enabled); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update audit policy: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListPolicies 处理列出审计策略的gRPC请求。
func (s *Server) ListPolicies(ctx context.Context, req *pb.ListPoliciesRequest) (*pb.ListPoliciesResponse, error) {
	page := max(int(req.PageNum), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	policies, total, err := s.app.ListPolicies(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list audit policies: %v", err))
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

// CreateReport 处理创建审计报告的gRPC请求。
func (s *Server) CreateReport(ctx context.Context, req *pb.CreateReportRequest) (*pb.CreateReportResponse, error) {
	report, err := s.app.CreateReport(ctx, req.Title, req.Description)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create audit report: %v", err))
	}
	return &pb.CreateReportResponse{
		Report: convertAuditReportToProto(report),
	}, nil
}

// GenerateReport 处理生成审计报告内容的gRPC请求。
func (s *Server) GenerateReport(ctx context.Context, req *pb.GenerateReportRequest) (*emptypb.Empty, error) {
	if err := s.app.GenerateReport(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to generate audit report: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListReports 处理列出审计报告的gRPC请求。
func (s *Server) ListReports(ctx context.Context, req *pb.ListReportsRequest) (*pb.ListReportsResponse, error) {
	page := max(int(req.PageNum), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	reports, total, err := s.app.ListReports(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list audit reports: %v", err))
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

func convertAuditLogToProto(l *domain.AuditLog) *pb.AuditLog {
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

func convertAuditPolicyToProto(p *domain.AuditPolicy) *pb.AuditPolicy {
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

func convertAuditReportToProto(r *domain.AuditReport) *pb.AuditReport {
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
