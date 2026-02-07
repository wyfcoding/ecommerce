// 生成摘要：实现合并订单读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/orderoptimization/domain"
)

const mergedOrderDetailPrefix = "orderoptimization:merged:detail:"

type mergedOrderReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewMergedOrderReadRepository 创建合并订单读模型仓储。
func NewMergedOrderReadRepository(client redis.UniversalClient, ttl time.Duration) domain.MergedOrderReadRepository {
	return &mergedOrderReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *mergedOrderReadRepository) Save(ctx context.Context, order *domain.MergedOrder) error {
	if order == nil {
		return nil
	}
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(uint64(order.ID)), data, r.ttl).Err()
}

func (r *mergedOrderReadRepository) GetByID(ctx context.Context, id uint64) (*domain.MergedOrder, error) {
	data, err := r.client.Get(ctx, r.key(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var order domain.MergedOrder
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *mergedOrderReadRepository) Delete(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.key(id)).Err()
}

func (r *mergedOrderReadRepository) key(id uint64) string {
	return fmt.Sprintf("%s%d", mergedOrderDetailPrefix, id)
}
