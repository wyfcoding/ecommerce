// 生成摘要：定义评论读模型仓储接口（Redis），用于高频查询。
// 假设：读模型以 review_id 为主键缓存。
package domain

import "context"

// ReviewReadRepository 定义评论读模型的高性能访问接口。
type ReviewReadRepository interface {
	// Save 保存或更新评论读模型。
	Save(ctx context.Context, review *Review) error
	// GetByID 根据评论ID获取读模型。
	GetByID(ctx context.Context, reviewID uint64) (*Review, error)
	// Delete 删除读模型数据。
	Delete(ctx context.Context, reviewID uint64) error
}
