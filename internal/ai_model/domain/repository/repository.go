package repository

import (
	"context"
	"ecommerce/internal/ai_model/domain/entity"
	"time"
)

// AIModelRepository 模型仓储接口
type AIModelRepository interface {
	Create(ctx context.Context, model *entity.AIModel) error
	GetByID(ctx context.Context, id uint64) (*entity.AIModel, error)
	GetByNo(ctx context.Context, no string) (*entity.AIModel, error)
	Update(ctx context.Context, model *entity.AIModel) error
	List(ctx context.Context, query *ModelQuery) ([]*entity.AIModel, int64, error)
	Delete(ctx context.Context, id uint64) error

	// Training Log methods
	CreateTrainingLog(ctx context.Context, log *entity.ModelTrainingLog) error
	ListTrainingLogs(ctx context.Context, modelID uint64) ([]*entity.ModelTrainingLog, error)

	// Prediction methods
	CreatePrediction(ctx context.Context, prediction *entity.ModelPrediction) error
	ListPredictions(ctx context.Context, modelID uint64, startTime, endTime time.Time, page, pageSize int) ([]*entity.ModelPrediction, int64, error)
}

// ModelQuery 查询条件
type ModelQuery struct {
	Status    entity.ModelStatus
	Type      string
	Algorithm string
	CreatorID uint64
	Page      int
	PageSize  int
}
