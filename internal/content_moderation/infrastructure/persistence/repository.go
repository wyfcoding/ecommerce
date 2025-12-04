package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/content_moderation/domain/entity"     // 导入内容审核模块的领域实体定义。
	"github.com/wyfcoding/ecommerce/internal/content_moderation/domain/repository" // 导入内容审核模块的领域仓储接口。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

// moderationRepository 是 ModerationRepository 接口的GORM实现。
// 它负责将内容审核模块的领域实体映射到数据库，并执行持久化操作。
type moderationRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewModerationRepository 创建并返回一个新的 moderationRepository 实例。
// db: GORM数据库连接实例。
func NewModerationRepository(db *gorm.DB) repository.ModerationRepository {
	return &moderationRepository{db: db}
}

// --- ModerationRecord methods ---

// CreateRecord 在数据库中创建一个新的内容审核记录。
func (r *moderationRepository) CreateRecord(ctx context.Context, record *entity.ModerationRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// GetRecord 根据ID从数据库获取内容审核记录。
func (r *moderationRepository) GetRecord(ctx context.Context, id uint64) (*entity.ModerationRecord, error) {
	var record entity.ModerationRecord
	if err := r.db.WithContext(ctx).First(&record, id).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// UpdateRecord 更新数据库中的内容审核记录。
func (r *moderationRepository) UpdateRecord(ctx context.Context, record *entity.ModerationRecord) error {
	return r.db.WithContext(ctx).Save(record).Error
}

// ListRecords 从数据库列出所有内容审核记录，支持通过状态过滤和分页。
func (r *moderationRepository) ListRecords(ctx context.Context, status entity.ModerationStatus, offset, limit int) ([]*entity.ModerationRecord, int64, error) {
	var list []*entity.ModerationRecord
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.ModerationRecord{})

	// 根据状态过滤审核记录。
	db = db.Where("status = ?", status)

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// --- SensitiveWord methods ---

// CreateWord 在数据库中创建一个新的敏感词记录。
func (r *moderationRepository) CreateWord(ctx context.Context, word *entity.SensitiveWord) error {
	return r.db.WithContext(ctx).Create(word).Error
}

// GetWord 根据ID从数据库获取敏感词记录。
func (r *moderationRepository) GetWord(ctx context.Context, id uint64) (*entity.SensitiveWord, error) {
	var word entity.SensitiveWord
	if err := r.db.WithContext(ctx).First(&word, id).Error; err != nil {
		return nil, err
	}
	return &word, nil
}

// UpdateWord 更新数据库中的敏感词记录。
func (r *moderationRepository) UpdateWord(ctx context.Context, word *entity.SensitiveWord) error {
	return r.db.WithContext(ctx).Save(word).Error
}

// DeleteWord 根据ID从数据库删除敏感词记录。
func (r *moderationRepository) DeleteWord(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.SensitiveWord{}, id).Error
}

// ListWords 从数据库列出所有敏感词记录，支持分页。
func (r *moderationRepository) ListWords(ctx context.Context, offset, limit int) ([]*entity.SensitiveWord, int64, error) {
	var list []*entity.SensitiveWord
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.SensitiveWord{})

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// FindWord 根据敏感词字符串从数据库查找敏感词记录。
func (r *moderationRepository) FindWord(ctx context.Context, word string) (*entity.SensitiveWord, error) {
	var w entity.SensitiveWord
	// 查找匹配的敏感词。
	if err := r.db.WithContext(ctx).Where("word = ?", word).First(&w).Error; err != nil {
		return nil, err
	}
	return &w, nil
}
