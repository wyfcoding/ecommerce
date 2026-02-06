package application

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/wyfcoding/ecommerce/internal/aimodel/domain"
	algorithm "github.com/wyfcoding/pkg/algorithm/ml"
	"github.com/wyfcoding/pkg/idgen"
)

// AIModelCommandService 负责AI模型模块的写操作和业务逻辑。
type AIModelCommandService struct {
	repo         domain.AIModelRepository
	publisher    domain.EventPublisher
	idGenerator  idgen.Generator
	logger       *slog.Logger
	loadedModels map[uint64]*algorithm.NaiveBayes
	modelsMu     sync.RWMutex
}

// NewAIModelCommandService 创建一个新的 AIModelCommandService 实例。
func NewAIModelCommandService(repo domain.AIModelRepository, publisher domain.EventPublisher, idGenerator idgen.Generator, logger *slog.Logger) *AIModelCommandService {
	return &AIModelCommandService{
		repo:         repo,
		publisher:    publisher,
		idGenerator:  idGenerator,
		logger:       logger,
		loadedModels: make(map[uint64]*algorithm.NaiveBayes),
	}
}

// CreateModel 创建一个新的AI模型记录。
func (m *AIModelCommandService) CreateModel(ctx context.Context, name, description, modelType, algorithmName string, creatorID uint64) (*domain.AIModel, error) {
	modelNo := fmt.Sprintf("AIM%d", m.idGenerator.Generate())
	model := domain.NewAIModel(modelNo, name, description, modelType, algorithmName, creatorID)

	err := m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveModelInTx(ctx, tx, model); err != nil {
			return err
		}
		event := &domain.AIModelCreatedEvent{
			ModelID:   uint64(model.ID),
			ModelNo:   model.ModelNo,
			Status:    model.Status,
			Timestamp: time.Now(),
		}
		if m.publisher == nil {
			return nil
		}
		return m.publisher.PublishInTx(ctx, tx, domain.AIModelCreatedEventType, fmt.Sprintf("%d", model.ID), event)
	})
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to create model", "name", name, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "model created successfully", "model_id", model.ID, "name", name)

	return model, nil
}

// StartTraining 启动训练。
func (m *AIModelCommandService) StartTraining(ctx context.Context, id uint64) error {
	model, err := m.repo.GetModel(ctx, id)
	if err != nil {
		return err
	}
	oldStatus := model.Status
	model.StartTraining()

	if err := m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveModelInTx(ctx, tx, model); err != nil {
			return err
		}
		return m.publishStatusUpdated(ctx, tx, model.ID, oldStatus, model.Status)
	}); err != nil {
		return err
	}

	// 异步执行模型训练任务
	go m.runTrainingTask(id)

	return nil
}

func (m *AIModelCommandService) runTrainingTask(modelID uint64) {
	bgCtx := context.Background()
	m.logger.Info("starting iterative training pipeline", "model_id", modelID)

	// 1. 模拟多轮迭代 (Epochs)
	numEpochs := 5
	finalAccuracy := 0.0
	for epoch := 1; epoch <= numEpochs; epoch++ {
		time.Sleep(500 * time.Millisecond) // 模拟计算开销

		// 模拟指标演进
		loss := 1.0 / float64(epoch)
		acc := 0.6 + (0.35 * (1.0 - loss)) // 从 0.6 提升到 0.95 左右
		finalAccuracy = acc

		// 记录详细训练日志 (Olap Analytics Ready)
		_ = m.AddTrainingLog(bgCtx, modelID, int32(epoch), loss, acc, loss*1.1, acc*0.98)
		m.logger.Debug("training epoch finished", "model_id", modelID, "epoch", epoch, "accuracy", acc)
	}

	// 2. 使用 NaiveBayes 训练最终推断实例
	nb := algorithm.NewNaiveBayes()
	docs := [][]string{
		{"good", "great", "awesome", "fantastic"},
		{"bad", "terrible", "awful", "worst"},
		{"happy", "joy", "love"},
		{"hate", "sad", "angry"},
	}
	labels := []string{"positive", "negative", "positive", "negative"}
	nb.Train(docs, labels)

	// 3. 缓存并更新状态
	m.modelsMu.Lock()
	m.loadedModels[modelID] = nb
	m.modelsMu.Unlock()

	if err := m.CompleteTraining(bgCtx, modelID, finalAccuracy, fmt.Sprintf("/models/%d.bin", modelID)); err != nil {
		m.logger.Error("failed to complete training", "model_id", modelID, "error", err)
		_ = m.FailTraining(bgCtx, modelID, err.Error())
	} else {
		m.logger.Info("training pipeline finished successfully", "model_id", modelID, "accuracy", finalAccuracy)
	}
}

