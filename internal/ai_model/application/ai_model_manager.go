package application

import (
	"context"
	"fmt"
	"time"

	"github.com/wyfcoding/ecommerce/internal/ai_model/domain"
	"github.com/wyfcoding/pkg/idgen"

	"log/slog"
)

// AIModelManager 负责AI模型模块的写操作和业务逻辑。
type AIModelManager struct {
	repo        domain.AIModelRepository
	idGenerator idgen.Generator
	logger      *slog.Logger
}

// NewAIModelManager 创建一个新的 AIModelManager 实例。
func NewAIModelManager(repo domain.AIModelRepository, idGenerator idgen.Generator, logger *slog.Logger) *AIModelManager {
	return &AIModelManager{
		repo:        repo,
		idGenerator: idGenerator,
		logger:      logger,
	}
}

// CreateModel 创建一个新的AI模型记录。
func (m *AIModelManager) CreateModel(ctx context.Context, name, description, modelType, algorithm string, creatorID uint64) (*domain.AIModel, error) {
	modelNo := fmt.Sprintf("AIM%d", m.idGenerator.Generate())
	model := domain.NewAIModel(modelNo, name, description, modelType, algorithm, creatorID)

	if err := m.repo.Create(ctx, model); err != nil {
		m.logger.ErrorContext(ctx, "failed to create model", "name", name, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "model created successfully", "model_id", model.ID, "name", name)

	return model, nil
}

// StartTraining 启动训练。
func (m *AIModelManager) StartTraining(ctx context.Context, id uint64) error {
	model, err := m.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	model.StartTraining()
	return m.repo.Update(ctx, model)
}

// CompleteTraining 完成训练。
func (m *AIModelManager) CompleteTraining(ctx context.Context, id uint64, accuracy float64, modelPath string) error {
	model, err := m.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	model.CompleteTraining(accuracy, modelPath)
	return m.repo.Update(ctx, model)
}

// FailTraining 训练失败。
func (m *AIModelManager) FailTraining(ctx context.Context, id uint64, reason string) error {
	model, err := m.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	model.FailTraining(reason)
	return m.repo.Update(ctx, model)
}

// Deploy 部署模型。
func (m *AIModelManager) Deploy(ctx context.Context, id uint64) error {
	model, err := m.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	model.Deploy()
	return m.repo.Update(ctx, model)
}

// CompleteDeployment 完成部署。
func (m *AIModelManager) CompleteDeployment(ctx context.Context, id uint64) error {
	model, err := m.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	model.CompleteDeployment()
	return m.repo.Update(ctx, model)
}

// Archive 归档模型。
func (m *AIModelManager) Archive(ctx context.Context, id uint64) error {
	model, err := m.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	model.Archive()
	return m.repo.Update(ctx, model)
}

// AddTrainingLog 添加训练日志。
func (m *AIModelManager) AddTrainingLog(ctx context.Context, modelID uint64, iteration int32, loss, accuracy, valLoss, valAccuracy float64) error {
	log := &domain.ModelTrainingLog{
		ModelID:            modelID,
		Iteration:          iteration,
		Loss:               loss,
		Accuracy:           accuracy,
		ValidationLoss:     valLoss,
		ValidationAccuracy: valAccuracy,
	}
	return m.repo.CreateTrainingLog(ctx, log)
}

// Predict 预测。
func (m *AIModelManager) Predict(ctx context.Context, modelID uint64, input string, userID uint64) (string, float64, error) {
	model, err := m.repo.GetByID(ctx, modelID)
	if err != nil {
		return "", 0, err
	}

	if model.Status != domain.ModelStatusDeployed {
		return "", 0, fmt.Errorf("model is not deployed")
	}

	output := "mock_prediction_result"
	confidence := 0.95

	prediction := &domain.ModelPrediction{
		ModelID:        modelID,
		Input:          input,
		Output:         output,
		Confidence:     confidence,
		UserID:         userID,
		PredictionTime: time.Now(),
	}

	if err := m.repo.CreatePrediction(ctx, prediction); err != nil {
		m.logger.WarnContext(ctx, "failed to save prediction record", "model_id", modelID, "error", err)
	}

	return output, confidence, nil
}

// --- DTOs ---

type ProductRecommendationDTO struct {
	ProductID uint64
	Score     float64
	Reason    string
}

type FeedItemDTO struct {
	ItemType  string
	ItemID    string
	Title     string
	ImageURL  string
	TargetURL string
	Score     float64
}

type ProductSearchResultDTO struct {
	ProductID       uint64
	SimilarityScore float64
}

type FraudScoreDTO struct {
	FraudScore   float64
	IsFraudulent bool
	Reasons      []string
}
