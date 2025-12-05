package grpc

import (
	"context" // 导入上下文。
	"fmt"     // 导入格式化库。

	pb "github.com/wyfcoding/ecommerce/go-api/scheduler/v1"           // 导入调度模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/scheduler/application"   // 导入调度模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/scheduler/domain/entity" // 导入调度模块的领域实体。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/emptypb"     // 导入空消息类型。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 SchedulerService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedSchedulerServiceServer                               // 嵌入生成的UnimplementedSchedulerServiceServer，确保前向兼容性。
	app                                    *application.SchedulerService // 依赖Scheduler应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 Scheduler gRPC 服务端实例。
func NewServer(app *application.SchedulerService) *Server {
	return &Server{app: app}
}

// CreateJob 处理创建任务的gRPC请求。
// req: 包含任务名称、描述、Cron表达式、处理器和参数的请求体。
// 返回创建成功的任务响应和可能发生的gRPC错误。
func (s *Server) CreateJob(ctx context.Context, req *pb.CreateJobRequest) (*pb.CreateJobResponse, error) {
	job, err := s.app.CreateJob(ctx, req.Name, req.Description, req.CronExpr, req.Handler, req.Params)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create job: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return &pb.CreateJobResponse{
		Job: convertJobToProto(job),
	}, nil
}

// UpdateJob 处理更新任务的gRPC请求。
// req: 包含任务ID、新的Cron表达式和参数的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) UpdateJob(ctx context.Context, req *pb.UpdateJobRequest) (*emptypb.Empty, error) {
	if err := s.app.UpdateJob(ctx, req.Id, req.CronExpr, req.Params); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update job: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ToggleJobStatus 处理切换任务状态的gRPC请求。
// req: 包含任务ID和启用/禁用标志的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) ToggleJobStatus(ctx context.Context, req *pb.ToggleJobStatusRequest) (*emptypb.Empty, error) {
	if err := s.app.ToggleJobStatus(ctx, req.Id, req.Enable); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to toggle job status: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// RunJob 处理立即运行任务的gRPC请求。
// req: 包含任务ID的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) RunJob(ctx context.Context, req *pb.RunJobRequest) (*emptypb.Empty, error) {
	if err := s.app.RunJob(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to run job: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListJobs 处理列出任务的gRPC请求。
// req: 包含分页参数和状态过滤的请求体。
// 返回任务列表响应和可能发生的gRPC错误。
func (s *Server) ListJobs(ctx context.Context, req *pb.ListJobsRequest) (*pb.ListJobsResponse, error) {
	// 获取分页参数。
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 根据Proto的状态值构建过滤器。
	var statusPtr *int
	if req.Status != -1 { // -1通常表示不进行状态过滤。
		st := int(req.Status)
		statusPtr = &st
	}

	// 调用应用服务层获取任务列表。
	jobs, total, err := s.app.ListJobs(ctx, statusPtr, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list jobs: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbJobs := make([]*pb.Job, len(jobs))
	for i, j := range jobs {
		pbJobs[i] = convertJobToProto(j)
	}

	return &pb.ListJobsResponse{
		Jobs:       pbJobs,
		TotalCount: total, // 总记录数。
	}, nil
}

// ListJobLogs 处理列出任务日志的gRPC请求。
// req: 包含任务ID过滤和分页参数的请求体。
// 返回任务日志列表响应和可能发生的gRPC错误。
func (s *Server) ListJobLogs(ctx context.Context, req *pb.ListJobLogsRequest) (*pb.ListJobLogsResponse, error) {
	// 获取分页参数。
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 调用应用服务层获取任务日志列表。
	logs, total, err := s.app.ListJobLogs(ctx, req.JobId, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list job logs: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbLogs := make([]*pb.JobLog, len(logs))
	for i, l := range logs {
		pbLogs[i] = convertJobLogToProto(l)
	}

	return &pb.ListJobLogsResponse{
		Logs:       pbLogs,
		TotalCount: total, // 总记录数。
	}, nil
}

// convertJobToProto 是一个辅助函数，将领域层的 Job 实体转换为 protobuf 的 Job 消息。
func convertJobToProto(j *entity.Job) *pb.Job {
	if j == nil {
		return nil
	}
	// 转换可选的时间字段。
	var lastRunTime, nextRunTime *timestamppb.Timestamp
	if j.LastRunTime != nil {
		lastRunTime = timestamppb.New(*j.LastRunTime)
	}
	if j.NextRunTime != nil {
		nextRunTime = timestamppb.New(*j.NextRunTime)
	}

	return &pb.Job{
		Id:          uint64(j.ID),                 // ID。
		Name:        j.Name,                       // 名称。
		Description: j.Description,                // 描述。
		CronExpr:    j.CronExpr,                   // Cron表达式。
		Handler:     j.Handler,                    // 处理器。
		Params:      j.Params,                     // 参数。
		Status:      int32(j.Status),              // 状态。
		LastRunTime: lastRunTime,                  // 上次运行时间。
		NextRunTime: nextRunTime,                  // 下次运行时间。
		RunCount:    j.RunCount,                   // 运行次数。
		FailCount:   j.FailCount,                  // 失败次数。
		CreatedAt:   timestamppb.New(j.CreatedAt), // 创建时间。
		UpdatedAt:   timestamppb.New(j.UpdatedAt), // 更新时间。
	}
}

// convertJobLogToProto 是一个辅助函数，将领域层的 JobLog 实体转换为 protobuf 的 JobLog 消息。
func convertJobLogToProto(l *entity.JobLog) *pb.JobLog {
	if l == nil {
		return nil
	}
	// 转换可选的结束时间字段。
	var endTime *timestamppb.Timestamp
	if l.EndTime != nil {
		endTime = timestamppb.New(*l.EndTime)
	}

	return &pb.JobLog{
		Id:        uint64(l.ID),                 // ID。
		JobId:     l.JobID,                      // 任务ID。
		JobName:   l.JobName,                    // 任务名称。
		Handler:   l.Handler,                    // 处理器。
		Params:    l.Params,                     // 参数。
		Status:    l.Status,                     // 状态。
		Result:    l.Result,                     // 结果。
		Error:     l.Error,                      // 错误信息。
		Duration:  l.Duration,                   // 耗时。
		StartTime: timestamppb.New(l.StartTime), // 开始时间。
		EndTime:   endTime,                      // 结束时间。
		CreatedAt: timestamppb.New(l.CreatedAt), // 创建时间。
		UpdatedAt: timestamppb.New(l.UpdatedAt), // 更新时间。
	}
}
