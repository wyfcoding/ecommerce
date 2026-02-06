package mysql

import (
	"context"
	"errors"
	"time"

	"github.com/wyfcoding/ecommerce/internal/aimodel/domain"
	"gorm.io/gorm"
)

type aiModelRepository struct {
	db *gorm.DB
}

// NewAIModelRepository 创建并返回一个新的 AIModelRepository 实例。
func NewAIModelRepository(db *gorm.DB) domain.AIModelRepository {
	return &aiModelRepository{db: db}
}

func (r *aiModelRepository) BeginTx(ctx context.Context) any {
	return r.db.WithContext(ctx).Begin()
}

func (r *aiModelRepository) CommitTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Commit().Error
}

func (r *aiModelRepository) RollbackTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Rollback().Error
}

func (r *aiModelRepository) WithTx(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// --- AIModel ---

func (r *aiModelRepository) SaveModel(ctx context.Context, model *domain.AIModel) error {
	return r.saveModelWithTx(ctx, r.db, model)
}

func (r *aiModelRepository) SaveModelInTx(ctx context.Context, tx any, model *domain.AIModel) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveModelWithTx(ctx, gormTx, model)
}

func (r *aiModelRepository) GetModel(ctx context.Context, id uint64) (*domain.AIModel, error) {
	var model AIModelModel
	if err := r.db.WithContext(ctx).Preload("TrainingLogs").First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrModelNotFound
		}
		return nil, err
	}
	return toAIModel(&model), nil
}

func (r *aiModelRepository) GetModelByNo(ctx context.Context, no string) (*domain.AIModel, error) {
	var model AIModelModel
	if err := r.db.WithContext(ctx).Preload("TrainingLogs").Where("model_no = ?", no).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrModelNotFound
		}
		return nil, err
	}
	return toAIModel(&model), nil
}

func (r *aiModelRepository) ListModels(ctx context.Context, query *domain.ModelQuery) ([]*domain.AIModel, int64, error) {
	var list []*AIModelModel
	var total int64

	db := r.db.WithContext(ctx).Model(&AIModelModel{})
	if query != nil {
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
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := 1
	pageSize := 10
	if query != nil {
		if query.Page > 0 {
			page = query.Page
		}
		if query.PageSize > 0 {
			pageSize = query.PageSize
		}
	}
	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*domain.AIModel, len(list))
	for i, model := range list {
		items[i] = toAIModel(model)
	}
	return items, total, nil
}

func (r *aiModelRepository) DeleteModel(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&AIModelModel{}, id).Error
}

// --- Training Logs ---

func (r *aiModelRepository) SaveTrainingLog(ctx context.Context, log *domain.ModelTrainingLog) error {
	return r.saveTrainingLogWithTx(ctx, r.db, log)
}

func (r *aiModelRepository) SaveTrainingLogInTx(ctx context.Context, tx any, log *domain.ModelTrainingLog) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveTrainingLogWithTx(ctx, gormTx, log)
}

func (r *aiModelRepository) ListTrainingLogs(ctx context.Context, modelID uint64) ([]*domain.ModelTrainingLog, error) {
	var logs []*ModelTrainingLogModel
	if err := r.db.WithContext(ctx).Where("model_id = ?", modelID).Order("iteration asc").Find(&logs).Error; err != nil {
		return nil, err
	}
	items := make([]*domain.ModelTrainingLog, len(logs))
	for i, log := range logs {
		items[i] = toTrainingLog(log)
	}
	return items, nil
}

// --- Predictions ---

func (r *aiModelRepository) SavePrediction(ctx context.Context, prediction *domain.ModelPrediction) error {
	return r.savePredictionWithTx(ctx, r.db, prediction)
}

func (r *aiModelRepository) SavePredictionInTx(ctx context.Context, tx any, prediction *domain.ModelPrediction) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.savePredictionWithTx(ctx, gormTx, prediction)
}

func (r *aiModelRepository) ListPredictions(ctx context.Context, modelID uint64, startTime, endTime time.Time, page, pageSize int) ([]*domain.ModelPrediction, int64, error) {
	var list []*ModelPredictionModel
	var total int64

	db := r.db.WithContext(ctx).Model(&ModelPredictionModel{}).Where("model_id = ?", modelID)

	if !startTime.IsZero() {
		db = db.Where("prediction_time >= ?", startTime)
	}
	if !endTime.IsZero() {
		db = db.Where("prediction_time <= ?", endTime)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Order("prediction_time desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*domain.ModelPrediction, len(list))
	for i, pred := range list {
		items[i] = toPrediction(pred)
	}
	return items, total, nil
}

// --- internal helpers ---

func (r *aiModelRepository) saveModelWithTx(ctx context.Context, tx *gorm.DB, model *domain.AIModel) error {
	if model == nil {
		return nil
	}
	gormTx := tx.WithContext(ctx)
	data := toAIModelModel(model)
	if err := gormTx.Omit("TrainingLogs", "Predictions").Save(data).Error; err != nil {
		return err
	}
	if synced := toAIModel(data); synced != nil {
		*model = *synced
	}
	return nil
}

func (r *aiModelRepository) saveTrainingLogWithTx(ctx context.Context, tx *gorm.DB, log *domain.ModelTrainingLog) error {
	if log == nil {
		return nil
	}
	gormTx := tx.WithContext(ctx)
	data := toTrainingLogModel(log)
	if err := gormTx.Save(data).Error; err != nil {
		return err
	}
	if synced := toTrainingLog(data); synced != nil {
		*log = *synced
	}
	return nil
}

func (r *aiModelRepository) savePredictionWithTx(ctx context.Context, tx *gorm.DB, prediction *domain.ModelPrediction) error {
	if prediction == nil {
		return nil
	}
	gormTx := tx.WithContext(ctx)
	data := toPredictionModel(prediction)
	if err := gormTx.Save(data).Error; err != nil {
		return err
	}
	if synced := toPrediction(data); synced != nil {
		*prediction = *synced
	}
	return nil
}
