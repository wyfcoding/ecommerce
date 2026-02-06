// 生成摘要：定义AI模型读模型仓储接口（Redis），用于高频查询。
package domain

import "context"

// AIModelReadRepository 定义AI模型读模型的高性能访问接口。
type AIModelReadRepository interface {
	// Save 保存或更新读模型。
	Save(ctx context.Context, model *AIModel) error
	// GetByID 根据模型ID获取读模型。
	GetByID(ctx context.Context, id uint64) (*AIModel, error)
	// GetByNo 根据模型编号获取读模型。
	GetByNo(ctx context.Context, modelNo string) (*AIModel, error)
	// Delete 删除读模型数据。
	Delete(ctx context.Context, id uint64, modelNo string) error
}
