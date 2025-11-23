package entity

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrModelNotFound = errors.New("model not found")
)

// ModelStatus AI模型状态
type ModelStatus string

const (
	ModelStatusDraft     ModelStatus = "draft"     // 草稿
	ModelStatusTraining  ModelStatus = "training"  // 训练中
	ModelStatusReady     ModelStatus = "ready"     // 就绪
	ModelStatusDeploying ModelStatus = "deploying" // 部署中
	ModelStatusDeployed  ModelStatus = "deployed"  // 已部署
	ModelStatusFailed    ModelStatus = "failed"    // 失败
	ModelStatusArchived  ModelStatus = "archived"  // 已归档
)

// AIModel AI模型聚合根
type AIModel struct {
	gorm.Model
	ModelNo      string                 `gorm:"type:varchar(64);uniqueIndex;not null;comment:模型编号" json:"model_no"`
	Name         string                 `gorm:"type:varchar(128);not null;comment:模型名称" json:"name"`
	Description  string                 `gorm:"type:text;comment:描述" json:"description"`
	Type         string                 `gorm:"type:varchar(64);not null;comment:类型" json:"type"`
	Algorithm    string                 `gorm:"type:varchar(64);not null;comment:算法" json:"algorithm"`
	Version      string                 `gorm:"type:varchar(32);not null;comment:版本" json:"version"`
	Status       ModelStatus            `gorm:"type:varchar(32);not null;default:'draft';comment:状态" json:"status"`
	Accuracy     float64                `gorm:"type:decimal(10,4);default:0;comment:准确率" json:"accuracy"`
	Parameters   map[string]interface{} `gorm:"type:json;serializer:json;comment:参数" json:"parameters"`
	TrainingData string                 `gorm:"type:text;comment:训练数据路径" json:"training_data"`
	ModelPath    string                 `gorm:"type:varchar(255);comment:模型文件路径" json:"model_path"`
	CreatorID    uint64                 `gorm:"not null;index;comment:创建人ID" json:"creator_id"`
	DeployedAt   *time.Time             `gorm:"comment:部署时间" json:"deployed_at"`
	FailedReason string                 `gorm:"type:text;comment:失败原因" json:"failed_reason"`
	TrainingLogs []*ModelTrainingLog    `gorm:"foreignKey:ModelID" json:"training_logs"`
	Predictions  []*ModelPrediction     `gorm:"foreignKey:ModelID" json:"predictions"`
}

// ModelTrainingLog 模型训练日志实体
type ModelTrainingLog struct {
	gorm.Model
	ModelID            uint64  `gorm:"not null;index;comment:模型ID" json:"model_id"`
	Iteration          int32   `gorm:"not null;comment:迭代轮次" json:"iteration"`
	Loss               float64 `gorm:"type:decimal(10,6);comment:损失值" json:"loss"`
	Accuracy           float64 `gorm:"type:decimal(10,4);comment:准确率" json:"accuracy"`
	ValidationLoss     float64 `gorm:"type:decimal(10,6);comment:验证集损失值" json:"validation_loss"`
	ValidationAccuracy float64 `gorm:"type:decimal(10,4);comment:验证集准确率" json:"validation_accuracy"`
}

// ModelPrediction 模型预测记录
type ModelPrediction struct {
	gorm.Model
	ModelID        uint64    `gorm:"not null;index;comment:模型ID" json:"model_id"`
	Input          string    `gorm:"type:text;not null;comment:输入数据" json:"input"`
	Output         string    `gorm:"type:text;not null;comment:输出结果" json:"output"`
	Confidence     float64   `gorm:"type:decimal(10,4);comment:置信度" json:"confidence"`
	UserID         uint64    `gorm:"not null;index;comment:调用用户ID" json:"user_id"`
	PredictionTime time.Time `gorm:"not null;comment:预测时间" json:"prediction_time"`
}

// NewAIModel 创建AI模型
func NewAIModel(modelNo, name, description, modelType, algorithm string, creatorID uint64) *AIModel {
	return &AIModel{
		ModelNo:     modelNo,
		Name:        name,
		Description: description,
		Type:        modelType,
		Algorithm:   algorithm,
		Version:     "1.0.0",
		Status:      ModelStatusDraft,
		Accuracy:    0.0,
		Parameters:  make(map[string]interface{}),
		CreatorID:   creatorID,
	}
}

// StartTraining 开始训练
func (m *AIModel) StartTraining() {
	m.Status = ModelStatusTraining
}

// CompleteTraining 完成训练
func (m *AIModel) CompleteTraining(accuracy float64, modelPath string) {
	m.Status = ModelStatusReady
	m.Accuracy = accuracy
	m.ModelPath = modelPath
}

// FailTraining 训练失败
func (m *AIModel) FailTraining(reason string) {
	m.Status = ModelStatusFailed
	m.FailedReason = reason
}

// Deploy 部署模型
func (m *AIModel) Deploy() {
	m.Status = ModelStatusDeploying
}

// CompleteDeployment 完成部署
func (m *AIModel) CompleteDeployment() {
	m.Status = ModelStatusDeployed
	now := time.Now()
	m.DeployedAt = &now
}

// Archive 归档模型
func (m *AIModel) Archive() {
	m.Status = ModelStatusArchived
}
