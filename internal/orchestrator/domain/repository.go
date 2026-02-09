package domain

import (
	"context"
	"time"
)

// SagaInstance 事务实例领域对象
type SagaInstance struct {
	SagaID      string      `json:"saga_id"`
	SagaType    string      `json:"saga_type"`
	BusinessKey string      `json:"business_key"`
	Payload     string      `json:"payload"`
	Status      string      `json:"status"`
	Steps       []*SagaStep `json:"steps"`
	CreatedAt   time.Time   `json:"created_at"`
}

type SagaStep struct {
	StepName   string    `json:"step_name"`
	Status     string    `json:"status"`
	Error      string    `json:"error"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
}

// OrchestratorRepository 仓储接口
type OrchestratorRepository interface {
	SaveInstance(ctx context.Context, instance *SagaInstance) error
	FindInstanceByID(ctx context.Context, sagaID string) (*SagaInstance, error)
	UpdateStep(ctx context.Context, sagaID string, step *SagaStep) error
}
