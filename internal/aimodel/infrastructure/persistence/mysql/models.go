package mysql

import (
	"time"

	"github.com/wyfcoding/ecommerce/internal/aimodel/domain"
	"gorm.io/gorm"
)

// AIModelModel AI 模型写模型。
type AIModelModel struct {
	gorm.Model
	ModelNo      string             `gorm:"column:model_no;type:varchar(64);uniqueIndex;not null;comment:模型编号"`
	Name         string             `gorm:"column:name;type:varchar(128);not null;comment:模型名称"`
	Description  string             `gorm:"column:description;type:text;comment:描述"`
	Type         string             `gorm:"column:type;type:varchar(64);not null;index;comment:类型"`
	Algorithm    string             `gorm:"column:algorithm;type:varchar(64);not null;comment:算法"`
	Version      string             `gorm:"column:version;type:varchar(32);not null;comment:版本"`
	Status       domain.ModelStatus `gorm:"column:status;type:varchar(32);not null;default:'draft';comment:状态"`
	Accuracy     float64            `gorm:"column:accuracy;type:decimal(10,4);default:0;comment:准确率"`
	Parameters   map[string]any     `gorm:"column:parameters;type:json;serializer:json;comment:参数"`
	TrainingData string             `gorm:"column:training_data;type:text;comment:训练数据路径"`
	ModelPath    string             `gorm:"column:model_path;type:varchar(255);comment:模型文件路径"`
	CreatorID    uint64             `gorm:"column:creator_id;not null;index;comment:创建人ID"`
	DeployedAt   *time.Time         `gorm:"column:deployed_at;comment:部署时间"`
	FailedReason string             `gorm:"column:failed_reason;type:text;comment:失败原因"`

	TrainingLogs []ModelTrainingLogModel `gorm:"foreignKey:ModelID"`
	Predictions  []ModelPredictionModel  `gorm:"foreignKey:ModelID"`
}

// ModelTrainingLogModel 训练日志写模型。
type ModelTrainingLogModel struct {
	gorm.Model
	ModelID            uint64  `gorm:"column:model_id;not null;index;comment:模型ID"`
	Iteration          int32   `gorm:"column:iteration;not null;comment:迭代轮次"`
	Loss               float64 `gorm:"column:loss;type:decimal(10,6);comment:损失值"`
	Accuracy           float64 `gorm:"column:accuracy;type:decimal(10,4);comment:准确率"`
	ValidationLoss     float64 `gorm:"column:validation_loss;type:decimal(10,6);comment:验证集损失值"`
	ValidationAccuracy float64 `gorm:"column:validation_accuracy;type:decimal(10,4);comment:验证集准确率"`
}

// ModelPredictionModel 预测记录写模型。
type ModelPredictionModel struct {
	gorm.Model
	ModelID        uint64    `gorm:"column:model_id;not null;index;comment:模型ID"`
	Input          string    `gorm:"column:input;type:text;not null;comment:输入数据"`
	Output         string    `gorm:"column:output;type:text;not null;comment:输出结果"`
	Confidence     float64   `gorm:"column:confidence;type:decimal(10,4);comment:置信度"`
	UserID         uint64    `gorm:"column:user_id;not null;index;comment:调用用户ID"`
	PredictionTime time.Time `gorm:"column:prediction_time;not null;comment:预测时间"`
}

func (AIModelModel) TableName() string          { return "ai_models" }
func (ModelTrainingLogModel) TableName() string { return "ai_model_training_logs" }
func (ModelPredictionModel) TableName() string  { return "ai_model_predictions" }

