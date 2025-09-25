package service

import (
	"context"
	v1 "ecommerce/api/data_processing/v1"
	"ecommerce/internal/data_processing/biz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DataProcessingService is the gRPC service implementation for data processing.
type DataProcessingService struct {
	v1.UnimplementedDataProcessingServiceServer
	uc *biz.DataProcessingUsecase
}

// NewDataProcessingService creates a new DataProcessingService.
func NewDataProcessingService(uc *biz.DataProcessingUsecase) *DataProcessingService {
	return &DataProcessingService{uc: uc}
}

// TriggerProcessingJob implements the TriggerProcessingJob RPC.
func (s *DataProcessingService) TriggerProcessingJob(ctx context.Context, req *v1.TriggerProcessingJobRequest) (*v1.TriggerProcessingJobResponse, error) {
	if req.JobType == "" {
		return nil, status.Error(codes.InvalidArgument, "job_type is required")
	}

	parameters := make(map[string]string)
	for k, v := range req.Parameters {
		parameters[k] = v
	}

	job, err := s.uc.TriggerProcessingJob(ctx, req.JobType, parameters)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to trigger processing job: %v", err)
	}

	return &v1.TriggerProcessingJobResponse{
		JobId:  job.JobID,
		Status: job.Status,
	}, nil
}

// TriggerSparkFlinkJob implements the TriggerSparkFlinkJob RPC.
func (s *DataProcessingService) TriggerSparkFlinkJob(ctx context.Context, req *v1.TriggerSparkFlinkJobRequest) (*v1.TriggerSparkFlinkJobResponse, error) {
	if req.JobName == "" || req.Platform == "" {
		return nil, status.Error(codes.InvalidArgument, "job_name and platform are required")
	}

	jobParameters := make(map[string]string)
	for k, v := range req.JobParameters {
		jobParameters[k] = v
	}

	job, err := s.uc.TriggerSparkFlinkJob(ctx, req.JobName, jobParameters, req.Platform)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to trigger Spark/Flink job: %v", err)
	}

	return &v1.TriggerSparkFlinkJobResponse{
		JobId:   job.JobID,
		Status:  job.Status,
		Message: "Job submitted successfully",
	}, nil
}
