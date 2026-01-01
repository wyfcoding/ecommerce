package application

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/wyfcoding/ecommerce/internal/aimodel/domain"
	"github.com/wyfcoding/pkg/algorithm"
	"github.com/wyfcoding/pkg/idgen"
)

// AIModelManager 负责AI模型模块的写操作和业务逻辑。
type AIModelManager struct {
	repo         domain.AIModelRepository
	idGenerator  idgen.Generator
	logger       *slog.Logger
	loadedModels map[uint64]*algorithm.NaiveBayes
	modelsMu     sync.RWMutex
}

// NewAIModelManager 创建一个新的 AIModelManager 实例。
func NewAIModelManager(repo domain.AIModelRepository, idGenerator idgen.Generator, logger *slog.Logger) *AIModelManager {
	return &AIModelManager{
		repo:         repo,
		idGenerator:  idGenerator,
		logger:       logger,
		loadedModels: make(map[uint64]*algorithm.NaiveBayes),
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
	if err := m.repo.Update(ctx, model); err != nil {
		return err
	}

	// 异步执行模型训练任务
	go m.runTrainingTask(id)

	return nil
}

func (m *AIModelManager) runTrainingTask(modelID uint64) {
	// 模拟训练耗时
	time.Sleep(2 * time.Second)

	// 使用 pkg/algorithm/naive_bayes 进行实际训练 (示例数据)
	nb := algorithm.NewNaiveBayes()

	// 模拟一些简单的文本分类数据 (例如：情感分析 Positive/Negative)
	docs := [][]string{
		{"good", "great", "awesome", "fantastic"},
		{"bad", "terrible", "awful", "worst"},
		{"happy", "joy", "love"},
		{"hate", "sad", "angry"},
		{"like", "enjoy", "recommend"},
		{"dislike", "avoid", "refund"},
	}
	labels := []string{
		"positive", "negative", "positive", "negative", "positive", "negative",
	}

	nb.Train(docs, labels)

	// 训练完成后，将模型加载到内存缓存
	m.modelsMu.Lock()
	m.loadedModels[modelID] = nb
	m.modelsMu.Unlock()

	// 更新数据库状态
	// 注意：这里使用 Background Context
	bgCtx := context.Background()
	if err := m.CompleteTraining(bgCtx, modelID, 0.95, fmt.Sprintf("/models/%d.bin", modelID)); err != nil {
		m.logger.Error("failed to complete training", "model_id", modelID, "error", err)
		// 尝试标记为失败
		_ = m.FailTraining(bgCtx, modelID, err.Error())
	} else {
		m.logger.Info("training task completed successfully", "model_id", modelID)
	}
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
	modelMeta, err := m.repo.GetByID(ctx, modelID)
	if err != nil {
		return "", 0, err
	}

	if modelMeta.Status != domain.ModelStatusDeployed && modelMeta.Status != domain.ModelStatusReady {
		return "", 0, fmt.Errorf("model is not deployed or ready (status: %s)", modelMeta.Status)
	}

	// 尝试从内存中获取模型
	m.modelsMu.RLock()
	nb, exists := m.loadedModels[modelID]
	m.modelsMu.RUnlock()

	if !exists {
		// 如果内存中不存在（可能是重启后），则重新初始化并“加载”
		// 实际场景应从文件系统加载序列化的模型
		m.logger.Warn("model not in memory, re-initializing dummy model", "model_id", modelID)

		// 重新训练一个 dummy 模型
		nb = algorithm.NewNaiveBayes()
		docs := [][]string{
			{"good", "great", "awesome", "fantastic"},
			{"bad", "terrible", "awful", "worst"},
		}
		labels := []string{"positive", "negative"}
		nb.Train(docs, labels)

		m.modelsMu.Lock()
		m.loadedModels[modelID] = nb
		m.modelsMu.Unlock()
	}

	// 执行预测
	inputTokens := strings.Fields(strings.ToLower(input))
	output, confidence := nb.PredictWithConfidence(inputTokens)

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

// --- 模块分段 ---

// ProductRecommendationDTO 结构体定义。
type ProductRecommendationDTO struct {
	ProductID uint64
	Score     float64
	Reason    string
}

// FeedItemDTO 结构体定义。
type FeedItemDTO struct {
	ItemType  string
	ItemID    string
	Title     string
	ImageURL  string
	TargetURL string
	Score     float64
}

// ProductSearchResultDTO 结构体定义。
type ProductSearchResultDTO struct {
	ProductID       uint64
	SimilarityScore float64
}

// FraudScoreDTO 结构体定义。
type FraudScoreDTO struct {
	FraudScore   float64
	IsFraudulent bool
	Reasons      []string
}
