package infrastructure

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/orchestrator/domain"
	"gorm.io/gorm"
)

type OrchestratorRepositoryImpl struct {
	db *gorm.DB
}

func NewOrchestratorRepository(db *gorm.DB) domain.OrchestratorRepository {
	return &OrchestratorRepositoryImpl{db: db}
}

func (r *OrchestratorRepositoryImpl) SaveInstance(ctx context.Context, instance *domain.SagaInstance) error {
	return r.db.WithContext(ctx).Save(instance).Error
}

func (r *OrchestratorRepositoryImpl) FindInstanceByID(ctx context.Context, sagaID string) (*domain.SagaInstance, error) {
	var instance domain.SagaInstance
	if err := r.db.WithContext(ctx).Preload("Steps").Where("saga_id = ?", sagaID).First(&instance).Error; err != nil {
		return nil, err
	}
	return &instance, nil
}

func (r *OrchestratorRepositoryImpl) UpdateStep(ctx context.Context, sagaID string, step *domain.SagaStep) error {
	// 实现复杂的步骤更新逻辑，通常涉及步骤表的保存
	return nil
}
