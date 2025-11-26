package grpc

import (
	"context"

	pb "github.com/wyfcoding/ecommerce/api/scheduler/v1"
	"github.com/wyfcoding/ecommerce/internal/scheduler/application"
	"github.com/wyfcoding/ecommerce/internal/scheduler/domain/entity"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedSchedulerServiceServer
	app *application.SchedulerService
}

func NewServer(app *application.SchedulerService) *Server {
	return &Server{app: app}
}

func (s *Server) CreateJob(ctx context.Context, req *pb.CreateJobRequest) (*pb.CreateJobResponse, error) {
	job, err := s.app.CreateJob(ctx, req.Name, req.Description, req.CronExpr, req.Handler, req.Params)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateJobResponse{
		Job: convertJobToProto(job),
	}, nil
}

func (s *Server) UpdateJob(ctx context.Context, req *pb.UpdateJobRequest) (*emptypb.Empty, error) {
	if err := s.app.UpdateJob(ctx, req.Id, req.CronExpr, req.Params); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ToggleJobStatus(ctx context.Context, req *pb.ToggleJobStatusRequest) (*emptypb.Empty, error) {
	if err := s.app.ToggleJobStatus(ctx, req.Id, req.Enable); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) RunJob(ctx context.Context, req *pb.RunJobRequest) (*emptypb.Empty, error) {
	if err := s.app.RunJob(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListJobs(ctx context.Context, req *pb.ListJobsRequest) (*pb.ListJobsResponse, error) {
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	var statusPtr *int
	if req.Status != -1 {
		st := int(req.Status)
		statusPtr = &st
	}

	jobs, total, err := s.app.ListJobs(ctx, statusPtr, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbJobs := make([]*pb.Job, len(jobs))
	for i, j := range jobs {
		pbJobs[i] = convertJobToProto(j)
	}

	return &pb.ListJobsResponse{
		Jobs:       pbJobs,
		TotalCount: total,
	}, nil
}

func (s *Server) ListJobLogs(ctx context.Context, req *pb.ListJobLogsRequest) (*pb.ListJobLogsResponse, error) {
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	logs, total, err := s.app.ListJobLogs(ctx, req.JobId, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbLogs := make([]*pb.JobLog, len(logs))
	for i, l := range logs {
		pbLogs[i] = convertJobLogToProto(l)
	}

	return &pb.ListJobLogsResponse{
		Logs:       pbLogs,
		TotalCount: total,
	}, nil
}

func convertJobToProto(j *entity.Job) *pb.Job {
	if j == nil {
		return nil
	}
	var lastRunTime, nextRunTime *timestamppb.Timestamp
	if j.LastRunTime != nil {
		lastRunTime = timestamppb.New(*j.LastRunTime)
	}
	if j.NextRunTime != nil {
		nextRunTime = timestamppb.New(*j.NextRunTime)
	}

	return &pb.Job{
		Id:          uint64(j.ID),
		Name:        j.Name,
		Description: j.Description,
		CronExpr:    j.CronExpr,
		Handler:     j.Handler,
		Params:      j.Params,
		Status:      int32(j.Status),
		LastRunTime: lastRunTime,
		NextRunTime: nextRunTime,
		RunCount:    j.RunCount,
		FailCount:   j.FailCount,
		CreatedAt:   timestamppb.New(j.CreatedAt),
		UpdatedAt:   timestamppb.New(j.UpdatedAt),
	}
}

func convertJobLogToProto(l *entity.JobLog) *pb.JobLog {
	if l == nil {
		return nil
	}
	var endTime *timestamppb.Timestamp
	if l.EndTime != nil {
		endTime = timestamppb.New(*l.EndTime)
	}

	return &pb.JobLog{
		Id:        uint64(l.ID),
		JobId:     l.JobID,
		JobName:   l.JobName,
		Handler:   l.Handler,
		Params:    l.Params,
		Status:    l.Status,
		Result:    l.Result,
		Error:     l.Error,
		Duration:  l.Duration,
		StartTime: timestamppb.New(l.StartTime),
		EndTime:   endTime,
		CreatedAt: timestamppb.New(l.CreatedAt),
		UpdatedAt: timestamppb.New(l.UpdatedAt),
	}
}