func toAIModelModel(model *domain.AIModel) *AIModelModel {
	if model == nil {
		return nil
	}
	return &AIModelModel{
		Model: gorm.Model{
			ID:        model.ID,
			CreatedAt: model.CreatedAt,
			UpdatedAt: model.UpdatedAt,
		},
		ModelNo:      model.ModelNo,
		Name:         model.Name,
		Description:  model.Description,
		Type:         model.Type,
		Algorithm:    model.Algorithm,
		Version:      model.Version,
		Status:       model.Status,
		Accuracy:     model.Accuracy,
		Parameters:   model.Parameters,
		TrainingData: model.TrainingData,
		ModelPath:    model.ModelPath,
		CreatorID:    model.CreatorID,
		DeployedAt:   model.DeployedAt,
		FailedReason: model.FailedReason,
	}
}

func toAIModel(model *AIModelModel) *domain.AIModel {
	if model == nil {
		return nil
	}
	entity := &domain.AIModel{
		ID:           model.ID,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
		ModelNo:      model.ModelNo,
		Name:         model.Name,
		Description:  model.Description,
		Type:         model.Type,
		Algorithm:    model.Algorithm,
		Version:      model.Version,
		Status:       model.Status,
		Accuracy:     model.Accuracy,
		Parameters:   model.Parameters,
		TrainingData: model.TrainingData,
		ModelPath:    model.ModelPath,
		CreatorID:    model.CreatorID,
		DeployedAt:   model.DeployedAt,
		FailedReason: model.FailedReason,
	}
	if len(model.TrainingLogs) > 0 {
		logs := make([]*domain.ModelTrainingLog, len(model.TrainingLogs))
		for i := range model.TrainingLogs {
			logs[i] = toTrainingLog(&model.TrainingLogs[i])
		}
		entity.TrainingLogs = logs
	}
	if len(model.Predictions) > 0 {
		predictions := make([]*domain.ModelPrediction, len(model.Predictions))
		for i := range model.Predictions {
			predictions[i] = toPrediction(&model.Predictions[i])
		}
		entity.Predictions = predictions
	}
	return entity
}

func toTrainingLogModel(log *domain.ModelTrainingLog) *ModelTrainingLogModel {
	if log == nil {
		return nil
	}
	return &ModelTrainingLogModel{
		Model: gorm.Model{
			ID:        log.ID,
			CreatedAt: log.CreatedAt,
			UpdatedAt: log.UpdatedAt,
		},
		ModelID:            log.ModelID,
		Iteration:          log.Iteration,
		Loss:               log.Loss,
		Accuracy:           log.Accuracy,
		ValidationLoss:     log.ValidationLoss,
		ValidationAccuracy: log.ValidationAccuracy,
	}
}

func toTrainingLog(model *ModelTrainingLogModel) *domain.ModelTrainingLog {
	if model == nil {
		return nil
	}
	return &domain.ModelTrainingLog{
		ID:                 model.ID,
		CreatedAt:          model.CreatedAt,
		UpdatedAt:          model.UpdatedAt,
		ModelID:            model.ModelID,
		Iteration:          model.Iteration,
		Loss:               model.Loss,
		Accuracy:           model.Accuracy,
		ValidationLoss:     model.ValidationLoss,
		ValidationAccuracy: model.ValidationAccuracy,
	}
}

func toPredictionModel(pred *domain.ModelPrediction) *ModelPredictionModel {
	if pred == nil {
		return nil
	}
	return &ModelPredictionModel{
		Model: gorm.Model{
			ID:        pred.ID,
			CreatedAt: pred.CreatedAt,
			UpdatedAt: pred.UpdatedAt,
		},
		ModelID:        pred.ModelID,
		Input:          pred.Input,
		Output:         pred.Output,
		Confidence:     pred.Confidence,
		UserID:         pred.UserID,
		PredictionTime: pred.PredictionTime,
	}
}

func toPrediction(model *ModelPredictionModel) *domain.ModelPrediction {
	if model == nil {
		return nil
	}
	return &domain.ModelPrediction{
		ID:             model.ID,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
		ModelID:        model.ModelID,
		Input:          model.Input,
		Output:         model.Output,
		Confidence:     model.Confidence,
		UserID:         model.UserID,
		PredictionTime: model.PredictionTime,
	}
}
