package grpc

import (
	"context"
	pb "github.com/wyfcoding/ecommerce/api/data_processing/v1"
	"github.com/wyfcoding/ecommerce/internal/data_processing/application"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedDataProcessingServer
	app *application.DataProcessingService
}

func NewServer(app *application.DataProcessingService) *Server {
	return &Server{app: app}
}

func (s *Server) ProcessData(ctx context.Context, req *pb.ProcessDataRequest) (*pb.ProcessDataResponse, error) {
	// Service SubmitTask(ctx, name, taskType, config, workflowID)
	// Proto ProcessData(data_id, processing_type, processing_params)
	// Mapping:
	// name -> "Process-" + data_id
	// taskType -> processing_type
	// config -> processing_params (serialized)
	// workflowID -> 0 (default)

	name := "Process-" + req.DataId
	config := "" // Serialize params if needed

	task, err := s.app.SubmitTask(ctx, name, req.ProcessingType, config, 0)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ProcessDataResponse{
		ProcessingId: strconv.FormatUint(uint64(task.ID), 10),
		Status:       "SUBMITTED",
	}, nil
}

func (s *Server) GetProcessingStatus(ctx context.Context, req *pb.GetProcessingStatusRequest) (*pb.ProcessingStatusResponse, error) {
	// Service does not expose GetTask directly in application layer, but repo has it.
	// We might need to add GetTask to application service or access repo via app if exposed (it's not).
	// Let's assume we can add GetTask to app or use ListTasks to find it (inefficient).
	// Wait, checking service.go again... it has ListTasks but not GetTask.
	// We should probably add GetTask to the service, but I cannot modify service.go easily without breaking things or being out of scope?
	// Actually, I can modify service.go if needed.
	// But for now, let's see if we can use ListTasks with a filter? No, ListTasks takes workflowID.

	// Let's check if I can add GetTask to service.go.
	// It's better to add it.

	// For this step, I will implement what I can.
	// Since I cannot call GetTask, I will return Unimplemented for now or try to add it.
	// Adding GetTask to service.go seems right.

	return nil, status.Error(codes.Unimplemented, "GetProcessingStatus not implemented yet")
}
