// 生成摘要：实现订阅计划读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/subscription/domain"
)

const subscriptionPlanDetailPrefix = "subscription:plan:detail:"

type subscriptionPlanReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewSubscriptionPlanReadRepository 创建订阅计划读模型仓储。
func NewSubscriptionPlanReadRepository(client redis.UniversalClient, ttl time.Duration) domain.SubscriptionPlanReadRepository {
	return &subscriptionPlanReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *subscriptionPlanReadRepository) Save(ctx context.Context, plan *domain.SubscriptionPlan) error {
	if plan == nil {
		return nil
	}
	data, err := json.Marshal(plan)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(uint64(plan.ID)), data, r.ttl).Err()
}

func (r *subscriptionPlanReadRepository) GetByID(ctx context.Context, id uint64) (*domain.SubscriptionPlan, error) {
	data, err := r.client.Get(ctx, r.key(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var plan domain.SubscriptionPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *subscriptionPlanReadRepository) Delete(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.key(id)).Err()
}

func (r *subscriptionPlanReadRepository) key(id uint64) string {
	return fmt.Sprintf("%s%d", subscriptionPlanDetailPrefix, id)
}
