package mysql

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/contentmoderation/domain"
	"gorm.io/gorm"
)

type moderationRepository struct {
	db *gorm.DB
}

// NewModerationRepository 创建并返回一个新的 moderationRepository 实例。
func NewModerationRepository(db *gorm.DB) domain.ModerationRepository {
	return &moderationRepository{db: db}
}

// --- tx helpers ---

func (r *moderationRepository) BeginTx(ctx context.Context) any {
	return r.db.WithContext(ctx).Begin()
}

func (r *moderationRepository) CommitTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Commit().Error
}

func (r *moderationRepository) RollbackTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Rollback().Error
}

func (r *moderationRepository) WithTx(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// --- ModerationRecord methods ---

func (r *moderationRepository) SaveRecord(ctx context.Context, record *domain.ModerationRecord) error {
	return r.saveRecordWithTx(ctx, r.db, record)
}

func (r *moderationRepository) SaveRecordInTx(ctx context.Context, tx any, record *domain.ModerationRecord) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveRecordWithTx(ctx, gormTx, record)
}

func (r *moderationRepository) GetRecord(ctx context.Context, id uint64) (*domain.ModerationRecord, error) {
	var record ModerationRecordModel
	if err := r.db.WithContext(ctx).First(&record, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toModerationRecord(&record), nil
}

func (r *moderationRepository) ListRecords(ctx context.Context, query *domain.ModerationRecordQuery) ([]*domain.ModerationRecord, int64, error) {
	var list []*ModerationRecordModel
	var total int64

	db := r.db.WithContext(ctx).Model(&ModerationRecordModel{})
	if query != nil {
		if query.UserID > 0 {
			db = db.Where("user_id = ?", query.UserID)
		}
		if query.ContentType != "" {
			db = db.Where("content_type = ?", query.ContentType)
		}
		if query.ContentID > 0 {
			db = db.Where("content_id = ?", query.ContentID)
		}
		if query.Status != nil {
			db = db.Where("status = ?", *query.Status)
		}
		if query.ModeratorID > 0 {
			db = db.Where("moderator_id = ?", query.ModeratorID)
		}
		if !query.StartTime.IsZero() {
			db = db.Where("created_at >= ?", query.StartTime)
		}
		if !query.EndTime.IsZero() {
			db = db.Where("created_at <= ?", query.EndTime)
		}
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := 1
	pageSize := 20
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

	items := make([]*domain.ModerationRecord, len(list))
	for i, record := range list {
		items[i] = toModerationRecord(record)
	}
	return items, total, nil
}

func (r *moderationRepository) DeleteRecord(ctx context.Context, id uint64) error {
	return r.deleteRecordWithTx(ctx, r.db, id)
}

func (r *moderationRepository) DeleteRecordInTx(ctx context.Context, tx any, id uint64) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.deleteRecordWithTx(ctx, gormTx, id)
}

// --- SensitiveWord methods ---

func (r *moderationRepository) SaveWord(ctx context.Context, word *domain.SensitiveWord) error {
	return r.saveWordWithTx(ctx, r.db, word)
}

func (r *moderationRepository) SaveWordInTx(ctx context.Context, tx any, word *domain.SensitiveWord) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveWordWithTx(ctx, gormTx, word)
}

func (r *moderationRepository) GetWord(ctx context.Context, id uint64) (*domain.SensitiveWord, error) {
	var word SensitiveWordModel
	if err := r.db.WithContext(ctx).First(&word, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toSensitiveWord(&word), nil
}

func (r *moderationRepository) ListWords(ctx context.Context, offset, limit int) ([]*domain.SensitiveWord, int64, error) {
	var list []*SensitiveWordModel
	var total int64

	db := r.db.WithContext(ctx).Model(&SensitiveWordModel{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*domain.SensitiveWord, len(list))
	for i, word := range list {
		items[i] = toSensitiveWord(word)
	}
	return items, total, nil
}

func (r *moderationRepository) DeleteWord(ctx context.Context, id uint64) error {
	return r.deleteWordWithTx(ctx, r.db, id)
}

func (r *moderationRepository) DeleteWordInTx(ctx context.Context, tx any, id uint64) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.deleteWordWithTx(ctx, gormTx, id)
}

func (r *moderationRepository) FindWord(ctx context.Context, word string) (*domain.SensitiveWord, error) {
	var model SensitiveWordModel
	if err := r.db.WithContext(ctx).Where("word = ?", word).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toSensitiveWord(&model), nil
}

// --- internal helpers ---

func (r *moderationRepository) saveRecordWithTx(ctx context.Context, tx *gorm.DB, record *domain.ModerationRecord) error {
	if record == nil {
		return nil
	}
	gormTx := tx.WithContext(ctx)
	model := toModerationRecordModel(record)
	if err := gormTx.Save(model).Error; err != nil {
		return err
	}
	if synced := toModerationRecord(model); synced != nil {
		*record = *synced
	}
	return nil
}

func (r *moderationRepository) deleteRecordWithTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	return tx.WithContext(ctx).Delete(&ModerationRecordModel{}, id).Error
}

func (r *moderationRepository) saveWordWithTx(ctx context.Context, tx *gorm.DB, word *domain.SensitiveWord) error {
	if word == nil {
		return nil
	}
	gormTx := tx.WithContext(ctx)
	model := toSensitiveWordModel(word)
	if err := gormTx.Save(model).Error; err != nil {
		return err
	}
	if synced := toSensitiveWord(model); synced != nil {
		*word = *synced
	}
	return nil
}

func (r *moderationRepository) deleteWordWithTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	return tx.WithContext(ctx).Delete(&SensitiveWordModel{}, id).Error
}
