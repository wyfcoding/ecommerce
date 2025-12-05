package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/go-api/audit/v1"               // 导入审计模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/audit/application"       // 导入审计模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/audit/domain/entity"     // 导入审计模块的领域实体。
	"github.com/wyfcoding/ecommerce/internal/audit/domain/repository" // 导入审计模块的仓储层查询对象。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/emptypb"     // 导入空消息类型。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 AuditService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedAuditServiceServer                           // 嵌入生成的UnimplementedAuditServiceServer，确保前向兼容性。
	app                                *application.AuditService // 依赖Audit应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 Audit gRPC 服务端实例。
func NewServer(app *application.AuditService) *Server {
	return &Server{app: app}
}

// LogEvent 处理记录审计事件的gRPC请求。
// req: 包含审计事件信息的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) LogEvent(ctx context.Context, req *pb.LogEventRequest) (*emptypb.Empty, error) {
	var opts []application.LogOption
	// 根据请求中提供的字段，动态构建LogOption列表。
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

	// 调用应用服务层记录审计事件。
	err := s.app.LogEvent(
		ctx,
		req.UserId,
		req.Username,
		entity.AuditEventType(req.EventType), // 映射protobuf事件类型到领域实体事件类型。
		req.Module,
		req.Action,
		opts..., // 传递动态构建的LogOption列表。
	)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to log audit event: %v", err))
	}

	return &emptypb.Empty{}, nil
}

// QueryLogs 处理查询审计日志的gRPC请求。
// req: 包含查询条件和分页参数的请求体。
// 返回审计日志列表响应和可能发生的gRPC错误。
func (s *Server) QueryLogs(ctx context.Context, req *pb.QueryLogsRequest) (*pb.QueryLogsResponse, error) {
	// 将protobuf请求中的查询参数映射到应用服务层使用的查询对象。
	query := &repository.AuditLogQuery{
		UserID:    req.UserId,
		Module:    req.Module,
		EventType: entity.AuditEventType(req.EventType),
		StartTime: req.StartTime.AsTime(),
		EndTime:   req.EndTime.AsTime(),
		Page:      int(req.PageNum),
		PageSize:  int(req.PageSize),
	}
	// 确保分页参数有效。
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 10
	}

	// 调用应用服务层查询审计日志。
	logs, total, err := s.app.QueryLogs(ctx, query)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to query audit logs: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbLogs := make([]*pb.AuditLog, len(logs))
	for i, l := range logs {
		pbLogs[i] = convertAuditLogToProto(l) // 使用辅助函数进行转换。
	}

	return &pb.QueryLogsResponse{
		Logs:       pbLogs,
		TotalCount: uint64(total), // 总记录数。
	}, nil
}

// CreatePolicy 处理创建审计策略的gRPC请求。
// req: 包含策略名称和描述的请求体。
// 返回created successfully的审计策略响应和可能发生的gRPC错误。
func (s *Server) CreatePolicy(ctx context.Context, req *pb.CreatePolicyRequest) (*pb.CreatePolicyResponse, error) {
	// 调用应用服务层创建审计策略。
	policy, err := s.app.CreatePolicy(ctx, req.Name, req.Description)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create audit policy: %v", err))
	}
	return &pb.CreatePolicyResponse{
		Policy: convertAuditPolicyToProto(policy), // 使用辅助函数进行转换。
	}, nil
}

// UpdatePolicy 处理更新审计策略的gRPC请求。
// req: 包含策略ID、事件类型列表、模块列表和启用状态的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) UpdatePolicy(ctx context.Context, req *pb.UpdatePolicyRequest) (*emptypb.Empty, error) {
	// 调用应用服务层更新审计策略。
	if err := s.app.UpdatePolicy(ctx, req.Id, req.EventTypes, req.Modules, req.Enabled); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update audit policy: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListPolicies 处理列出审计策略的gRPC请求。
// req: 包含分页参数的请求体。
// 返回审计策略列表响应和可能发生的gRPC错误。
func (s *Server) ListPolicies(ctx context.Context, req *pb.ListPoliciesRequest) (*pb.ListPoliciesResponse, error) {
	// 获取分页参数。
	page := int(req.PageNum)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 调用应用服务层获取审计策略列表。
	policies, total, err := s.app.ListPolicies(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list audit policies: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbPolicies := make([]*pb.AuditPolicy, len(policies))
	for i, p := range policies {
		pbPolicies[i] = convertAuditPolicyToProto(p) // 使用辅助函数进行转换。
	}

	return &pb.ListPoliciesResponse{
		Policies:   pbPolicies,
		TotalCount: uint64(total), // 总记录数。
	}, nil
}

// CreateReport 处理创建审计报告的gRPC请求。
// req: 包含报告标题和描述的请求体。
// 返回created successfully的审计报告响应和可能发生的gRPC错误。
func (s *Server) CreateReport(ctx context.Context, req *pb.CreateReportRequest) (*pb.CreateReportResponse, error) {
	// 调用应用服务层创建审计报告。
	report, err := s.app.CreateReport(ctx, req.Title, req.Description)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create audit report: %v", err))
	}
	return &pb.CreateReportResponse{
		Report: convertAuditReportToProto(report), // 使用辅助函数进行转换。
	}, nil
}

// GenerateReport 处理生成审计报告内容的gRPC请求。
// req: 包含报告ID的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) GenerateReport(ctx context.Context, req *pb.GenerateReportRequest) (*emptypb.Empty, error) {
	// 调用应用服务层生成报告内容。
	if err := s.app.GenerateReport(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to generate audit report: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListReports 处理列出审计报告的gRPC请求。
// req: 包含分页参数的请求体。
// 返回审计报告列表响应和可能发生的gRPC错误。
func (s *Server) ListReports(ctx context.Context, req *pb.ListReportsRequest) (*pb.ListReportsResponse, error) {
	// 获取分页参数。
	page := int(req.PageNum)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 调用应用服务层获取审计报告列表。
	reports, total, err := s.app.ListReports(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list audit reports: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbReports := make([]*pb.AuditReport, len(reports))
	for i, r := range reports {
		pbReports[i] = convertAuditReportToProto(r) // 使用辅助函数进行转换。
	}

	return &pb.ListReportsResponse{
		Reports:    pbReports,
		TotalCount: uint64(total), // 总记录数。
	}, nil
}

// convertAuditLogToProto 是一个辅助函数，将领域层的 AuditLog 实体转换为 protobuf 的 AuditLog 消息。
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

// convertAuditPolicyToProto 是一个辅助函数，将领域层的 AuditPolicy 实体转换为 protobuf 的 AuditPolicy 消息。
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

// convertAuditReportToProto 是一个辅助函数，将领域层的 AuditReport 实体转换为 protobuf 的 AuditReport 消息。
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
