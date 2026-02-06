package domain

import (
	"errors"
	"time"
)

// 定义AI模型模块的业务错误。
var (
	ErrModelNotFound = errors.New("model not found") // 模型未找到。
)

// ModelStatus 定义了AI模型的生命周期状态。
type ModelStatus string

const (
	ModelStatusDraft     ModelStatus = "draft"     // 草稿：模型已创建，但尚未训练。
	ModelStatusTraining  ModelStatus = "training"  // 训练中：模型正在进行训练。
	ModelStatusReady     ModelStatus = "ready"     // 就绪：模型已训练完成，可供部署。
	ModelStatusDeploying ModelStatus = "deploying" // 部署中：模型正在被部署到生产环境。
	ModelStatusDeployed  ModelStatus = "deployed"  // 已部署：模型已成功部署并运行。
	ModelStatusFailed    ModelStatus = "failed"    // 失败：模型训练或部署失败。
	ModelStatusArchived  ModelStatus = "archived"  // 已归档：模型已不再使用，被归档。
)

// AIModel 实体是AI模型模块的聚合根。
// 它代表一个完整的AI模型，包括其元数据、训练状态、部署信息和性能指标。
type AIModel struct {
	ID           uint           `json:"id"`          // 主键ID
	CreatedAt    time.Time      `json:"created_at"`  // 创建时间
	UpdatedAt    time.Time      `json:"updated_at"`  // 更新时间
	ModelNo      string         `json:"model_no"`    // 模型编号（唯一）
	Name         string         `json:"name"`        // 模型名称
	Description  string         `json:"description"` // 模型描述
	Type         string         `json:"type"`        // 模型类型（例如，推荐模型、分类模型）
	Algorithm    string         `json:"algorithm"`   // 使用的算法（例如，RandomForest、NeuralNetwork）
	Version      string         `json:"version"`     // 模型版本
	Status       ModelStatus    `json:"status"`
	Accuracy     float64        `json:"accuracy"`      // 模型准确率（训练完成后）
	Parameters   map[string]any `json:"parameters"`    // 模型参数（JSON格式存储）
	TrainingData string         `json:"training_data"` // 训练数据路径
	ModelPath    string         `json:"model_path"`    // 模型文件路径
	CreatorID    uint64         `json:"creator_id"`    // 创建模型的用户ID
	DeployedAt   *time.Time     `json:"deployed_at"`   // 模型部署时间
	FailedReason string         `json:"failed_reason"` // 失败原因

	TrainingLogs []*ModelTrainingLog `json:"training_logs"` // 训练日志
	Predictions  []*ModelPrediction  `json:"predictions"`   // 预测记录
}

// ModelTrainingLog 实体记录了AI模型训练过程中的关键指标和事件。
type ModelTrainingLog struct {
	ID                 uint      `json:"id"`                  // 主键ID
	CreatedAt          time.Time `json:"created_at"`          // 创建时间
	UpdatedAt          time.Time `json:"updated_at"`          // 更新时间
	ModelID            uint64    `json:"model_id"`            // 关联的AI模型ID
	Iteration          int32     `json:"iteration"`           // 迭代轮次
	Loss               float64   `json:"loss"`                // 损失值
	Accuracy           float64   `json:"accuracy"`            // 准确率
	ValidationLoss     float64   `json:"validation_loss"`     // 验证集损失值
	ValidationAccuracy float64   `json:"validation_accuracy"` // 验证集准确率
}

// ModelPrediction 实体记录了AI模型每次预测的输入、输出、置信度等信息。
type ModelPrediction struct {
	ID             uint      `json:"id"`              // 主键ID
	CreatedAt      time.Time `json:"created_at"`      // 创建时间
	UpdatedAt      time.Time `json:"updated_at"`      // 更新时间
	ModelID        uint64    `json:"model_id"`        // 关联的AI模型ID
	Input          string    `json:"input"`           // 预测输入
	Output         string    `json:"output"`          // 预测输出
	Confidence     float64   `json:"confidence"`      // 置信度
	UserID         uint64    `json:"user_id"`         // 调用用户ID
	PredictionTime time.Time `json:"prediction_time"` // 预测时间
}

// NewAIModel 创建并返回一个新的 AIModel 实体实例。
// modelNo: 模型编号。
// name, description, modelType, algorithm: 模型的基本元数据。
// creatorID: 创建模型的用户ID。
func NewAIModel(modelNo, name, description, modelType, algorithm string, creatorID uint64) *AIModel {
	return &AIModel{
		ModelNo:      modelNo,
		Name:         name,
		Description:  description,
		Type:         modelType,
		Algorithm:    algorithm,
		Version:      "1.0.0",              // 默认版本号。
		Status:       ModelStatusDraft,     // 初始状态为草稿。
		Accuracy:     0.0,                  // 初始准确率为0。
		Parameters:   make(map[string]any), // 初始化模型参数map。
		CreatorID:    creatorID,
		TrainingLogs: []*ModelTrainingLog{},
		Predictions:  []*ModelPrediction{},
	}
}

// StartTraining 将模型状态设置为“训练中”。
func (m *AIModel) StartTraining() {
	m.Status = ModelStatusTraining
}

// CompleteTraining 将模型状态设置为“就绪”，并记录训练结果。
func (m *AIModel) CompleteTraining(accuracy float64, modelPath string) {
	m.Status = ModelStatusReady
	m.Accuracy = accuracy
	m.ModelPath = modelPath
}

// FailTraining 将模型状态设置为“失败”，并记录失败原因。
func (m *AIModel) FailTraining(reason string) {
	m.Status = ModelStatusFailed
	m.FailedReason = reason
}

// Deploy 将模型状态设置为“部署中”。
func (m *AIModel) Deploy() {
	m.Status = ModelStatusDeploying
}

// CompleteDeployment 将模型状态设置为“已部署”，并记录部署时间。
func (m *AIModel) CompleteDeployment() {
	m.Status = ModelStatusDeployed
	now := time.Now()
	m.DeployedAt = &now
}

// Archive 将模型状态设置为“已归档”。
func (m *AIModel) Archive() {
	m.Status = ModelStatusArchived
}
