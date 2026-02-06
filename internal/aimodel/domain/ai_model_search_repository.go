// 生成摘要：定义AI模型搜索仓储接口（Elasticsearch），用于分页与过滤查询。
package domain

import "context"

// AIModelSearchRepository 定义AI模型搜索仓储接口。
type AIModelSearchRepository interface {
	// Index 将模型写入搜索索引。
	Index(ctx context.Context, model *AIModel) error
	// Delete 从索引中删除模型。
	Delete(ctx context.Context, modelID uint64) error
	// Search 按条件检索模型（支持状态/类型/算法/创建人过滤、分页）。
	Search(ctx context.Context, status *ModelStatus, modelType, algorithm string, creatorID *uint64, offset, limit int) ([]*AIModel, int64, error)
}
