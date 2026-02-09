package domain

import (
	"context"
)

// OrchestratorRepository 仓储接口
type OrchestratorRepository interface {
	SaveInstance(ctx context.Context, instance *SagaInstance) error
	FindInstanceByID(ctx context.Context, sagaID string) (*SagaInstance, error)
	UpdateStep(ctx context.Context, sagaID string, step *SagaStep) error
}
