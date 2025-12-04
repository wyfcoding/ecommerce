package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/ai_model/domain/entity"     // 导入AI模型模块的领域实体定义。
	"github.com/wyfcoding/ecommerce/internal/ai_model/domain/repository" // 导入AI模型模块的领域仓储接口。
	"time"                                                               // 导入时间包，用于查询条件。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

// aiModelRepository 是 AIModelRepository 接口的GORM实现。
// 它负责将AI模型模块的领域实体映射到数据库，并执行持久化操作。
type aiModelRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewAIModelRepository 创建并返回一个新的 aiModelRepository 实例。
// db: GORM数据库连接实例。
func NewAIModelRepository(db *gorm.DB) repository.AIModelRepository {
	return &aiModelRepository{db: db}
}

// Create 在数据库中创建一个新的AI模型记录。
func (r *aiModelRepository) Create(ctx context.Context, model *entity.AIModel) error {
	return r.db.WithContext(ctx).Create(model).Error
}

// GetByID 根据ID从数据库获取AI模型记录，并预加载其关联的训练日志。
func (r *aiModelRepository) GetByID(ctx context.Context, id uint64) (*entity.AIModel, error) {
	var model entity.AIModel
	if err := r.db.WithContext(ctx).Preload("TrainingLogs").First(&model, id).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

// GetByNo 根据模型编号从数据库获取AI模型记录，并预加载其关联的训练日志。
func (r *aiModelRepository) GetByNo(ctx context.Context, no string) (*entity.AIModel, error) {
	var model entity.AIModel
	if err := r.db.WithContext(ctx).Preload("TrainingLogs").Where("model_no = ?", no).First(&model).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

// Update 更新数据库中的AI模型记录。
func (r *aiModelRepository) Update(ctx context.Context, model *entity.AIModel) error {
	return r.db.WithContext(ctx).Save(model).Error
}

// List 从数据库列出所有AI模型记录，支持通过查询条件进行过滤和分页。
func (r *aiModelRepository) List(ctx context.Context, query *repository.ModelQuery) ([]*entity.AIModel, int64, error) {
	var list []*entity.AIModel
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.AIModel{})

	// 根据查询条件构建WHERE子句。
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

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	offset := (query.Page - 1) * query.PageSize
	if err := db.Offset(offset).Limit(query.PageSize).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// Delete 根据ID从数据库删除AI模型记录（软删除，因为使用了gorm.Model）。
func (r *aiModelRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.AIModel{}, id).Error
}

// CreateTrainingLog 在数据库中创建一个新的模型训练日志记录。
func (r *aiModelRepository) CreateTrainingLog(ctx context.Context, log *entity.ModelTrainingLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// ListTrainingLogs 列出指定模型的所有训练日志，按迭代轮次升序排列。
func (r *aiModelRepository) ListTrainingLogs(ctx context.Context, modelID uint64) ([]*entity.ModelTrainingLog, error) {
	var logs []*entity.ModelTrainingLog
	if err := r.db.WithContext(ctx).Where("model_id = ?", modelID).Order("iteration asc").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// CreatePrediction 在数据库中创建一个新的模型预测记录。
func (r *aiModelRepository) CreatePrediction(ctx context.Context, prediction *entity.ModelPrediction) error {
	return r.db.WithContext(ctx).Create(prediction).Error
}

// ListPredictions 列出指定模型的所有预测记录，支持时间范围过滤和分页，按预测时间降序排列。
func (r *aiModelRepository) ListPredictions(ctx context.Context, modelID uint64, startTime, endTime time.Time, page, pageSize int) ([]*entity.ModelPrediction, int64, error) {
	var list []*entity.ModelPrediction
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.ModelPrediction{}).Where("model_id = ?", modelID)

	// 根据时间范围构建WHERE子句。
	if !startTime.IsZero() {
		db = db.Where("prediction_time >= ?", startTime)
	}
	if !endTime.IsZero() {
		db = db.Where("prediction_time <= ?", endTime)
	}

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Order("prediction_time desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
