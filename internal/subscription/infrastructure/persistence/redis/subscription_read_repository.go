// 生成摘要：实现订阅记录读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/subscription/domain"
)

const subscriptionDetailPrefix = "subscription:detail:"

type subscriptionReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewSubscriptionReadRepository 创建订阅记录读模型仓储。
func NewSubscriptionReadRepository(client redis.UniversalClient, ttl time.Duration) domain.SubscriptionReadRepository {
	return &subscriptionReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *subscriptionReadRepository) Save(ctx context.Context, sub *domain.Subscription) error {
	if sub == nil {
		return nil
	}
	data, err := json.Marshal(sub)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(uint64(sub.ID)), data, r.ttl).Err()
}

func (r *subscriptionReadRepository) GetByID(ctx context.Context, id uint64) (*domain.Subscription, error) {
	data, err := r.client.Get(ctx, r.key(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var sub domain.Subscription
	if err := json.Unmarshal(data, &sub); err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *subscriptionReadRepository) Delete(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.key(id)).Err()
}

func (r *subscriptionReadRepository) key(id uint64) string {
	return fmt.Sprintf("%s%d", subscriptionDetailPrefix, id)
}
