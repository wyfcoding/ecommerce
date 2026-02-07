// 生成摘要：实现评论读模型 Redis 仓储，提供按评论ID的快速读取。
// 假设：评论以 review_id 为主键缓存。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/review/domain"
)

const (
	reviewDetailPrefix = "review:detail:"
	statsPrefix        = "review:stats:"
)

type reviewReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewReviewReadRepository 创建评论读模型仓储。
func NewReviewReadRepository(client redis.UniversalClient, ttl time.Duration) domain.ReviewReadRepository {
	return &reviewReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *reviewReadRepository) Save(ctx context.Context, review *domain.Review) error {
	if review == nil {
		return nil
	}
	data, err := json.Marshal(review)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.reviewKey(review.ID), data, r.ttl).Err()
}

func (r *reviewReadRepository) GetByID(ctx context.Context, reviewID uint64) (*domain.Review, error) {
	data, err := r.client.Get(ctx, r.reviewKey(reviewID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var review domain.Review
	if err := json.Unmarshal(data, &review); err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *reviewReadRepository) Delete(ctx context.Context, reviewID uint64) error {
	return r.client.Del(ctx, r.reviewKey(reviewID)).Err()
}

func (r *reviewReadRepository) SaveProductStats(ctx context.Context, stats *domain.ProductRatingStats) error {
	if stats == nil {
		return nil
	}
	data, err := json.Marshal(stats)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.statsKey(stats.ProductID), data, r.ttl).Err()
}

func (r *reviewReadRepository) GetProductStats(ctx context.Context, productID uint64) (*domain.ProductRatingStats, error) {
	data, err := r.client.Get(ctx, r.statsKey(productID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var stats domain.ProductRatingStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

func (r *reviewReadRepository) reviewKey(reviewID uint64) string {
	return fmt.Sprintf("%s%d", reviewDetailPrefix, reviewID)
}

func (r *reviewReadRepository) statsKey(productID uint64) string {
	return fmt.Sprintf("%s%d", statsPrefix, productID)
}
