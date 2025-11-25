package application

import (
	"context"
	"fmt"
	"time"

	"github.com/wyfcoding/ecommerce/internal/ai_model/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/ai_model/domain/repository"
	"github.com/wyfcoding/ecommerce/pkg/idgen"

	"log/slog"
)

type AIModelService struct {
	repo        repository.AIModelRepository
	idGenerator idgen.Generator
	logger      *slog.Logger
}

func NewAIModelService(repo repository.AIModelRepository, idGenerator idgen.Generator, logger *slog.Logger) *AIModelService {
	return &AIModelService{
		repo:        repo,
		idGenerator: idGenerator,
		logger:      logger,
	}
}

// CreateModel 创建模型
func (s *AIModelService) CreateModel(ctx context.Context, name, description, modelType, algorithm string, creatorID uint64) (*entity.AIModel, error) {
	modelNo := fmt.Sprintf("AIM%d", s.idGenerator.Generate())
	model := entity.NewAIModel(modelNo, name, description, modelType, algorithm, creatorID)

	if err := s.repo.Create(ctx, model); err != nil {
		s.logger.Error("failed to create model", "error", err)
		return nil, err
	}

	return model, nil
}

// StartTraining 开始训练
func (s *AIModelService) StartTraining(ctx context.Context, id uint64) error {
	model, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	model.StartTraining()
	return s.repo.Update(ctx, model)
}

// CompleteTraining 完成训练
func (s *AIModelService) CompleteTraining(ctx context.Context, id uint64, accuracy float64, modelPath string) error {
	model, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	model.CompleteTraining(accuracy, modelPath)
	return s.repo.Update(ctx, model)
}

// FailTraining 训练失败
func (s *AIModelService) FailTraining(ctx context.Context, id uint64, reason string) error {
	model, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	model.FailTraining(reason)
	return s.repo.Update(ctx, model)
}

// Deploy 部署模型
func (s *AIModelService) Deploy(ctx context.Context, id uint64) error {
	model, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	model.Deploy()
	return s.repo.Update(ctx, model)
}

// CompleteDeployment 完成部署
func (s *AIModelService) CompleteDeployment(ctx context.Context, id uint64) error {
	model, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	model.CompleteDeployment()
	return s.repo.Update(ctx, model)
}

// Archive 归档模型
func (s *AIModelService) Archive(ctx context.Context, id uint64) error {
	model, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	model.Archive()
	return s.repo.Update(ctx, model)
}

// ListModels 获取模型列表
func (s *AIModelService) ListModels(ctx context.Context, query *repository.ModelQuery) ([]*entity.AIModel, int64, error) {
	return s.repo.List(ctx, query)
}

// GetModelDetails 获取模型详情
func (s *AIModelService) GetModelDetails(ctx context.Context, id uint64) (*entity.AIModel, error) {
	return s.repo.GetByID(ctx, id)
}

// AddTrainingLog 添加训练日志
func (s *AIModelService) AddTrainingLog(ctx context.Context, modelID uint64, iteration int32, loss, accuracy, valLoss, valAccuracy float64) error {
	log := &entity.ModelTrainingLog{
		ModelID:            modelID,
		Iteration:          iteration,
		Loss:               loss,
		Accuracy:           accuracy,
		ValidationLoss:     valLoss,
		ValidationAccuracy: valAccuracy,
	}
	return s.repo.CreateTrainingLog(ctx, log)
}

// Predict 预测
func (s *AIModelService) Predict(ctx context.Context, modelID uint64, input string, userID uint64) (string, float64, error) {
	model, err := s.repo.GetByID(ctx, modelID)
	if err != nil {
		return "", 0, err
	}

	if model.Status != entity.ModelStatusDeployed {
		return "", 0, fmt.Errorf("model is not deployed")
	}

	// Mock prediction logic
	output := "mock_prediction_result"
	confidence := 0.95

	prediction := &entity.ModelPrediction{
		ModelID:        modelID,
		Input:          input,
		Output:         output,
		Confidence:     confidence,
		UserID:         userID,
		PredictionTime: time.Now(),
	}

	if err := s.repo.CreatePrediction(ctx, prediction); err != nil {
		s.logger.Warn("failed to save prediction record", "error", err)
	}

	return output, confidence, nil
}
