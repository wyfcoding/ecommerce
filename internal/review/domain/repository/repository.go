package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/review/domain/entity" // 导入评论领域的实体定义。
)

// ReviewRepository 是评论模块的仓储接口。
// 它定义了对 Review 和 ProductRatingStats 实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type ReviewRepository interface {
	// Save 将评论实体保存到数据存储中。
	// 如果评论已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// review: 待保存的评论实体。
	Save(ctx context.Context, review *entity.Review) error
	// Get 根据ID获取评论实体。
	Get(ctx context.Context, id uint64) (*entity.Review, error)
	// List 列出指定商品ID的所有评论实体，支持通过状态过滤和分页。
	List(ctx context.Context, productID uint64, status *entity.ReviewStatus, offset, limit int) ([]*entity.Review, int64, error)
	// Delete 根据ID删除评论实体。
	Delete(ctx context.Context, id uint64) error
	// GetProductStats 获取指定商品的评分统计数据。
	GetProductStats(ctx context.Context, productID uint64) (*entity.ProductRatingStats, error)
}
