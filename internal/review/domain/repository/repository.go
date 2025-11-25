package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/review/domain/entity"
)

// ReviewRepository 评论仓储接口
type ReviewRepository interface {
	Save(ctx context.Context, review *entity.Review) error
	Get(ctx context.Context, id uint64) (*entity.Review, error)
	List(ctx context.Context, productID uint64, status *entity.ReviewStatus, offset, limit int) ([]*entity.Review, int64, error)
	Delete(ctx context.Context, id uint64) error
	GetProductStats(ctx context.Context, productID uint64) (*entity.ProductRatingStats, error)
}
