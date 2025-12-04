package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/file/domain/entity" // 导入文件领域的实体定义。
)

// FileRepository 是文件模块的仓储接口。
// 它定义了对文件元数据实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type FileRepository interface {
	// Save 将文件元数据实体保存到数据存储中。
	// 如果文件元数据已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// file: 待保存的文件元数据实体。
	Save(ctx context.Context, file *entity.FileMetadata) error
	// Get 根据ID获取文件元数据实体。
	// ctx: 上下文。
	// id: 文件的唯一标识符。
	Get(ctx context.Context, id uint64) (*entity.FileMetadata, error)
	// Delete 根据ID删除文件元数据实体。
	// ctx: 上下文。
	// id: 文件的唯一标识符。
	Delete(ctx context.Context, id uint64) error
	// List 列出所有文件元数据实体，支持分页。
	// ctx: 上下文。
	// offset: 偏移量。
	// limit: 限制数量。
	List(ctx context.Context, offset, limit int) ([]*entity.FileMetadata, int64, error)
}
