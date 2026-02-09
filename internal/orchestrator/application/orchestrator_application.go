package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/orchestrator/domain"
	"github.com/wyfcoding/pkg/saga"
)

type StartSagaCommand struct {
	SagaType    string
	BusinessKey string
	Payload     string
}

type OrchestratorApplicationService struct {
	repo   domain.OrchestratorRepository
	engine *saga.Engine // 引用 pkg/saga 中的核心逻辑
	logger *slog.Logger
}

func NewOrchestratorApplicationService(repo domain.OrchestratorRepository, engine *saga.Engine, logger *slog.Logger) *OrchestratorApplicationService {
	return &OrchestratorApplicationService{
		repo:   repo,
		engine: engine,
		logger: logger,
	}
}

func (s *OrchestratorApplicationService) StartSaga(ctx context.Context, cmd StartSagaCommand) (string, error) {
	sagaID := fmt.Sprintf("SAGA-%d", time.Now().UnixNano())
	s.logger.Info("starting new saga transaction", "saga_id", sagaID, "type", cmd.SagaType)

	instance := &domain.SagaInstance{
		ID:            sagaID,
		SagaType:      cmd.SagaType,
		OriginalRefID: cmd.BusinessKey,
		ContextData:   cmd.Payload,
		Status:        domain.SagaStarted,
		StartTime:     time.Now(),
	}

	if err := s.repo.SaveInstance(ctx, instance); err != nil {
		return "", err
	}

	// 调用 Saga 引擎开始执行
	go func() {
		// 引擎执行逻辑...
		s.logger.Info("saga execution started in background", "saga_id", sagaID)
	}()

	return sagaID, nil
}

func (s *OrchestratorApplicationService) GetStatus(ctx context.Context, sagaID string) (*domain.SagaInstance, error) {
	return s.repo.FindInstanceByID(ctx, sagaID)
}