// CompleteTraining 完成训练。
func (m *AIModelCommandService) CompleteTraining(ctx context.Context, id uint64, accuracy float64, modelPath string) error {
	model, err := m.repo.GetModel(ctx, id)
	if err != nil {
		return err
	}
	oldStatus := model.Status
	model.CompleteTraining(accuracy, modelPath)

	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveModelInTx(ctx, tx, model); err != nil {
			return err
		}
		return m.publishStatusUpdated(ctx, tx, model.ID, oldStatus, model.Status)
	})
}

// FailTraining 训练失败。
func (m *AIModelCommandService) FailTraining(ctx context.Context, id uint64, reason string) error {
	model, err := m.repo.GetModel(ctx, id)
	if err != nil {
		return err
	}
	oldStatus := model.Status
	model.FailTraining(reason)

	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveModelInTx(ctx, tx, model); err != nil {
			return err
		}
		return m.publishStatusUpdated(ctx, tx, model.ID, oldStatus, model.Status)
	})
}

// Deploy 部署模型。
func (m *AIModelCommandService) Deploy(ctx context.Context, id uint64) error {
	model, err := m.repo.GetModel(ctx, id)
	if err != nil {
		return err
	}
	oldStatus := model.Status
	model.Deploy()

	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveModelInTx(ctx, tx, model); err != nil {
			return err
		}
		return m.publishStatusUpdated(ctx, tx, model.ID, oldStatus, model.Status)
	})
}

// CompleteDeployment 完成部署。
func (m *AIModelCommandService) CompleteDeployment(ctx context.Context, id uint64) error {
	model, err := m.repo.GetModel(ctx, id)
	if err != nil {
		return err
	}
	oldStatus := model.Status
	model.CompleteDeployment()

	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveModelInTx(ctx, tx, model); err != nil {
			return err
		}
		return m.publishStatusUpdated(ctx, tx, model.ID, oldStatus, model.Status)
	})
}

// Archive 归档模型。
func (m *AIModelCommandService) Archive(ctx context.Context, id uint64) error {
	model, err := m.repo.GetModel(ctx, id)
	if err != nil {
		return err
	}
	oldStatus := model.Status
	model.Archive()

	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveModelInTx(ctx, tx, model); err != nil {
			return err
		}
		return m.publishStatusUpdated(ctx, tx, model.ID, oldStatus, model.Status)
	})
}

// AddTrainingLog 添加训练日志。
func (m *AIModelCommandService) AddTrainingLog(ctx context.Context, modelID uint64, iteration int32, loss, accuracy, valLoss, valAccuracy float64) error {
	logEntry := &domain.ModelTrainingLog{
		ModelID:            modelID,
		Iteration:          iteration,
		Loss:               loss,
		Accuracy:           accuracy,
		ValidationLoss:     valLoss,
		ValidationAccuracy: valAccuracy,
	}
	return m.repo.SaveTrainingLog(ctx, logEntry)
}

// Predict 预测。
func (m *AIModelCommandService) Predict(ctx context.Context, modelID uint64, input string, userID uint64) (string, float64, error) {
	modelMeta, err := m.repo.GetModel(ctx, modelID)
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
		// 真实化执行：从持久化存储加载模型权重
		m.logger.Warn("model not in memory, attempting to load from storage", "model_id", modelID, "path", modelMeta.ModelPath)

		// 这里模拟从文件加载并反序列化 NaiveBayes 实例
		// 在顶级架构中，模型应由专门的 ModelLoader 进行生命周期管理
		if modelMeta.ModelPath == "" {
			return "", 0, fmt.Errorf("model weight path is empty, cannot perform inference")
		}

		// 模拟加载逻辑
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

	if err := m.repo.SavePrediction(ctx, prediction); err != nil {
		m.logger.WarnContext(ctx, "failed to save prediction record", "model_id", modelID, "error", err)
	}

	return output, confidence, nil
}

func (m *AIModelCommandService) publishStatusUpdated(ctx context.Context, tx any, modelID uint, oldStatus, newStatus domain.ModelStatus) error {
	if m.publisher == nil {
		return nil
	}
	event := &domain.AIModelStatusUpdatedEvent{
		ModelID:   uint64(modelID),
		OldStatus: oldStatus,
		NewStatus: newStatus,
		Timestamp: time.Now(),
	}
	return m.publisher.PublishInTx(ctx, tx, domain.AIModelStatusUpdatedEventType, fmt.Sprintf("%d", modelID), event)
}
