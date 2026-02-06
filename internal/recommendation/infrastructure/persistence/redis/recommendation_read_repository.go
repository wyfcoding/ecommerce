// 生成摘要：实现推荐读模型 Redis 仓储，提供按用户与类型的快速读取。
// 假设：推荐列表以 user_id + type 为主键缓存。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/recommendation/domain"
)

const (
	recommendationListPrefix = "recommendation:list:"
)

type recommendationReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewRecommendationReadRepository 创建推荐读模型仓储。
func NewRecommendationReadRepository(client redis.UniversalClient, ttl time.Duration) domain.RecommendationReadRepository {
	return &recommendationReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *recommendationReadRepository) SaveRecommendations(ctx context.Context, userID uint64, recType *domain.RecommendationType, recs []*domain.Recommendation) error {
	data, err := json.Marshal(recs)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.listKey(userID, recType), data, r.ttl).Err()
}

func (r *recommendationReadRepository) GetRecommendations(ctx context.Context, userID uint64, recType *domain.RecommendationType, limit int) ([]*domain.Recommendation, error) {
	data, err := r.client.Get(ctx, r.listKey(userID, recType)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var recs []*domain.Recommendation
	if err := json.Unmarshal(data, &recs); err != nil {
		return nil, err
	}
	if limit > 0 && len(recs) > limit {
		recs = recs[:limit]
	}
	return recs, nil
}

func (r *recommendationReadRepository) DeleteRecommendations(ctx context.Context, userID uint64, recType *domain.RecommendationType) error {
	return r.client.Del(ctx, r.listKey(userID, recType)).Err()
}

func (r *recommendationReadRepository) listKey(userID uint64, recType *domain.RecommendationType) string {
	t := "all"
	if recType != nil && *recType != "" {
		t = string(*recType)
	}
	return fmt.Sprintf("%s%d:%s", recommendationListPrefix, userID, t)
}
