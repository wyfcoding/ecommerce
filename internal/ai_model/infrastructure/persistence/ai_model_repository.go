package persistence

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/ai_model/domain"

	"gorm.io/gorm"
)

type aiModelRepository struct {
	db *gorm.DB
}

// NewAIModelRepository 创建并返回一个新的 aiModelRepository 实例。
func NewAIModelRepository(db *gorm.DB) domain.AIModelRepository {
	return &aiModelRepository{db: db}
}

// Create 在数据库中创建一个新的AI模型记录。
func (r *aiModelRepository) Create(ctx context.Context, model *domain.AIModel) error {
	return r.db.WithContext(ctx).Create(model).Error
}

// GetByID 根据ID从数据库获取AI模型记录，并预加载其关联的训练日志。
func (r *aiModelRepository) GetByID(ctx context.Context, id uint64) (*domain.AIModel, error) {
	var model domain.AIModel
	if err := r.db.WithContext(ctx).Preload("TrainingLogs").First(&model, id).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

// GetByNo 根据模型编号从数据库获取AI模型记录，并预加载其关联的训练日志。
func (r *aiModelRepository) GetByNo(ctx context.Context, no string) (*domain.AIModel, error) {
	var model domain.AIModel
	if err := r.db.WithContext(ctx).Preload("TrainingLogs").Where("model_no = ?", no).First(&model).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

// Update 更新数据库中的AI模型记录。
func (r *aiModelRepository) Update(ctx context.Context, model *domain.AIModel) error {
	return r.db.WithContext(ctx).Save(model).Error
}

// List 从数据库列出所有AI模型记录，支持通过查询条件进行过滤和分页。
func (r *aiModelRepository) List(ctx context.Context, query *domain.ModelQuery) ([]*domain.AIModel, int64, error) {
	var list []*domain.AIModel
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.AIModel{})

	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.Type != "" {
		db = db.Where("type = ?", query.Type)
	}
	if query.Algorithm != "" {
		db = db.Where("algorithm = ?", query.Algorithm)
	}
	if query.CreatorID > 0 {
		db = db.Where("creator_id = ?", query.CreatorID)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := db.Offset(offset).Limit(query.PageSize).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// Delete 根据ID从数据库删除AI模型记录。
func (r *aiModelRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.AIModel{}, id).Error
}

// CreateTrainingLog 在数据库中创建一个新的模型训练日志记录。
func (r *aiModelRepository) CreateTrainingLog(ctx context.Context, log *domain.ModelTrainingLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// ListTrainingLogs 列出指定模型的所有训练日志。
func (r *aiModelRepository) ListTrainingLogs(ctx context.Context, modelID uint64) ([]*domain.ModelTrainingLog, error) {
	var logs []*domain.ModelTrainingLog
	if err := r.db.WithContext(ctx).Where("model_id = ?", modelID).Order("iteration asc").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// CreatePrediction 在数据库中创建一个新的模型预测记录。
func (r *aiModelRepository) CreatePrediction(ctx context.Context, prediction *domain.ModelPrediction) error {
	return r.db.WithContext(ctx).Create(prediction).Error
}

// ListPredictions 列出指定模型的所有预测记录。
func (r *aiModelRepository) ListPredictions(ctx context.Context, modelID uint64, startTime, endTime time.Time, page, pageSize int) ([]*domain.ModelPrediction, int64, error) {
	var list []*domain.ModelPrediction
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.ModelPrediction{}).Where("model_id = ?", modelID)

	if !startTime.IsZero() {
		db = db.Where("prediction_time >= ?", startTime)
	}
	if !endTime.IsZero() {
		db = db.Where("prediction_time <= ?", endTime)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Order("prediction_time desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
