package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/data_ingestion/domain/entity"     // 导入数据摄取模块的领域实体定义。
	"github.com/wyfcoding/ecommerce/internal/data_ingestion/domain/repository" // 导入数据摄取模块的领域仓储接口。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

// dataIngestionRepository 是 DataIngestionRepository 接口的GORM实现。
// 它负责将数据摄取模块的领域实体映射到数据库，并执行持久化操作。
type dataIngestionRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewDataIngestionRepository 创建并返回一个新的 dataIngestionRepository 实例。
// db: GORM数据库连接实例。
func NewDataIngestionRepository(db *gorm.DB) repository.DataIngestionRepository {
	return &dataIngestionRepository{db: db}
}

// --- Source methods ---

// SaveSource 将数据源实体保存到数据库。
// 如果数据源已存在（通过ID），则更新其信息；如果不存在，则创建。
func (r *dataIngestionRepository) SaveSource(ctx context.Context, source *entity.IngestionSource) error {
	return r.db.WithContext(ctx).Save(source).Error
}

// GetSource 根据ID从数据库获取数据源记录。
func (r *dataIngestionRepository) GetSource(ctx context.Context, id uint64) (*entity.IngestionSource, error) {
	var source entity.IngestionSource
	if err := r.db.WithContext(ctx).First(&source, id).Error; err != nil {
		return nil, err
	}
	return &source, nil
}

// ListSources 从数据库列出所有数据源记录，支持分页。
func (r *dataIngestionRepository) ListSources(ctx context.Context, offset, limit int) ([]*entity.IngestionSource, int64, error) {
	var list []*entity.IngestionSource
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.IngestionSource{})

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

// UpdateSource 更新数据库中的数据源记录。
func (r *dataIngestionRepository) UpdateSource(ctx context.Context, source *entity.IngestionSource) error {
	return r.db.WithContext(ctx).Save(source).Error
}

// DeleteSource 根据ID从数据库删除数据源记录。
func (r *dataIngestionRepository) DeleteSource(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.IngestionSource{}, id).Error
}

// --- Job methods ---

// SaveJob 将摄取任务实体保存到数据库。
func (r *dataIngestionRepository) SaveJob(ctx context.Context, job *entity.IngestionJob) error {
	return r.db.WithContext(ctx).Save(job).Error
}

// GetJob 根据ID从数据库获取摄取任务记录。
func (r *dataIngestionRepository) GetJob(ctx context.Context, id uint64) (*entity.IngestionJob, error) {
	var job entity.IngestionJob
	if err := r.db.WithContext(ctx).First(&job, id).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

// ListJobs 从数据库列出指定数据源ID的所有摄取任务记录，支持分页。
func (r *dataIngestionRepository) ListJobs(ctx context.Context, sourceID uint64, offset, limit int) ([]*entity.IngestionJob, int64, error) {
	var list []*entity.IngestionJob
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.IngestionJob{})
	if sourceID != 0 { // 如果提供了数据源ID，则按数据源ID过滤。
		db = db.Where("source_id = ?", sourceID)
	}

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

// UpdateJob 更新数据库中的摄取任务记录。
func (r *dataIngestionRepository) UpdateJob(ctx context.Context, job *entity.IngestionJob) error {
	return r.db.WithContext(ctx).Save(job).Error
}
