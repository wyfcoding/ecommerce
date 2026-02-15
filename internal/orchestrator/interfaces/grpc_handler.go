package interfaces

import (
	"context"

	pb "github.com/wyfcoding/ecommerce/go-api/orchestrator/v1"
	"github.com/wyfcoding/ecommerce/internal/orchestrator/application"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrchestratorHandler struct {
	pb.UnimplementedOrchestratorServiceServer
	appService *application.OrchestratorApplicationService
}

func NewOrchestratorHandler(appService *application.OrchestratorApplicationService) *OrchestratorHandler {
	return &OrchestratorHandler{
		appService: appService,
	}
}

func (h *OrchestratorHandler) StartSaga(ctx context.Context, req *pb.StartSagaRequest) (*pb.StartSagaResponse, error) {
	cmd := application.StartSagaCommand{
		SagaType:    req.SagaType,
		BusinessKey: req.BusinessKey,
		Payload:     req.Payload,
	}

	sagaID, err := h.appService.StartSaga(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return &pb.StartSagaResponse{
		SagaId: sagaID,
		Status: "STARTED",
	}, nil
}

func (h *OrchestratorHandler) GetSagaStatus(ctx context.Context, req *pb.GetSagaStatusRequest) (*pb.GetSagaStatusResponse, error) {
	instance, err := h.appService.GetStatus(ctx, req.SagaId)
	if err != nil {
		return nil, err
	}

	steps := make([]*pb.SagaStep, 0, len(instance.Steps))
	for _, s := range instance.Steps {
		var startedAt, finishedAt *timestamppb.Timestamp
		if s.ScheduledAt != nil {
			startedAt = timestamppb.New(*s.ScheduledAt)
		}
		if s.FinishedAt != nil {
			finishedAt = timestamppb.New(*s.FinishedAt)
		}

		steps = append(steps, &pb.SagaStep{
			StepName:   s.Name,
			Status:     string(s.Status),
			Error:      s.Error,
			StartedAt:  startedAt,
			FinishedAt: finishedAt,
		})
	}

	return &pb.GetSagaStatusResponse{
		SagaId:      instance.SagaID,
		SagaType:    instance.SagaType,
		BusinessKey: instance.OriginalRefID,
		Status:      string(instance.Status),
		Steps:       steps,
		CreatedAt:   timestamppb.New(instance.StartTime),
	}, nil
}
