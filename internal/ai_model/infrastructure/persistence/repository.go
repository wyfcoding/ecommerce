package persistence

import (
	"context"
	"ecommerce/internal/ai_model/domain/entity"
	"ecommerce/internal/ai_model/domain/repository"
	"time"

	"gorm.io/gorm"
)

type aiModelRepository struct {
	db *gorm.DB
}

func NewAIModelRepository(db *gorm.DB) repository.AIModelRepository {
	return &aiModelRepository{db: db}
}

func (r *aiModelRepository) Create(ctx context.Context, model *entity.AIModel) error {
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *aiModelRepository) GetByID(ctx context.Context, id uint64) (*entity.AIModel, error) {
	var model entity.AIModel
	if err := r.db.WithContext(ctx).Preload("TrainingLogs").First(&model, id).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

func (r *aiModelRepository) GetByNo(ctx context.Context, no string) (*entity.AIModel, error) {
	var model entity.AIModel
	if err := r.db.WithContext(ctx).Preload("TrainingLogs").Where("model_no = ?", no).First(&model).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

func (r *aiModelRepository) Update(ctx context.Context, model *entity.AIModel) error {
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *aiModelRepository) List(ctx context.Context, query *repository.ModelQuery) ([]*entity.AIModel, int64, error) {
	var list []*entity.AIModel
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.AIModel{})

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

func (r *aiModelRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.AIModel{}, id).Error
}

func (r *aiModelRepository) CreateTrainingLog(ctx context.Context, log *entity.ModelTrainingLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *aiModelRepository) ListTrainingLogs(ctx context.Context, modelID uint64) ([]*entity.ModelTrainingLog, error) {
	var logs []*entity.ModelTrainingLog
	if err := r.db.WithContext(ctx).Where("model_id = ?", modelID).Order("iteration asc").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *aiModelRepository) CreatePrediction(ctx context.Context, prediction *entity.ModelPrediction) error {
	return r.db.WithContext(ctx).Create(prediction).Error
}

func (r *aiModelRepository) ListPredictions(ctx context.Context, modelID uint64, startTime, endTime time.Time, page, pageSize int) ([]*entity.ModelPrediction, int64, error) {
	var list []*entity.ModelPrediction
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.ModelPrediction{}).Where("model_id = ?", modelID)

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
