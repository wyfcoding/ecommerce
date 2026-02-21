// 生成摘要：实现用户行为读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/risk/domain"
)

const userBehaviorPrefix = "risk:behavior:user:"

type userBehaviorReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewUserBehaviorReadRepository 创建用户行为读模型仓储。
func NewUserBehaviorReadRepository(client redis.UniversalClient, ttl time.Duration) domain.UserBehaviorReadRepository {
	return &userBehaviorReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *userBehaviorReadRepository) Save(ctx context.Context, behavior *domain.UserBehavior) error {
	if behavior == nil {
		return nil
	}
	data, err := json.Marshal(behavior)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(behavior.UserID), data, r.ttl).Err()
}

func (r *userBehaviorReadRepository) GetByUserID(ctx context.Context, userID uint64) (*domain.UserBehavior, error) {
	data, err := r.client.Get(ctx, r.key(userID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var behavior domain.UserBehavior
	if err := json.Unmarshal(data, &behavior); err != nil {
		return nil, err
	}
	return &behavior, nil
}

func (r *userBehaviorReadRepository) DeleteByUserID(ctx context.Context, userID uint64) error {
	return r.client.Del(ctx, r.key(userID)).Err()
}

func (r *userBehaviorReadRepository) key(userID uint64) string {
	return fmt.Sprintf("%s%d", userBehaviorPrefix, userID)
}
