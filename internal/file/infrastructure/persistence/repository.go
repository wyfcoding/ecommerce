package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/file/domain/entity"     // 导入文件模块的领域实体定义。
	"github.com/wyfcoding/ecommerce/internal/file/domain/repository" // 导入文件模块的领域仓储接口。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

// fileRepository 是 FileRepository 接口的GORM实现。
// 它负责将文件模块的领域实体映射到数据库，并执行持久化操作。
type fileRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewFileRepository 创建并返回一个新的 fileRepository 实例。
// db: GORM数据库连接实例。
func NewFileRepository(db *gorm.DB) repository.FileRepository {
	return &fileRepository{db: db}
}

// Save 将文件元数据实体保存到数据库。
// 如果文件元数据已存在（通过ID），则更新其信息；如果不存在，则创建。
func (r *fileRepository) Save(ctx context.Context, file *entity.FileMetadata) error {
	return r.db.WithContext(ctx).Save(file).Error
}

// Get 根据ID从数据库获取文件元数据记录。
func (r *fileRepository) Get(ctx context.Context, id uint64) (*entity.FileMetadata, error) {
	var file entity.FileMetadata
	if err := r.db.WithContext(ctx).First(&file, id).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

// Delete 根据ID从数据库删除文件元数据记录。
func (r *fileRepository) Delete(ctx context.Context, id uint64) error {
	// GORM的Delete方法会根据主键进行删除。
	return r.db.WithContext(ctx).Delete(&entity.FileMetadata{}, id).Error
}

// List 从数据库列出所有文件元数据记录，支持分页。
func (r *fileRepository) List(ctx context.Context, offset, limit int) ([]*entity.FileMetadata, int64, error) {
	var list []*entity.FileMetadata
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.FileMetadata{})

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
