package domain

import (
	"errors"
	"time"

	"gorm.io/gorm" // 导入GORM库。
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
	gorm.Model                       // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	ModelNo      string              `gorm:"type:varchar(64);uniqueIndex;not null;comment:模型编号" json:"model_no"` // 模型编号，唯一索引，不允许为空。
	Name         string              `gorm:"type:varchar(128);not null;comment:模型名称" json:"name"`                // 模型名称。
	Description  string              `gorm:"type:text;comment:描述" json:"description"`                            // 模型描述。
	Type         string              `gorm:"type:varchar(64);not null;comment:类型" json:"type"`                   // 模型类型（例如，推荐模型、分类模型）。
	Algorithm    string              `gorm:"type:varchar(64);not null;comment:算法" json:"algorithm"`              // 使用的算法（例如，RandomForest、NeuralNetwork）。
	Version      string              `gorm:"type:varchar(32);not null;comment:版本" json:"version"`                // 模型版本。
	Status       ModelStatus         `gorm:"type:varchar(32);not null;default:'draft';comment:状态" json:"status"` // 模型状态，默认为“草稿”。
	Accuracy     float64             `gorm:"type:decimal(10,4);default:0;comment:准确率" json:"accuracy"`           // 模型准确率（训练完成后）。
	Parameters   map[string]any      `gorm:"type:json;serializer:json;comment:参数" json:"parameters"`             // 模型参数（JSON格式存储）。
	TrainingData string              `gorm:"type:text;comment:训练数据路径" json:"training_data"`                      // 训练数据在存储中的路径。
	ModelPath    string              `gorm:"type:varchar(255);comment:模型文件路径" json:"model_path"`                 // 模型文件在存储中的路径。
	CreatorID    uint64              `gorm:"not null;index;comment:创建人ID" json:"creator_id"`                     // 创建模型的用户ID，索引字段。
	DeployedAt   *time.Time          `gorm:"comment:部署时间" json:"deployed_at"`                                    // 模型部署时间。
	FailedReason string              `gorm:"type:text;comment:失败原因" json:"failed_reason"`                        // 模型训练或部署失败的原因。
	TrainingLogs []*ModelTrainingLog `gorm:"foreignKey:ModelID" json:"training_logs"`                            // 模型的训练日志，一对多关系。
	Predictions  []*ModelPrediction  `gorm:"foreignKey:ModelID" json:"predictions"`                              // 模型的预测记录，一对多关系。
}

// ModelTrainingLog 实体记录了AI模型训练过程中的关键指标和事件。
type ModelTrainingLog struct {
	gorm.Model                 // 嵌入gorm.Model。
	ModelID            uint64  `gorm:"not null;index;comment:模型ID" json:"model_id"`                  // 关联的AI模型ID，索引字段。
	Iteration          int32   `gorm:"not null;comment:迭代轮次" json:"iteration"`                       // 训练的迭代或Epoch轮次。
	Loss               float64 `gorm:"type:decimal(10,6);comment:损失值" json:"loss"`                   // 训练过程中的损失值。
	Accuracy           float64 `gorm:"type:decimal(10,4);comment:准确率" json:"accuracy"`               // 训练过程中的准确率。
	ValidationLoss     float64 `gorm:"type:decimal(10,6);comment:验证集损失值" json:"validation_loss"`     // 验证集上的损失值。
	ValidationAccuracy float64 `gorm:"type:decimal(10,4);comment:验证集准确率" json:"validation_accuracy"` // 验证集上的准确率。
}

// ModelPrediction 实体记录了AI模型每次预测的输入、输出、置信度等信息。
type ModelPrediction struct {
	gorm.Model               // 嵌入gorm.Model。
	ModelID        uint64    `gorm:"not null;index;comment:模型ID" json:"model_id"`      // 关联的AI模型ID，索引字段。
	Input          string    `gorm:"type:text;not null;comment:输入数据" json:"input"`     // 预测的输入数据。
	Output         string    `gorm:"type:text;not null;comment:输出结果" json:"output"`    // 预测的输出结果。
	Confidence     float64   `gorm:"type:decimal(10,4);comment:置信度" json:"confidence"` // 预测的置信度。
	UserID         uint64    `gorm:"not null;index;comment:调用用户ID" json:"user_id"`     // 调用此模型进行预测的用户ID。
	PredictionTime time.Time `gorm:"not null;comment:预测时间" json:"prediction_time"`     // 预测发生的时间。
}

// NewAIModel 创建并返回一个新的 AIModel 实体实例。
// modelNo: 模型编号。
// name, description, modelType, algorithm: 模型的基本元数据。
// creatorID: 创建模型的用户ID。
func NewAIModel(modelNo, name, description, modelType, algorithm string, creatorID uint64) *AIModel {
	return &AIModel{
		ModelNo:     modelNo,
		Name:        name,
		Description: description,
		Type:        modelType,
		Algorithm:   algorithm,
		Version:     "1.0.0",              // 默认版本号。
		Status:      ModelStatusDraft,     // 初始状态为草稿。
		Accuracy:    0.0,                  // 初始准确率为0。
		Parameters:  make(map[string]any), // 初始化模型参数map。
		CreatorID:   creatorID,
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
